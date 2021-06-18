package db

import (
	"fmt"
	"strconv"
	"strings"
)

// ErrInvalidQuery :
// Used in case the query to perform in the DB is either
// `nil` or empty.
var ErrInvalidQuery = fmt.Errorf("invalid nil query")

// ErrInvalidDB :
// Provides an abstract way to define that an error
// occurred while accessing some properties on the
// DB. This is especially useful to define whether
// an error returned by the `Init` method of some
// `DBModule` originates in a failure of the DB.
var ErrInvalidDB = fmt.Errorf("invalid nil DB")

// ErrInvalidData :
// Used to indicate that a marshalling error has
// been detected when trying to import some item
// in the DB.
var ErrInvalidData = fmt.Errorf("invalid data to insert to DB")

// ErrInvalidScan :
// Used to indicate that a `Scan` operation on a
// `QueryResult` has failec.
var ErrInvalidScan = fmt.Errorf("invalid scan operation on DB")

// ErrNoSQLCode :
// Defines that the error message provided in input
// does not define any SQL error code.
var ErrNoSQLCode = fmt.Errorf("no SQL code found in error message")

// Defines the possible error code as returned by
// the SQL driver.
const (
	nonNullConstraint   int = 23502
	foreignKeyViolation int = 23503
	duplicatedElement   int = 23505
)

// Error :
// Defines a generic error type which is associated to a
// SQL error. It basically defines the code that was set
// as return value for the SQL query along with the init
// error.
//
// The `SQLCode` defines the SQL error code returned by
// the query.
//
// The `Err` defines the initial error that produced
// this `DBError`.
type Error struct {
	SQLCode int
	Err     error
}

// Error :
// Implementation of the `error` interface to provide a
// description of the error.
func (e Error) Error() string {
	return fmt.Sprintf("SQL query failed with code was %d (err: %v)", e.SQLCode, e.Err.Error())
}

// NonNullConstrainteError :
// Used to define an error indicating that a non null
// constraint was violated by an insert request.
//
// The `Column` defines the name of the column that
// was meant to be populated with a null value which
// was not possible due to a constraint.
//
// The `Err` defines the initial error that caused
// the non null violation error.
type NonNullConstrainteError struct {
	Column string
	Err    error
}

// Error :
// Implementation of the `error` interface.
func (e NonNullConstrainteError) Error() string {
	return fmt.Sprintf("Query violates non null constraint on column \"%s\"", e.Column)
}

// DuplicatedElementError :
// Used to define a duplicated element in a table which
// lead to a unique key error.
//
// The `Constraint` defines the name of the unique key
// constraint that was violated by the request.
//
// The `Err` defines the initial error that caused the
// duplicated element error.
type DuplicatedElementError struct {
	Constraint string
	Err        error
}

// Error :
// Implementation of the `error` interface.
func (e DuplicatedElementError) Error() string {
	return fmt.Sprintf("Query violates unique constraint \"%s\"", e.Constraint)
}

// ForeignKeyViolationError :
// Used to define a foreign key violation in a table
// which leads to inconsistent data.
//
// The `Table` defines the name of the table attached
// to the error.
//
// The `ForeignKey` defines the key that was actually
// violated by the request.
//
// The `Err` defines the initial error that cause the
// foreign key violation to be raised.
type ForeignKeyViolationError struct {
	Table      string
	ForeignKey string
	Err        error
}

// Error :
// Implementation of the `error` interface.
func (e ForeignKeyViolationError) Error() string {
	return fmt.Sprintf("Query violates foreign key \"%s\" existence on table \"%s\"", e.ForeignKey, e.Table)
}

// newNonNullConstraintError :
// Used to perform the creation of an error describing
// a non null constraint being violated. It will use
// the input error to extract information about the
// actual column that was violated.
//
// The `err` defines the error from which this error
// is to be built.
//
// Returns the created error.
func newNonNullConstraintError(err error) error {
	// The error message in case of a duplicated element
	// looks something like below.
	msg := err.Error()

	cue := "null value in column \""

	id := strings.Index(msg, cue)
	if id < 0 {
		return err
	}

	end := msg[id+len(cue):]

	id = strings.Index(end, "\"")
	if id < 0 {
		return err
	}

	nnce := Error{
		SQLCode: nonNullConstraint,
		Err: NonNullConstrainteError{
			Column: end[:id],
			Err:    err,
		},
	}

	return nnce
}

