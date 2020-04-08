package dispatcher

import (
	"net/http"
	"oglike_server/pkg/logger"
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
// The `name` of the route defines the actual endpoint to
// target to reach the route. We consider only absolute
// path and perform no cleaning upon creating the route.
// One can however sanitize the route through the method
// of the same name so that it at least starts with a '/'
// character.
//
// The `handler` defines the actual processing to call in
// case this route is triggered. It will be initialized
// to a default `NoOp` handler.
//
// The `log` will be used in case anything is requiring
// to notify the user of an error.
type Route struct {
	methods map[string]bool
	name    string
	handler http.Handler
	log     logger.Logger
}

// NewRoute :
// Used to create a new route with no associated methods
// and the sepcified path. In case the path is empty, the
// route is still created.
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
	return &Route{
		methods: make(map[string]bool, 0),
		name:    path,
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
func (r *Route) match(req *http.Request) matching {
	// Check whether the path at least starts correctly to
	// be registered in the route.
	path := req.URL.String()

	if !strings.HasPrefix(path, r.name) {
		// The route does not match the path of the request,
		// it cannot be matched.
		return notFound
	}

	// Check the method of the request.
	_, ok := r.methods[req.Method]
	if !ok {
		// The method does not match the type requested by
		// the route, it cannot be matched.
		return methodNotAllowed
	}

	// The route seems to match the input request.
	return matched
}
