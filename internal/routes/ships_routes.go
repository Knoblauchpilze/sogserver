package routes

import (
	"net/http"
	"oglike_server/internal/data"
)

// listShips :
// Used to perform the creation of a handler allowing to serve
// the requests on ships.
//
// Returns the handler that can be executed to serve said reqs.
func (s *server) listShips() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewEndpointDesc("ships")

	allowed := map[string]string{
		"ship_id":   "id",
		"ship_name": "name",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithIDFilter("id")
	ed.WithDataFunc(
		func(filters []data.DBFilter) (interface{}, error) {
			return s.ships.Ships(filters)
		},
	)

	return ed.ServeRoute(s.log)
}
