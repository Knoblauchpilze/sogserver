package routes

import (
	"net/http"
	"oglike_server/internal/data"
)

// listTechnologies :
// Used to perform the creation of a handler allowing to serve
// the requests on technologies.
//
// Returns the handler that can be executed to serve said reqs.
func (s *server) listTechnologies() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("technologies")

	allowed := map[string]string{
		"technology_id":   "id",
		"technology_name": "name",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id")
	ed.WithDataFunc(
		func(filters []data.DBFilter) (interface{}, error) {
			return s.technologies.Technologies(filters)
		},
	)

	return ed.ServeRoute(s.log)
}
