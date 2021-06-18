package dispatcher

import (
	"fmt"
	"net/http"
	"oglike_server/pkg/logger"
	"regexp"
	"strings"
)

// Convenience define allowing to reference the possible
// matching state for a route. It is used to precisely
// determine the best match for an input requets.
type matching int

// Definition of the possible match state for a route.
const (
	methodNotAllowed matching = iota
	notFound
	matchedPartial
	matched
)

// Route :
// Defines a generic route which is a path that can be used
// to target a server. The route is for now composed of a
// string and a method, which allows to only react to some
// specific CRUDE behavior on a dedicated route, and also
// to serve multiple request types on a single endpoint.
// This works well with the REST paradigm where a endpoint
// is typically assigned with all the operations that can
// be performed on a collection.
// The route also defines a handler which is called in
// case a request is directed towards this route. This
// handler can bypass some of the verifications related
// to the route because it has already been handled by
// the route itself.
//
// The `methods` defines the HTTP verbs associated to this
// route. No request that doesn't match one of these verbs
// will be directed towards this route.
//
// The `elems` of the route defines the individual route
// elements that should be matched for a request to be
// targeting a route.
// Typically a request would target the `/path/to/route`
// path and the elements would contain `path`, `to` and
// `route`. Each one will be matched if possible which
// will in the end allow to match the route entirely.
// These paths will be converted into regular expressions
// so that we can also handle things like below:
// `/path/to/route/[a-z]+`.
//
// The `handler` defines the actual processing to call in
// case this route is triggered. It will be initialized
// to a default `NoOp` handler.
//
// The `log` will be used in case anything is requiring
// to notify the user of an error.
type Route struct {
	methods map[string]bool
	elems   []*regexp.Regexp
	handler http.Handler
	log     logger.Logger
}

// ErrRouteNotValid :
// Indicates that the expression provided to define a
// route is not valid.
var ErrRouteNotValid = fmt.Errorf("invalid expression provided for route")

// routeMatch :
// Stores the information about a matched route. Notably
// it indicates whether the route could be matched or not
// and some more info about how the route failed to match.
//
// The `handler` defines the actual handler that should be
// used to process the request. Should never be `nil` if
// a `NotFoundHandler` is provided by the router.
//
// The `match` allows to precisely determine which kind
// of matching was possible among all the routes that are
// managed by this router.
//
// The `length` defins the length that has been matched
// in the route's definition. It allows to provide some
// sort of measure of how good this route is at marching
// the input request in a quantitative way.
// For now the length does not actually means the number
// of characters matched but the number of segments that
// have been matched. This allows to abstract away the
// actual length of the route elements and focus on the
// number of tokens that have been matched.
type routeMatch struct {
	handler http.Handler
	match   matching
	length  int
}

// buildRouteElements :
// Used to separate the input route in a set of regular
// expressions that will be traversed sequentially when
// performing the matching.
//
// The `route` defines the input route to analyze. The
// route will be split up on '/' character and each of
// the token will be transformed into a regexp where a
// special `^...$` part is added to make sure that the
// regexp only matches for the full token (and not a
// part of it).
//
// Returns an array of regular expressions describing
// the input route in order to allow easy matching for
// the route along with any error.
func buildRouteElements(route string) ([]*regexp.Regexp, error) {
	// Remove the first and last '/' characters from the
	// input route if any. Do it unconditionnally as the
	// functions are handling case where the string does
	// not contain the element.
	route = strings.TrimPrefix(route, "/")
	route = strings.TrimSuffix(route, "/")

	// Make sure that the route is not empty. If this
	// is the case we will return an empty array of
	// regexp. This will be consistent with what is
	// expected when matching the route.
	if route == "" {
		return []*regexp.Regexp{}, nil
	}

	// Split the route on '/' characters and build the
	// list of regexp representing them.
	tokens := strings.Split(route, "/")
	elems := make([]*regexp.Regexp, 0)

	// For each token, convert it into a regexp. We
	// will also add the `^...$` statements to each
	// one of them.
	for _, token := range tokens {
		// Format the string to include the additional
		// control flow operators if it's not already
		// appended to the token.
		str := token
		if !strings.HasPrefix(str, "^") {
			str = fmt.Sprintf("^%s", str)
		}
		if !strings.HasSuffix(str, "$") {
			str = fmt.Sprintf("%s$", str)
		}

		// Try to convert this token into a valid
		// regular expression to use.
		exp, err := regexp.Compile(str)

		if err != nil {
			return elems, ErrRouteNotValid
		}

		elems = append(elems, exp)
	}

	return elems, nil
}

// NewRoute :
// Used to create a new route with no associated methods
// and the sepcified path. In case the path is empty, the
// route is still created.
// Note that if the route contains an invalid element
// that cannot be converted to a regular expression a
// panic will be issued.
//
// The `path` indicates the path that is associated to the
// route to create. It will be used by the route to make
// sure that only requests intended for a route are served
// to it.
//
// The `log` is used to create the default `NoOp` handler
// associated to this route.
//
// Returns the created route.
func NewRoute(path string, log logger.Logger) *Route {
	// Transform the route into regexp elements that can be
	// parsed and used to match the requests.
	tokens, err := buildRouteElements(path)
	if err != nil {
		log.Trace(logger.Error, "route", fmt.Sprintf("Unable to create route tokens for \"%s\" (err: %v)", path, err))

		panic(ErrRouteNotValid)
	}

	return &Route{
		methods: make(map[string]bool),
		elems:   tokens,
		handler: http.Handler(NoOp(log)),
		log:     log,
	}
}

