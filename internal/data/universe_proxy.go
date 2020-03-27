package data

import (
	"encoding/json"
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"

	"github.com/google/uuid"
)

// UniverseProxy :
// Intended as a wrapper to access properties of universes
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
type UniverseProxy struct {
	dbase *db.DB
	log   logger.Logger
}

// NewUniverseProxy :
// Create a new proxy on the input `dbase` to access the
// properties of universes as registered in the DB. In
// case the provided DB is `nil` a panic is issued.
// Information in the following thread helped shape this
// component:
// https://www.reddit.com/r/golang/comments/9i5cpg/good_approach_to_interacting_with_databases/
//
// The `dbase` represents the database whose accesses are
// to be wrapped.
//
// The `log` will be used to notify information so that
// we can have an idea of the activity of this component.
// One possible example is for timing the requests.
//
// Returns the created proxy.
func NewUniverseProxy(dbase *db.DB, log logger.Logger) UniverseProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create universes proxy from invalid DB"))
	}

	return UniverseProxy{dbase, log}
}

// Universes :
// Allows to fetch the list of universes currently available
// for a player to create an account. Universes should only
// be created when needed and are not typically something a
// player can do.
//
// Returns the list of universes along with any errors. Note
// that in case the error is not `nil` the returned list is
// to be ignored.
func (p *UniverseProxy) Universes() ([]Universe, error) {
	// Create the query and execute it.
	props := []string{
		"id",
		"name",
		"economic_speed",
		"fleet_speed",
		"research_speed",
		"fleets_to_ruins_ratio",
		"defenses_to_ruins_ratio",
		"consumption_ratio",
		"galaxy_count",
		"solar_system_size",
	}

	query := fmt.Sprintf("select %s from universes", strings.Join(props, ", "))
	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch universes (err: %v)", err)
	}

	// Populate the return value.
	universes := make([]Universe, 0)
	var uni Universe

	for rows.Next() {
		err = rows.Scan(
			&uni.ID,
			&uni.Name,
			&uni.EcoSpeed,
			&uni.FleetSpeed,
			&uni.ResearchSpeed,
			&uni.FleetsToRuins,
			&uni.DefensesToRuins,
			&uni.FleetConsumption,
			&uni.GalaxiesCount,
			&uni.SolarSystemSize,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for universe (err: %v)", err))
			continue
		}

		universes = append(universes, uni)
	}

	return universes, nil
}

// Planets :
// Return a list of planets associated registered in the
// universe specified in input. It queries the DB to fetch
// the relevant data. Note that only the created planets
// will be returned.
//
// The `uni` describes the universe for which planetes
// should be returned. In case it does not represent a
// valid universe the returned planets list will most
// likely be empty. It is only represented using its
// identifier.
//
// Returns the list of planets for this universe along
// with any error. In case the error is not `nil` the
// value of the array should be ignored.
func (p *UniverseProxy) Planets(uni string) ([]Planet, error) {
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
	where := fmt.Sprintf("pl.uni='%s'", uni)

	query := fmt.Sprintf("select %s from %s on %s where %s", strings.Join(props, ", "), table, joinCond, where)
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

		planets = append(planets, planet)
	}

	return planets, nil
}

// Planet :
// Attempts to retrieve the planet with an identifier in
// concordance with the input value in the universe that
// is described by the `uni` string.
// If no such planet exist an error is returned.
//
// The `uni` defines the identifier of the universe into
// which the planet should be searched.
//
// The `planet` defines the supposed identifier of the
// planet to fetch.
//
// Returns the corresponding planet or an error in case
// the planet cannot be found in the provided universe.
func (p *UniverseProxy) Planet(uni string, planet string) (Planet, error) {
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
	where := fmt.Sprintf("pl.uni='%s' and p.id='%s'", uni, planet)

	query := fmt.Sprintf("select %s from %s on %s where %s", strings.Join(props, ", "), table, joinCond, where)
	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return Planet{}, fmt.Errorf("Could not query DB to fetch planets (err: %v)", err)
	}

	// Populate the return value: we should obtain a single
	// result, otherwise it's an issue.
	var pl Planet

	galaxy := 0
	system := 0
	position := 0

	// Scan the first row.
	err = rows.Scan(
		&pl.ID,
		&pl.PlayerID,
		&pl.Name,
		&pl.Fields,
		&pl.MinTemp,
		&pl.MaxTemp,
		&pl.Diameter,
		&galaxy,
		&system,
		&position,
	)

	// Check for errors.
	if err != nil {
		return pl, fmt.Errorf("Could not retrieve info for planet \"%s\" from universe \"%s\" (err: %v)", planet, uni, err)
	}

	pl.Coords = Coordinate{
		galaxy,
		system,
		position,
	}

	// Skip remaining values (if any), and indicate the problem
	// if there are more.
	count := 0
	for rows.Next() {
		count++
	}
	if count > 0 {
		err = fmt.Errorf("Found %d values for planet \"%s\" in universe \"%s\"", count, planet, uni)
	}

	return pl, err
}

