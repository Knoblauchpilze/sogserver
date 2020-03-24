package handlers

import (
	"fmt"
	"net/http"
	"oglike_server/pkg/logger"
)

// NoOpHandler :
// Describe an empty `HTTP` handler which will only log a message through the
// provided logger whenever a request is received on the associated route. It
// can be used as a way to handle elements not yet implemented by the server
// as it serves a `404` status code with no data.
//
// The `log` represents the logger object to use to notify of any connexion
// request on this endpoint.
//
// Returns a callable function that will log a message and return a `200` code
// in case of an incoming connection.
func NotFound(log logger.Logger) http.HandlerFunc {
	// The return value is a callable `HTTP` handler.
	return func(w http.ResponseWriter, r *http.Request) {
		// Notify from this connection.
		log.Trace(logger.Warning, fmt.Sprintf("Handling request from \"%v\" in no op handler", r.URL))

		// Resource not found is the only answer we can provide.
		http.NotFound(w, r)
	}
}
