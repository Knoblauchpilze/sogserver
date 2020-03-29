package routes

import (
	"fmt"
	"net/http"
	"net/url"
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
// The `buildings` represents the proxy to use to perform requests
// concerning buildings and to access information about this topic.
//
// The `technologies` fills a similar purpose as `buildings` but
// for technologies related requests.
//
// The `ships` fills a similar purpose as `buildings` but for ships
// related requests.
//
// The `defenses` fills a similar purpose as `buildings` but for
// defenses related requests.
//
// The `logger` allows to perform most of the logging on any action
// done by the server such as logging clients' connections, errors
// and generally some elements useful to track the activity of the
// server.
type server struct {
	port         int
	universes    data.UniverseProxy
	accounts     data.AccountProxy
	buildings    data.BuildingProxy
	technologies data.TechnologyProxy
	ships        data.ShipProxy
	defenses     data.DefenseProxy
	log          logger.Logger
}

// Values :
// A convenience define to allow for easy manipulation of a list
// of strings as a single element. This is mostly used to be able
// to interpret multiple values for a single query parameter in
// an easy way.
type Values []string

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
	params map[string]Values
}

// routeData :
// Used to define the data that can be passed to a route based on
// its name. It is constructed by appending the "-data" suffix to
// the route and getting the corresponding value.
// It is useful to group common behavior of all the interfaces on
// this server.
//
// The `value` represents the data extracted from ``
type routeData struct {
	value string
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
		data.NewBuildingProxy(dbase, log),
		data.NewTechnologyProxy(dbase, log),
		data.NewShipProxy(dbase, log),
		data.NewDefenseProxy(dbase, log),
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
		vars.path = extra

		return vars, nil
	}

	// Extract query parameters and the route (which is basically the part of
	// the string before the beginning of the query params).
	vars.path = extra[:beginQueryParams]
	queryStr := extra[beginQueryParams+1:]

	params, err := url.ParseQuery(queryStr)
	if err != nil {
		return vars, fmt.Errorf("Unable to parse query parameters in route \"%s\" (err: %v)", route, err)
	}

	// Analyze the retrieved query parameters.
	for key, values := range params {
		// Handle cases where several parameter are provided for the same key:
		// we will only keep the first one for now. We will also handle cases
		// where no value is provided for the key.
		if len(values) == 0 {
			s.log.Trace(logger.Warning, fmt.Sprintf("Key \"%s\" does not have any value in route \"%s\"", key, route))
			continue
		}

		vars.params[key] = values
	}

	return vars, nil
}

// extractRouteData :
// Used to extract the values passed to a request assuming its method
// allows it. We don't enforce very strictly to call this method only
// if the request can define such values but the result will be empty
// if this is not the case.
// The key used to form values from the request is directly built in
// association with the route's name: basically if the route is set
// to `/path/to/route` the key will be "route-data" that is the last
// part of the route where we added a "-data" string.
//
// The `r` defines the request from which the route values should be
// extracted.
//
// Returns the route's data along with any errors.
func (s *server) extractRouteData(r *http.Request) (routeData, error) {
	data := routeData{
		"",
	}

	route := r.URL.String()

	// Keep only the last part of the route (i.e. the data that is
	// after the last occurrence of the '/' character). We assume
	// that the data is registered under the route's name unless we
	// can find some information indicating that it can be refined.
	cutID := strings.LastIndex(route, "/")

	dataKey := route + "-data"
	if cutID >= 0 && cutID < len(route)-1 {
		dataKey = route[cutID+1:] + "-data"
	}

	// Fetch the data from the input request.
	data.value = r.FormValue(dataKey)

	return data, nil
}
