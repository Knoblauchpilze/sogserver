package routes

import (
	"net/http"
	"oglike_server/pkg/handlers"
)

// routes :
// Used to setup all the routes able to be served by this server.
// All the routes are set up with the adequate handler but no
// actual binding is done.
func (s *server) routes() {
	// Handle known routes.
	s.routeUniverses()
	s.routeAccounts()
	s.routeBuildings()
	s.routeTechnologies()
	s.routeShips()
	s.routeDefenses()
	s.routePlanets()
	s.routePlayers()
	s.routeFleets()
	s.routeActions()

	// Default to `NotFound` in any other case.
	http.HandleFunc("/", handlers.NotFound(s.log))
}

// route :
// Used to perform the necessary wrapping around the specified
// handler provided that it should be binded to the input route
// and only respond to said method.
//
// The `name` of the route define the binding that should be
// performed for the input handler.
//
// The `method` indicates the method for which the handler is
// sensible.
//
// The `handler` defines the element that will serve input req
// and which should be wrapped to provide more security.
func (s *server) route(name string, method string, handler http.HandlerFunc) {
	http.HandleFunc(
		name,
		handlers.Method(
			s.log,
			method,
			handlers.WithSafetyNet(s.log, handler),
		),
	)
}

// routeUniverses :
// Used to set up the routes needed to offer the features related
// to universes.
func (s *server) routeUniverses() {
	s.route("/universes", "GET", s.listUniverses())
	s.route("/universes/", "GET", s.listUniverses())

	s.route("/universe", "POST", s.createUniverse())
}

// routeAccounts :
// Similar to the `routeUniverses` facet but sets up the routes to
// serve the functionalities related to the accounts registered in
// the server.
func (s *server) routeAccounts() {
	s.route("/accounts", "GET", s.listAccounts())
	s.route("/accounts/", "GET", s.listAccounts())

	s.route("/account", "POST", s.createAccount())
}

// routeBuildings :
// Similar to the `routeUniverses` facet but sets up the routes to
// serve the functionalities related to the buildings registered in
// the server.
func (s *server) routeBuildings() {
	s.route("/buildings", "GET", s.listBuildings())
	s.route("/buildings/", "GET", s.listBuildings())
}

// routeTechnologies :
// Similar to the `routeUniverses` facet but sets up the routes to
// serve the functionalities related to the technologies registered
// in the server.
func (s *server) routeTechnologies() {
	s.route("/technologies", "GET", s.listTechnologies())
	s.route("/technologies/", "GET", s.listTechnologies())
}

// routeShips :
// Similar to the `routeUniverses` facet but sets up the routes to
// serve the functionalities related to the ships registered in the
// server.
func (s *server) routeShips() {
	s.route("/ships", "GET", s.listShips())
	s.route("/ships/", "GET", s.listShips())
}

// routeDefenses :
// Similar to the `routeUniverses` facet but sets up the routes to
// serve the functionalities related to the defenses registered in
// the server.
func (s *server) routeDefenses() {
	s.route("/defenses", "GET", s.listDefenses())
	s.route("/defenses/", "GET", s.listDefenses())
}

// routePlanets :
// Similar to the `routeUniverses` facet but sets up the routes to
// serve the functionalities related to the planets registered in
// the server.
func (s *server) routePlanets() {
	s.route("/planets", "GET", s.listPlanets())
	s.route("/planets/", "GET", s.listPlanets())
}

// routePlayers :
// Similar to the `routeUniverses` facet but sets up the routes to
// serve functionalities related to the players registered in each
// universe.
func (s *server) routePlayers() {
	s.route("/players", "GET", s.listPlayers())
	s.route("/players/", "GET", s.listPlayers())

	s.route("/player", "POST", s.createPlayer())
}

// routeFleets :
// Similar to the `routeUniverses` facet but sets up the routes to
// serve functionalities related to the fleets registered in each
// universe.
func (s *server) routeFleets() {
	s.route("/fleets", "GET", s.listFleets())
	s.route("/fleets/", "GET", s.listFleets())

	// TODO: We should create a new post route to add components to
	// an existing fleet.
	s.route("/fleet", "POST", s.createFleet())
}

// routeActions :
// Similar to the `routeUniverses` facet but sets up the routes to
// serve functionalities related to the actions to be performed for
// buildings, technologies, ships and defenses.
func (s *server) routeActions() {
	s.route("/action/building", "POST", s.registerBuildingAction())
	s.route("/action/technology", "POST", s.registerTechnologyAction())
	s.route("/action/ship", "POST", s.registerShipAction())
	s.route("/action/defense", "POST", s.registerDefenseAction())
}
