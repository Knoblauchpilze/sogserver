package data

import (
	"encoding/json"
	"fmt"
	"oglike_server/internal/locker"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"

	"github.com/jackc/pgx"
)

// commonProxy :
// Intended as a common wrapper to access the main DB
// through a convenience way. It holds most of the
// common resources needed to acces the DB and notify
// errors/information to the user about processes that
// may occur while fetching data. This helps hiding
// the complexity of how the data is laid out in the
// `DB` and the precise name of tables from the rest
// of the application.
// The following link contains useful information on
// the paradigm we're following with this object:
// https://www.reddit.com/r/golang/comments/9i5cpg/good_approach_to_interacting_with_databases/
//
// The `dbase` is the database that is wrapped by this
// object. It is checked for consistency upon building
// the wrapper.
//
// The `log` allows to perform display to the user so as
// to inform of potential issues and debug information to
// the outside world.
//
// The `lock` allows to lock specific resources when some
// data should be retrieved. For example in the case of
// a planet, one might first want to update the upgrade
// actions that are built on this planet in order to be
// sure to get up-to-date content.
// This typically include checking whether some buildings
// have reached completion and upgrading the resources
// that are present on the planet.
// Each of these actions are executed in a lazy update
// fashion where the action is created and then performed
// only when needed: we store the completion time and if
// the data needs to be accessed we upgrade it.
// This mechanism requires that when the data needs to be
// fetched for a planet an update operation should first
// be performed to ensure that the data is up-to-date.
// As no data is usually shared across elements of a same
// collection we don't see the need to lock all of them
// when a single one should be updated. Using the structure
// defined in the `ConcurrentLock` we have a way to lock
// only some elements which is exactly what we need.
type commonProxy struct {
	dbase *db.DB
	log   logger.Logger
	lock  *locker.ConcurrentLocker
}

// queryDesc :
// Defines an abstract query where some fields can be
// configured to adapt in a certain extent to various
// queries.
// The produced query will be something like below:
// `select [props] from [table] where [filters]`.
//
// The `props` define the list of properties to select
// by the query. Each property will be listed in order
// compared to the order defined in this slice. They
// will be joined by a ',' character and not prefixed
// by any table.
//
// The `table` defines the table into which the props
// should be queried. Note that it is perfectly valid
// to have a composed table in here as long as the
// props account for that (typically if the table is
// `aTable a inner join anotherTable b on a.id=b.id`
// the properties should either be unique or prefixed
// with the name of the table).
//
// The `filters` will be appended in the `where` clause
// of the generated SQL query. Each filter is added
// as a `and` statement to the others.
type queryDesc struct {
	props   []string
	table   string
	filters []DBFilter
}

// valid :
// Used to determine whether the query is obviously
// not valid.
//
// Returns `true` if the query is not obviously wrong.
func (q queryDesc) valid() bool {
	return len(q.props) > 0 && len(q.table) > 0
}

// generate :
// Used to perform the generation of a valid SQL query
// from the data registered in this query. This method
// assumes that the query is valid (which is verified
// with the `valid` method of this object) and does not
// perform additional checks.
//
// Returns a string representing the query. The string
// is only guaranteed to be valid if `q.valid()` is
// `true`.
func (q queryDesc) generate() string {
	// Generate base query.
	str := fmt.Sprintf("select %s from %s", strings.Join(q.props, ", "), q.table)

	// Append filters if any.
	if len(q.filters) > 0 {
		str += " where"

		for id, filter := range q.filters {
			if id > 0 {
				str += " and"
			}
			str += fmt.Sprintf(" %s", filter)
		}
	}

	return str
}

// queryRows :
// Defines the result of a query as executed by the
// main DB. This small wrapper allows to automatically
// cycle through remaining rows when it goes out of
// scope through the `Closer` interface.
//
// The `rows` defines the low level rows returned by
// the execution of the query.
//
// The `err` defines the error that was associated
// with the query itself.
type queryResult struct {
	rows *pgx.Rows
	err  error
}

// next :
// Forward the call to the internal rows object so
// that we move to the next row of the result.
//
// Returns `true` if there are more rows.
func (q queryResult) next() bool {
	return q.rows.Next()
}

