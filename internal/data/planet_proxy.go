package data

import (
	"fmt"
	"math"
	"math/rand"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"

	"github.com/google/uuid"
)

// getDefaultPlanetName :
// Used to retrieve a default name for a planet. The
// generated name will be different based on whether
// the planet is a homeworld or no.
//
// The `isHomeworld` is `true` if we should generate
// a name for the homeworld and `false` otherwise.
//
// Returns a string corresponding to the name of the
// planet.
func getDefaultPlanetName(isHomeWorld bool) string {
	if isHomeWorld {
		return "homeworld"
	}

	return "planet"
}

// getPlanetTemperatureAmplitude :
// Used to retrieve the default planet temperature's
// amplitude. Basically the interval between the min
// and max temperature will always be equal to this
// value.
//
// Returns the default temperature amplitude for the
// planets.
func getPlanetTemperatureAmplitude() int {
	return 50
}

// PlanetProxy :
// Intended as a wrapper to access properties of planets
// and retrieve data from the database. This helps hiding
// the complexity of how the data is laid out in the `DB`
// and the precise name of tables from the exterior world.
//
// The `dbase` is the database that is wrapped by this
// object. It is checked for consistency upon building the
// wrapper.
//
// The `log` allows to perform display to the user so as
// to inform of potential issues and debug information to
// the outside world.
type PlanetProxy struct {
	dbase *db.DB
	log   logger.Logger
}

// NewPlanetProxy :
// Create a new proxy on the input `dbase` to access the
// properties of planets as registered in the DB. In
// case the provided DB is `nil` a panic is issued.
// Information in the following thread helped shape this
// component:
// https://www.reddit.com/r/golang/comments/9i5cpg/good_approach_to_interacting_with_databases/
//
// The `dbase` represents the database to use to fetch
// data related to planets.
//
// The `log` will be used to notify information so that
// we can have an idea of the activity of this component.
// One possible example is for timing the requests.
//
// Returns the created proxy.
func NewPlanetProxy(dbase *db.DB, log logger.Logger) PlanetProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create planets proxy from invalid DB"))
	}

	// TODO: Include init in here ? This could be done for
	// all proxys and thus allow to fetch information as
	// needed for each one of them.

	return PlanetProxy{
		dbase,
		log,
	}
}

// GetIdentifierDBColumnName :
// Used to retrieve the string literal defining the name of the
// identifier column in the `planets` table in the database.
//
// Returns the name of the `identifier` column in the database.
func (p *PlanetProxy) GetIdentifierDBColumnName() string {
	return "p.id"
}

// Planets :
// Return a list of planets registered so far in all the planets
// defined in the DB. The input filters might help to narrow the
// search a bit by providing coordinates to look for and a uni to
// look into.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// accounts available. Each one is appended `as-is` to the
// query.
//
// Returns the list of planets registered in the DB and matching
// the input list of filters. In case the error is not `nil` the
// value of the array should be ignored.
func (p *PlanetProxy) Planets(filters []DBFilter) ([]Planet, error) {
	// Create the query and execute it.
	props := []string{
		"p.id",
		"p.player",
		"p.name",
		"p.fields",
		"p.min_temperature",
		"p.max_temperature",
		"p.diameter",
		"p.galaxy",
		"p.solar_system",
		"p.position",
	}

	table := "planets p inner join players pl"
	joinCond := "p.player=pl.id"

	query := fmt.Sprintf("select %s from %s on %s", strings.Join(props, ", "), table, joinCond)

	if len(filters) > 0 {
		query += " where"

		for id, filter := range filters {
			if id > 0 {
				query += " and"
			}
			query += fmt.Sprintf(" %s", filter)
		}
	}

	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch planets (err: %v)", err)
	}

	// Populate the return value.
	planets := make([]Planet, 0)
	var planet Planet

	galaxy := 0
	system := 0
	position := 0

	for rows.Next() {
		err = rows.Scan(
			&planet.ID,
			&planet.PlayerID,
			&planet.Name,
			&planet.Fields,
			&planet.MinTemp,
			&planet.MaxTemp,
			&planet.Diameter,
			&galaxy,
			&system,
			&position,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for planet (err: %v)", err))
			continue
		}

		planet.Coords = Coordinate{
			galaxy,
			system,
			position,
		}

		// Fetch buildings, ships and defenses for this planet.
		err = p.fetchPlanetData(&planet)
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch data for planet \"%s\" (err: %v)", planet.ID, err))
			continue
		}

		planets = append(planets, planet)
	}

	return planets, nil
}

