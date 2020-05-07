package data

import (
	"fmt"
	"math/rand"
	"oglike_server/internal/locker"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// PlanetProxy :
// Intended as a wrapper to access properties of planets
// and retrieve data from the database. In most cases we
// need to access some properties of the planets for a
// given identifier. A planet is a set of resources, a
// set of ships and defenses that are currently built on
// it and some upgrade actions (meaning that a building
// will have one more level soon or that there will be
// more ships deployed on this planet).
type PlanetProxy struct {
	commonProxy
}

// planetGenerationMaxTrials :
// Used to define the maximum number of trials that can
// be performed to create a planet. This is used as a
// way to rapidly exhaust possibilities when trying to
// create a new player and not cycle through all the
// possible coordinates. Most of the time it should be sufficient
var planetGenerationMaxTrials = 10

// ErrTooManyTrials :
// Used to indicate that the planet was not generated due
// to too many failure when trying to select cooridnates.
var ErrTooManyTrials = fmt.Errorf("Could not create planet after %d trial(s)", planetGenerationMaxTrials)

// NewPlanetProxy :
// Create a new proxy allowing to serve the requests
// related to planets.
//
// The `data` defines the data model to use to fetch
// information and verify requests.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewPlanetProxy(data model.Instance, log logger.Logger) PlanetProxy {
	return PlanetProxy{
		commonProxy: newCommonProxy(data, log, "planets"),
	}
}

// Planets :
// Return a list of planets registered so far in all the planets
// defined in the DB. The input filters might help to narrow the
// search a bit by providing coordinates to look for and a uni
// to look into.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// planets available. Each one is appended `as-is` to the
// query.
//
// Returns the list of planets registered in the DB and matching
// the input list of filters. In case the error is not `nil` the
// value of the array should be ignored.
func (p *PlanetProxy) Planets(filters []db.Filter) ([]model.Planet, error) {
	// We need to retrieve all the identifiers that matched
	// the input filters and then build the corresponding
	// planet objects for each one of them.
	// Note that we should protect the access to the players
	// by using a locker on each player owning a planet to
	// return. The only way we have at this step to fetch
	// the ids of all players is through the filters. So we
	// will traverse them and lock each player.
	// In order to have some fail-safe mechanism we will
	// surround this by a dedicated function.
	locks := make([]*locker.Lock, 0)

	defer func() {
		for _, l := range locks {
			err := l.Unlock()

			if err != nil {
				p.trace(logger.Error, fmt.Sprintf("Could not release lock on player (err: %v)", err))
			}
		}
	}()

	for _, filter := range filters {
		if filter.Key == "player" {
			for _, player := range filter.Values {
				l, err := p.data.Locker.Acquire(player)

				if err != nil {
					p.trace(logger.Error, fmt.Sprintf("Unable to lock player \"%s\" to fetch planet (err: %v)", player, err))
					return []model.Planet{}, err
				}

				l.Lock()

				locks = append(locks, l)
			}
		}
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table:   "planets",
		Filters: filters,
	}

	dbRes, err := p.proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not query DB to fetch planets (err: %v)", err))
		return []model.Planet{}, err
	}
	if dbRes.Err != nil {
		p.trace(logger.Error, fmt.Sprintf("Invalid query to fetch planets (err: %v)", dbRes.Err))
		return []model.Planet{}, dbRes.Err
	}

	// Fetch the data for each planet.
	var ID string
	IDs := make([]string, 0)

	for dbRes.Next() {
		err = dbRes.Scan(&ID)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Error while fetching planet ID (err: %v)", err))
			continue
		}

		IDs = append(IDs, ID)
	}

	planets := make([]model.Planet, 0)

	for _, ID = range IDs {
		// Protect the fetching of the planet's data with a
		// lock on the player.
		pla, err := model.NewReadOnlyPlanet(ID, p.data)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Unable to fetch planet \"%s\" data from DB (err: %v)", ID, err))
			continue
		}

		planets = append(planets, pla)
	}

	return planets, nil
}

// generateResources :
// Used to perform the creation of the default resources for
// a planet when it's being created. This translate the fact
// that a planet is never really `empty` in the game.
// This function will create all the necessary entries in the
// planet input object but does not create anything in the DB.
//
// The `planet` defines the element for which resources need
// to be generated.
//
// Returns any error.
func (p *PlanetProxy) generateResources(planet *model.Planet) error {
	// A planet always has the base amount defined in the DB.
	resources, err := p.data.Resources.Resources(p.proxy, []db.Filter{})
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Unable to generate resources for planet (err: %v)", err))
		return err
	}

	// We will consider that we have a certain number of each
	// resources readily available on each planet upon its
	// creation. The values are hard-coded there and we use
	// the identifier of the resources retrieved from the DB
	// to populate the planet.
	if planet.Resources == nil {
		planet.Resources = make([]model.ResourceInfo, 0)
	}

	for _, res := range resources {
		desc := model.ResourceInfo{
			Resource:   res.ID,
			Amount:     float32(res.BaseAmount),
			Storage:    float32(res.BaseStorage),
			Production: float32(res.BaseProd),
		}

		planet.Resources = append(planet.Resources, desc)
	}

	return nil
}

