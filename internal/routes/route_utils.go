package routes

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Values :
// A convenience define to allow for easy manipulation of a list
// of strings as a single element. This is mostly used to be able
// to interpret multiple values for a single query parameter in
// an easy way.
type Values []string

// RouteVars :
// Define common information to be passed in the route to contact
// the server. We handle extra path that can be added to the route
// (typically to refine the behavior expected from the base route)
// and some query parameters.
// An object of this type is extracted for each single request to
// put some sort of format when contacting the server. This is
// then passed to the underlying interface implementation to set
// and generate some filters from these variables. These values
// can be extracted from the input request through calling the
// `ExtractRouteVars` method.
//
// The `RouteElems` represents the extra path added to the route
// as it was provided to target the server. Typically if the server
// receives a request on `/universes`, the `RouteElems` will be set
// to the empty slice. On the other hand `/universes/oberon` will
// yield a single element `oberon` in the `RouteElems` slice.
//
// The `ExtraElems` represent some more route elements which are
// not part of the route as defined by the user. Typically in the
// case of a route like `/path/to/collection`, if the input req
// is actually defined by `/path/to/collection/res-id`, the route
// elems will contain "path", "to" and "collection" while the
// extra elems will contain the "res-id".
//
// The `Params` define the query parameters associated to the input
// request. Note that in some case no parameters are provided.
type RouteVars struct {
	RouteElems []string
	ExtraElems []string
	Params     map[string]Values
}

// RouteData :
// Used to define the data that can be passed to a route based on
// its name. It is constructed by appending the "-data" suffix to
// the route and getting the corresponding value.
// It is useful to group common behavior of all the interfaces on
// this server.
//
// The `RouteElems` represents the extra path added to the route
// as it was provided to target the server. Typically if the server
// receives a request on `/universe`, the `RouteElems` will be set
// to the empty slice. On the other hand `/universe/oberon` will
// yield a single element `oberon` in the `RouteElems` slice.
//
// The `ExtraElems` represent some more route elements which are
// not part of the route as defined by the user. Typically in the
// case of a route like `/path/to/collection`, if the input req
// is actually defined by `/path/to/collection/res-id`, the route
// elems will contain "path", "to" and "collection" while the
// extra elems will contain the "res-id".
//
// The `Data` represents the data extracted from the route itself.
// It is represented as an array of raw strings which are usually
// unmarshalled into meaningful structures by the data creation
// process.
type RouteData struct {
	RouteElems []string
	ExtraElems []string
	Data       Values
}

// ErrInvalidRequest :
// Used to indicate that the request provided in input is not
// valid (typically `nil`).
var ErrInvalidRequest = fmt.Errorf("Invalid request provided in input")

// ErrInvalidRoute :
// Used to indicate that the route provided in input is not
// consistent with the expected data. Typically it might be
// used to indicate that the route does not start with the
// adequate prefix.
var ErrInvalidRoute = fmt.Errorf("Invalid route provided in input")

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
// input route string.
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
		return "", ErrInvalidRequest
	}

	route := r.URL.String()

	if !strings.HasPrefix(route, prefix) {
		return "", ErrInvalidRoute
	}

	// The route itself is actually the value we want to return.
	return prefix, nil
}

// extractExtraPath :
// Convenience method allowing to strip the provided prefix to keep
// only the extra information provided by the route. Typically in
// the case of a request like `/path/to/collection/res-id` if the
// route is defined as `/path/to/collection`, the extra path will
// be assigned to `res-id`.
// Note that we also manage the case where some query parameters
// are provided in the request. We don't include them in the extra
// path.
//
// The `req` defines the input request as it was used to contact
// the server: it should be composed of the route, some extra
// (optional) pathes and some (optional) query parameters.
//
// The `route` defines the known prefix that should be right at
// the start of the `req` string, excluding the additional pathes.
//
// Returns the list of extra elements added to the `req` when
// compared to the `route` and the query parameters string.
func extractExtraPath(r *http.Request, route string) (string, error) {
	if r == nil {
		return "", ErrInvalidRequest
	}

	req := r.URL.String()

	if !strings.HasPrefix(req, route) {
		return "", ErrInvalidRoute
	}

	// Return the request without the `route` prefix (note: this will
	// still include any query parameters).
	return strings.TrimPrefix(req, route), nil
}

// tokenizeRoute :
// Used to extract the meaningful components of the route provided
// in input and separate them on the '/' character. We will also
// ignore the query parameters that might be existing at the end of
// the route.
// The second string corresponds to the query parameters stripped
// from their '?' leading character if any.
//
// The `route` defines a string that might contain query parameters
// and '/' characters which will be used to tokenize it.
//
// Returns the list of tokens extracted from the route.
func tokenizeRoute(route string) ([]string, string) {
	// The extra path for the route is specified until we reach a '?' character.
	// After that come the query parameters.
	beginQueryParams := strings.Index(route, "?")
	if beginQueryParams < 0 {
		// No query parameters found for this request: the `route` path defines
		// the extra route path.
		return splitRouteElements(route), ""
	}

	return splitRouteElements(route[:beginQueryParams]), route[beginQueryParams+1:]
}

// ExtractRouteVars :
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
		make([]string, 0),
		make(map[string]Values),
	}

	// Extract the route from the input request.
	path, err := extractRoute(r, route)
	if err != nil {
		return vars, err
	}

	// The `path` should not contain any query parameters in case there are
	// some additional elements in the route. Even if the route does not
	// define such pathes, we will extract the query parameters with the
	// `extractExtraPath` method so we don't care about it right now.
	vars.RouteElems, _ = tokenizeRoute(path)

	// Also extract the addition elements provided in the route.
	extra, err := extractExtraPath(r, route)
	if err != nil {
		return vars, err
	}

	// The extra path for the route is specified until we reach a '?' character.
	// After that come the query parameters.
	var queryStr string
	vars.ExtraElems, queryStr = tokenizeRoute(extra)

	if len(queryStr) == 0 {
		// No query parameters found for this request: the `extra` path defines
		// the extra route path.
		return vars, nil
	}

	// Some query parameters are provided in the input route: analyze them.
	params, err := url.ParseQuery(queryStr)
	if err != nil {
		return vars, ErrInvalidRequest
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
// The `route` represents the common route prefix that should be ignored
// to extract parameters. We will try to match this pattern in the route
// and then extract information after that.
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
func extractRouteData(route string, dataKey string, r *http.Request) (RouteData, error) {
	elems := RouteData{
		make([]string, 0),
		make([]string, 0),
		make([]string, 0),
	}

	// Extract the route from the input request.
	path, err := extractRoute(r, route)
	if err != nil {
		return elems, err
	}

	// The `path` should not contain any query parameters in case there are
	// some additional elements in the route. Even if the route does not
	// define such pathes, we will extract the query parameters with the
	// `extractExtraPath` method so we don't care about it right now.
	elems.RouteElems, _ = tokenizeRoute(path)

	// Also extract the addition elements provided in the route.
	extra, err := extractExtraPath(r, route)
	if err != nil {
		return elems, err
	}

	// Fetch the extra elements of the route, but we don't care about the
	// query parameters for now.
	elems.ExtraElems, _ = tokenizeRoute(extra)

	// Fetch the data from the input request: as we want to allow
	// for multiple instances of the same key we need to call the
	// `ParseForm` method (as described in the documentation of
	// the `FormValue` method).
	err = r.ParseForm()
	if err != nil {
		return elems, ErrInvalidRequest
	}

	// Search for the relevant key.
	value, ok := r.Form[dataKey]
	if ok {
		elems.Data = append(elems.Data, value...)
	}

	return elems, nil
}
