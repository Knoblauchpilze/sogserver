package internal

import (
	"fmt"
	"net/http"
	"oglike_server/pkg/logger"
)

// listUniverses :
// Queries the internal database to get a list of the universes and
// some common properties and serve these values through a `json`
// syntax to the client.
//
// Returns the handler that can be executed to serve such requests.
func (s *server) listUniverses() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route, err := s.extractRoute(r, "/universes")
		if err != nil {
			panic(fmt.Errorf("Error while serving universes (err: %v)", err))
		}

		s.log.Trace(logger.Warning, fmt.Sprintf("Should serve list of all universes (route: \"%s\")", route))
	}
}

// listUniverse :
// Provide detailed information about a single universe referenced
// by the input route. If no such universe exist this will be sent
// back to the client.
//
// Returns the handler to be executed to serve these requests.
func (s *server) listUniverse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uni, err := s.extractRoute(r, "/universes/")
		if err != nil {
			panic(fmt.Errorf("Error while serving universe (err: %v)", err))
		}

		s.log.Trace(logger.Warning, fmt.Sprintf("Should serve info for universe \"%s\"", uni))
	}
}
