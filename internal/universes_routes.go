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
		vars, err := s.extractRouteVars("/universes", r)
		if err != nil {
			panic(fmt.Errorf("Error while serving universes (err: %v)", err))
		}

		// We have to assume that no `extra route` is provided on this
		// endpoint.
		if vars.path != "" {
			s.log.Trace(logger.Warning, fmt.Sprintf("Detected ignored extra route \"%s\" when serving universes", vars.path))
		}

		s.log.Trace(logger.Warning, fmt.Sprintf("Should serve universes: vars are %v", vars))
	}
}

// listUniverse :
// Analyze the route provided in input to retrieve the properties of
// all universes matching the requested information. This is usually
// used in coordination with the `listUniverses` method where the
// user will first fetch a list of all universes and then maybe use
// this list to query specific properties of some universe.
// The return value includes the list of properties using a `json`
// format.
//
// Returns the handler that can be executed to serve such requests.
func (s *server) listUniverse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars, err := s.extractRouteVars("/universes", r)
		if err != nil {
			panic(fmt.Errorf("Error while serving universe (err: %v)", err))
		}

		s.log.Trace(logger.Warning, fmt.Sprintf("Should serve universe: vars are %v", vars))
		// 	http.HandleFunc("/universes/universe_id/coordinates", handlers.NotFound(s.log))
		// 	http.HandleFunc("/universes/universe_id/coordinates/galaxy_id", handlers.NotFound(s.log))
		// 	http.HandleFunc("/universes/universe_id/coordinates/galaxy_id/system_id", handlers.NotFound(s.log))

		// 	http.HandleFunc("/universes/universe_id/planets", handlers.NotFound(s.log))
		// 	http.HandleFunc("/universes/universe_id/planet_id", handlers.NotFound(s.log))
		// 	http.HandleFunc("/universes/universe_id/planet_id/buildings", handlers.NotFound(s.log))
		// 	http.HandleFunc("/universes/universe_id/planet_id/ships", handlers.NotFound(s.log))
		// 	http.HandleFunc("/universes/universe_id/planet_id/fleets", handlers.NotFound(s.log))
		// 	http.HandleFunc("/universes/universe_id/planet_id/defenses", handlers.NotFound(s.log))
	}
}
