package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/handlers"
	"oglike_server/pkg/logger"
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
		"fleet_id":     "id",
		"fleet_name":   "name",
		"galaxy":       "target_galaxy",
		"solar_system": "target_solar_system",
		"position":     "target_position",
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

	// Finally we should determine whether one is requesting to
	// access to the components of a single fleet. In this case
	// the input route will contain a `components` part after
	// the identifier of the fleet.
	// If this is the case we will append a last filter to tell
	// that the components of the fleet described by the filters
	// should be fetched. This will be recognized during the
	// fetching phase provided by this adapter.
	if len(vars.RouteElems) == 2 && vars.RouteElems[1] == "components" {
		filters = append(
			filters,
			handlers.Filter{
				Key:     "components",
				Options: []string{""},
			},
		)
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
	// Convert the input request filters into DB filters. We only do
	// that for filters that are distinct from the `components` that
	// indicates a request to fetch components of a fleet.
	dbFilters := make([]data.DBFilter, 0)
	for _, filter := range filters {
		if filter.Key != "components" {
			dbFilters = append(
				dbFilters,
				data.DBFilter{
					Key:    filter.Key,
					Values: filter.Options,
				},
			)
		}
	}

	// Check whether we're requesting the components of a fleet
	// or a list of fleets. This is indicated by the fact that
	// the last filter has a key of `components`.
	if len(filters) > 0 && filters[len(filters)-1].Key == "components" {
		return fa.proxy.FleetComponents(dbFilters)
	}

	// Otherwise we're requesting the fleets and not its comps.
	return fa.proxy.Fleets(dbFilters)
}

// fleetCreator :
// Implements the interface requested by the creation handler in
// the `handlers` package. The main functions are describing the
// interface to create a new fleet from the data fetched from the
// input request.
//
// The `proxy` defines the proxy to use to interact with the DB
// when creating the data.
//
// The `log` allows to notify problems and information during a
// fleet's creation.
type fleetCreator struct {
	proxy data.FleetProxy
	log   logger.Logger
}

// Route :
// Implementation of the method to get the route name to create some
// new fleets.
// Returns the name of the route.
func (fc *fleetCreator) Route() string {
	return "fleet"
}

// AccessRoute :
// Implementation of the method to get the route name to access to
// the data created by this handler. This is basically the `fleet`
// route.
func (fc *fleetCreator) AccessRoute() string {
	return "fleets"
}

// DataKey :
// Implementation of the method to get the name of the key used to
// pass data to the server.
// Returns the name of the key.
func (fc *fleetCreator) DataKey() string {
	return "fleet-data"
}

// Create :
// Implementation of the method to perform the creation of the data
// related to the new fleets. We will use the internal proxy to
// request the DB to create a new fleet.
//
// The `input` represent the data fetched from the input request and
// should contain the properties of the fleets to create.
//
// Return the targets of the created resources along with any error.
func (fc *fleetCreator) Create(input handlers.RouteData) ([]string, error) {
	// We need to iterate over the data retrieved from the route and
	// create fleets from it.
	var fleet data.Fleet
	resources := make([]string, 0)

	// Prevent request with no data.
	if len(input.Data) == 0 {
		return resources, fmt.Errorf("Could not perform creation of fleet with no data")
	}

	for _, rawData := range input.Data {
		// Try to unmarshal the data into a valid `Fleet` struct.
		err := json.Unmarshal([]byte(rawData), &fleet)
		if err != nil {
			fc.log.Trace(logger.Error, fmt.Sprintf("Could not create fleet from data \"%s\" (err: %v)", rawData, err))
			continue
		}

		// Create the fleet.
		err = fc.proxy.Create(&fleet)
		if err != nil {
			fc.log.Trace(logger.Error, fmt.Sprintf("Could not register fleet from data \"%s\" (err: %v)", rawData, err))
			continue
		}

		// Successfully created an fleet.
		fc.log.Trace(logger.Notice, fmt.Sprintf("Created new fleet \"%s\" with name \"%s\"", fleet.ID, fleet.Name))
		resources = append(resources, fleet.ID)
	}

	// Return the path to the resources created during the process.
	return resources, nil
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

// createFleet :
// Creates a handler allowing to server requests to create new
// fleets in the main DB. This rely on the handler structure
// provided by the `handlers` package which allows to mutualize
// the extraction of the data from the input request and the
// general flow to perform the creation.
//
// Returns the handler which can be executed to perform such
// requests.
func (s *server) createFleet() http.HandlerFunc {
	return handlers.ServeCreationRoute(
		&fleetCreator{
			s.fleets,
			s.log,
		},
		s.log,
	)
}