// fetchPlanetData :
// Used to fetch data built on the planet provided in input.
// This typically include the buildings, the ships deployed
// and the defenses installed.
//
// The `planet` references the planet for which data should
// be fetched. We assume that the internal fields (and more
// specifically the identifier) are already populated.
//
// Returns any error.
func (p *PlanetProxy) fetchPlanetData(planet *Planet) error {
	// Check whether the planet has an identifier assigned.
	if planet.ID == "" {
		return fmt.Errorf("Unable to fetch data from planet with invalid identifier")
	}

	// Fetch resources.
	query := fmt.Sprintf("select res, amount from planets_resources where planet='%s'", planet.ID)
	rows, err := p.dbase.DBQuery(query)

	if err != nil {
		return fmt.Errorf("Could not fetch resources for planet \"%s\" (err: %v)", planet.ID, err)
	}

	planet.Resources = make([]Resource, 0)
	var resource Resource

	for rows.Next() {
		err = rows.Scan(
			&resource.ID,
			&resource.Amount,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve resource for planet \"%s\" (err: %v)", planet.ID, err))
			continue
		}

		planet.Resources = append(planet.Resources, resource)
	}

	// Fetch buildings.
	query = fmt.Sprintf("select building, level from planets_buildings where planet='%s'", planet.ID)
	rows, err = p.dbase.DBQuery(query)

	if err != nil {
		return fmt.Errorf("Could not fetch buildings for planet \"%s\" (err: %v)", planet.ID, err)
	}

	planet.Buildings = make([]Building, 0)
	var building Building

	for rows.Next() {
		err = rows.Scan(
			&building.ID,
			&building.Level,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve building for planet \"%s\" (err: %v)", planet.ID, err))
			continue
		}

		planet.Buildings = append(planet.Buildings, building)
	}

	// Fetch ships.
	query = fmt.Sprintf("select ship, count from planets_ships where planet='%s'", planet.ID)
	rows, err = p.dbase.DBQuery(query)

	if err != nil {
		return fmt.Errorf("Could not fetch ships for planet \"%s\" (err: %v)", planet.ID, err)
	}

	planet.Ships = make([]Ship, 0)
	var ship Ship

	for rows.Next() {
		err = rows.Scan(
			&ship.ID,
			&ship.Count,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve ship for planet \"%s\" (err: %v)", planet.ID, err))
			continue
		}

		planet.Ships = append(planet.Ships, ship)
	}

	// Fetch defenses.
	query = fmt.Sprintf("select defense, count from planets_defenses where planet='%s'", planet.ID)
	rows, err = p.dbase.DBQuery(query)

	if err != nil {
		return fmt.Errorf("Could not fetch defenses for planet \"%s\" (err: %v)", planet.ID, err)
	}

	planet.Defenses = make([]Defense, 0)
	var defense Defense

	for rows.Next() {
		err = rows.Scan(
			&defense.ID,
			&defense.Count,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve defense for planet \"%s\" (err: %v)", planet.ID, err))
			continue
		}

		planet.Defenses = append(planet.Defenses, defense)
	}

	return nil
}

