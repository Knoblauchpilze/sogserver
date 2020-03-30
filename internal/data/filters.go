package data

import (
	"fmt"
	"strings"
)

// DBFilter :
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
type DBFilter struct {
	Key    string
	Values []string
}

// String :
// Implementation of the `Stringer` interface for a filter.
// It allows to automatically transform it into a value to
// use in a SQL query.
//
// Returns the equivalent string for this filter.
func (f DBFilter) String() string {
	// We need to quote the values first and then join them.
	quoted := make([]string, len(f.Values))
	for id, str := range f.Values {
		quoted[id] = fmt.Sprintf("'%s'", str)
	}

	return fmt.Sprintf("%s in (%s)", f.Key, strings.Join(quoted, ","))
}
