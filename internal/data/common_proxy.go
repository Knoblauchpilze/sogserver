package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"strings"
)

// fetchElementDependency :
// Used to fetch the dependencies for the input element. We don't
// need to know the element in itself, we will just use it as a
// filter when fetching elements from the table described in input.
//
// The `dbase` defines the DB that should be used to fetch data.
// We assume that this value is valid (i.e. not `nil`). A panic
// may be issued if this is not the case.
//
// The `element` represents an identifier that can be used to only
// get some matching in the `table` where dependencies should be
// fetched.
//
// The `filterName` defines the name of the column to which the
// `element` should be applied. This will basically be translated
// into something like `where filterName='element'` in the SQL
// query.
//
// The `table` describes the name of the table from which deps are
// to be fetched.
//
// Returns the list of dependencies as described in the DB for
// the input `element` along with any error (in which case the
// list of dependencies should be ignored).
func fetchElementDependency(dbase *db.DB, element string, filterName string, table string) ([]TechDependency, error) {
	// Check consistency.
	if element == "" {
		return []TechDependency{}, fmt.Errorf("Cannot fetch dependencies for invalid element")
	}

	// Build and execute the query.
	props := []string{
		"requirement",
		"level",
	}

	query := fmt.Sprintf("select %s from %s where %s='%s'", strings.Join(props, ", "), table, filterName, element)

	// Execute the query.
	rows, err := dbase.DBQuery(query)

	if err != nil {
		return []TechDependency{}, fmt.Errorf("Could not retrieve dependencies for \"%s\" (err: %v)", element, err)
	}

	// Populate the dependencies.
	var gError error

	deps := make([]TechDependency, 0)
	var dep TechDependency

	for rows.Next() {
		err = rows.Scan(
			&dep.ID,
			&dep.Level,
		)

		if err != nil {
			gError = fmt.Errorf("Could not retrieve info for dependency of \"%s\" (err: %v)", element, err)
		}

		deps = append(deps, dep)
	}

	return deps, gError
}

// fetchElementCost :
// Used to fetch the cost to build the input element. We don't
// really need to interpret the element, we will just fetch the
// table indicated by the input arguments and search for elems
// matching the `element` key.
//
// The `dbase` defines the DB that should be queried to retrieve
// data. We assume that this value is valid (i.e. not `nil`) and
// a panic may be issued if this is not the case.
//
// The `element` defines the filtering key that will be searched
// in the corresponding table of the DB.
//
// The `filterName` defines the name of the column to which the
// `element` should be applied. This will basically be translated
// into something like `where filterName='element'` in the SQL
// query.
//
// The `table` describes the name of the table from which the
// costs should be retrieved.
//
// Returns the list of costs registered for the element in the
// DB. In case the error value is not `nil` the list should be
// ignored.
func fetchElementCost(dbase *db.DB, element string, filterName string, table string) ([]ResourceAmount, error) {
	// Check consistency.
	if element == "" {
		return []ResourceAmount{}, fmt.Errorf("Cannot fetch costs for invalid element")
	}

	// Build and execute the query.
	props := []string{
		"res",
		"cost",
	}

	query := fmt.Sprintf("select %s from %s where %s='%s'", strings.Join(props, ", "), table, filterName, element)

	// Execute the query.
	rows, err := dbase.DBQuery(query)

	if err != nil {
		return []ResourceAmount{}, fmt.Errorf("Could not retrieve costs for \"%s\" (err: %v)", element, err)
	}

	// Populate the costs.
	var gError error

	costs := make([]ResourceAmount, 0)
	var cost ResourceAmount

	for rows.Next() {
		err = rows.Scan(
			&cost.Resource,
			&cost.Amount,
		)

		if err != nil {
			gError = fmt.Errorf("Could not retrieve info for cost of \"%s\" (err: %v)", element, err)
		}

		costs = append(costs, cost)
	}

	return costs, gError
}
