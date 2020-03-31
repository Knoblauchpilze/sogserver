package routes

import (
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/handlers"
)

// fleetAdapter :
// Implements the interface requested by the general handler in
// the `handlers` package. The main functions are describing the
// interface to retrieve information about the fleets from a DB.
//
// The `proxy` defines the proxy to use to interact with the DB
// when fetching the data.
type fleetAdapter struct {
	proxy data.FleetProxy
}

// Route :
// Implementation of the method to get the route name to access
// the fleets' data.
// Returns the name of the route.
func (fa *fleetAdapter) Route() string {
	return "fleets"
}

// ParseFilters :
// Implementation of the method to get the filters from the route
// variables. This will allow to fetch fleets through a precise
// identifier.
//
// The `vars` are the variables (both route element and the query
// parameters) retrieved from the input request.
//
// Returns the list of filters extracted from the input info.
func (fa *fleetAdapter) ParseFilters(vars handlers.RouteVars) []handlers.Filter {
	filters := make([]handlers.Filter, 0)

	// Traverse the input parameters and select only the ones relevant
	// to fleets querying.
	allowed := map[string]string{
		"fleet_id":   fa.proxy.GetIdentifierDBColumnName(),
		"fleet_name": "name",
		"player_id":  "player",
	}

	for key, values := range vars.Params {
		// Check whether this key corresponds to an fleet filter. If
		// this is the case we can register it.
		filterName, ok := allowed[key]

		if ok && len(values) > 0 {
			// None of the filters associated to the fleet are numeric
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
	// provide a filter on the fleet's identifier. More precisely
	// the route can define something like:
	// `/fleets/fleet-id` which we will interpret as a filter on
	// the fleet's identifier.
	// Note that we assume that if the route contains more than `1`
	// element it *always* contains an identifier as second token.
	if len(vars.RouteElems) > 0 {
		def := vars.RouteElems[0]

		// Append the identifier filter to the existing list.
		found := false
		for id := range filters {
			if filters[id].Key == fa.proxy.GetIdentifierDBColumnName() {
				found = true
				filters[id].Options = append(filters[id].Options, def)
			}
		}

		if !found {
			filters = append(
				filters,
				handlers.Filter{
					Key:     fa.proxy.GetIdentifierDBColumnName(),
					Options: []string{def},
				},
			)
		}
	}

	return filters
}

// Data :
// Implementation of the method to get the data related to fleets
// from the internal DB. We will use the internal DB proxy to get
// the info while still applying the filters.
//
// The `filters` represent the filters extracted from the route and
// as provided by the `ParseFilters` method. We need to convert it
// into a semantic that can be interpreted by the DB.
//
// Returns the data related to the fleets along with any errors.
func (fa *fleetAdapter) Data(filters []handlers.Filter) (interface{}, error) {
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
	return fa.proxy.Fleets(dbFilters)
}

// listFleets :
// Creates a handler allowing to serve requests on the fleets
// by interrogating the main DB. We uses the handler structure
// in the `handlers` package and provide the needed endpoint
// desc as requested.
//
// Returns the handler that can be executed to serve such requests.
func (s *server) listFleets() http.HandlerFunc {
	return handlers.ServeRoute(
		&fleetAdapter{
			s.fleets,
		},
		s.log,
	)
}
