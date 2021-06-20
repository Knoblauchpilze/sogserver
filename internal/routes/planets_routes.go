package routes

import (
	"encoding/json"
	"net/http"
	"oglike_server/internal/game"
	"oglike_server/pkg/db"
)

// listPlanets :
// Used to perform the creation of a handler allowing to serve
// the requests on planets.
//
// Returns the handler that can be executed to serve said reqs.
func (s *Server) listPlanets() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("planets")

	allowed := map[string]string{
		"id":           "p.id",
		"name":         "p.name",
		"galaxy":       "p.galaxy",
		"solar_system": "p.solar_system",
		"position":     "p.position",
		"player":       "p.player",
		"universe":     "pl.universe",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("p.id").WithModule("planets").WithLocker(s.og)
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.planets.Planets(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// listMoons :
// Used to perform the creation of a handler allowing to serve
// the requests on moons.
//
// Returns the handler that can be executed to serve said reqs.
func (s *Server) listMoons() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("moons")

	allowed := map[string]string{
		"id":           "m.id",
		"planet":       "m.planet",
		"name":         "m.name",
		"galaxy":       "p.galaxy",
		"solar_system": "p.solar_system",
		"position":     "p.position",
		"player":       "p.player",
		"universe":     "p.universe",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("m.id").WithModule("moons").WithLocker(s.og)
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.planets.Moons(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// listDebris :
// Used to perform the creation of a handler allowing to serve
// the requests on debris.
//
// Returns the handler that can be executed to serve said reqs.
func (s *Server) listDebris() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("debris")

	allowed := map[string]string{
		"id":           "d.id",
		"universe":     "d.universe",
		"galaxy":       "d.galaxy",
		"solar_system": "d.solar_system",
		"position":     "d.position",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("d.id").WithModule("debris").WithLocker(s.og)
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.planets.Debris(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// changePlanets :
// Used to perform the creation of a handler allowing to serve
// the requests to change a planet.
//
// Returns the handler to execute to perform said requests.
func (s *Server) changePlanets() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewCreateResourceEndpoint("planets")

	// Configure the endpoint.
	ed.WithDataKey("planet-data").WithModule("planets").WithLocker(s.og)
	ed.WithCreationFunc(
		func(input RouteData) ([]string, error) {
			// We need to iterate over the data retrieved from the route and
			// create planets from it.
			var planet game.Planet
			resources := make([]string, 0)

			// Make sure that there's a route element.
			if len(input.ExtraElems) == 0 {
				return resources, ErrNoData
			}

			planetID := input.ExtraElems[0]

			// Prevent request with no data.
			if len(input.Data) == 0 {
				return resources, ErrNoData
			}

			for _, rawData := range input.Data {
				// Try to unmarshal the data into a valid `Planet` struct.
				err := json.Unmarshal([]byte(rawData), &planet)
				if err != nil {
					return resources, ErrInvalidData
				}

				// Force the planet's identifier with the route's data.
				planet.ID = planetID

				// Update the planet.
				res, err := s.planets.Update(planet)
				if err != nil {
					return resources, err
				}

				// Successfully updated a planet.
				resources = append(resources, res)
			}

			// Return the path to the resources updated during the process.
			return resources, nil
		},
	)

	return ed.ServeRoute(s.log)
}

// changeProduction :
// Used to perform the creation of a handler allowing to server
// the requests to change the production of a particular resource
// of a planet.
//
// Returns the handler to execute to perform said requests.
func (s *Server) changeProduction() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewCreateResourceEndpoint("planets")

	// Configure the endpoint.
	ed.WithDataKey("planet-data").WithModule("planets").WithLocker(s.og)
	ed.WithCreationFunc(
		func(input RouteData) ([]string, error) {
			// We need to iterate over the data retrieved from the route and
			// create planets from it.
			var production []game.BuildingInfo
			resources := make([]string, 0)

			// Make sure that there's a route element.
			if len(input.ExtraElems) == 0 {
				return resources, ErrNoData
			}

			planetID := input.ExtraElems[0]

			// Prevent request with no data.
			if len(input.Data) == 0 {
				return resources, ErrNoData
			}

			for _, rawData := range input.Data {
				// Try to unmarshal the data into a valid `ResourceInfo` struct.
				err := json.Unmarshal([]byte(rawData), &production)
				if err != nil {
					return resources, ErrInvalidData
				}

				// Update the planet.
				res, err := s.planets.UpdateProduction(planetID, production)
				if err != nil {
					return resources, err
				}

				// Successfully updated a planet.
				resources = append(resources, res)
			}

			// Return the path to the resources updated during the process.
			return resources, nil
		},
	)

	return ed.ServeRoute(s.log)
}

// changeMoons :
// Used to perform the creation of a handler allowing to serve
// the requests to change a moon.
//
// Returns the handler to execute to perform said requests.
func (s *Server) changeMoons() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewCreateResourceEndpoint("moons")

	// Configure the endpoint.
	ed.WithDataKey("moon-data").WithModule("moons").WithLocker(s.og)
	ed.WithCreationFunc(
		func(input RouteData) ([]string, error) {
			// We need to iterate over the data retrieved from the route and
			// create moons from it.
			var moon game.Planet
			resources := make([]string, 0)

			// Make sure that there's a route element.
			if len(input.ExtraElems) == 0 {
				return resources, ErrNoData
			}

			planetID := input.ExtraElems[0]

			// Prevent request with no data.
			if len(input.Data) == 0 {
				return resources, ErrNoData
			}

			for _, rawData := range input.Data {
				// Try to unmarshal the data into a valid `Planet` struct.
				err := json.Unmarshal([]byte(rawData), &moon)
				if err != nil {
					return resources, ErrInvalidData
				}

				// Force the moon's identifier with the route's data and
				// force the request to be applied on moon.
				moon.ID = planetID
				moon.Moon = true

				// Update the planet.
				res, err := s.planets.Update(moon)
				if err != nil {
					return resources, err
				}

				// Successfully updated a moon.
				resources = append(resources, res)
			}

			// Return the path to the resources updated during the process.
			return resources, nil
		},
	)

	return ed.ServeRoute(s.log)
}

// deletePlanet :
// Used to perform the creation of a handler allowing to serve
// the requests to delete a planet.
//
// Returns the handler to execute to perform said requests.
func (s *Server) deletePlanet() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewDeleteResourceEndpoint("planets")

	// Configure the endpoint.
	ed.WithModule("planets").WithLocker(s.og)
	ed.WithDeleterFunc(
		func(resource string) error {
			return s.planets.Delete(resource)
		},
	)

	return ed.ServeRoute(s.log)
}
