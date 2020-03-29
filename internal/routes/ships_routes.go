package routes

import (
	"fmt"
	"net/http"
	"oglike_server/pkg/logger"
)

// listShips :
// Queries the internal database to get a list of the ships
// available in the data model of the game. The return value
// is served through a `json` array to the client.
//
// Returns the handler to execute to serve such requests.
func (s *server) listShips() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars, err := s.extractRouteVars("/ships", r)
		if err != nil {
			panic(fmt.Errorf("Error while serving ships (err: %v)", err))
		}

		// We have to assume that no `extra route` is provided on this
		// endpoint.
		if vars.path != "" {
			s.log.Trace(logger.Warning, fmt.Sprintf("Detected ignored extra route \"%s\" when serving ships", vars.path))
		}

		// Retrieve the filtering options (to potentially retrieve only
		// some ships and not all of them).
		filters := parseShipsFilters(vars)

		// Retrieve the ships by querying the database.
		ships, err := s.ships.Ships(filters)
		if err != nil {
			s.log.Trace(logger.Error, fmt.Sprintf("Unexpected error while fetching ships (err: %v)", err))
			http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

			return
		}

		// Marshal the content of the ships.
		err = marshalAndSend(ships, w)
		if err != nil {
			s.log.Trace(logger.Error, fmt.Sprintf("Error while sending ships to client (err: %v)", err))
		}
	}
}
