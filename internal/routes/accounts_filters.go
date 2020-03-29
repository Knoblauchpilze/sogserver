package routes

import "oglike_server/internal/data"

// getAccountsFilters :
// Used to retrieve a table defining for each account property
// the name of the associated key to provide as query parameter
// to set it.
// Typically a `key` from the return value will be the name of
// a query parameter while its value will be the name of the
// corresponding filter in the internal database.
// The returned mapping applies to convert query parameters to
// filtering for accounts.
//
// Returns a mapping between input query parameters key and
// internal database columns to use to filter on said keys.
func getAccountsFilters() map[string]string {
	return map[string]string{
		"account_id":   "id",
		"account_name": "name",
		"account_mail": "mail",
	}
}

// parseAccountFilters :
// Used to extract the relevant filters to apply on accounts from
// the input route variables. The return value is an array of said
// filters that can be used in the SQL query.
// We use common values to identify prefix for filters.
//
// The `vars` represents the variables provided through the route
// with which the user contacted the server. Only some parameters
// will actually represent the filters to apply on the request.
//
// Returns the array of filters that were relevant with an account
// request. Might be empty if no such filters are defined.
func parseAccountFilters(vars routeVars) []data.Filter {
	filters := make([]data.Filter, 0)

	// Traverse the input parameters and select only the ones relevant
	// to accounts querying.
	allowed := getAccountsFilters()

	for key, values := range vars.params {
		// Check whether this key corresponds to an account filter. If
		// this is the case we can register it.
		filterName, ok := allowed[key]

		if ok {
			// None of the filters associated to the account are numeric
			// for now. If it's the case later on it would have to be
			// modified.
			filter := data.Filter{
				Key:     filterName,
				Values:  values,
				Numeric: false,
			}

			filters = append(filters, filter)
		}
	}

	return filters
}
