package routes

import (
	"net/http"
	"oglike_server/pkg/dispatcher"
)

// routes :
// Used to setup all the routes able to be served by this server.
// All the routes are set up with the adequate handler but no
// actual binding is done.
func (s *server) routes() {
	// Handle known routes.
	s.route("GET", "/universes", s.listUniverses())
	s.route("GET", "/accounts", s.listAccounts())
	// TODO: Should update the technology upgrade actions.
	s.route("GET", "/players", s.listPlayers())
	s.route("GET", "/buildings", s.listBuildings())
	s.route("GET", "/technologies", s.listTechnologies())
	s.route("GET", "/ships", s.listShips())
	s.route("GET", "/defenses", s.listDefenses())
	// TODO: Update upgrade actions (from resources, buildings, ships,
	// defenses, fleets) for *this* planet.
	s.route("GET", "/planets", s.listPlanets())
	s.route("GET", "/fleets", s.listFleets())
	s.route("GET", "/fleets/{a-zA-Z0-9-}+/components", s.listFleets())

	s.route("POST", "/universes", s.createUniverse())
	s.route("POST", "/accounts", s.createAccount())
	s.route("POST", "/players", s.createPlayer())
	s.route("POST", "/fleets", s.createFleet())
	// TODO: Should propose the equivalent GET method so that the client
	// can display the relevant timers.
	s.route("POST", "/actions/buildings", s.registerBuildingAction())
	s.route("POST", "/actions/technologies", s.registerTechnologyAction())
	s.route("POST", "/actions/ships", s.registerShipAction())
	s.route("POST", "/actions/defenses", s.registerDefenseAction())

}

// route :
// Used to perform the necessary wrapping around the specified
// handler provided that it should be binded to the input route
// and only respond to said method.
//
// The `method` indicates the method for which the handler is
// sensible.
//
// The `name` of the route define the binding that should be
// performed for the input handler.
//
// The `handler` defines the element that will serve input req
// and which should be wrapped to provide more security.
func (s *server) route(method string, name string, handler http.HandlerFunc) {
	s.router.HandleFunc(
		name,
		dispatcher.WithSafetyNet(
			s.log,
			handler,
		),
	).Methods(method)
}
