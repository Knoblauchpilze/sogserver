package routes

import (
	"fmt"
	"net/http"
	"strings"
)

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
func extractRoute(r *http.Request, prefix string) (string, error) {
	if r == nil {
		return "", fmt.Errorf("Cannot strip prefix \"%s\" from invalid route", prefix)
	}

	route := r.URL.String()

	if !strings.HasPrefix(route, prefix) {
		return "", fmt.Errorf("Cannot strip prefix \"%s\" from route \"%s\"", prefix, route)
	}

	return strings.TrimPrefix(route, prefix), nil
}
