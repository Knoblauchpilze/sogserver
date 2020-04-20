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
		"ship_id":   "id",
		"ship_name": "name",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id").WithModule("ships")
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.ships.Ships(s.dbase, filters)
		},
	)

	return ed.ServeRoute(s.log)
}
