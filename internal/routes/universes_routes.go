package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/handlers"
	"oglike_server/pkg/logger"
)

// universeAdapter :
// Implements the interface requested by the general handler in
// the `handlers` package. The main functions are describing the
// interface to retrieve information about the universes from a
// database.
//
// The `proxy` defines the proxy to use to interact with the DB
// when fetching the data.
type universeAdapter struct {
	proxy data.UniverseProxy
}

// Route :
// Implementation of the method to get the route name to access
// the universes' data.
// Returns the name of the route.
func (ua *universeAdapter) Route() string {
	return "universes"
}

// ParseFilters :
// Implementation of the method to get the filters from the route
// variables. This will allow to fetch universes through a precise
// identifier.
//
// The `vars` are the variables (both route element and the query
// parameters) retrieved from the input request.
//
// Returns the list of filters extracted from the input info.
func (ua *universeAdapter) ParseFilters(vars handlers.RouteVars) []handlers.Filter {
	filters := make([]handlers.Filter, 0)

	// Traverse the input parameters and select only the ones relevant
	// to universes querying.
	allowed := map[string]string{
		"universe_id":   "id",
		"universe_name": "name",
	}

	for key, values := range vars.Params {
		// Check whether this key corresponds to an universe filter. If
		// this is the case we can register it.
		filterName, ok := allowed[key]

		if ok && len(values) > 0 {
			// None of the filters associated to the universe are numeric
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
	// provide a filter on the universe's identifier. More precisely
	// the route can define something like:
	// `/universes/uni-id` which we will interpret as a filter on
	// the universe's identifier.
	// Note that we assume that if the route contains more than `1`
	// element it *always* contains an identifier as second token.
	if len(vars.RouteElems) > 0 {
		uni := vars.RouteElems[0]

		// Append the identifier filter to the existing list.
		found := false
		for id := range filters {
			if filters[id].Key == "id" {
				found = true
				filters[id].Options = append(filters[id].Options, uni)
			}
		}

		if !found {
			filters = append(
				filters,
				handlers.Filter{
					Key:     "id",
					Options: []string{uni},
				},
			)
		}
	}

	return filters
}

// Data :
// Implementation of the method to get the data related to universes
// from the internal DB. We will use the internal DB proxy to get the
// info while still applying the filters.
//
// The `filters` represent the filters extracted from the route and
// as provided by the `ParseFilters` method. We need to convert it
// into a semantic that can be interpreted by the DB.
//
// Returns the data related to the universes along with any errors.
func (ua *universeAdapter) Data(filters []handlers.Filter) (interface{}, error) {
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
	return ua.proxy.Universes(dbFilters)
}

// universeCreator :
// Implements the interface requested by the creation handler in
// the `handlers` package. The main functions are describing the
// interface to perform the creation of a new universe into the
// DB.
//
// The `proxy` defines the proxy to use to interact with the DB
// when creating data.
//
// The `log` allows to notify problems and information during a
// universe's creation.
type universeCreator struct {
	proxy data.UniverseProxy
	log   logger.Logger
}

// Route :
// Implementation of the method to get the route name to create some
// new universes.
// Returns the name of the route.
func (uc *universeCreator) Route() string {
	return "universe"
}

// AccessRoute :
// Implementation of the method to get the route name to access to
// the data created by this handler. This is basically the `players`
// route.
func (uc *universeCreator) AccessRoute() string {
	return "universes"
}

// DataKey :
// Implementation of the method to get the name of the key used to
// pass data to the server.
// Returns the name of the key.
func (uc *universeCreator) DataKey() string {
	return "universe-data"
}

// Create :
// Implementation of the method to perform the creation of the data
// related to the new universes. We will use the internal proxy to
// request the DB to create a new universe.
//
// The `input` represent the data fetched from the input request and
// should contain the properties of the universes to create.
//
// Return the targets of the created resources along with any error.
func (uc *universeCreator) Create(input handlers.RouteData) ([]string, error) {
	// We need to iterate over the data retrieved from the route and
	// create universes from it.
	var uni data.Universe
	resources := make([]string, 0)

	// Prevent request with no data.
	if len(input.Data) == 0 {
		return resources, fmt.Errorf("Could not perform creation of universe with no data")
	}

	for _, rawData := range input.Data {
		// Try to unmarshal the data into a valid `Universe` struct.
		err := json.Unmarshal([]byte(rawData), &uni)
		if err != nil {
			uc.log.Trace(logger.Error, fmt.Sprintf("Could not create universe from data \"%s\" (err: %v)", rawData, err))
			continue
		}

		// Create the universe.
		err = uc.proxy.Create(&uni)
		if err != nil {
			uc.log.Trace(logger.Error, fmt.Sprintf("Could not register universe from data \"%s\" (err: %v)", rawData, err))
			continue
		}

		// Successfully created a universe.
		uc.log.Trace(logger.Notice, fmt.Sprintf("Created new universe \"%s\" with id \"%s\"", uni.Name, uni.ID))
		resources = append(resources, uni.ID)
	}

	// Return the path to the resources created during the process.
	return resources, nil
}

// listUniverses :
// Creates a handler allowing to serve requests on the universes
// by interrogating the main DB. We uses the handler structure in
// the `handlers` package and provide the needed endpoint desc as
// requested.
//
// Returns the handler that can be executed to serve such requests.
func (s *server) listUniverses() http.HandlerFunc {
	return handlers.ServeRoute(
		&universeAdapter{
			s.universes,
		},
		s.log,
	)
}

// createUniverse :
// Creates a handler allowing to server requests to create new
// universes in the main DB. This rely on the handler structure
// provided by the `handlers` package which allows to mutualize
// the extraction of the data from the input request and the
// general flow to perform the creation.
//
// Returns the handler which can be executed to perform such
// requests.
func (s *server) createUniverse() http.HandlerFunc {
	return handlers.ServeCreationRoute(
		&universeCreator{
			s.universes,
			s.log,
		},
		s.log,
	)
}
