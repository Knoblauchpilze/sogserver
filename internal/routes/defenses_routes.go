package routes

import (
	"fmt"
	"net/http"
	"oglike_server/pkg/logger"
)

// listDefenses :
// Queries the internal database to get a list of all the
// defenses available in the data model of the game. The
// return value is served through a `json` array to the
// client.
//
// Returns the handler to execute to serve such requests.
func (s *server) listDefenses() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars, err := s.extractRouteVars("/defenses", r)
		if err != nil {
			panic(fmt.Errorf("Error while serving defenses (err: %v)", err))
		}

		// We have to assume that no `extra route` is provided on this
		// endpoint.
		if vars.path != "" {
			s.log.Trace(logger.Warning, fmt.Sprintf("Detected ignored extra route \"%s\" when serving defenses", vars.path))
		}

		// Retrieve the filtering options (to potentially retrieve only
		// some defenses and not all of them).
		filters := parseDefensesFilters(vars)

		// Retrieve the defenses by querying the database.
		defenses, err := s.defenses.Defenses(filters)
		if err != nil {
			s.log.Trace(logger.Error, fmt.Sprintf("Unexpected error while fetching defenses (err: %v)", err))
			http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

			return
		}

		// Marshal the content of the defenses.
		err = marshalAndSend(defenses, w)
		if err != nil {
			s.log.Trace(logger.Error, fmt.Sprintf("Error while sending defenses to client (err: %v)", err))
		}
	}
}
