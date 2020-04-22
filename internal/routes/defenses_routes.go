package routes

import (
	"net/http"
	"oglike_server/pkg/db"
)

// listDefenses :
// Used to perform the creation of a handler allowing to serve
// the requests on defenses.
//
// Returns the handler that can be executed to serve said reqs.
func (s *Server) listDefenses() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("defenses")

	allowed := map[string]string{
		"defense_id":   "id",
		"defense_name": "name",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id").WithModule("defenses")
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.og.Defenses.Defenses(s.proxy, filters)
		},
	)

	return ed.ServeRoute(s.log)
}
