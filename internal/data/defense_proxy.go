package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"
)

// DefenseProxy :
// Intended as a wrapper to access properties of defenses
// and retrieve data from the database. This helps hiding
// the complexity of how the data is laid out in the `DB`
// and the precise name of tables from the exterior world.
//
// The `dbase` is the database that is wrapped by this
// object. It is checked for consistency upon creating the
// wrapper.
//
// The `log` allows to perform display to the user so as
// to inform of potential issues and debug information to
// the outside world.
type DefenseProxy struct {
	dbase *db.DB
	log   logger.Logger
}

// NewDefenseProxy :
// Create a new proxy on the input `dbase` to access the
// properties of defenses as registered in the DB.
// In case the provided DB is `nil` a panic is issued.
//
// The `dbase` represents the database to use to fetch
// data related to defenses.
//
// The `log` will be used to notify information so that
// we can have an idea of the activity of this component.
// One possible example is for timing the requests.
//
// Returns the created proxy.
func NewDefenseProxy(dbase *db.DB, log logger.Logger) DefenseProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create defenses proxy from invalid DB"))
	}

	return DefenseProxy{dbase, log}
}

// Defenses :
// Allows to fetch the list of defenses currently available for
// a player to build on a planet. This list is normally never
// changed as it's very unlikely to create a new defense.
// The user can choose to filter parts of the defenses using
// an array of filters that will be applied to the SQL query.
// No controls is enforced on the filters so one should make
// sure that it's consistent with the underlying table.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// defenses available. Each one is appended `as-is` to the
// query.
//
// Returns the list of defenses along with any errors. Note
// that in case the error is not `nil` the returned list is to
// be ignored.
func (p *DefenseProxy) Defenses(filters []DBFilter) ([]DefenseDesc, error) {
	// Create the query and execute it.
	props := []string{
		"id",
		"name",
	}

	query := fmt.Sprintf("select %s from defenses", strings.Join(props, ", "))
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
		return nil, fmt.Errorf("Could not query DB to fetch defenses (err: %v)", err)
	}

	// Populate the return value.
	defenses := make([]DefenseDesc, 0)
	var def DefenseDesc

	for rows.Next() {
		err = rows.Scan(
			&def.ID,
			&def.Name,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for defense (err: %v)", err))
			continue
		}

		def.Cost, err = fetchElementCost(p.dbase, def.ID, "defense", "defenses_costs")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch cost for defense \"%s\" (err: %v)", def.ID, err))
		}

		def.BuildingsDeps, err = fetchElementDependency(p.dbase, def.ID, "defense", "tech_tree_defenses_vs_buildings")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch building dependencies for defense \"%s\" (err: %v)", def.ID, err))
		}

		def.TechnologiesDeps, err = fetchElementDependency(p.dbase, def.ID, "defense", "tech_tree_defenses_vs_technologies")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch technologies dependencies for defense \"%s\" (err: %v)", def.ID, err))
		}

		defenses = append(defenses, def)
	}

	return defenses, nil
}
