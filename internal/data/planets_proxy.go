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

	return PlanetProxy{dbase, log}
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

		planets = append(planets, planet)
	}

	return planets, nil
}
