package routes

import "oglike_server/internal/data"

// getUniverseIdentifierDBColumnName :
// Used to retrieve the string literal defining the name of the
// identifier column in the `universes` table in the database.
//
// Returns the name of the `identifier` column in the database.
func getUniverseIdentifierDBColumnName() string {
	return "id"
}

// getUniversesFilters :
// Used to retrieve a table defining for each universe property
// the name of the associated key to provide as query parameter
// to set it.
// Typically a `key` from the return value will be the name of
// a query parameter while its value will be the name of the
// corresponding filter in the internal database.
// The returned mapping applies to convert query parameters to
// filtering for universes.
//
// Returns a mapping between input query parameters key and
// internal database columns to use to filter on said keys.
func getUniversesFilters() map[string]string {
	return map[string]string{
		"universe_id":   getUniverseIdentifierDBColumnName(),
		"universe_name": "name",
	}
}

// parseUniverseFilters :
// Used to extract the relevant filters to apply on universes from
// the input route variables. The return value is an array of said
// filters that can be used in the SQL query.
// We use common values to identify prefix for filters.
//
// The `vars` represents the variables provided through the route
// with which the user contacted the server. Only some parameters
// will actually represent the filters to apply on the request.
//
// Returns the array of filters that were relevant with a universe
// request. Might be empty if no such filters are defined.
func parseUniverseFilters(vars routeVars) []data.Filter {
	filters := make([]data.Filter, 0)

	// Traverse the input parameters and select only the ones relevant
	// to universes querying.
	allowed := getUniversesFilters()

	for key, values := range vars.params {
		// Check whether this key corresponds to an universe filter. If
		// this is the case we can register it.
		filterName, ok := allowed[key]

		if ok {
			// None of the filters associated to the universe are numeric
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

// generateUniverseFilterFromID :
// Used to create a valid filter from the universe's identifier in
// argument. Allows to keep the definition of the string matching
// the universe's id in the DB internal to this file.
//
// The `uni` represents the identifier of the universe for which a
// filter should be generated.
//
// Returns the generated array of filter (for convenience purposes).
func generateUniverseFilterFromID(uni string) data.Filter {
	return data.Filter{
		Key:     getUniverseIdentifierDBColumnName(),
		Values:  []string{uni},
		Numeric: false,
	}
}
