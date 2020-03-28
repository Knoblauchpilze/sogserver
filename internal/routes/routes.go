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

	// Default to `NotFound` in any other case.
	http.HandleFunc("/", handlers.NotFound(s.log))
}

// routeUniverses :
// Used to set up the routes needed to offer the features related
// to universes.
func (s *server) routeUniverses() {
	// List existing universes.
	// GET, `/universes`
	http.HandleFunc(
		"/universes",
		handlers.Method(
			s.log,
			"GET",
			handlers.WithSafetyNet(s.log, s.listUniverses()),
		),
	)

	// List properties of a specific universe.
	// GET, `/universes/universe_id`
	// GET, `/universes/universe_id/planets`

	// GET, `/universes/universe_id/planet_id/buildings`
	// GET, `/universes/universe_id/planet_id/defense`
	// GET, `/universes/universe_id/planet_id/ships`
	http.HandleFunc(
		"/universes/",
		handlers.Method(
			s.log,
			"GET",
			handlers.WithSafetyNet(s.log, s.listUniverse()),
		),
	)

	// Create a new universe.
	// POST, `/universe`, `universe-data`
	http.HandleFunc(
		"/universe",
		handlers.Method(
			s.log,
			"POST",
			handlers.WithSafetyNet(s.log, s.createUniverse()),
		),
	)
}

// routeAccounts :
// Similar to the `routeUniverses` facet but sets up the routes to
// serve the functionalities related to the accounts registered in
// the server.
func (s *server) routeAccounts() {
	// List existing accounts.
	// GET, `/accounts/account_id`
	// GET, `/accounts/account_id/players`

	// GET, `/accounts/account_id/player_id/planets`
	// GET, `/accounts/account_id/player_id/technologies`
	// GET, `/accounts/account_id/player_id/fleets`
	http.HandleFunc(
		"/accounts",
		handlers.Method(
			s.log,
			"GET",
			handlers.WithSafetyNet(s.log, s.listAccounts()),
		),
	)

	// List properties of a specific account.
	// GET, `/accounts`
	http.HandleFunc(
		"/accounts/",
		handlers.Method(
			s.log,
			"GET",
			handlers.WithSafetyNet(s.log, s.listAccount()),
		),
	)

	// Create a new universe.
	// POST, `/account`, `account-data`
	http.HandleFunc(
		"/account",
		handlers.Method(
			s.log,
			"POST",
			handlers.WithSafetyNet(s.log, s.createAccount()),
		),
	)
}
