package routes

import (
	"fmt"
	"net/http"
	"oglike_server/pkg/logger"
)

// listTechnologies :
// Queries the internal database to get a list of all the
// technologies available in the data model of the game.
// The return value is served through a `json` array to the
// client.
//
// Returns the handler to execute to serve such requests.
func (s *server) listTechnologies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars, err := s.extractRouteVars("/technologies", r)
		if err != nil {
			panic(fmt.Errorf("Error while serving technologies (err: %v)", err))
		}

		// We have to assume that no `extra route` is provided on this
		// endpoint.
		if vars.path != "" {
			s.log.Trace(logger.Warning, fmt.Sprintf("Detected ignored extra route \"%s\" when serving technologies", vars.path))
		}

		// Retrieve the filtering options (to potentially retrieve only
		// some technologies and not all of them).
		filters := parseTechnologiesFilters(vars)

		// Retrieve the technologies by querying the database.
		technologies, err := s.technologies.Technologies(filters)
		if err != nil {
			s.log.Trace(logger.Error, fmt.Sprintf("Unexpected error while fetching technologies (err: %v)", err))
			http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

			return
		}

		// Marshal the content of the technologies.
		err = marshalAndSend(technologies, w)
		if err != nil {
			s.log.Trace(logger.Error, fmt.Sprintf("Error while sending technologies to client (err: %v)", err))
		}
	}
}
