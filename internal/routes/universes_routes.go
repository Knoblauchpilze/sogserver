package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/handlers"
	"oglike_server/pkg/logger"
)

// listUniverses :
// Used to perform the creation of a handler allowing to serve
// the requests on universes.
//
// Returns the handler that can be executed to serve said reqs.
func (s *server) listUniverses() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewEndpointDesc("universes")

	allowed := map[string]string{
		"universe_id":   "id",
		"universe_name": "name",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithIDFilter("id")
	ed.WithDataFunc(
		func(filters []data.DBFilter) (interface{}, error) {
			return s.universes.Universes(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// universeCreator :
// Implements the interface requested by the creation handler in
// the `handlers` package. The main functions are describing the
// interface to create a new universe from the data fetched from
// the input request.
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
