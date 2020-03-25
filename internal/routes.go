package internal

import (
	"net/http"
	"oglike_server/pkg/handlers"
)

// routes :
// Used to setup all the routes able to be served by this server.
// All the routes are set up with the adequate handler but no
// actual binding is done.
func (s *server) routes() {
	s.routeUniverses()
	s.routeAccounts()
	s.routeNotImplemented()
}

// routeUniverses :
// Used to set up the routes needed to offer the features related
// to universes.
func (s *server) routeUniverses() {
	// List existing universes.
	http.HandleFunc("/universes", handlers.WithSafetyNet(s.log, s.listUniverses()))

	// List properties of a specific universe.
	http.HandleFunc("/universes/", handlers.WithSafetyNet(s.log, s.listUniverse()))
}

// routeAccounts :
// Similar to the `routeUniverses` facet but sets up the routes to
// serve the functionalities related to the accounts registered in
// the server.
func (s *server) routeAccounts() {
	// List existing accounts.
	http.HandleFunc("/accounts", handlers.WithSafetyNet(s.log, s.listAccounts()))

	// List properties of a specific account.
	http.HandleFunc("/accounts/", handlers.WithSafetyNet(s.log, s.listAccount()))
}

// routeNotImplemented :
// Currently not implemented but will be soon !
func (s *server) routeNotImplemented() {
	http.HandleFunc("/accounts/account_id/player_id/planets", handlers.NotFound(s.log))
	http.HandleFunc("/accounts/account_id/player_id/researches", handlers.NotFound(s.log))
	http.HandleFunc("/accounts/account_id/player_id/fleets", handlers.NotFound(s.log))

	http.HandleFunc("/universes/universe_id/coordinates", handlers.NotFound(s.log))
	http.HandleFunc("/universes/universe_id/coordinates/galaxy_id", handlers.NotFound(s.log))
	http.HandleFunc("/universes/universe_id/coordinates/galaxy_id/system_id", handlers.NotFound(s.log))

	http.HandleFunc("/universes/universe_id/planets", handlers.NotFound(s.log))
	http.HandleFunc("/universes/universe_id/planet_id", handlers.NotFound(s.log))
	http.HandleFunc("/universes/universe_id/planet_id/buildings", handlers.NotFound(s.log))
	http.HandleFunc("/universes/universe_id/planet_id/ships", handlers.NotFound(s.log))
	http.HandleFunc("/universes/universe_id/planet_id/fleets", handlers.NotFound(s.log))
	http.HandleFunc("/universes/universe_id/planet_id/defenses", handlers.NotFound(s.log))
}
