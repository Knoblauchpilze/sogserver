package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// DefenseProxy :
// Intended as a wrapper to access properties of defenses
// and retrieve data from the database. Internally uses the
// common proxy defined in this package.
type DefenseProxy struct {
	commonProxy
}

// NewDefenseProxy :
// Create a new proxy allowing to serve the requests
// related to defenses.
//
// The `dbase` represents the database to use to fetch
// data related to defenses.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewDefenseProxy(dbase *db.DB, log logger.Logger) DefenseProxy {
	return DefenseProxy{
		newCommonProxy(dbase, log),
	}
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
	query := queryDesc{
		props: []string{
			"id",
			"name",
		},
		table:   "defenses",
		filters: filters,
	}

	// Create the query and execute it.
	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch defenses (err: %v)", err)
	}

	// Populate the return value.
	defenses := make([]DefenseDesc, 0)
	var desc DefenseDesc

	for res.next() {
		err = res.scan(
			&desc.ID,
			&desc.Name,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for defense (err: %v)", err))
			continue
		}

		desc.Cost, err = fetchElementCost(p.dbase, desc.ID, "defense", "defenses_costs")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch cost for defense \"%s\" (err: %v)", desc.ID, err))
		}

		desc.BuildingsDeps, err = fetchElementDependency(p.dbase, desc.ID, "defense", "tech_tree_defenses_vs_buildings")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch building dependencies for defense \"%s\" (err: %v)", desc.ID, err))
		}

		desc.TechnologiesDeps, err = fetchElementDependency(p.dbase, desc.ID, "defense", "tech_tree_defenses_vs_technologies")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch technologies dependencies for defense \"%s\" (err: %v)", desc.ID, err))
		}

		defenses = append(defenses, desc)
	}

	return defenses, nil
}
