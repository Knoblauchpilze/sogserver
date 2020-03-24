package handlers

import (
	"fmt"
	"net/http"
	"oglike_server/pkg/logger"
)

// WithSafetyNet :
// Wrap the call to the `HTTP` handler func in argument with an error handling
// mechanism which will recover from any panic issued by the handler. It will
// answer with an internal error code to the client, indicating that something
// went wrong and log a message to indicate the failure.
//
// The `log` represents the logger object to use to notify of any failure so
// that it is not lost and can be analyzed later.
//
// The `next` represents the `HTTP` handler which execution should be wrapped
// with some error protection mechanism.
//
// Returns a callable function that will execute the `next` handler when called.
func WithSafetyNet(log logger.Logger, next http.HandlerFunc) http.HandlerFunc {
	// The return value is a callable `HTTP` handler.
	return func(w http.ResponseWriter, r *http.Request) {

		// Decorate the `next` handler with a protection mechanism.
		func() {
			// Recover from any leaking panic.
			defer func() {
				err := recover()

				if err != nil {
					// Log the error and answer with an internal server error.
					log.Trace(logger.Error, fmt.Sprintf("Recovering from unexpected panic (err: %v)", err))

					http.Error(w, "Unexpected error while processing request", http.StatusInternalServerError)
				}
			}()

			// Execute the input `HTTP` handler.
			next.ServeHTTP(w, r)
		}()
	}
}
