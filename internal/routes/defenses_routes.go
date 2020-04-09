package routes

import (
	"net/http"
	"oglike_server/internal/data"
)

// listDefenses :
// Used to perform the creation of a handler allowing to serve
// the requests on defenses.
//
// Returns the handler that can be executed to serve said reqs.
func (s *server) listDefenses() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewEndpointDesc("defenses")

	allowed := map[string]string{
		"defense_id":   "id",
		"defense_name": "name",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithIDFilter("id")
	ed.WithDataFunc(
		func(filters []data.DBFilter) (interface{}, error) {
			return s.defenses.Defenses(filters)
		},
	)

	return ed.ServeRoute(s.log)
}
