package db

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx"
)

// QueryDesc :
// Defines an abstract query where some fields can be
// configured to adapt in a certain extent to various
// queries.
// The produced query will be something like below:
// `select [props] from [table] where [filters]`.
//
// The `Props` define the list of properties to select
// by the query. Each property will be listed in order
// compared to the order defined in this slice. They
// will be joined by a ',' character and not prefixed
// by any table.
//
// The `Table` defines the table into which the props
// should be queried. Note that it is perfectly valid
// to have a composed table in here as long as the
// props account for that (typically if the table is
// `aTable a inner join anotherTable b on a.id=b.id`
// the properties should either be unique or prefixed
// with the name of the table).
//
// The `v` will be appended in the `where` clause
// of the generated SQL query. Each filter is added
// as a `and` statement to the others.
type QueryDesc struct {
	Props   []string
	Table   string
	Filters []Filter
}

// valid :
// Used to determine whether the query is obviously
// not valid.
//
// Returns `true` if the query is not obviously wrong.
func (q QueryDesc) valid() bool {
	return len(q.Props) > 0 && len(q.Table) > 0
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
func (q QueryDesc) generate() string {
	// Generate base query.
	str := fmt.Sprintf("select %s from %s", strings.Join(q.Props, ", "), q.Table)

	// Append filters if any.
	if len(q.Filters) > 0 {
		str += " where"

		for id, filter := range q.Filters {
			if id > 0 {
				str += " and"
			}
			str += fmt.Sprintf(" %s", filter)
		}
	}

	return str
}

// QueryResult :
// Defines the result of a query as executed by the
// main DB. This small wrapper allows to automatically
// cycle through remaining rows when it goes out of
// scope through the `Closer` interface.
//
// The `rows` defines the low level rows returned by
// the execution of the query.
//
// The `Err` defines the error that was associated
// with the query itself.
type QueryResult struct {
	rows *pgx.Rows
	Err  error
}

// Next :
// Forward the call to the internal rows object so
// that we move to the next row of the result.
//
// Returns `true` if there are more rows.
func (q QueryResult) Next() bool {
	return q.rows.Next()
}

// Scan :
// Forward the call to the internal rows object so
// that the content of the row is retrieved.
//
// The `dest` defines the destination elements where
// the current row should be queried.
//
// Returns any error.
func (q QueryResult) Scan(dest ...interface{}) error {
	return q.rows.Scan(dest...)
}

// Close :
// Implementation of the `Closer` interface which will
// release the remaining rows described by this object
// if any, making sure that the connection to the DB
// is released.
func (q QueryResult) Close() {
	if q.rows != nil {
		q.rows.Close()
	}
}

// InsertReq :
// Used to describe the data to be inserted to the DB
// through abstract common properties. Just like the
// `queryDesc` it allows to mutualize most of the code
// to perform the formatting of the data in order to
// insert it into the DB.
//
// The `Script` defines the name of the function that
// should be called to perform the insertion. This
// function should accept a number of arguments that
// matches the number provided in `args`.
//
// The `Args` represent an array of interface that
// should be marshalled and send as positionnal params
// of the insertion script. The arguments will be
// passed to the script in the order they are defined
// in this slice.
//
// The `SkipReturn` boolean indicates whether the
// insertion request expects a return a value or not.
// This allows to precise the syntax to use to perform
// the query.
type InsertReq struct {
	Script     string
	Args       []interface{}
	SkipReturn bool
}

// Convertible :
// Used as an interface allowing to convert an element
// into a specialized version that should be used in a
// request to insert this data in the DB. Typically it
// allows for types to define a custom facet when being
// exported to the DB (by ignoring some field or even
// restructuring the fields marshalled to the DB).
// This type will be searched in the input interface
// and be called before performing the seralization to
// insert the data in the DB.
type Convertible interface {
	Convert() interface{}
}

// Proxy :
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
type Proxy struct {
	dbase *DB
}

// NewProxy :
// Performs the creation of a new common proxy from the
// input database and logger.
//
// The `dbase` defines the main DB that should be wrapped
// by this object.
//
// Returns the created object.
func NewProxy(dbase *DB) Proxy {
	return Proxy{
		dbase: dbase,
	}
}

// FetchFromDB :
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
func (p Proxy) FetchFromDB(query QueryDesc) (QueryResult, error) {
	// Check for invalid DB.
	if p.dbase == nil {
		return QueryResult{}, ErrInvalidDB
	}

	// Check the query to make sure it is valid.
	if !query.valid() {
		return QueryResult{}, ErrInvalidQuery
	}

	// Generate the string from the input query properties.
	queryStr := query.generate()

	// Execute it and return the produced data.
	var res QueryResult
	res.rows, res.Err = p.dbase.DBQuery(queryStr)

	return res, nil
}

// InsertToDB :
// Used to perform the insertion of the input data to the
// DB by marshalling it and using the provided insertion
// script to perform the DB request.
//
// The `req` defines an abstract description of the req
// to perform in the DB.
//
// Returns any error occuring while performing the DB
// operation.
func (p Proxy) InsertToDB(req InsertReq) error {
	// Check for invalid DB.
	if p.dbase == nil {
		return ErrInvalidDB
	}

	// Marshal all the elements provided as arguments of
	// the insertion script.
	argsAsStr := make([]string, 0)

	for _, arg := range req.Args {
		// Check whether this argument can be converted
		// into a more meaningful type.
		cvrt, ok := arg.(Convertible)

		var raw []byte
		var err error

		// Marshal the argument using either the converted
		// value or the base value.
		if ok {
			raw, err = json.Marshal(cvrt.Convert())
		} else {
			// Make sure that the string is not `double quoted`:
			// this would not work well with the SQL syntax and
			// can happen in case the argument is a string in
			// itself.
			str, ok := arg.(string)

			if ok {
				raw = []byte(str)
			} else {
				raw, err = json.Marshal(arg)
			}
		}

		if err != nil {
			return ErrInvalidData
		}

		// Quote the string to be consistent with the SQL
		// syntax and register it.
		argAsStr := fmt.Sprintf("'%s'", string(raw))

		argsAsStr = append(argsAsStr, argAsStr)
	}

	// Create the DB request.
	var query string

	switch req.SkipReturn {
	case false:
		query = fmt.Sprintf("SELECT * from %s(%s)", req.Script, strings.Join(argsAsStr, ", "))
	default: // true
		query = fmt.Sprintf("SELECT %s(%s)", req.Script, strings.Join(argsAsStr, ", "))
	}

	_, err := p.dbase.DBExecute(query)

	return formatDBError(err)
}
