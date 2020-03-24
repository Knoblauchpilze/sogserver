package internal

import (
	"net/http"
	"oglike_server/pkg/handlers"
	"strconv"
)

// routes :
// Used to setup all the routes able to be served by this server.
// All the routes are set up with the adequate handler but no
// actual binding is done.
func (s *server) routes() {
	// Existing universes.
	http.HandleFunc("/universes", handlers.NotFound(s.log))

	// Players accounts.
	http.HandleFunc("/accounts", handlers.NotFound(s.log))

	// Planets.
	http.HandleFunc("/planets", handlers.NotFound(s.log))
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
