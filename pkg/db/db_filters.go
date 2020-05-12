package db

import (
	"fmt"
	"strings"
	"time"
)

// Filter :
// Generic filter that can be used to restrain the number
// of results returned by a query. This allows to narrow
// a search and keep only relevant information.
// A filter is combined into a SQL query through a syntax
// that uses both the `Key` and a set of `Values` in the
// following way:
// `Key in ('Values[0]', 'Values[1]', ...)`
// Note that if the `Values` array contains several values
// they should be combined in a OR fashion (so the filter
// will match if the `Key` is any of the specified values).
//
// The `Key` describes the value that should be filtered.
// It usually corresponds to the name of a column in the
// database.
//
// The `Values` represents the specific instances of the
// key that should be kept. Anything that is not part of
// the list of value will be ignored.
type Filter struct {
	Key      string
	Values   []interface{}
	Operator Operation
}

// Operation :
// Defines the possible operations to use to combine
// the values available in a filter. Depending on the
// operation value the key will be compared differently
// to the list of values associated to the filter.
type Operation int

// List of possible operators for a filter.
const (
	In Operation = iota
	LessThan
	GreaterThan
)

// String :
// Implementation of the `Stringer` interface for a filter.
// It allows to automatically transform it into a value to
// use in a SQL query.
//
// Returns the equivalent string for this filter.
func (f Filter) String() string {
	// Depending on the operation associated to the filter
	// we will interpret differently the values.
	switch f.Operator {
	case LessThan:
		return f.stringifyLessThan()
	case GreaterThan:
		return f.stringifyGreaterThan()
	case In:
		fallthrough
	default: // Assume `In` semantic.
		return f.stringifyBelong()
	}
}

// stringifyBelong :
// Used to stringify the `Key` and `Values` associated to
// this filter with a `Belongs to` semantic.
//
// Returns the corresponding string.
func (f Filter) stringifyBelong() string {
	// We need to quote the values first and then join them.
	quoted := make([]string, len(f.Values))
	for id, v := range f.Values {
		// In case the filter is a `time.Time` we will use
		// the `RFC3339` syntax. This topic helped to solve
		// the issue:
		// https://stackoverflow.com/questions/37782278/fully-parsing-timestamps-in-golang
		t, ok := v.(time.Time)
		if ok {
			quoted[id] = fmt.Sprintf("'%v'", t.Format(time.RFC3339))
			continue
		}

		quoted[id] = fmt.Sprintf("'%v'", v)
	}

	return fmt.Sprintf("%s in (%s)", f.Key, strings.Join(quoted, ","))
}

// stringifyLessThan :
// Used to stringify the `Key` and `Values` associated to
// this filter with a `Less than` semantic.
//
// Returns the corresponding string.
func (f Filter) stringifyLessThan() string {
	return f.stringifyOperator("<")
}

// stringifyGreaterThan :
// Used to stringify the `Key` and `Values` associated to
// this filter with a `Greater than` semantic.
//
// Returns the corresponding string.
func (f Filter) stringifyGreaterThan() string {
	return f.stringifyOperator(">")
}

// stringifyOperator :
// Used as a generic operation grouping the values of
// this filter with a `and` semantic and comparing the
// `Key` with the provided operator to the `Values`.
//
// The `op` defines the string to compare the `Key`
// with each element of the `Values`.
//
// Returns the produced string.
func (f Filter) stringifyOperator(op string) string {
	// Traverse the list of values and append each one
	// to build the filter's representation.
	out := ""

	for id, filter := range f.Values {
		if id > 0 {
			out += " and "
		}

		// Apply a similar processing to the `time` values
		// as in the case of the `stringifyBelong` method.
		t, ok := filter.(time.Time)
		if ok {
			out += fmt.Sprintf("%s %s '%v'", f.Key, op, t.Format(time.RFC3339))

			continue
		}

		out += fmt.Sprintf("%s %s '%v'", f.Key, op, filter)
	}

	return out
}
