package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// BuildingProxy :
// Intended as a wrapper to access properties of buildings
// and retrieve data from the database. Internally uses the
// common proxy defined in this package.
type BuildingProxy struct {
	commonProxy
}

// NewBuildingProxy :
// Create a new proxy allowing to serve the requests
// related to buildings.
//
// The `dbase` represents the database to use to fetch
// data related to buildings.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewBuildingProxy(dbase *db.DB, log logger.Logger) BuildingProxy {
	return BuildingProxy{
		newCommonProxy(dbase, log),
	}
}

// Buildings :
// Allows to fetch the list of buildings currently available
// for a player to build. This list normally never changes as
// it's very unlikely to create a new building.
// The user can choose to filter parts of the buildings using
// an array of filters that will be applied to the SQL query.
// No controls is enforced on the filters so one should make
// sure that it's consistent with the underlying table.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// buildings available. Each one is appended `as-is` to the
// query.
//
// Returns the list of buildings along with any errors. Note
// that in case the error is not `nil` the returned list is
// to be ignored.
func (p *BuildingProxy) Buildings(filters []DBFilter) ([]BuildingDesc, error) {
	// Create the query and execute it.
	query := queryDesc{
		props: []string{
			"id",
			"name",
		},
		table:   "buildings",
		filters: filters,
	}

	// Create the query and execute it.
	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch buildings (err: %v)", err)
	}

	// Populate the return value.
	buildings := make([]BuildingDesc, 0)
	var desc BuildingDesc

	for res.next() {
		err = res.scan(
			&desc.ID,
			&desc.Name,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for building (err: %v)", err))
			continue
		}

		desc.BuildingsDeps, err = fetchElementDependency(p.dbase, desc.ID, "building", "tech_tree_buildings_dependencies")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch building dependencies for building \"%s\" (err: %v)", desc.ID, err))
		}

		desc.TechnologiesDeps, err = fetchElementDependency(p.dbase, desc.ID, "building", "tech_tree_buildings_vs_technologies")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch technologies dependencies for building \"%s\" (err: %v)", desc.ID, err))
		}

		buildings = append(buildings, desc)
	}

	return buildings, nil
}
