package routes

import (
	"encoding/json"
	"net/http"
	"oglike_server/internal/game"
	"oglike_server/pkg/db"
)

// listUniverses :
// Used to perform the creation of a handler allowing to serve
// the requests on universes.
//
// Returns the handler that can be executed to serve said reqs.
func (s *Server) listUniverses() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("universes")

	allowed := map[string]string{
		"id":   "id",
		"name": "name",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id").WithModule("universes").WithLocker(s.og)
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.universes.Universes(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// listUniverseRankings :
// Used to perform the creation of a handler allowing to server
// the requests related to rankings in universes.
//
// Returns the handler that can be executed to server said reqs.
func (s *Server) listUniverseRankings() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("universes")

	allowed := map[string]string{
		"id": "p.universe",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithIDFilter("p.universe").WithResourceFilter("p.id").WithModule("universes").WithLocker(s.og)
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.universes.Rankings(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// createUniverse :
// Used to perform the creation of a handler allowing to serve
// the requests to create universes.
//
// Returns the handler to execute to perform said requests.
func (s *Server) createUniverse() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewCreateResourceEndpoint("universes")

	// Configure the endpoint.
	ed.WithDataKey("universe-data").WithModule("universes").WithLocker(s.og)
	ed.WithCreationFunc(
		func(input RouteData) ([]string, error) {
			// We need to iterate over the data retrieved from the route and
			// create universes from it.
			var uni game.Universe
			resources := make([]string, 0)

			// Prevent request with no data.
			if len(input.Data) == 0 {
				return resources, ErrNoData
			}

			for _, rawData := range input.Data {
				// Try to unmarshal the data into a valid `Universe` struct.
				err := json.Unmarshal([]byte(rawData), &uni)
				if err != nil {
					return resources, ErrInvalidData
				}

				// Create the universe.
				res, err := s.universes.Create(uni)
				if err != nil {
					return resources, err
				}

				// Successfully created a universe.
				resources = append(resources, res)
			}

			// Return the path to the resources created during the process.
			return resources, nil
		},
	)

	return ed.ServeRoute(s.log)
}
