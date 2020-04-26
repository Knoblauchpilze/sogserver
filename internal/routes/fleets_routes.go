package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/model"
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
		"fleet_id":     "id",
		"fleet_name":   "name",
		"galaxy":       "target_galaxy",
		"solar_system": "target_solar_system",
		"position":     "target_position",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id").WithModule("fleets")
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.fleets.Fleets(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// createFleet :
// Used to perform the creation of a handler allowing to server
// the requests to create fleets.
//
// Returns the handler to execute to perform said requests.
func (s *Server) createFleet() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewCreateResourceEndpoint("fleets")

	// Configure the endpoint.
	ed.WithDataKey("fleet-data").WithModule("fleets")
	ed.WithCreationFunc(
		func(input RouteData) ([]string, error) {
			// We need to iterate over the data retrieved from the route and
			// create fleets from it.
			var fleet model.Fleet
			resources := make([]string, 0)

			// Prevent request with no data.
			if len(input.Data) == 0 {
				return resources, ErrNoData
			}

			// Iterate over the provided data to create the corresponding
			// fleets in the main DB.
			for _, rawData := range input.Data {
				// Try to unmarshal the data into a valid `Fleet` struct.
				err := json.Unmarshal([]byte(rawData), &fleet)
				if err != nil {
					return resources, ErrInvalidData
				}

				// Create the fleet.
				res, err := s.fleets.Create(fleet)
				if err != nil {
					return resources, ErrDBError
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

// createFleetComponent :
// Used to perform the creation of a handler allowing to server
// the requests to create fleet components.
//
// Returns the handler to execute to perform said requests.
func (s *Server) createFleetComponent() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewCreateResourceEndpoint("fleets")

	// Configure the endpoint.
	ed.WithDataKey("fleet-data").WithModule("fleets")
	ed.WithCreationFunc(
		func(input RouteData) ([]string, error) {
			// We need to iterate over the data retrieved from the route and
			// create fleets from it.
			var comp model.Component
			resources := make([]string, 0)

			// Prevent request with no data.
			if len(input.Data) == 0 {
				return resources, ErrNoData
			}

			// Make sure that we can retrieve the identifier of the
			// fleet for which the component should be created from
			// the route's data.
			if len(input.RouteElems) != 2 || input.RouteElems[1] != "components" {
				return resources, ErrInvalidData
			}

			fleetID := input.RouteElems[0]

			for _, rawData := range input.Data {
				// Try to unmarshal the data into a valid `Component` struct.
				err := json.Unmarshal([]byte(rawData), &comp)
				if err != nil {
					return resources, ErrInvalidData
				}

				// Create the fleet component.
				res, err := s.fleets.CreateComponent(fleetID, comp)
				if err != nil {
					return resources, ErrDBError
				}

				// Successfully created a fleet component: we should prefix
				// the resource by a `components/` string in order to have
				// consistency with the input route. We should also prefix
				// with the fleet's identifier.
				fullRes := fmt.Sprintf("%s/components/%s", fleetID, res)
				resources = append(resources, fullRes)
			}

			// Return the path to the resources created during the process.
			return resources, nil
		},
	)

	return ed.ServeRoute(s.log)
}
