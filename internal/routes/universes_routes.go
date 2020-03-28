package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/logger"
	"strings"
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
		err = marshalAndSend(unis, w)
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

		// Extract parts of the route: we first need to remove the first
		// '/' character.
		purged := vars.path[1:]
		parts := strings.Split(purged, "/")

		// Depending on the number of parts in the route, we will call
		// the suited handler.
		switch len(parts) {
		case 2:
			// The second argument should be `planets`
			if parts[1] != "planets" {
				s.log.Trace(logger.Warning, fmt.Sprintf("Detected ignored extra route \"%s\" when serving planets for universe \"%s\"", parts[1], parts[0]))
			}
			s.listPlanetsForUniverse(w, parts[0], vars)
			return
		case 3:
			s.listPlanetsProps(w, parts)
			return
		case 1:
			fallthrough
		default:
			// Can't do anything.
		}

		s.log.Trace(logger.Error, fmt.Sprintf("Unhandled request for universe \"%s\"", purged))
		http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)
	}
}

// createUniverse :
// Produce a handler that can be used to perform the creation of the
// universe. This operation should only be performed by an admin and
// is usually not intended to be executed very often.
//
// Returns the handler to execute to handle universes' creation.
func (s *server) createUniverse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars, err := s.extractRouteVars("/universe", r)
		if err != nil {
			panic(fmt.Errorf("Error while creating universe (err: %v)", err))
		}

		// The route should not contain anymore data.
		if vars.path != "" {
			s.log.Trace(logger.Warning, fmt.Sprintf("Detected ignored extra route \"%s\" when creating universe", vars.path))
		}

		// Extract the data to use to create the universe from the input
		// request: this can be done conveniently through the server's
		// base method.
		var uniData routeData
		uniData, err = s.extractRouteData(r)
		if err != nil {
			panic(fmt.Errorf("Error while fetching data to create universe (err: %v)", err))
		}

		// Unmarshal the content into a valid universe.
		var uni data.Universe
		err = json.Unmarshal([]byte(uniData.value), &uni)
		if err != nil {
			panic(fmt.Errorf("Error while parsing data to create universe (err: %v)", err))
		}

		err = s.universes.Create(&uni)
		if err != nil {
			s.log.Trace(logger.Error, fmt.Sprintf("Could not create universe from name \"%s\" (err: %v)", uni.Name, err))
			http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

			return
		}
	}
}

// listPlanetsForUniverse :
// Used to forrmat the list of planets registered in the universe set
// as input argument. The resulting list is marshalled into a `json`
// structure and returned to the client through the provided response
// writer.
//
// The `w` argument defines the response writer to send back data to
// the client.
//
// The `universe` represents the identifier of the universe for which
// planet should be fetched.
//
// The `vars` represent the query parameters that were passed to the
// server with the input request. This is useful to extract filtering
// options to use when fetching planets.
func (s *server) listPlanetsForUniverse(w http.ResponseWriter, universe string, vars routeVars) {
	// Fetch filtering properties to narrow the planets returned.
	filters := parsePlanetsFilters(vars)

	// Fetch planets.
	planets, err := s.universes.Planets(universe, filters)
	if err != nil {
		s.log.Trace(logger.Error, fmt.Sprintf("Unexpected error while fetching universe \"%s\" (err: %v)", universe, err))
		http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

		return
	}

	// Marshal and send the content.
	err = marshalAndSend(planets, w)
	if err != nil {
		s.log.Trace(logger.Error, fmt.Sprintf("Unexpected error while sending data for \"%s\" (err: %v)", universe, err))
	}
}

// listPlanetsProps :
// Used to retrieve information about a certain planet and provide
// the corresponding info to the client through the response writer
// given as input.
//
// The `w` is the response writer to use to send the response back
// to the client.
//
// The `params` represents the aprameters provided to filter the
// data to retrieve for this planet. The first element of this
// array is guaranteed to correspond to the identifier of the
// planet for which the data should be retrieved.
func (s *server) listPlanetsProps(w http.ResponseWriter, params []string) {
	//  We know that the first elements of the `params` array should
	// correspond to the planet's identifier (i.e. the root value
	// where specific information should be fetched. The rest of the
	// values correspond to filtering properties to query only some
	// information.
	// We only consider routes that try to access specific props of
	// the planet: the route `universe_id/planet_id` is not valid
	// and will be served an error.
	uni := params[0]

	if len(params) == 1 {
		s.log.Trace(logger.Error, fmt.Sprintf("Unhandled request for universe \"%s\"", uni))
		http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

		return
	}

	// We need to feetch the planet's data from its identifier.
	planetID := params[1]

	planet, err := s.universes.Planet(uni, planetID)
	if err != nil {
		s.log.Trace(logger.Error, fmt.Sprintf("Unable to find planet \"%s\" associated to universe \"%s\" (err: %v)", planetID, uni, err))
		http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

		return
	}

	// Retrieve specific information of the player.
	var errSend error
	var buildings []data.Building
	var ships []data.Ship
	var defenses []data.Defense
	var resources []data.Resource

	switch params[2] {
	case "buildings":
		buildings, err = s.universes.Buildings(planet.ID)
		if err == nil {
			errSend = marshalAndSend(buildings, w)
		}
	case "ships":
		ships, err = s.universes.Ships(planet.ID)
		if err == nil {
			errSend = marshalAndSend(ships, w)
		}
	case "defenses":
		defenses, err = s.universes.Defenses(planet.ID)
		if err == nil {
			errSend = marshalAndSend(defenses, w)
		}
	case "resources":
		resources, err = s.universes.Resources(planet.ID)
		if err == nil {
			errSend = marshalAndSend(resources, w)
		}
	}

	// Notify errors.
	if err != nil {
		s.log.Trace(logger.Error, fmt.Sprintf("Unable to fetch data for planet \"%s\" (err: %v)", planet.ID, err))
		http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

		return
	}

	if errSend != nil {
		s.log.Trace(logger.Error, fmt.Sprintf("Unexpected error while sending data for planet \"%s\" (err: %v)", planet.ID, err))
	}
}
