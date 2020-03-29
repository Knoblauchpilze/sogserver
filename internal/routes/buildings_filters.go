package routes

import "oglike_server/internal/data"

// getBuildingsFilters :
// Used to retrieve a table defining for each property defining
// a building the name of the associated key to provide as query
// parameter to set a filter on it.
// Typically a `key` from the return value will be the name of
// a query parameter while its value will be the name of the
// corresponding filter in the internal database.
// The returned mapping applies to convert query parameters to
// filtering for planets.
//
// Returns a mapping between input query parameters key and
// internal database columns to use to filter on said keys.
func getBuildingsFilters() map[string]string {
	return map[string]string{
		"building_name": "name",
		"building_id":   "id",
	}
}

// parseBuildingsFilters :
// Used to extract the relevant filters to apply on buildings from
// the input route variables. The return value is an array of the
// filters that can be used in the SQL query.
// We use common values to identify prefix for filters.
//
// The `vars` represents the variables provided through the route
// with which the user contacted the server. Only some parameters
// will actually represent the filters to apply on the request.
//
// Returns the array of filters that were relevant with a building
// request. Might be empty if no such filters are defined.
func parseBuildingsFilters(vars routeVars) []data.Filter {
	filters := make([]data.Filter, 0)

	// Traverse the input parameters and select only the ones relevant
	// to buildings querying.
	allowed := getBuildingsFilters()

	for key, values := range vars.params {
		// Check whether this key corresponds to a building filter. If
		// this is the case we can register it.
		filterName, ok := allowed[key]

		if ok {
			// All the filters associated to the buildings aren't
			// numeric for now. If it's not the case later on it
			// would have to be modified.
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
