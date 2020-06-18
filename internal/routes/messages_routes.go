package routes

import (
	"net/http"
	"oglike_server/pkg/db"
)

// listMessages :
// Used to perform the creation of a handler allowing to serve
// the requests on messages.
//
// Returns the handler that can be executed to serve said reqs.
func (s *Server) listMessages() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("messages")

	allowed := map[string]string{
		"id":   "mi.id",
		"name": "mi.name",
		"type": "mt.type",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id").WithModule("messages")
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.og.Messages.Messages(s.proxy, filters)
		},
	)

	return ed.ServeRoute(s.log)
}
