package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/game"
	"oglike_server/pkg/db"
)

// listFleetObjectives :
// Used to perform the creation of a handler allowing to server
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
// Used to perform the creation of a handler allowing to server
// the requests to create fleet components.
//
// Returns the handler to execute to perform said requests.
func (s *Server) createFleetComponent() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewCreateResourceEndpoint("players")

	// Configure the endpoint.
	ed.WithDataKey("fleet-data").WithoutPrefix().WithModule("fleets")
	ed.WithCreationFunc(
		func(input RouteData) ([]string, error) {
			// We need to iterate over the data retrieved from the route and
			// create fleets from it.
			var comp game.Component
			resources := make([]string, 0)

			// Prevent request with no data.
			if len(input.Data) == 0 {
				return resources, ErrNoData
			}

			// Make sure that we can retrieve the identifier of the
			// planet for which the component should be added from
			// the route's data.
			if len(input.ExtraElems) != 4 || input.ExtraElems[1] != "planets" || input.ExtraElems[3] != "fleets" {
				return resources, ErrInvalidData
			}

			player := input.ExtraElems[0]
			planet := input.ExtraElems[2]

			for _, rawData := range input.Data {
				// Try to unmarshal the data into a valid `Component` struct.
				err := json.Unmarshal([]byte(rawData), &comp)
				if err != nil {
					return resources, ErrInvalidData
				}

				// Make sure that this component is linked to the
				// planet and player described in the route.
				comp.Source = planet
				comp.SourceType = game.World
				comp.Player = player

				// Create the fleet component.
				res, err := s.fleets.CreateComponent(comp)
				if err != nil {
					return resources, err
				}

				// Successfully created a fleet component: we should prefix
				// the resource by a `components/` string in order to have
				// consistency with the input route. We should also prefix
				// with the fleet's identifier.
				fullRes := fmt.Sprintf("fleets/%s", res)
				resources = append(resources, fullRes)
			}

			// Return the path to the resources created during the process.
			return resources, nil
		},
	)

	return ed.ServeRoute(s.log)
}
