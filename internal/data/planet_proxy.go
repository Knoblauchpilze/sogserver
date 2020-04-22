package data

import (
	"fmt"
	"math/rand"
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

// NewPlanetProxy :
// Create a new proxy allowing to serve the requests
// related to planets.
//
// The `dbase` represents the database to use to fetch
// data related to planets.
//
// The `data` defines the data model to use to fetch
// information and verify actions.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewPlanetProxy(dbase *db.DB, data model.Instance, log logger.Logger) PlanetProxy {
	return PlanetProxy{
		commonProxy: newCommonProxy(dbase, data, log, "planets"),
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
// accounts available. Each one is appended `as-is` to the
// query.
//
// Returns the list of planets registered in the DB and matching
// the input list of filters. In case the error is not `nil` the
// value of the array should be ignored.
func (p *PlanetProxy) Planets(filters []db.Filter) ([]model.Planet, error) {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"p.id",
		},
		Table:   "planets p inner join players pl on p.player=pl.id",
		Filters: filters,
	}

	res, err := p.proxy.FetchFromDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not query DB to fetch planets (err: %v)", err))
		return []model.Planet{}, err
	}

	// We now need to retrieve all the identifiers that matched
	// the input filters and then build the corresponding planets
	// object for each one of them.
	var ID string
	IDs := make([]string, 0)

	for res.Next() {
		err = res.Scan(&ID)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Error while fetching planet ID (err: %v)", err))
			continue
		}

		IDs = append(IDs, ID)
	}

	planets := make([]model.Planet, 0)

	for _, ID = range IDs {
		pla, err := model.NewPlanetFromDB(ID, p.data)

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
			Amount:     res.BaseAmount,
			Storage:    res.BaseStorage,
			Production: res.BaseProd,
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
// Returns any error arised during the creation process.
func (p *PlanetProxy) CreateFor(player model.Player) error {
	// Check consistency.
	if player.Valid() {
		return model.ErrInvalidPlayer
	}

	// First we need to fetch the universe related to the
	// planet to create.
	uni, err := p.fetchUniverse(player.Universe)
	if err != nil {
		return fmt.Errorf("Could not create planet for \"%s\" (err: %v)", player.ID, err)
	}

	// Retrieve the list of coordinates that are already
	// used in the universe the player's in.
	usedCoords, err := p.generateUsedCoords(uni)
	totalPlanets := uni.GalaxiesCount * uni.GalaxySize * uni.SolarSystemSize

	// Try to insert the planet in the DB while we have some
	// untested coordinates and we didn't suceed in inserting
	// it.
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
		coord := model.Coordinate{
			Galaxy:   rand.Int() % uni.GalaxiesCount,
			System:   rand.Int() % uni.GalaxySize,
			Position: rand.Int() % uni.SolarSystemSize,
		}

		exists := true
		for exists {
			key := coord.Linearize(uni)

			if _, ok := usedCoords[key]; !ok {
				// We found a not yet used coordinate.
				exists = false
			} else {
				// Pick some new coordinates.
				coord = model.Coordinate{
					Galaxy:   rand.Int() % uni.GalaxiesCount,
					System:   rand.Int() % uni.GalaxySize,
					Position: rand.Int() % uni.SolarSystemSize,
				}
			}
		}

		// Generate a new planet. We also need to associate
		// some resources to it.
		planet := model.NewPlanet(player.ID, coord)
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
		} else {
			p.trace(logger.Warning, fmt.Sprintf("Could not import planet at %s for \"%s\" (err: %v)", coord, player.ID, err))

			// Register this coordinate as being used as we can't
			// successfully use it to create the planet anyways.
			usedCoords[coord.Linearize(uni)] = coord
		}

		trials++
	}

	// Check whether we could insert the element in the DB: if
	// this is not the case we couldn't create the planet.
	if !inserted {
		return fmt.Errorf("Could not insert planet for player \"%s\" in DB after %d trial(s)", player.ID, trials)
	}

	return nil
}

// generateUsedCoords :
// Used to find and generate a list of the used coordinates
// in the corresponding universe. Note that the list is only
// some snapshot of the state of the coordinates which can
// evolve through time. Typically if some pending requests to
// insert a planet are pending or some actions require some
// action to create/destroy a planet this list will be changed
// and might not be accurate.
// We figure it's not really a problem to insert elements in
// the DB as it's unlikely to ever failed a lot of times in
// a row. What can maybe happen is that the first try fails
// to insert a planet but the second one with a different set
// of coordinates it will most likely succeed.
//
// The `uni` defines the universe for which available coords
// should be fetched. This will be fetched from the DB.
//
// The return value includes all the user coordinates in the
// universe along with any errors.
func (p *PlanetProxy) generateUsedCoords(uni Universe) (map[int]Coordinate, error) {
	// Create the query allowing to fetch all the planets of
	// a specific universe. This will consistute the list of
	// used planets for this universe.
	query := p.buildQuery(
		[]string{
			"p.galaxy",
			"p.solar_system",
			"p.position",
		},
		"planets p inner join players pl on p.player=pl.id",
		"pl.uni",
		uni.ID,
	)

	// Create the query and execute it.
	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return map[int]Coordinate{}, fmt.Errorf("Could not fetch used coordinates for universe \"%s\" (err: %v)", uni.ID, err)
	}

	// Traverse all the coordinates and populate the list.
	coords := make(map[int]Coordinate)
	var coord Coordinate

	for res.next() {
		err = res.scan(
			&coord.Galaxy,
			&coord.System,
			&coord.Position,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for coordinate in universe \"%s\" (err: %v)", uni.ID, err))
			continue
		}

		key := coord.Linearize(uni)

		// Check whether it's the first time we encounter
		// this used location.
		if _, ok := coords[key]; ok {
			p.log.Trace(logger.Error, fmt.Sprintf("Overriding used coordinate %v in universe \"%s\"", coord, uni.ID))
		}

		coords[key] = coord
	}

	return coords, nil
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
