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
	ed := NewEndpointDesc("planets")

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
	ed.WithFilters(allowed).WithIDFilter("id")
	ed.WithDataFunc(
		func(filters []data.DBFilter) (interface{}, error) {
			return s.planets.Planets(filters)
		},
	)

	return ed.ServeRoute(s.log)
}
