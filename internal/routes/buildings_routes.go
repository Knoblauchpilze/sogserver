package routes

import (
	"net/http"
	"oglike_server/internal/data"
)

// listBuildings :
// Used to perform the creation of a handler allowing to serve
// the requests on buildings.
//
// Returns the handler that can be executed to serve said reqs.
func (s *server) listBuildings() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("buildings")

	allowed := map[string]string{
		"building_id":   "id",
		"building_name": "name",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id")
	ed.WithDataFunc(
		func(filters []data.DBFilter) (interface{}, error) {
			return s.buildings.Buildings(filters)
		},
	)

	return ed.ServeRoute(s.log)
}