// newDuplicatedElementError :
// Used to perform the creation of a duplicated element
// error from the input error. It will analyze the msg
// associated to the error and extract the constraint
// that was violated.
//
// The `err` defines the error from which this error is
// to be built.
//
// Returns the created error.
func newDuplicatedElementError(err error) error {
	// The error message in case of a duplicated element
	// looks something like below.
	msg := err.Error()

	cue := "duplicate key value violates unique constraint \""

	id := strings.Index(msg, cue)
	if id < 0 {
		return err
	}

	end := msg[id+len(cue):]

	id = strings.Index(end, "\"")
	if id < 0 {
		return err
	}

	dee := Error{
		SQLCode: duplicatedElement,
		Err: DuplicatedElementError{
			Constraint: end[:id],
			Err:        err,
		},
	}

	return dee
}

// newForeignKeyViolation :
// Used to perform the creation of a foreign key issue
// error from the input data. It will analyze the msg
// associated to the error and extract the foreign key
// that was violated.
//
// The `err` defines the error from which this error
// is to be built.
//
// Returns the created error.
func newForeignKeyViolation(err error) error {
	// The error message in case of a duplicated element
	// looks something like below.
	msg := err.Error()

	// First fetch the name of the table.
	cue := "insert or update on table \""

	id := strings.Index(msg, cue)
	if id < 0 {
		return err
	}

	end := msg[id+len(cue):]

	id = strings.Index(end, "\"")
	if id < 0 {
		return err
	}

	table := end[:id]

	// Then analyze the foreign key that was violated.
	cue = "violates foreign key constraint \""

	id = strings.Index(end, cue)
	if id < 0 {
		return err
	}

	constraint := end[id+len(cue):]

	// The error message is structured like so:
	// `tableName_field_fkey`.
	// We want to extract the `field` to provide
	// the best possible info.
	if len(constraint) <= len(table) {
		return err
	}

	constraint = constraint[len(table)+1:]

	id = strings.Index(constraint, "_fkey")
	if id < 0 {
		return err
	}

	// Build and return the error.
	fkve := Error{
		SQLCode: foreignKeyViolation,
		Err: ForeignKeyViolationError{
			Table:      table,
			ForeignKey: constraint[:id],
			Err:        err,
		},
	}

	return fkve
}

// parseSQLCode :
// Used to parse the SQL code defined in an error message
// assuming it looks something like the following:
// `error msg (SQLSTATE : CODE)`.
// In case it cannot parse the corresponding code an error
// is returned.
func parseSQLCode(msg string) (int, error) {
	sqlCue := "SQLSTATE "

	// Analyze the input error to retrieve at least the
	// SQL error code.
	codeIndex := strings.Index(msg, sqlCue)
	if codeIndex < 0 {
		return 0, ErrNoSQLCode
	}

	end := msg[codeIndex+len(sqlCue):]

	id := strings.Index(end, ")")
	if id < 0 {
		return 0, ErrNoSQLCode
	}

	codeStr := end[:id]

	code, err := strconv.ParseInt(codeStr, 10, 32)
	if err != nil {
		return 0, ErrNoSQLCode
	}

	return int(code), nil
}

// formatDBError :
// Used to extract some information about the DB error
// provided in input. It will typically define whether
// the code refer to a foreign key violation, a `null`
// value where it should not be, etc.
//
// The `dbErr` defines the DB error to analyze.
//
// Returns the formatted DB error (in case all else
// fails, the initial error is returned).
func formatDBError(err error) error {
	// In case no error occurred, do nothing.
	if err == nil {
		return err
	}

	// Retrieve the SQL code for this request. In case
	// we can't find a valid code we will return the
	// input error not to create more errors.
	code, pErr := parseSQLCode(err.Error())
	if pErr != nil {
		return err
	}

	// Otherwise we can start building the error.
	var e error

	switch code {
	case nonNullConstraint:
		e = newNonNullConstraintError(err)
	case foreignKeyViolation:
		e = newForeignKeyViolation(err)
	case duplicatedElement:
		e = newDuplicatedElementError(err)
	default:
		// Base error will do.
		e = Error{
			SQLCode: code,
			Err:     err,
		}
	}

	return e
}
