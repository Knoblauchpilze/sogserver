package routes

import "oglike_server/internal/data"

// getTechnologiesFilters :
// Used to retrieve a table defining for each property defining
// a technology the name of the associated key to provide as
// query parameter to set a filter on it.
// Typically a `key` from the return value will be the name of
// a query parameter while its value will be the name of the
// corresponding filter in the internal database.
// The returned mapping applies to convert query parameters to
// filtering for planets.
//
// Returns a mapping between input query parameters key and
// internal database columns to use to filter on said keys.
func getTechnologiesFilters() map[string]string {
	return map[string]string{
		"technology_name": "name",
		"technology_id":   "id",
	}
}

// parseTechnologiesFilters :
// Used to extract the relevant filters to apply on technologies
// from the input route variables. The return value is an array
// of the filters that can be used in the SQL query.
// We use common values to identify prefix for filters.
//
// The `vars` represents the variables provided through the route
// with which the user contacted the server. Only some parameters
// will actually represent the filters to apply on the request.
//
// Returns the array of filters that were relevant with a techno
// request. Might be empty if no such filters are defined.
func parseTechnologiesFilters(vars routeVars) []data.Filter {
	filters := make([]data.Filter, 0)

	// Traverse the input parameters and select only the ones relevant
	// to technologies querying.
	allowed := getTechnologiesFilters()

	for key, values := range vars.params {
		// Check whether this key corresponds to a technology filter. If
		// this is the case we can register it.
		filterName, ok := allowed[key]

		if ok {
			// All the filters associated to the technologies are not
			// numeric for now. If it's not the case later on it would
			// have to be modified.
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
