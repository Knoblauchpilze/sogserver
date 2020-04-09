package dispatcher

import (
	"net/http"
	"oglike_server/pkg/logger"
)

// Router :
// Defines a generic router that can be used to simplify the
// handling of multiple routes for a server. It helps with
// the organization of the routes by providing some means to
// register routes with a specific name and method.
//
// The `log` allows to nofity errors and information to the
// user in case a wrong route is called for example.
//
// The `notFoundHandler` defines the handler to use in case
// no route can be matched for a request. The defautl value
// is using the default object defined by this package that
// just prints an error message indicating the route that
// was accessed. It can be useful when coupled with a logs
// system to analyze the failures.
//
// The `methodNotAllowedHandler` defines a handler that is
// called whenver a route is matched for a request but the
// method does not correspond to the defined route. This is
// also provided with a default handler which indicates the
// failure.
//
// The `routes` register all the routes defined for this
// router to handle so far. It basically is used when a
// new request is received to route it to the element that
// best matches the pathes defined by the routes.
//
// The `log` allows to notify the user of information and
// various errors that can be produced by this element.
type Router struct {
	notFoundHandler         http.Handler
	methodNotAllowedHandler http.Handler
	routes                  []*Route
	log                     logger.Logger
}

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
// TODO: Add a mechanism for the matching length. This
// probably involves some sort of regexp matching.
type routeMatch struct {
	handler http.Handler
	match   matching
}

// NewRouter :
// Creates a new router with default handlers for not found
// and method not allowed and no route to match.
//
// The `log` will be passed on to the routes handled by this
// router in order to allow notification of the user when a
// route has trouble being routed.
//
// Returns the created router.
func NewRouter(log logger.Logger) *Router {
	return &Router{
		notFoundHandler:         NotFound(log),
		methodNotAllowedHandler: NotAllowed(log),
		routes:                  make([]*Route, 0),
		log:                     log,
	}
}

// addRoute :
// Registers a new empty route in this router. It will not be
// associated to any method and will have the specified path
// which is mandatory.
// If the provided path is empty, the route will be associated
// to the '/' path.
//
// Returns the created route.
func (r *Router) addRoute(path string) *Route {
	// Sanitize the path in case it is empty.
	if len(path) == 0 {
		path = "/"
	}

	// Create a new route.
	route := NewRoute(path, r.log)

	// Register it internally.
	r.routes = append(r.routes, route)

	// Return the route to allow chain calling.
	return route
}

// HandleFunc :
// Registers a new route in the internal list of served routes
// with the provided path and associated handler.
// Note that the route will still be registered in case another
// route with a similar path is available.
//
// The `path` defines the path to access to the route. It is
// transformed (in case it is empty) to a "/" default path.
//
// The `handler` defines the processing unit associated to the
// route.
//
// Returns the created route.
func (r *Router) HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *Route {
	return r.addRoute(path).HandlerFunc(f)
}

// ServeHTTP :
// Used to dispatch the input request to the best suited
// handler as registered in the internal routes. If none
// of the handlers are able to receive the request the
// `NotFound` handler will be called.
//
// The `w` represent the response writer to use in case
// some data should be returned back to the client.
//
// The `req` defines the input request which should be
// routed through the internal handlers.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Try to match the input request against the internal
	// registered routes.
	var match routeMatch
	matched := r.Match(req, &match)

	// In case we didn't match anything, use the `NotFound`
	// handler.
	if !matched {
		r.notFoundHandler.ServeHTTP(w, req)
		return
	}

	// Otherwise, use the matched handler and server the
	// response.
	match.handler.ServeHTTP(w, req)
}

// Match attempts to match the given request against the
// router's registered routes. It returns whether or not
// a route could be matched and the actual route if any.
//
// The `req` defines the input request to match against
// the internal routes.
//
// The `m` will be populated with the best matching route
// (if any). Note that in case no registered routes can
// be matched, the `NotFound` handler is returned as this
// value. In case the route could be matched but the method
// was not valid, the `NotAllowed` handler is returned.
//
// Returns `true` in case a route could be matched and
// `false` otherwise.
func (r *Router) Match(req *http.Request, m *routeMatch) bool {
	// Traverse the internal list of routes and check for
	// a match.
	for _, route := range r.routes {
		m.match = route.match(req)

		if m.match == matched {
			// Select this route.
			m.handler = route.Handler()
			return true
		}
	}

	// The route could not be matched. Check whether we could
	// match a route but the method was wrong, in which case
	// we will select the `NotAllowed` handler.
	if m.match == methodNotAllowed {
		// We assume that we always have a method not allowed
		// handler as we create it when building this router.
		m.handler = r.methodNotAllowedHandler
		return true
	}

	// We could not match anything, but we can rely on the
	// not found handler which is always defined for this
	// router.
	m.match = notFound
	m.handler = r.notFoundHandler

	return true
}
