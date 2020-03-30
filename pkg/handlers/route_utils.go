package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// InternalServerErrorString :
// Used to provide a unique string that can be used in case an
// error occurs while serving a client request and we need to
// provide an answer.
//
// Returns a common string to indicate an error.
func InternalServerErrorString() string {
	return "Unexpected server error"
}

// sanitizeRoute :
// Used to remove any '/' characters leading or trailing the
// input route string.*
//
// The `route` is the string to be sanitized.
//
// A string stripped from any leading or trailing '/' items.
func sanitizeRoute(route string) string {
	if strings.HasPrefix(route, "/") {
		route = strings.TrimPrefix(route, "/")
	}
	if strings.HasSuffix(route, "/") {
		route = strings.TrimSuffix(route, "/")
	}

	return route
}

// splitRoutElements :
// Used to transform part of the route into its composing single
// elements. Typically a value of `/universes/oberon` will be
// split into `universes` and `oberon`. Any '/' character will
// be stripped from the input string and used as a separator for
// tokens in the route.
// In case no '/' character is found the output array will have
// a length of `1` element representing the input string.
//
// The `route` is the element which should be split on the '/'
// characters.
//
// Returns an array of all tokens formed by the '/' characters
// in the string.
func splitRouteElements(route string) []string {
	// Remove prefix for the route and suffix.
	if strings.HasPrefix(route, "/") {
		route = strings.TrimPrefix(route, "/")
	}
	if strings.HasSuffix(route, "/") {
		route = strings.TrimSuffix(route, "/")
	}

	// Handle for empty string.
	if len(route) == 0 {
		return make([]string, 0)
	}

	// Split on '/' characters.
	return strings.Split(route, "/")
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

// extractRouteVars :
// This facet of the server allows to conveniently extract the information
// available in the route used to contact the server. Using the input route
// it will try to detect the query parameters defined for this route along
// with information about the actual extra path that may have been provided
// in the input route.
// In case the route used to contact the server does not start with the input
// `route` value an error is returned.
//
// The `route` represents the common route prefix that should be ignored to
// extract parameters. We will try to match this pattern in the route and
// then extract information after that.
//
// The `r` represents the request that should be parsed to extract query
// parameters.
//
// Returns a map containing the query parameters as defined in the route.
// The map may be empty but should not be `nil`. Also returns any error
// that might have been encountered. The returned map should not be used
// in case the error is not `nil`.
func extractRouteVars(route string, r *http.Request) (RouteVars, error) {
	vars := RouteVars{
		make([]string, 0),
		make(map[string]Values),
	}

	// Extract the route from the input request.
	extra, err := extractRoute(r, route)
	if err != nil {
		return vars, fmt.Errorf("Could not extract vars from route \"%s\" (err: %v)", route, err)
	}

	// The extra path for the route is specified until we reach a '?' character.
	// After that come the query parameters.
	beginQueryParams := strings.Index(extra, "?")
	if beginQueryParams < 0 {
		// No query parameters found for this request: the `extra` path defines
		// the extra route path.
		vars.RouteElems = splitRouteElements(extra)

		return vars, nil
	}

	// Extract query parameters and the route (which is basically the part of
	// the string before the beginning of the query params).
	vars.RouteElems = splitRouteElements(extra[:beginQueryParams])
	queryStr := extra[beginQueryParams+1:]

	params, err := url.ParseQuery(queryStr)
	if err != nil {
		return vars, fmt.Errorf("Unable to parse query parameters in route \"%s\" (err: %v)", route, err)
	}

	// Analyze the retrieved query parameters.
	for key, values := range params {
		// Make sure that at least a value exists for the key.
		if values == nil {
			vars.Params[key] = make([]string, 0)
		} else {
			vars.Params[key] = values
		}
	}

	return vars, nil
}

// extractRouteData :
// Used to extract the values passed to a request assuming its method
// allows it. We don't enforce very strictly to call this method only
// if the request can define such values but the result will be empty
// if this is not the case.
//
// The `dataKey` is used to form values from the request. The input
// request may define several data which might be parsed by different
// parts of the server. This method only extracts the ones that are
// relevant relatively to the provided key.
// It also fetches *all* the instances of the values matching the key.
//
// The `r` defines the request from which the route values should be
// extracted.
//
// Returns the route's data along with any errors.
func extractRouteData(dataKey string, r *http.Request) (RouteData, error) {
	elems := RouteData{
		make([]string, 0),
	}

	route := r.URL.String()

	// Fetch the data from the input request: as we want to allow
	// for multiple instances of the same key we need to call the
	// `ParseForm` method (as described in the documentation of
	// the `FormValue` method).
	err := r.ParseForm()
	if err != nil {
		return elems, fmt.Errorf("Could not parse data for key \"%s\" from route (err: %v)", route, err)
	}

	// Search for the relevant key.
	value, ok := r.Form[dataKey]
	if ok {
		elems.Data = append(elems.Data, value...)
	}

	return elems, nil
}
