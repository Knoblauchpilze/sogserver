package routes

import (
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/handlers"
)

// buildingAdapter :
// Implements the interface requested by the general handler in
// the `handlers` package. The main functions are describing the
// interface to retrieve information about the buildings from a
// database.
//
// The `proxy` defines the proxy to use to interact with the DB
// when fetching the data.
type buildingAdapter struct {
	proxy data.BuildingProxy
}

// Route :
// Implementation of the method to get the route name to access
// the buildings' data.
// Returns the name of the route.
func (ba *buildingAdapter) Route() string {
	return "buildings"
}

// ParseFilters :
// Implementation of the method to get the filters from the route
// variables. This will allow to fetch buildings through a precise
// identifier.
//
// The `vars` are the variables (both route element and the query
// parameters) retrieved from the input request.
//
// Returns the list of filters extracted from the input info.
func (ba *buildingAdapter) ParseFilters(vars handlers.RouteVars) []handlers.Filter {
	filters := make([]handlers.Filter, 0)

	// Traverse the input parameters and select only the ones relevant
	// to buildings querying.
	allowed := map[string]string{
		"building_id":   ba.proxy.GetIdentifierDBColumnName(),
		"building_name": "name",
	}

	for key, values := range vars.Params {
		// Check whether this key corresponds to an building filter. If
		// this is the case we can register it.
		filterName, ok := allowed[key]

		if ok && len(values) > 0 {
			// None of the filters associated to the building are numeric
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
	// provide a filter on the building's identifier. More precisely
	// the route can define something like:
	// `/buildings/building-id` which we will interpret as a filter
	// on the building's identifier.
	// Note that we assume that if the route contains more than `1`
	// element it *always* contains an identifier as second token.
	if len(vars.RouteElems) > 0 {
		building := vars.RouteElems[0]

		// Append the identifier filter to the existing list.
		found := false
		for id := range filters {
			if filters[id].Key == ba.proxy.GetIdentifierDBColumnName() {
				found = true
				filters[id].Options = append(filters[id].Options, building)
			}
		}

		if !found {
			filters = append(
				filters,
				handlers.Filter{
					Key:     ba.proxy.GetIdentifierDBColumnName(),
					Options: []string{building},
				},
			)
		}
	}

	return filters
}

// Data :
// Implementation of the method to get the data related to buildings
// from the internal DB. We will use the internal DB proxy to get the
// info while still applying the filters.
//
// The `filters` represent the filters extracted from the route and
// as provided by the `ParseFilters` method. We need to convert it
// into a semantic that can be interpreted by the DB.
//
// Returns the data related to the buildings along with any errors.
func (ba *buildingAdapter) Data(filters []handlers.Filter) (interface{}, error) {
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
	return ba.proxy.Buildings(dbFilters)
}

// listBuildings :
// Creates a handler allowing to serve requests on the buildings
// by interrogating the main DB. We uses the handler structure in
// the `handlers` package and provide the needed endpoint desc as
// requested.
//
// Returns the handler that can be executed to serve such requests.
func (s *server) listBuildings() http.HandlerFunc {
	return handlers.ServeRoute(
		&buildingAdapter{
			s.buildings,
		},
		s.log,
	)
}