// CreateFor :
// Used to handle the creation of a planet for the specified
// player at the input coordinate. In case the coordinates are
// `nil` we will assume that we are creating the homeworld for
// the player and thus we can choose the coordinates randomly.
// Otherwise we will try to create the planet at the specified
// coordinates and fail if the coordinates are not available.
// The universe to create the planet in is described by the
// `UniverseID` of the player.
//
// The `player` represents the account for which the planet is
// to be created. We assume that the universe and the player's
// identifiers are valid (otherwise we won't be able to attach
// the planet to a valid account).
//
// The `coord` represents the desired coordinates where the
// planet should be created. In case this value is `nil` we
// assume that the homeworld of the player should be created
// and thus we will choose randomly some coordinates among
// the available locations.
//
// Returns any error arised during the creation process.
func (p *PlanetProxy) CreateFor(player Player, coord *Coordinate) error {
	// Check consistency.
	if player.ID == "" {
		return fmt.Errorf("Cannot create planet for invalid player")
	}
	if player.UniverseID == "" {
		return fmt.Errorf("Cannot create planet for player \"%s\" in invalid universe", player.ID)
	}

	// First we need to fetch the universe related to the
	// planet to create.
	// TODO: Fetch universe from player's data.
	var uni Universe

	// Create the planet from the available data.
	planet, err := p.generatePlanet(player.ID, coord, uni)
	if err != nil {
		return fmt.Errorf("Could not create planet for \"%s\" (err: %v)", player.ID, err)
	}

	// We will now try to insert the planet into the DB if
	// we have valid coordinates.
	availableCoords := make([]Coordinate, 0)
	if coord != nil {
		availableCoords = append(availableCoords, *coord)
	} else {
		// TODO: Generate list of available coordinates from the DB.
	}

	// Try to insert the planet in the DB while we have some
	// untested coordinates and we didn't suceed in inserting
	// it.
	inserted := false
	trials := 0
	for !inserted && len(availableCoords) > 0 {
		// Pick a random coordinate from the list and use it to
		// try to insert the planet in the DB. We also need to
		// remove it from the available coordinates list so as
		// not to pick it again.
		size := len(availableCoords)

		coordID := rand.Int() % size
		coord := availableCoords[coordID]

		if size == 1 {
			availableCoords = nil
		} else {
			availableCoords[size-1], availableCoords[coordID] = availableCoords[coordID], availableCoords[size-1]
			availableCoords = availableCoords[:size-1]
		}

		// Try to create the planet at the specified coordinates.
		// TODO: Handle this.
		fmt.Println(fmt.Sprintf("Trying to insert planet \"%s\" for \"%s\" at %s", planet.ID, player.ID, coord))

		trials++
	}

	// Check whether we could insert the element in the DB: if
	// this is not the case we couldn't create the planet.
	if !inserted {
		return fmt.Errorf("Could not insert planet for player \"%s\" in DB after %d trial(s)", player.ID, trials)
	}

	return nil
}

// generatePlanet :
// Used to perform the generation of the properties of a planet
// based on the input player and coordinates. All the info to
// actually define the planet will be generated including the
// resources.
// The universe in which the planet should be provided in case
// the coordinates are to be determined by this function. As we
// will need information as to the positions that are possible
// for a planet in this case.
//
// The `player` defines the identifier of the player to which
// the planet belongs.
//
// The `coord` defines the coordinates of the planet to create.
// If the value is `nil` no data is generated.
//
// The `uni` argument defines the universe in which the planet
// is to be created. This helps with defining valid coordinates
// in case none are provided.
//
// Returns the created planet.
func (p *PlanetProxy) generatePlanet(player string, coord *Coordinate, uni Universe) (Planet, error) {
	trueCoords := Coordinate{0, 0, 0}
	if coord != nil {
		trueCoords = *coord
	}

	// Create the planet and generate base information.
	planet := Planet{
		player,
		uuid.New().String(),
		trueCoords,
		getDefaultPlanetName(coord == nil),
		0,
		0,
		0,
		0,
		make([]Resource, 0),
		make([]Building, 0),
		make([]Ship, 0),
		make([]Defense, 0),
	}

	p.generateResources(&planet)

	return planet, nil
}

