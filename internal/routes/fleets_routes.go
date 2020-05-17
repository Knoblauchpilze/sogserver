package routes

import (
	"encoding/json"
	"net/http"
	"oglike_server/internal/game"
	"oglike_server/pkg/db"
)

// fleetCreationFunc :
// Convenience define allowing to refer to the creation
// process of a fleet. Depending on whether the fleet to
// create should be integrated into an ACS fleet or not
// the creation will be slightly different.
// Using this define allow to mutualize all the fetching
// part of the code and only specialize things at the
// very last step.
//
// The `fleet` defines the fleet's data fetched from
// the route. This represent the resource to create.
//
// The return value includes both any error and the ID
// of the fleet that was created from the `fleet` input
// data.
type fleetCreationFunc func(fleet game.Fleet) (string, error)

// listFleets :
// Used to perform the creation of a handler allowing to serve
// the requests on fleets.
//
// Returns the handler that can be executed to serve said reqs.
func (s *Server) listFleets() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("fleets")

	allowed := map[string]string{
		"id":           "f.id",
		"universe":     "f.uni",
		"objective":    "f.objective",
		"source":       "f.source",
		"target":       "f.target",
		"galaxy":       "f.target_galaxy",
		"solar_system": "f.target_solar_system",
		"position":     "f.target_position",
		"acs":          "fac.acs",
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

// listACSFleets :
// Used to perform the creation of a handler allowing to serve
// the requests for ACS fleets.
//
// Returns the handler than can be executed to serve said reqs.
func (s *Server) listACSFleets() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("fleets/acs")

	allowed := map[string]string{
		"id":          "id",
		"universe":    "universe",
		"objective":   "objective",
		"target":      "target",
		"target_type": "target_type",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id").WithModule("fleets").WithLocker(s.og)
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.fleets.ACSFleets(filters)
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

// createGenericFleet :
// Used to mutualize the common code used to create fleets and
// ACS fleets. Most of the code is similar except the final
// creation functio so we figured it would make sense to kinda
// factor the rest of the code creation.
//
// The `route` defines the route associated to the endpoint.
//
// The `creator` defines the creation function to call once
// the fleet has been unmarshalled from input data.
//
// Returns the created handler.
func (s *Server) createGenericFleet(route string, create fleetCreationFunc) http.HandlerFunc {
	// Create the endpoint from the route.
	ed := NewCreateResourceEndpoint(route)

	// Configure the endpoint.
	ed.WithDataKey("fleet-data").WithModule("fleets").WithLocker(s.og)

	// The unmarshalling process is always the same, only the
	// last creation step is actually specific.
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
				res, err := create(fleet)
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

// createFleet :
// Used to perform the creation of a handler allowing to serve
// the requests to create a fleet.
//
// Returns the handler to execute to perform said requests.
func (s *Server) createFleet() http.HandlerFunc {
	return s.createGenericFleet(
		"fleets",
		func(fleet game.Fleet) (string, error) {
			return s.fleets.CreateFleet(fleet)
		},
	)
}

// createACSFleet :
// Used to perform the creation of a handler allowing to serve
// the requets to create an ACS fleet.
//
// Returns the handler to execute to perform said requests.
func (s *Server) createACSFleet() http.HandlerFunc {
	return s.createGenericFleet(
		"fleets/acs",
		func(fleet game.Fleet) (string, error) {
			return s.fleets.CreateACSFleet(fleet)
		},
	)
}
