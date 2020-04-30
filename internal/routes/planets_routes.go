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
	ed := NewGetResourceEndpoint("players")

	allowed := map[string]string{
		"id":           "id",
		"name":         "name",
		"galaxy":       "galaxy",
		"solar_system": "solar_system",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id").WithIDFilter("player").WithModule("planets")
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.planets.Planets(filters)
		},
	)

	return ed.ServeRoute(s.log)
}
