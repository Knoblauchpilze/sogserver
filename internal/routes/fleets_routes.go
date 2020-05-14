package routes

import (
	"encoding/json"
	"net/http"
	"oglike_server/internal/game"
	"oglike_server/pkg/db"
)

// listFleets :
// Used to perform the creation of a handler allowing to serve
// the requests on fleets.
//
// Returns the handler that can be executed to serve said reqs.
func (s *Server) listFleets() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("fleets")

	allowed := map[string]string{
		"id":           "id",
		"universe":     "uni",
		"objective":    "objective",
		"source":       "source",
		"target":       "target",
		"galaxy":       "target_galaxy",
		"solar_system": "target_solar_system",
		"position":     "target_position",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id").WithModule("fleets").WithLocker(s.og)
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.fleets.Fleets(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// listFleetObjectives :
// Used to perform the creation of a handler allowing to serve
// the requests on fleet objectives.
//
// Returns the created handler.
func (s *Server) listFleetObjectives() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("fleets/objectives")

	allowed := map[string]string{
		"id":   "id",
		"name": "name",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id").WithModule("fleets")
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.og.Objectives.Objectives(s.proxy, filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// createFleetComponent :
// Used to perform the creation of a handler allowing to serve
// the requests to create fleet components.
//
// Returns the handler to execute to perform said requests.
func (s *Server) createFleet() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewCreateResourceEndpoint("fleets")

	// Configure the endpoint.
	ed.WithDataKey("fleet-data").WithModule("fleets").WithLocker(s.og)
	ed.WithCreationFunc(
		func(input RouteData) ([]string, error) {
			// We need to iterate over the data retrieved from the route and
			// create fleets from it.
			var fleet game.Fleet
			resources := make([]string, 0)

			// Prevent request with no data.
			if len(input.Data) == 0 {
				return resources, ErrNoData
			}

			for _, rawData := range input.Data {
				// Try to unmarshal the data into a valid `Fleet` struct.
				err := json.Unmarshal([]byte(rawData), &fleet)
				if err != nil {
					return resources, ErrInvalidData
				}

				// Create the fleet component.
				res, err := s.fleets.CreateFleet(fleet)
				if err != nil {
					return resources, err
				}

				// Successfully created a fleet.
				resources = append(resources, res)
			}

			// Return the path to the resources created during the process.
			return resources, nil
		},
	)

	return ed.ServeRoute(s.log)
}
