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
	s.route("GET", "/players", s.listPlayers())
	s.route("GET", "/players/[a-zA-Z0-9-]+/actions/technologies", s.listPlayerTechnologies())
	s.route("GET", "/buildings", s.listBuildings())
	s.route("GET", "/technologies", s.listTechnologies())
	s.route("GET", "/ships", s.listShips())
	s.route("GET", "/defenses", s.listDefenses())
	// TODO: Update upgrade actions (from fleets) for *this* planet.
	s.route("GET", "/planets", s.listPlanets())
	s.route("GET", "/planets/[a-zA-Z0-9-]+/actions/buildings", s.listPlanetBuildings())
	s.route("GET", "/planets/[a-zA-Z0-9-]+/actions/ships", s.listPlanetShips())
	s.route("GET", "/planets/[a-zA-Z0-9-]+/actions/defenses", s.listPlanetDefenses())
	s.route("GET", "/fleets", s.listFleets())
	s.route("GET", "/fleets/[a-zA-Z0-9-]+/components", s.listFleetComponents())

	s.route("POST", "/universes", s.createUniverse())
	s.route("POST", "/accounts", s.createAccount())
	s.route("POST", "/players", s.createPlayer())
	s.route("POST", "/players/[a-zA-Z0-9-]+/actions/technologies", s.registerTechnologyAction())
	s.route("POST", "/planets/[a-zA-Z0-9-]+/actions/buildings", s.registerBuildingAction())
	s.route("POST", "/planets/[a-zA-Z0-9-]+/actions/ships", s.registerShipAction())
	s.route("POST", "/planets/[a-zA-Z0-9-]+/actions/defenses", s.registerDefenseAction())
	s.route("POST", "/fleets", s.createFleet())
	s.route("POST", "/fleets/[a-zA-Z0-9-]+/components", s.createFleetComponent())
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