// CreateFor :
// Used to handle the creation of a planet for the specified
// player. This method is only used when a new player needs
// to be registered in the universe so the coordinates of the
// new planet to create are determine directly in this method.
//
// The `player` represents the account for which the planet is
// to be created. We assume that the universe and the player's
// identifiers are valid (otherwise we won't be able to attach
// the planet to a valid account).
//
// Returns any error arised during the creation process along
// with the identifier of the planet that was created. The
// identifier might be empty in case the planet could not be
// created.
func (p *PlanetProxy) CreateFor(player model.Player) (string, error) {
	// Check consistency.
	if !player.Valid() {
		return "", model.ErrInvalidPlayer
	}

	// First we need to fetch the universe related to the
	// planet to create.
	uni, err := model.NewUniverseFromDB(player.Universe, p.data)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Unable to fetch universe \"%s\" to create planet (err: %v)", player.Universe, err))
		return "", err
	}

	// Retrieve the list of coordinates that are already
	// used in the universe the player's in.
	usedCoords, err := uni.UsedCoords(p.data.Proxy)
	totalPlanets := uni.GalaxiesCount * uni.GalaxySize * uni.SolarSystemSize

	// Try to insert the planet in the DB while we have some
	// untested coordinates and we didn't suceed in inserting
	// it.
	id := ""
	inserted := false
	trials := 0

	for !inserted && len(usedCoords) < totalPlanets && trials < planetGenerationMaxTrials {
		// Pick a random coordinate and check whether it belongs
		// to the already used coordinates. If this is the case
		// we will try to pick a new one. Otherwise we will try
		// to perform the insertion of the planet in the DB.
		// In case the insertion fails we will add the selected
		// coordinates to the list of used one so as not to try
		// to use it again.
		coord := model.NewPlanetCoordinate(
			rand.Int()%uni.GalaxiesCount,
			rand.Int()%uni.GalaxySize,
			rand.Int()%uni.SolarSystemSize,
		)

		exists := true
		for exists {
			key := coord.Linearize(uni.GalaxySize, uni.SolarSystemSize)

			if _, ok := usedCoords[key]; !ok {
				// We found a not yet used coordinate.
				exists = false
			} else {
				// Pick some new coordinates.
				coord = model.NewPlanetCoordinate(
					rand.Int()%uni.GalaxiesCount,
					rand.Int()%uni.GalaxySize,
					rand.Int()%uni.SolarSystemSize,
				)
			}
		}

		// Generate a new planet. We also need to associate
		// some resources to it.
		planet := model.NewPlanet(player.ID, coord, true)
		err := p.generateResources(planet)
		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Unable to generate resources for planet at %s for \"%s\" (err: %v)", coord, player.ID, err))
		}

		// Try to create the planet at the specified coordinates.
		err = p.createPlanet(planet)

		// Check for errors.
		if err == nil {
			p.trace(logger.Notice, fmt.Sprintf("Created planet at %s for \"%s\" in \"%s\" with %d field(s)", coord, player.ID, player.Universe, planet.Fields))
			inserted = true
			id = planet.ID
		} else {
			p.trace(logger.Warning, fmt.Sprintf("Could not import planet at %s for \"%s\" (err: %v)", coord, player.ID, err))

			// Register this coordinate as being used as we can't
			// successfully use it to create the planet anyways.
			usedCoords[coord.Linearize(uni.GalaxySize, uni.SolarSystemSize)] = coord
		}

		trials++
	}

	// Check whether we could insert the element in the DB: if
	// this is not the case we couldn't create the planet.
	if !inserted {
		return "", ErrTooManyTrials
	}

	return id, nil
}

// createPlanet :
// Used to attempt to create the planet with the specified
// data. We only try to perform the insertion of the input
// data into the DB and report in error if it cannot be
// performed for some reasons.
//
// The `planet` defines the data to insert in the DB.
//
// Returns an error in case the planet could not be used
// in the DB (for example because the coordinates already
// are used).
func (p *PlanetProxy) createPlanet(planet *model.Planet) error {
	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_planet",
		Args: []interface{}{
			planet,
			planet.Resources,
		},
	}

	err := p.proxy.InsertToDB(query)

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not import planet \"%s\" (err: %v)", planet.Name, err))
	}

	return err
}
