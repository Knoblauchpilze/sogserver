package db

import (
	"fmt"
	"strings"
)

// ErrInvalidQuery :
// Used in case the query to perform in the DB is either
// `nil` or empty.
var ErrInvalidQuery = fmt.Errorf("Invalid nil query")

// ErrInvalidDB :
// Provides an abstract way to define that an error
// occurred while accessing some properties on the
// DB. This is especially useful to define whether
// an error returned by the `Init` method of some
// `DBModule` originates in a failure of the DB.
var ErrInvalidDB = fmt.Errorf("Invalid nil DB")

// ErrInvalidData :
// Used to indicate that a marshalling error has
// been detected when trying to import some item
// in the DB.
var ErrInvalidData = fmt.Errorf("Invalid data to insert to DB")

// ErrInvalidScan :
// Used to indicate that a `Scan` operation on a
// `QueryResult` has failec.
var ErrInvalidScan = fmt.Errorf("Invalid scan operation on DB")

// ErrorType :
// Defines some convenience named values for common SQL
// errors.
type ErrorType int

// Defines the possible named SQL errors.
const (
	DuplicatedElement ErrorType = iota
	ForeignKeyViolation
	Unknown
)

// getDuplicatedElementErrorKey :
// Used to retrieve a string describing part of the error
// message issued by the database when trying to insert a
// duplicated element on a unique column. Can be used to
// standardize the definition of this error.
//
// Return part of the error string issued when inserting
// an already existing key.
func getDuplicatedElementErrorKey() string {
	return "SQLSTATE 23505"
}

// getForeignKeyViolationErrorKey :
// Used to retrieve a string describing part of the error
// message issued by the database when trying to insert an
// element that does not match a foreign key constraint.
// Can be used to standardize the definition of this error.
//
// Return part of the error string issued when violating a
// foreign key constraint.
func getForeignKeyViolationErrorKey() string {
	return "SQLSTATE 23503"
}

// GetSQLErrorCode :
// Performs an analysis of the input error string to extract
// a named error code if possible. In case the error does not
// seem to match anything known, the `Unknown` code is sent
// back.
//
// The `errStr` defines the error message to analyze.
//
// Returns the error code for this error or `Unknown` if it
// does not match any known error.
func GetSQLErrorCode(errStr string) ErrorType {
	// Check for all known keys.
	if strings.Contains(errStr, getDuplicatedElementErrorKey()) {
		return DuplicatedElement
	}

	if strings.Contains(errStr, getForeignKeyViolationErrorKey()) {
		return ForeignKeyViolation
	}

	return Unknown
}
