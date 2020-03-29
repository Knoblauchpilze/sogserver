package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"
)

// BuildingProxy :
// Intended as a wrapper to access properties of buildings
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
type BuildingProxy struct {
	dbase *db.DB
	log   logger.Logger
}

// NewTechnologyProxy :
// Create a new proxy on the input `dbase` to access the
// properties of buildings as registered in the DB.
// In case the provided DB is `nil` a panic is issued.
//
// The `dbase` represents the database to use to fetch
// data related to buildings.
//
// The `log` will be used to notify information so that
// we can have an idea of the activity of this component.
// One possible example is for timing the requests.
//
// Returns the created proxy.
func NewBuildingProxy(dbase *db.DB, log logger.Logger) BuildingProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create buildings proxy from invalid DB"))
	}

	return BuildingProxy{dbase, log}
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
func (p *BuildingProxy) Buildings(filters []Filter) ([]BuildingDesc, error) {
	// Create the query and execute it.
	props := []string{
		"id",
		"name",
	}

	query := fmt.Sprintf("select %s from buildings", strings.Join(props, ", "))
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
		return nil, fmt.Errorf("Could not query DB to fetch buildings (err: %v)", err)
	}

	// Populate the return value.
	buildings := make([]BuildingDesc, 0)
	var bui BuildingDesc

	for rows.Next() {
		err = rows.Scan(
			&bui.ID,
			&bui.Name,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for building (err: %v)", err))
			continue
		}

		bui.BuildingsDeps = make([]TechDependency, 0)
		bui.TechnologiesDeps = make([]TechDependency, 0)

		buildings = append(buildings, bui)
	}

	// We know need to populate the dependencies that need to be
	// met in order to be able to construct this building. This
	// is also retrieved from the DB.
	var techDep TechDependency

	buildingDepsQueryTemplate := "select requirement, level from tech_tree_buildings_dependencies where building='%s'"
	techDepsQueryTemplate := "select requirement, level from tech_tree_buildings_vs_technologies where building='%s'"

	for id := range buildings {
		// Fetch the building by value.
		bui := &buildings[id]

		// Replace the building's identifier in the query template.
		buildingDepsQuery := fmt.Sprintf(buildingDepsQueryTemplate, bui.ID)

		// Execute the query.
		rows, err = p.dbase.DBQuery(buildingDepsQuery)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve building dependencies for building \"%s\" (err: %v)", bui.ID, err))
			continue
		}

		// Populate the dependency.
		for rows.Next() {
			err = rows.Scan(
				&techDep.ID,
				&techDep.Level,
			)

			if err != nil {
				p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve building dependency for building \"%s\" (err: %v)", bui.ID, err))
				continue
			}

			bui.BuildingsDeps = append(bui.BuildingsDeps, techDep)
		}
	}

	// We prefer to handle two distinct for loops as failure in the
	// first dependency gathering does not mean failure in the other
	// one with this approach. It's not optimal (because the list of
	// dependencies will not be complete).
	for id := range buildings {
		// Fetch the building by value.
		bui := &buildings[id]

		// Replace the building's identifier in the query template.
		techDepsQuery := fmt.Sprintf(techDepsQueryTemplate, bui.ID)

		// Execute the query.
		rows, err = p.dbase.DBQuery(techDepsQuery)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve technology dependencies for building \"%s\" (err: %v)", bui.ID, err))
			continue
		}

		// Populate the dependency.
		for rows.Next() {
			err = rows.Scan(
				&techDep.ID,
				&techDep.Level,
			)

			if err != nil {
				p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve technology dependency for building \"%s\" (err: %v)", bui.ID, err))
				continue
			}

			bui.TechnologiesDeps = append(bui.TechnologiesDeps, techDep)
		}
	}

	return buildings, nil
}
