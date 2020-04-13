package data

import (
	"fmt"
	"oglike_server/internal/locker"
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

// performWithLock :
// Used to exectue the specified query on the provided DB by
// making sure that the lock on the specified resource is
// acquired and released when needed.
//
// The `resource` represents an identifier of the resoure to
// access with the query: this method makes sure that a lock
// on this resource is created and handled adequately.
//
// The `dbase` represents the DB into which the query should
// be performed. Should not be `nil` otheriwse a panic will
// be issued.
//
// The `query` represents the operation to perform on the DB
// which should be protected with a lock. It should consist
// in a valid SQL query.
//
// The `cl` allows to acquire a locker on the resource to
// make sure that a single routine is executing the update
// on the input resource at once.
//
// Returns any error occurring during the process.
func performWithLock(resource string, dbase *db.DB, query string, cl *locker.ConcurrentLocker) error {
	// Prevent invalid resource identifier.
	if resource == "" {
		return fmt.Errorf("Cannot update resources for invalid empty id")
	}

	// Acquire a lock on this resource.
	resLock := cl.Acquire(resource)
	defer cl.Release(resLock)

	// Perform the update: we will wrap the function inside
	// a dedicated handler to make sure that we don't lock
	// the resource more than necessary.
	var err error
	var errRelease error
	var errExec error

	func() {
		resLock.Lock()
		defer func() {
			if rawErr := recover(); rawErr != nil {
				err = fmt.Errorf("Error occured while executing query (err: %v)", rawErr)
			}
			errRelease = resLock.Release()
		}()

		// Perform the update.
		_, errExec = dbase.DBExecute(query)
	}()

	// Return any error.
	if errExec != nil {
		return fmt.Errorf("Could not perform operation on resource \"%s\" (err: %v)", resource, errExec)
	}
	if errRelease != nil {
		return fmt.Errorf("Could not release locker protecting resource \"%s\" (err: %v)", resource, err)
	}
	if err != nil {
		return fmt.Errorf("Could not execute operation on resource \"%s\" (err: %v)", resource, err)
	}

	return nil
}
