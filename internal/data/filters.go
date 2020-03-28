package data

import "fmt"

// Filter :
// Generic filter that can be used to restrain the number
// of results returned by a query. This allows to narrow
// a search and keep only relevant information.
// A filter is combined into a SQL auery through a syntax
// that uses both the `Key` and `Value` in the following
// ways:
// `Key = 'Value'`
// `Key = Value`
//
// Choosing between one or the other depens on whether the
// filter is numeric or not.
//
// The `Key` describes the value that should be filtered.
// It usually corresponds to the name of a column in the
// database.
//
// The `Value` represents the specific instance of the key
// that should be kept. Anything that is not this value
// will be ignored.
//
// The `Numeric` boolean allows to determine whether the
// filter is applied on a numeric column or not. It will
// change slightly the syntax used in the SQL query.
type Filter struct {
	Key     string
	Value   string
	Numeric bool
}

// String :
// Implementation of the `Stringer` interface for a filter.
// It allows to automatically transform it into a value to
// use in a SQL query.
//
// Returns the equivalent string for this filter.
func (f Filter) String() string {
	if f.Numeric {
		return fmt.Sprintf("%s=%s", f.Key, f.Value)
	}

	return fmt.Sprintf("%s='%s'", f.Key, f.Value)
}