// generatePlanetSize :
// Used to generate the size associated to a planet. The size
// is a general notion including both its actual diameter and
// also the temperature on the surface of the planet. Both
// values depend on the actual position of the planet in the
// parent solar system.
// We consider that if the planet has a solar system of `0`
// nothing will be generated.
//
// The `planet` defines the planet for which the size should
// be generated.
func (p *PlanetProxy) generatePlanetSize(planet *Planet) {
	// Check whether the planet and its coordinates are valid.
	if planet == nil || planet.Coords.Position == 0 {
		return
	}

	// Create a random source to be used for the generation of
	// the planet's properties. We will use a procedural algo
	// which will be based on the position of the planet in its
	// parent universe.
	source := rand.NewSource(int64(planet.Coords.generateSeed()))
	rng := rand.New(source)

	// The table of the dimensions of the planet are inspired
	// from this link:
	// https://ogame.fandom.com/wiki/Colonizing_in_Redesigned_Universes
	var min int
	var max int
	var stdDev int

	switch planet.Coords.Position {
	case 1:
		// Range [96; 172], average 134.
		min = 96
		max = 172
		stdDev = max - min
	case 2:
		// Range [104; 176], average 140.
		min = 104
		max = 176
		stdDev = max - min
	case 3:
		// Range [112; 182], average 147.
		min = 112
		max = 182
		stdDev = max - min
	case 4:
		// Range [118; 208], average 163.
		min = 118
		max = 208
		stdDev = max - min
	case 5:
		// Range [133; 232], average 182.
		min = 133
		max = 232
		stdDev = max - min
	case 6:
		// Range [152; 248], average 200.
		min = 152
		max = 248
		stdDev = max - min
	case 7:
		// Range [156; 262], average 204.
		min = 156
		max = 262
		stdDev = max - min
	case 8:
		// Range [150; 246], average 198.
		min = 150
		max = 246
		stdDev = max - min
	case 9:
		// Range [142; 232], average 187.
		min = 142
		max = 232
		stdDev = max - min
	case 10:
		// Range [136; 210], average 173.
		min = 136
		max = 210
		stdDev = max - min
	case 11:
		// Range [125; 186], average 156.
		min = 125
		max = 186
		stdDev = max - min
	case 12:
		// Range [114; 172], average 143.
		min = 114
		max = 172
		stdDev = max - min
	case 13:
		// Range [100; 168], average 134.
		min = 100
		max = 168
		stdDev = max - min
	case 14:
		// Range [90; 164], average 127.
		min = 96
		max = 164
		stdDev = max - min
	case 15:
		fallthrough
	default:
		// Assume default case if the `15th` position
		// Range [90; 164], average 134.
		min = 90
		max = 164
		stdDev = max - min
	}

	mean := (max + min) / 2
	planet.Fields = mean + int(math.Round(rng.NormFloat64()*float64(stdDev)))

	// The temperatures are described in the following link:
	// https://ogame.fandom.com/wiki/Temperature

	switch planet.Coords.Position {
	case 1:
		// Range [220; 260], average 240.
		min = 220
		max = 260
		stdDev = max - min
	case 2:
		// Range [170; 210], average 190.
		min = 170
		max = 210
		stdDev = max - min
	case 3:
		// Range [120; 160], average 140.
		min = 120
		max = 160
		stdDev = max - min
	case 4:
		// Range [70; 110], average 90.
		min = 70
		max = 110
		stdDev = max - min
	case 5:
		// Range [60; 100], average 80.
		min = 60
		max = 100
		stdDev = max - min
	case 6:
		// Range [50; 90], average 70.
		min = 50
		max = 90
		stdDev = max - min
	case 7:
		// Range [40; 80], average 60.
		min = 40
		max = 80
		stdDev = max - min
	case 8:
		// Range [30; 70], average 50.
		min = 30
		max = 70
		stdDev = max - min
	case 9:
		// Range [20; 60], average 40.
		min = 20
		max = 60
		stdDev = max - min
	case 10:
		// Range [10; 50], average 30.
		min = 10
		max = 50
		stdDev = max - min
	case 11:
		// Range [0; 40], average 20.
		min = 0
		max = 40
		stdDev = max - min
	case 12:
		// Range [-10; 30], average 10.
		min = -10
		max = 30
		stdDev = max - min
	case 13:
		// Range [-50; -10], average -30.
		min = -50
		max = -10
		stdDev = max - min
	case 14:
		// Range [-90; -50], average -70.
		min = -90
		max = -50
		stdDev = max - min
	case 15:
		fallthrough
	default:
		// Assume default case if the `15th` position
		// Range [-130; -90], average -110.
		min = -130
		max = -90
		stdDev = max - min
	}

	mean = (max + min) / 2
	planet.MaxTemp = mean + int(math.Round(rng.NormFloat64()*float64(stdDev)))
	planet.MinTemp = planet.MaxTemp - getPlanetTemperatureAmplitude()
}

// generateResources :
// Used to perform the creation of the default resources for
// a planet when it's being created. This translate the fact
// that a planet is never really `empty` in the game.
// This function will create all the necessary entries in the
// planet input object but does not create anything in the DB.
//
// The `planet` defines the planet for which resources should
// be generated.
func (p *PlanetProxy) generateResources(planet *Planet) {
	// Discard empty planets.
	if planet == nil {
		return
	}

	// TODO: To handle this we should have some sort of `init`
	// mechanism which would retrieve the base properties of
	// the DB such as the universes, the name of the resources,
	// technologies and ships (maybe ?) so that it's readily
	// available for computations like this one.
	// Indeed it's very rare that we will add a resource and
	// it's a good approximation to not have to query the DB
	// each time we need it.
	// We could also include the rapid fire tables, etc. We
	// could also include the table defining the planets'
	// sizes and temperature in a table and retrieve this
	// as well. This seems like a more powerful solution as
	// it combines the robustness of having it in the DB so
	// it's persisted and the flexibility of being able to
	// use it directly in the code.
	// The generation of resources would then just rely on
	// the identifier of the resources fetched from this
	// model and then populate the fields in the planet.
}
