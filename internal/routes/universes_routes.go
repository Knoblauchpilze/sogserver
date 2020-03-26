package routes

import (
	"encoding/json"
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

		// Retrieve the universes from the bridge.
		unis, err := s.universes.Universes()
		if err != nil {
			s.log.Trace(logger.Error, fmt.Sprintf("Unexpected error while fetching universes (err: %v)", err))
			http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

			return
		}

		// Marshal the content of the universes.
		out, err := json.Marshal(unis)
		if err != nil {
			s.log.Trace(logger.Error, fmt.Sprintf("Error while marshalling universes (err: %v)", err))
			http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

			return
		}

		// Notify the client.
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(out)

		if err != nil {
			s.log.Trace(logger.Error, fmt.Sprintf("Error while sending universes to client (err: %v)", err))
		}
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
	}
}