// Buildings :
// Return a list of the buildings currently built on the
// planet specified as input. This will automatically use
// the universe associated to the planet to fetch needed
// data.
//
// The `planet` defines the planet for which the list of
// buildings should be fetched. It is only described by
// its identifier.
//
// Returns a list of buildings for the specified planet.
// If the planet's identifier is not valid the return
// list will most likely be invalid. In case no buildings
// are built on the planet the output list will also be
// empty. It should be ignored in case the error is not
// `nil`.
func (p *UniverseProxy) Buildings(planet string) ([]Building, error) {
	// Create the query and execute it.
	props := []string{
		"building",
		"level",
	}

	query := fmt.Sprintf("select %s from planets_buildings where planet='%s'", strings.Join(props, ", "), planet)
	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch buildings for planet \"%s\" (err: %v)", planet, err)
	}

	// Populate the return value.
	buildings := make([]Building, 0)
	var building Building

	for rows.Next() {
		err = rows.Scan(
			&building.ID,
			&building.Level,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve buildings for planet \"%s\" (err: %v)", planet, err))
			continue
		}

		buildings = append(buildings, building)
	}

	return buildings, nil
}

// Defenses :
// Return a list of the defenses currently built on the
// planet specified as input. This will automatically use
// the universe associated to the planet to fetch needed
// data.
//
// The `planet` defines the planet for which the list of
// defenses should be fetched. It is only described by
// its identifier.
//
// Returns a list of defenses for the specified planet.
// If the planet's identifier is not valid the return
// list will most likely be invalid. In case no defenses
// are built on the planet the output list will also be
// empty. It should be ignored in case the error is not
// `nil`.
func (p *UniverseProxy) Defenses(planet string) ([]Defense, error) {
	// Create the query and execute it.
	props := []string{
		"defense",
		"count",
	}

	query := fmt.Sprintf("select %s from planets_defenses where planet='%s'", strings.Join(props, ", "), planet)
	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch defenses for planet \"%s\" (err: %v)", planet, err)
	}

	// Populate the return value.
	defenses := make([]Defense, 0)
	var defense Defense

	for rows.Next() {
		err = rows.Scan(
			&defense.ID,
			&defense.Count,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve defenses for planet \"%s\" (err: %v)", planet, err))
			continue
		}

		defenses = append(defenses, defense)
	}

	return defenses, nil
}

// Ships :
// Similar to `Defenses` but returns the number of ships
// currently available on the specified planet. Note that
// it does not include the fleets that are moving towards
// the planet or leaving from it.
//
// The `planet` defines the planet for which the list of
// ships should be fetched. It is only described by
// its identifier.
//
// Returns a list of the ships currently available on the
// specified planet. The list is empty if no ships are
// available and should be ignored if the associated error
// is not `nil`.
func (p *UniverseProxy) Ships(planet string) ([]Ship, error) {
	// Create the query and execute it.
	props := []string{
		"ship",
		"count",
	}

	query := fmt.Sprintf("select %s from planets_ships where planet='%s'", strings.Join(props, ", "), planet)
	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch ships for planet \"%s\" (err: %v)", planet, err)
	}

	// Populate the return value.
	ships := make([]Ship, 0)
	var ship Ship

	for rows.Next() {
		err = rows.Scan(
			&ship.ID,
			&ship.Count,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve ships for planet \"%s\" (err: %v)", planet, err))
			continue
		}

		ships = append(ships, ship)
	}

	return ships, nil
}

// Fleets :
// Similar to the `Ships` method but returns the list of
// fleets that are directed towards or start from this
// planet. Note that it accounts both for friendly but
// also enemy fleets.
//
// The `planet` defines the planet for which the list of
// fleets should be fetched.
//
// Returns a list of the fleets currently directed towards
// the planet or leaving from it no matter their objectives.
// The list should be ignored if the error is not `nil`.
func (p *UniverseProxy) Fleets(planet Planet) ([]Fleet, error) {
	// /universes/universe_id/planet_id/fleets
	return nil, fmt.Errorf("Not implemented")
}

// Create :
// Used to perform the creation of the universe described
// by the input data to the DB. In case the creation cannot
// be performed an error is returned.
//
// The `uni` describes the element to create in DB.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`.
func (p *UniverseProxy) Create(uni Universe) error {
	// Assign a valid identifier if this is not already the case.
	if uni.ID == "" {
		uni.ID = uuid.New().String()
	}

	// TODO: Handle controls to make sure that the universes are
	// not created with invalid value (such as negative galaxies
	// count, etc.).

	// Marshal the input universe to pass it to the import script.
	data, err := json.Marshal(uni)
	if err != nil {
		return fmt.Errorf("Could not import universe \"%s\" (err: %v)", uni.Name, err)
	}
	jsonToSend := string(data)

	query := fmt.Sprintf("select * from create_universe('%s')", jsonToSend)
	_, err = p.dbase.DBExecute(query)

	// Check for errors.
	if err != nil {
		return fmt.Errorf("Could not import universe \"%s\" (err: %v)", uni.Name, err)
	}

	// Successfully created a universe.
	p.log.Trace(logger.Notice, fmt.Sprintf("Created new universe \"%s\" with id \"%s\"", uni.Name, uni.ID))

	// All is well.
	return nil
}
