package dispatcher

import (
	"fmt"
	"net/http"
	"oglike_server/pkg/logger"
)

// NotFound :
// Describe an empty `HTTP` handler which will only log a message
// through the provided logger whenever a request is received on
// the associated route.
//
// The `log` represents the logger object to use to notify of any
// connexion request on this endpoint.
//
// Returns a callable function that will log a message and return
// a `404` code in case of an incoming connection.
func NotFound(log logger.Logger) http.HandlerFunc {
	// The return value is a callable `HTTP` handler.
	return func(w http.ResponseWriter, r *http.Request) {
		// Notify from this connection.
		log.Trace(logger.Warning, getModuleName(), fmt.Sprintf("Handling request from \"%v\" in not found handler", r.URL))

		// Resource not found is the only answer we can provide.
		http.NotFound(w, r)
	}
}

// NotAllowed :
// Describe an empty `HTTP` handler which will only log a message
// through the provided logger whenever a request is received on
// the associated route. It typically indicates that the method
// used to contact this endpoint is not supported for now.
//
// The `log` represents the logger object to use to notify of any
// connexion request on this endpoint.
//
// Returns a callable function that will log a message and return
// a `405` code in case of an incoming connection.
func NotAllowed(log logger.Logger) http.HandlerFunc {
	// The return value is a callable `HTTP` handler.
	return func(w http.ResponseWriter, r *http.Request) {
		// Notify from this connection.
		log.Trace(logger.Warning, getModuleName(), fmt.Sprintf("Handling request from \"%v\" in not allowed handler", r.URL))

		// Method not allowed is the answer we will provide.
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// NoOp :
// Describe an empty `HTTP` handler which will only log a message
// through the provided logger whenever a request is received on
// the associated route.
// The return code will indicate that the request was successful
// but nothing really happened.
//
// The `log` represents the logger object to use to notify of any
// connexion request on this endpoint.
//
// Returns a callable function that will log a message and return
// a `200` code in case of an incoming connection.
func NoOp(log logger.Logger) http.HandlerFunc {
	// The return value is a callable `HTTP` handler.
	return func(w http.ResponseWriter, r *http.Request) {
		// Notify from this connection.
		log.Trace(logger.Warning, getModuleName(), fmt.Sprintf("Handling request from \"%v\" in no op handler", r.URL))

		// The cleanup code of the `HandlerFunc` will automatically
		// write the headers to indicate a success so we don't have
		// to do anything here.
	}
}

// WithSafetyNet :
// Wrap the call to the `HTTP` handler func in argument with an
// error handling mechanism which will recover from any panic
// issued by the handler. It will answer with an internal error
// code to the client, indicating that something went wrong and
// log a message to indicate the failure.
//
// The `log` represents the logger object to use to notify of
// any failure so that it is not lost and can be analyzed later.
//
// The `next` represents the `HTTP` handler which execution is
// to be wrapped with some error protection mechanism.
//
// Returns a callable function that will execute the `next`
// handler when called.
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
					log.Trace(logger.Error, getModuleName(), fmt.Sprintf("Recovering from unexpected panic (err: %v)", err))

					http.Error(w, "Unexpected error while processing request", http.StatusInternalServerError)
				}
			}()

			// Execute the input `HTTP` handler.
			next.ServeHTTP(w, r)
		}()
	}
}
