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
// The `dbase` define the DB that should be used to fetch data.
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
// The `table` describe the name of the table from which deps are
// to be fetched.
//
// Returns the list of dependencies as described in the DB for
// the input `element` along with any error (in which case the
// list of dependencies should be ignored).
func fetchElementDependency(dbase *db.DB, element string, filterName string, table string) ([]TechDependency, error) {
	// Allocate the dependencies if needed and check consistency.
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
		return []TechDependency{}, fmt.Errorf("Could not retrieve technology dependencies for \"%s\" (err: %v)", element, err)
	}

	// Populate the dependency.
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
