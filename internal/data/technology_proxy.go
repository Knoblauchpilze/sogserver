package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// TechnologyProxy :
// Intended as a wrapper to access properties of technologies
// and retrieve data from the database. Internally uses the
// common proxy defined in this package.
type TechnologyProxy struct {
	commonProxy
}

// NewTechnologyProxy :
// Create a new proxy allowing to serve the requests
// related to technologies.
//
// The `dbase` represents the database to use to fetch
// data related to technologies.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewTechnologyProxy(dbase *db.DB, log logger.Logger) TechnologyProxy {
	return TechnologyProxy{
		newCommonProxy(dbase, log),
	}
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
	query := queryDesc{
		props: []string{
			"id",
			"name",
		},
		table:   "technologies",
		filters: filters,
	}

	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return []TechnologyDesc{}, fmt.Errorf("Could not query DB to fetch technologies (err: %v)", err)
	}

	// Populate the return value.
	technologies := make([]TechnologyDesc, 0)
	var desc TechnologyDesc

	for res.next() {
		err = res.scan(
			&desc.ID,
			&desc.Name,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for technology (err: %v)", err))
			continue
		}

		desc.BuildingsDeps, err = fetchElementDependency(p.dbase, desc.ID, "technology", "tech_tree_technologies_vs_buildings")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch building dependencies for technology \"%s\" (err: %v)", desc.ID, err))
		}

		desc.TechnologiesDeps, err = fetchElementDependency(p.dbase, desc.ID, "technology", "tech_tree_technologies_dependencies")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch technologies dependencies for technology \"%s\" (err: %v)", desc.ID, err))
		}

		technologies = append(technologies, desc)
	}

	return technologies, nil
}
