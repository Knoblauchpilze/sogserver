package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"
)

// TechnologyProxy :
// Intended as a wrapper to access properties of technologies
// and retrieve data from the database. This helps hiding the
// complexity of how the data is laid out in the `DB` and the
// precise name of tables from the exterior world.
//
// The `dbase` is the database that is wrapped by this object.
// It is checked for consistency upon creating the wrapper.
//
// The `log` allows to perform display to the user so as
// to inform of potential issues and debug information to
// the outside world.
type TechnologyProxy struct {
	dbase *db.DB
	log   logger.Logger
}

// NewTechnologyProxy :
// Create a new proxy on the input `dbase` to access the
// properties of technologies as registered in the DB.
// In case the provided DB is `nil` a panic is issued.
//
// The `dbase` represents the database to use to fetch
// data related to technologies.
//
// The `log` will be used to notify information so that
// we can have an idea of the activity of this component.
// One possible example is for timing the requests.
//
// Returns the created proxy.
func NewTechnologyProxy(dbase *db.DB, log logger.Logger) TechnologyProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create technologies proxy from invalid DB"))
	}

	return TechnologyProxy{dbase, log}
}

// Technologies :
// Allows to fetch the list of technologies currently available
// for a player to research. This list normally never changes as
// it's very unlikely to create a new technology.
// The user can choose to filter parts of the technologies using
// an array of filters that will be applied to the SQL query.
// No controls is enforced on the filters so one should make
// sure that it's consistent with the underlying table.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// technologies available. Each one is appended `as-is` to the
// query.
//
// Returns the list of technologies along with any errors. Note
// that in case the error is not `nil` the returned list is to
// be ignored.
func (p *TechnologyProxy) Technologies(filters []DBFilter) ([]TechnologyDesc, error) {
	// Create the query and execute it.
	props := []string{
		"id",
		"name",
	}

	query := fmt.Sprintf("select %s from technologies", strings.Join(props, ", "))
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
		return nil, fmt.Errorf("Could not query DB to fetch technologies (err: %v)", err)
	}

	// Populate the return value.
	technologies := make([]TechnologyDesc, 0)
	var tech TechnologyDesc

	for rows.Next() {
		err = rows.Scan(
			&tech.ID,
			&tech.Name,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for technology (err: %v)", err))
			continue
		}

		tech.BuildingsDeps, err = fetchElementDependency(p.dbase, tech.ID, "technology", "tech_tree_technologies_vs_buildings")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch building dependencies for technology \"%s\" (err: %v)", tech.ID, err))
		}

		tech.TechnologiesDeps, err = fetchElementDependency(p.dbase, tech.ID, "technology", "tech_tree_technologies_dependencies")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch technologies dependencies for technology \"%s\" (err: %v)", tech.ID, err))
		}

		technologies = append(technologies, tech)
	}

	return technologies, nil
}
