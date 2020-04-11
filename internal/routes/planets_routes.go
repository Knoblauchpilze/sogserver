package routes

import (
	"net/http"
	"oglike_server/internal/data"
)

// listPlanets :
// Used to perform the creation of a handler allowing to serve
// the requests on planets.
//
// Returns the handler that can be executed to serve said reqs.
func (s *server) listPlanets() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("planets")

	allowed := map[string]string{
		"planet_id":    "p.id",
		"planet_name":  "p.name",
		"galaxy":       "p.galaxy",
		"solar_system": "p.solar_system",
		"universe":     "pl.uni",
		"player_id":    "p.player",
		"account_id":   "pl.account",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithIDFilter("p.id")
	ed.WithDataFunc(
		func(filters []data.DBFilter) (interface{}, error) {
			return s.planets.Planets(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// listPlanetBuildings :
// Used to perform the creation of a handler allowing to serve
// the requests on building upgrade actions for a planet.
//
// Returns the handler that can be executed to serve said reqs.
func (s *server) listPlanetBuildings() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("planets")

	allowed := map[string]string{
		"action_id":     "id",
		"building_id":   "building",
		"current_level": "current_level",
		"desired_level": "desired_level",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithIDFilter("planet")
	ed.WithDataFunc(
		func(filters []data.DBFilter) (interface{}, error) {
			return s.upgradeAction.Buildings(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// listPlanetShips :
// Used to perform the creation of a handler allowing to serve
// the requests on ship upgrade actions for a planet.
//
// Returns the handler that can be executed to serve said reqs.
func (s *server) listPlanetShips() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("planets")

	allowed := map[string]string{
		"action_id":     "id",
		"ship_id":       "ship",
		"current_level": "current_level",
		"desired_level": "desired_level",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithIDFilter("planet")
	ed.WithDataFunc(
		func(filters []data.DBFilter) (interface{}, error) {
			return s.upgradeAction.Ships(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// listPlanetDefenses :
// Used to perform the creation of a handler allowing to serve
// the requests on defense upgrade actions for a planet.
//
// Returns the handler that can be executed to serve said reqs.
func (s *server) listPlanetDefenses() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("planets")

	allowed := map[string]string{
		"action_id":     "id",
		"defense_id":    "defense",
		"current_level": "current_level",
		"desired_level": "desired_level",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithIDFilter("planet")
	ed.WithDataFunc(
		func(filters []data.DBFilter) (interface{}, error) {
			return s.upgradeAction.Defenses(filters)
		},
	)

	return ed.ServeRoute(s.log)
}