// scan :
// Forward the call to the internal rows object so
// that the content of the row is retrieved.
//
// The `dest` defines the destination elements where
// the current row should be queried.
//
// Returns any error.
func (q queryResult) scan(dest ...interface{}) error {
	return q.rows.Scan(dest)
}

// Close :
// Implementation of the `Closer` interface which will
// release the remaining rows described by this object
// if any, making sure that the connection to the DB
// is released.
func (q queryResult) Close() {
	if q.rows != nil {
		q.rows.Close()
	}
}

// insertReq :
// Used to describe the data to be inserted to the DB
// through abstract common properties. Just like the
// `queryDesc` it allows to mutualize most of the code
// to perform the formatting of the data in order to
// insert it into the DB.
//
// The `script` defines the name of the function that
// should be called to perform the insertion. This
// function should accept a number of arguments that
// matches the number provided in `args`.
//
// The `args` represent an array of interface that
// should be marshalled and send as positionnal params
// of the insertion script. The arguments will be
// passed to the script in the order they are defined
// in this slice.
type insertReq struct {
	script string
	args   []interface{}
}

// newCommonProxy :
// Performs the creation of a new common proxy from the
// input database and logger.
//
// The `dbase` defines the main DB that should be wrapped
// by this object.
//
// The `log` defines the logger allowing to notify errors
// or info to the user.
//
// Returns the created object and panics if something is
// not right when creating the proxy.
func newCommonProxy(dbase *db.DB, log logger.Logger) commonProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create common proxy from invalid DB"))
	}

	return commonProxy{
		dbase: dbase,
		log:   log,
		lock:  locker.NewConcurrentLocker(log),
	}
}

// performWithLock :
// Used to exectue the specified query on the internal DB
// while making sure that the lock on the specified ID is
// acquired and released when needed.
//
// The `resource` represents an identifier of the resoure
// to access with the query: this method makes sure that
// a lock on this resource is created and handled so as
// to ensure that a single process is accessing it at any
// time.
//
// The `query` represents the operation to perform on the
// DB which should be protected with a lock. It should
// consist in a valid SQL query.
//
// Returns any error occurring during the process.
func (cp commonProxy) performWithLock(resource string, query string) error {
	// Prevent invalid resource identifier.
	if resource == "" {
		return fmt.Errorf("Cannot perform query \"%s\" for invalid empty resource ID", query)
	}

	// Acquire a lock on this resource.
	resLock := cp.lock.Acquire(resource)
	defer cp.lock.Release(resLock)

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
		_, errExec = cp.dbase.DBExecute(query)
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

// fetchDB :
// Used to perform the query defined by the input argument
// in the main DB. The return value is described through a
// local structure allowing to manipulate more easily the
// results.
//
// The `query` defines the query to perform.
//
// Returns the rows as fetched from the DB along with any
// errors. Note that we distinguish any errors that can
// have occurred during the execution of the query from an
// error that was returned *before* the execution of the
// query.
func (cp commonProxy) fetchDB(query queryDesc) (queryResult, error) {
	// Check the query to make sure it is valid.
	if !query.valid() {
		return queryResult{}, fmt.Errorf("Cannot perform invalid query in DB")
	}

	// Generate the string from the input query properties.
	queryStr := query.generate()

	// Execute it and return the produced data.
	var res queryResult
	res.rows, res.err = cp.dbase.DBQuery(queryStr)

	return res, nil
}

// insertToDB :
// Used to perform the insertion of the input data to the
// DB by marshalling it and using the provided insertion
// script to perform the DB request.
//
// The `req` defines an abstract description of the req
// to perform in the DB.
//
// Returns any error occuring while performing the DB
// operation.
func (cp commonProxy) insertToDB(req insertReq) error {
	// Marshal all the elements provided as arguments of
	// the insertion script.
	argsAsStr := make([]string, 0)

	for _, arg := range req.args {
		// Marshal the argument
		raw, err := json.Marshal(arg)

		if err != nil {
			return fmt.Errorf("Could not import data through \"%s\" (err: %v)", req.script, err)
		}

		// Quote the string to be consistent with the SQL
		// syntax and register it.
		argAsStr := fmt.Sprintf("'%s'", string(raw))

		argsAsStr = append(argsAsStr, argAsStr)
	}

	// Create the DB request.
	query := fmt.Sprintf("select * from %s()", strings.Join(argsAsStr, ", "))
	_, err := cp.dbase.DBExecute(query)

	return err
}
