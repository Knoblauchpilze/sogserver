package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/logger"
)

// listUniverses :
// Used to perform the creation of a handler allowing to serve
// the requests on universes.
//
// Returns the handler that can be executed to serve said reqs.
func (s *server) listUniverses() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("universes")

	allowed := map[string]string{
		"universe_id":   "id",
		"universe_name": "name",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id")
	ed.WithDataFunc(
		func(filters []data.DBFilter) (interface{}, error) {
			return s.universes.Universes(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// createUniverse :
// Used to perform the creation of a handler allowing to server
// the requests to create universes.
//
// Returns the handler to execute to perform said requests.
func (s *server) createUniverse() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewCreateResourceEndpoint("universes")

	// Configure the endpoint.
	ed.WithDataKey("universe-data")
	ed.WithCreationFunc(
		func(input RouteData) ([]string, error) {
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
					s.log.Trace(logger.Error, fmt.Sprintf("Could not create universe from data \"%s\" (err: %v)", rawData, err))
					continue
				}

				// Create the universe.
				err = s.universes.Create(&uni)
				if err != nil {
					s.log.Trace(logger.Error, fmt.Sprintf("Could not register universe from data \"%s\" (err: %v)", rawData, err))
					continue
				}

				// Successfully created a universe.
				s.log.Trace(logger.Notice, fmt.Sprintf("Created new universe \"%s\" with id \"%s\"", uni.Name, uni.ID))
				resources = append(resources, uni.ID)
			}

			// Return the path to the resources created during the process.
			return resources, nil
		},
	)

	return ed.ServeRoute(s.log)
}
