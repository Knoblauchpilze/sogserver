package routes

import (
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/handlers"
)

// defenseAdapter :
// Implements the interface requested by the general handler in
// the `handlers` package. The main functions are describing the
// interface to retrieve information about the defenses from a
// database.
//
// The `proxy` defines the proxy to use to interact with the DB
// when fetching the data.
type defenseAdapter struct {
	proxy data.DefenseProxy
}

// Route :
// Implementation of the method to get the route name to access
// the defenses' data.
// Returns the name of the route.
func (da *defenseAdapter) Route() string {
	return "defenses"
}

// ParseFilters :
// Implementation of the method to get the filters from the route
// variables. This will allow to fetch defenses through a precise
// identifier.
//
// The `vars` are the variables (both route element and the query
// parameters) retrieved from the input request.
//
// Returns the list of filters extracted from the input info.
func (da *defenseAdapter) ParseFilters(vars handlers.RouteVars) []handlers.Filter {
	filters := make([]handlers.Filter, 0)

	// Traverse the input parameters and select only the ones relevant
	// to defenses querying.
	allowed := map[string]string{
		"defense_id":   "id",
		"defense_name": "name",
	}

	for key, values := range vars.Params {
		// Check whether this key corresponds to an defense filter. If
		// this is the case we can register it.
		filterName, ok := allowed[key]

		if ok && len(values) > 0 {
			// None of the filters associated to the defense are numeric
			// for now. If it's the case later on it would have to be
			// modified.
			filter := handlers.Filter{
				Key:     filterName,
				Options: values,
			}

			filters = append(filters, filter)
		}
	}

	// We also need to fetch parts of the route that can be used to
	// provide a filter on the defense's identifier. More precisely
	// the route can define something like:
	// `/defenses/defense-id` which we will interpret as a filter on
	// the defense's identifier.
	// Note that we assume that if the route contains more than `1`
	// element it *always* contains an identifier as second token.
	if len(vars.RouteElems) > 0 {
		def := vars.RouteElems[0]

		// Append the identifier filter to the existing list.
		found := false
		for id := range filters {
			if filters[id].Key == "id" {
				found = true
				filters[id].Options = append(filters[id].Options, def)
			}
		}

		if !found {
			filters = append(
				filters,
				handlers.Filter{
					Key:     "id",
					Options: []string{def},
				},
			)
		}
	}

	return filters
}

// Data :
// Implementation of the method to get the data related to defenses
// from the internal DB. We will use the internal DB proxy to get the
// info while still applying the filters.
//
// The `filters` represent the filters extracted from the route and
// as provided by the `ParseFilters` method. We need to convert it
// into a semantic that can be interpreted by the DB.
//
// Returns the data related to the defenses along with any errors.
func (da *defenseAdapter) Data(filters []handlers.Filter) (interface{}, error) {
	// Convert the input request filters into DB filters.
	dbFilters := make([]data.DBFilter, 0)
	for _, filter := range filters {
		dbFilters = append(
			dbFilters,
			data.DBFilter{
				Key:    filter.Key,
				Values: filter.Options,
			},
		)
	}

	// Use the DB proxy to fetch the data.
	return da.proxy.Defenses(dbFilters)
}

// listDefenses :
// Creates a handler allowing to serve requests on the defenses
// by interrogating the main DB. We uses the handler structure in
// the `handlers` package and provide the needed endpoint desc as
// requested.
//
// Returns the handler that can be executed to serve such requests.
func (s *server) listDefenses() http.HandlerFunc {
	return handlers.ServeRoute(
		&defenseAdapter{
			s.defenses,
		},
		s.log,
	)
}
