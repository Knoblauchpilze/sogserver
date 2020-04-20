package routes

import (
	"net/http"
	"oglike_server/pkg/db"
)

// listBuildings :
// Used to perform the creation of a handler allowing to serve
// the requests on buildings.
//
// Returns the handler that can be executed to serve said reqs.
func (s *Server) listBuildings() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("buildings")

	allowed := map[string]string{
		"building_id":   "id",
		"building_name": "name",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id").WithModule("buildings")
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.buildings.Buildings(s.dbase, filters)
		},
	)

	return ed.ServeRoute(s.log)
}
