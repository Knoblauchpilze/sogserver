package routes

import (
	"net/http"
	"oglike_server/pkg/db"
)

// listResources :
// Used to perform the creation of a handler allowing to serve
// the requests on resources.
//
// Returns the handler that can be executed to serve said reqs.
func (s *Server) listResources() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("resources")

	allowed := map[string]string{
		"id":   "id",
		"name": "name",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id").WithModule("resources")
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.og.Resources.Resources(s.proxy, filters)
		},
	)

	return ed.ServeRoute(s.log)
}
