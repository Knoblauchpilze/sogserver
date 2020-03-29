package routes

import (
	"fmt"
	"net/http"
	"oglike_server/pkg/logger"
)

// listBuildings :
// Queries the internal database to get a list of the buildings
// available in the data model of the game. The return value is
// served through a `json` array to the client.
//
// Returns the handler to execute to serve such requests.
func (s *server) listBuildings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars, err := s.extractRouteVars("/buildings", r)
		if err != nil {
			panic(fmt.Errorf("Error while serving buildings (err: %v)", err))
		}

		// We have to assume that no `extra route` is provided on this
		// endpoint.
		if vars.path != "" {
			s.log.Trace(logger.Warning, fmt.Sprintf("Detected ignored extra route \"%s\" when serving buildings", vars.path))
		}

		// Retrieve the filtering options (to potentially retrieve only
		// some buildings and not all of them).
		filters := parseBuildingsFilters(vars)

		// Retrieve the buildings by querying the database.
		buildings, err := s.buildings.Buildings(filters)
		if err != nil {
			s.log.Trace(logger.Error, fmt.Sprintf("Unexpected error while fetching buildings (err: %v)", err))
			http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

			return
		}

		// Marshal the content of the buildings.
		err = marshalAndSend(buildings, w)
		if err != nil {
			s.log.Trace(logger.Error, fmt.Sprintf("Error while sending buildings to client (err: %v)", err))
		}
	}
}
