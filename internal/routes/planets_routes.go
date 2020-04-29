package routes

import (
	"net/http"
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
		"universe":     "pl.universe",
		"player":       "p.player",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("p.id").WithModule("planets")
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.planets.Planets(filters)
		},
	)

	return ed.ServeRoute(s.log)
}