// Handler :
// Returns the handler associated to this route. Should
// never be `nil`.
//
// Returns the processing handler for this route.
func (r *Route) Handler() http.Handler {
	return r.handler
}

// Methods :
// Register the set of methods provided in in put as valid
// methods to reach this route. Note that in case the method
// already exists, nothing happen.
// Note that the input methods are transformed into upper
// case verbs internally (so it's not mandatory to do so
// beforehand).
//
// The `methods` define the new methods to register as valid
// for this route.
//
// Returns a reference to this route which is interesting to
// chain calls on this route.
func (r *Route) Methods(methods ...string) *Route {
	// Traverse the input list of methods and register each
	// one of them internally. We want to perform a filter
	// of the input methods so as not to register anything.
	filtered := filterMethods(methods, r.log)

	for method := range filtered {
		r.methods[method] = true
	}

	return r
}

// HandlerFunc :
// Register the provided handler func as the main processing
// function for this route. It will be called whenever the
// route is actually executed.
//
// The `f` argument defines the processing unit to attach to
// the route.
//
// Returns this route, so that we can chain call.
func (r *Route) HandlerFunc(f func(http.ResponseWriter, *http.Request)) *Route {
	// Wrap the provided handler func into a valid handler.
	r.handler = http.HandlerFunc(f)

	return r
}

// Match :
// Used to verify whether this route can match the input
// request. It will check whether the path of the route
// corresponds to the path of the request and also perform
// a verification of the method of the request.
//
// The `req` represents the input request to match on this
// route.
//
// Returns the matching state for this route. Can be one
// of the available type which helps describe precisely
// how the request could be matched against this route.
func (r *Route) match(req *http.Request) routeMatch {
	// Check whether the path at least starts correctly to
	// be registered in the route.
	path := req.URL.String()

	// We need to strip the query parameters from the route.
	id := strings.Index(path, "?")
	if id >= 0 {
		path = path[:id]
	}

	m := routeMatch{}
	m.length = r.matchName(path)

	if m.length == 0 {
		// The route does not match the path of the request,
		// it cannot be matched.
		m.match = notFound

		return m
	}

	// Check the method of the request.
	_, ok := r.methods[req.Method]
	if !ok {
		// The method does not match the type requested by
		// the route, it cannot be matched.
		m.match = methodNotAllowed

		return m
	}

	// The route seems to match the input request. We will
	// either declare a match or a partial match depending
	// on the length matched.
	m.match = matchedPartial
	if m.length == len(r.elems) {
		m.match = matched
	}

	m.handler = r.handler

	return m
}

// mathcName :
// Used to determine whether the input `uri` can be used
// to match the route name. This method takes care of the
// processing needed to make sure that the `uri` not only
// defines the same path as the route but also that it is
// consistent with the route syntax.
// Typically we will try to prevent matching of cases as
// described below:
//  -route: `/path/to/route`
//  -uri  : `/path/to/routeeeee`
//
// The `uri` represents the string to match to the name
// of the route.
//
// Returns an integer representing the number of single
// characters from the input route that have been matched
// by this route. It provides some measure of how good of
// a fit this route is to the `uri`.
func (r *Route) matchName(uri string) int {
	// We want to match as much of the route as possible
	// given the elements that compose it. We need to first
	// analyze the input `uri` to extract individual tokens
	// that can then be matched against the route.
	// We will also prevent matching of empty routes and
	// other weird cases right away.

	// Sanitize the input `uri`. As the `TrimSuff/Prefix`
	// are handling the case where the prefix does not
	// exist we do it unconditionnally.
	uri = strings.TrimPrefix(uri, "/")
	uri = strings.TrimSuffix(uri, "/")

	if uri == "" {
		// We only match if there are no elements to match
		// in this route.
		if len(r.elems) == 0 {
			return 1
		}

		return 0
	}

	// Convert it to tokens.
	tokens := strings.Split(uri, "/")

	// Try to match each token of the route. To do so we
	// obviously need at least as many tokens in the input
	// `uri` as defined in the route.
	// Typically imagine the following situation:
	// - `route 1: /path/to/route/1`
	// - `route 2: /path/to/route/1/and/more`
	// - `route 3: /another/path/to/some/other/route`
	// - `uri    : /path/to/route/1/and`
	//
	// In this case we want the `route 1` to be selected
	// and not the `route 2`: indeed even though `route 2`
	// would have a longer length matched, it would not
	// be sufficient to describe the route entirely so it
	// is not considered a good match.
	// Same goes for the `route 3` which in addition to
	// not being a good match has a greater length than
	// the `uri`.
	// This is nice because it also means that we spare
	// some comparison right away by checking that there
	// are at least as many tokens in the `uri` than in
	// the `route` itself.
	if len(r.elems) > len(tokens) {
		return 0
	}

	length := 0
	matched := true

	// We know for sure that `len(r.elems) <= len(tokens)`
	// so it is safe to access elements this way.
	for id := 0; id < len(r.elems) && matched; id++ {
		matched = r.elems[id].Match([]byte(tokens[id]))

		if matched {
			length++
		}
	}

	return length
}
