package routes

import (
	"net/http"
	"oglike_server/pkg/db"
)

// listShips :
// Used to perform the creation of a handler allowing to serve
// the requests on ships.
//
// Returns the handler that can be executed to serve said reqs.
func (s *Server) listShips() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("ships")

	allowed := map[string]string{
		"id":   "id",
		"name": "name",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id").WithModule("ships")
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.og.Ships.Ships(s.proxy, filters)
		},
	)

	return ed.ServeRoute(s.log)
}
