package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"
)

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
	// TODO: Handle this.
	return fmt.Errorf("Not implemented")
}
