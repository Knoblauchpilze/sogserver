package routes

import (
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strconv"
	"strings"
)

// server :
// Defines a server that can be used to handle the interaction with
// the OG database. This server handles can be built from the input
// database and logger and will perform the listening to handle the
// clients' requests.
// This article helped a bit to set up and describe the data model
// and structures used to describe the server:
// https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years
//
// The `port` allows to determine which port should be used by the
// server to accept incoming requests. This is usually specified in
// the configuration so as not to conflict with any other API.
//
// The `universes` represents a proxy object allowing to interact
// and retrieve properties of universes from the main DB. It is used
// as a way to hide the complexity of the DB and only use high-level
// functions that do not rely on the internal schema of the DB to
// work.
//
// The `accounts` fills a similar role to `universes` but is related
// to accounts information.
//
// The `logger` allows to perform most of the logging on any action
// done by the server such as logging clients' connections, errors
// and generally some elements useful to track the activity of the
// server.
type server struct {
	port      int
	universes data.UniverseProxy
	accounts  data.AccountProxy
	log       logger.Logger
}

// routeVars :
// Define common information to be passed in the route to contact
// the server. We handle extra path that can be added to the route
// (typically to refine the behavior expected from the base route)
// and some query parameters.
// These values can be extracted from the input request through the
// `extractRouteVars` method on the server.
//
// The `path` represents the extra path added to the route that
// was provided to target the server. Typically if the server is
// able to serve the `/universes` route, a request handled on the
// `/universes/oberon` path will have `/oberon` in the `path`.
//
// The `params` define the query parameters associated to the input
// request. Note that in some case no parameters are provided.
type routeVars struct {
	path   string
	params map[string]string
}

// NewServer :
// Create a new server with the input elements to use internally to
// access data and perform logging.
// In case any of the arguments are not valid a panic is issued to
// indicate the failure.
//
// The `port` defines the port to listen to by the server.
//
// The `dbase` represents a pointer to the database to use to fetch
// data when needed to answer clients' requests.
//
// The `log` is used to notify from various processes in the server
// and keep track of the activity.
func NewServer(port int, dbase *db.DB, log logger.Logger) server {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create server from empty database"))
	}

	return server{
		port,
		data.NewUniverseProxy(dbase, log),
		data.NewAccountProxy(dbase, log),
		log,
	}
}

// Serve :
// Used to start listening to the port associated to this server
// and handle incoming requests. This will return an error in case
// something went wrong while listening to the port.
func (s *server) Serve() error {
	// Setup routes.
	s.routes()

	// Serve the root path.
	return http.ListenAndServe(":"+strconv.FormatInt(int64(s.port), 10), nil)
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
func (s *server) extractRouteVars(route string, r *http.Request) (routeVars, error) {
	vars := routeVars{
		"",
		make(map[string]string),
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
		vars.path = extra

		return vars, nil
	}

	// Extract query parameters and the route (which is basically the part of
	// the string before the beginning of the query params).
	vars.path = extra[:beginQueryParams]
	queryStr := extra[beginQueryParams+1:]

	// Query parameters should be separated by '&' characters. Each parameter is
	// separated between a key and a value through the '=' character.
	params := strings.Split(queryStr, "&")
	for _, param := range params {
		// Split the parameter into its key and value component if possible.
		tokens := strings.Split(param, "=")

		// Discard invalid tokens.
		if len(tokens) != 2 {
			s.log.Trace(logger.Warning, fmt.Sprintf("Unable to parse query parameter \"%s\" in route \"%s\"", param, route))
			continue
		}

		// Detect override.
		existing, ok := vars.params[tokens[0]]
		if ok {
			s.log.Trace(logger.Warning, fmt.Sprintf("Overriding query parameter \"%s\": \"%s\" replaced by \"%s\"", tokens[0], existing, tokens[1]))
		}

		vars.params[tokens[0]] = tokens[1]
	}

	return vars, nil
}
