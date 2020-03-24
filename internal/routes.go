package internal

import (
	"fmt"
	"net/http"
	"oglike_server/pkg/handlers"
	"strconv"
	"strings"
)

// routes :
// Used to setup all the routes able to be served by this server.
// All the routes are set up with the adequate handler but no
// actual binding is done.
func (s *server) routes() {
	s.routeUniverses()
	s.routeAccounts()
}

// routeUniverses :
// Used to set up the routes needed to offer the features related
// to universes. This is composed of several routes each one aiming
// at serving a single feature of the server.
func (s *server) routeUniverses() {
	// List existing universes.
	http.HandleFunc("/universes", handlers.WithSafetyNet(s.log, s.listUniverses()))

	// Get details about a universe.
	http.HandleFunc("/universes/", handlers.WithSafetyNet(s.log, s.listUniverse()))
}

// routeAccounts :
// Similar to the `routeUniverses` facet but sets up the routes to
// serve the functionalities related to the accounts registered in
// the server.
func (s *server) routeAccounts() {
	// List existing accounts.
	http.HandleFunc("/accounts", handlers.WithSafetyNet(s.log, s.listAccounts()))

	// Get details about a specific account.
	http.HandleFunc("/accounts/", handlers.WithSafetyNet(s.log, s.listAccount()))
}

// Serve :
// Used to start listening to the port associated to this server
// and handle incoming requests. This will return an error in case
// something went wrong while listening to the port.
func (s *server) Serve() error {
	// Setup routes.
	s.routes()

	// Serve the root path.
	return http.ListenAndServe(":"+strconv.FormatInt(int64(s.port), 10), nil)
}

// extractRoute :
// Convenience method allowing to strip the input prefix from the
// route defined in an input request to keep only the part that is
// specific to a server behavior.
//
// The `r` argument represents the request from which the route
// should be extracted. An error is raised in case this requets is
// not valid.
//
// The `prefix` represents the prefix to be stripped from the input
// request. If the prefix does not exist in the route an error is
// returned as well.
//
// Returns either the route stripped from the prefix or an error if
// something went wrong.
func (s *server) extractRoute(r *http.Request, prefix string) (string, error) {
	if r == nil {
		return "", fmt.Errorf("Cannot strip prefix \"%s\" from invalid route", prefix)
	}

	route := r.URL.String()

	if !strings.HasPrefix(route, prefix) {
		return "", fmt.Errorf("Cannot strip prefix \"%s\" from route \"%s\"", prefix, route)
	}

	return strings.TrimPrefix(route, prefix), nil
}
