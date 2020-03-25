package handlers

import (
	"fmt"
	"net/http"
	"oglike_server/pkg/logger"
	"strings"
)

// getSupportedMethods :
// Returns the list of `HTTP` verbs that can be used as valid filtering
// methods for such a handler.
func getSupportedMethods() map[string]bool {
	return map[string]bool{
		"GET":     true,
		"HEAD":    true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"CONNECT": true,
		"OPTIONS": true,
		"TRACE":   true,
		"PATCH":   true,
	}
}

// filterMethods :
func filterMethods(methods []string, log logger.Logger) map[string]bool {
	filtered := make(map[string]bool, 0)
	supported := getSupportedMethods()

	for _, method := range methods {
		consolidated := strings.ToUpper(method)
		_, ok := supported[consolidated]

		// Filter invalid methods.
		if !ok {
			log.Trace(logger.Error, fmt.Sprintf("Filtering invalid HTTP method \"%s\"", method))
			continue
		}

		filtered[consolidated] = true
	}

	return filtered
}

// Methods :
// Describe a `HTTP` handler which filters only certain methods to associate
// to the route. Any request not matching one of the registered route will be
// served a `404` with no data.
//
// The `log` represents the logger object to use to notify of any bad connexion
// request on this endpoint.
//
// The `methods` allow to filter only the type of interaction for the endpoint.
//
// The `next` handler will be called only if the input request defines a method
// that is consistent with the available types.
//
// Returns a callable function that will filter requests based on the specified
// set of methods that can be handled by the `next` handler.
func Methods(log logger.Logger, methods []string, next http.HandlerFunc) http.HandlerFunc {
	// Filter the input methods.
	filtered := filterMethods(methods, log)

	// The return value is a callable `HTTP` handler.
	return func(w http.ResponseWriter, r *http.Request) {
		// Check whether the input request respects the defined methods.
		_, ok := filtered[r.Method]
		if !ok {
			// Can't accept this method for connecting to the handler.
			log.Trace(logger.Error, fmt.Sprintf("Discarding request with wrong method \"%s\" on \"%s\"", r.Method, r.URL.String()))

			http.NotFound(w, r)
		}

		// Propagate to the next handler.
		next.ServeHTTP(w, r)
	}
}

// Method :
// Convenience wrapper that takes a single route in argument and handle on its
// own the call to the `Methods` with the correct array of strings.
//
// The `log` is the logger to be forwarded to `Methods`.
//
// The `method` is the single method that is allowed to contect the `next` one.
//
// The `next` represents the handler for which methods should be filtered.
//
// Returns a callable `HTTP` handle that can be used to wrap calls to the `next`
// handler and only transmit requests with a suited method.
func Method(log logger.Logger, method string, next http.HandlerFunc) http.HandlerFunc {
	return Methods(log, []string{method}, next)
}
