package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/game"
	"oglike_server/pkg/logger"
)

// creationFunc :
// Convenience define which allows to refer to the process
// to create data into the DB given some resources fetched
// from the request.a set of filters. It is used as a way
// to mutualize most of the parsing code allowing to fetch
// some data from the input request before proceeding to
// the insertion of the data into the DB.
//
// The `data` defines the actual data extracted from the
// route, including both the route itself (and its single
// elements) along with the data retrieved from a user
// defined key. Typically the key will represent a value
// like `universe-data` that the resource creator can
// look upon in the input request.
//
// The return value includes both any error and the list
// of resources that could be created from the input data
// so that it can be returned to the client as requested
// by the REST architecture.
type creationFunc func(data RouteData) ([]string, error)

// CreateResourceEndpoint :
// Defines the information to describe a endpoint which can be
// used to handle data creation. This interface allows some
// mutualization of the common extraction code which fetch the
// data from the request to format it into something that can
// easily be inserted in the DB.
//
// The `route` defines a string that should be contained in the
// input string: it helps separating the routes from the query
// parameters that can be provided to filter out some results in
// the output result.
//
// The `key` defines the data key to search for to fetch the
// data from the input request.
//
// The `creator` references the function to use to create the
// data into the DB once it has been retrieved from the input
// request and converted into something meaningful.
//
// The `module` defines a string that can be used to make the
// logs displayed more explicit by specifying this module's id.
// This string should be unique across the application and is
// used as a mean to easily distinguish between the different
// services composing the server.
//
// The `prefixRes` allows to determine whether the path that
// is returned to indicate the resources created by the func
// `creator` should be prefixed by the route's name. Note
// that the default value is `true`. One can disable this
// behavior by using the `WithoutPrefix` method. Note however
// that a '/' prefix is still added.
//
// The `lock` allows to define whether a locker should be
// applied when creating data into the DB. This lock will
// be automatically acquired before passing on the request
// to the `creator` function and release afterwards. If it
// is set to `nil` (default behavior) no lock is acquired.
type CreateResourceEndpoint struct {
	route     string
	key       string
	creator   creationFunc
	module    string
	prefixRes bool
	lock      *game.Instance
}

// ErrNoData :
// Used to indicate that there is not data provided to perform
// the creation of the resource.
var ErrNoData = fmt.Errorf("no data provided to perform resource creation")

// ErrInvalidData :
// Used as a generic error in case the data provided as input
// of a creation request could not be read.
var ErrInvalidData = fmt.Errorf("unable to parse input data")

// NewCreateResourceEndpoint :
// Creates a new empty endpoint description with the provided
// route. The fetcher func is defined as an empty element to
// avoid fetching anything and no filters are provided.
//
// The `route` will be sanitized and represent the path that
// is associated to this endpoint.
//
// Returns the created end point description.
func NewCreateResourceEndpoint(route string) *CreateResourceEndpoint {
	return &CreateResourceEndpoint{
		route:     sanitizeRoute(route),
		prefixRes: true,
	}
}

// WithDataKey :
// Defines that this endpoint should use the provided data
// key to fetch the data to insert in the DB.
//
// The `key` defines the string to use to fetch the data from
// the input request.
//
// Returns the endpoint to allow chain calling.
func (cre *CreateResourceEndpoint) WithDataKey(key string) *CreateResourceEndpoint {
	cre.key = key
	return cre
}

// WithCreationFunc :
// Assigns the input creation function as the main way to
// perform the insertion of data in the DB.
//
// The `f` represents the creation function that should be
// used by this endpoint to insert data in the DB.
//
// Returns this endpoint to allow chain calling.
func (cre *CreateResourceEndpoint) WithCreationFunc(f creationFunc) *CreateResourceEndpoint {
	cre.creator = f
	return cre
}

// WithModule :
// Assigns a new string as the module name for this object.
//
// The `module` defines the name of the module to assign to
// this object.
//
// Returns this endpoint to allow chain calling.
func (cre *CreateResourceEndpoint) WithModule(module string) *CreateResourceEndpoint {
	cre.module = module
	return cre
}

// WithLocker :
// Assigns a new locker element to use when fetching data.
//
// The `locker` defines the element to acquire before a
// request on DB data can be served.
//
// Returns this endpoint to allow chain calling.
func (cre *CreateResourceEndpoint) WithLocker(i game.Instance) *CreateResourceEndpoint {
	cre.lock = &i
	return cre
}

// WithoutPrefix :
// Used to define that the resources' paths created by the
// `creator` function should not be prefixed with the name
// of the route.
//
// Returns this endpoint to allow chain calling.
func (cre *CreateResourceEndpoint) WithoutPrefix() *CreateResourceEndpoint {
	cre.prefixRes = false
	return cre
}

// ServeRoute :
// Returns a handler using this endpoint description to be
// able to serve requests given the data present in this
// endpoint.
// Note that we don't actually start serving anything, we
// just create the necessary handler that can be used to
// do so.
//
// The `log` will be used to notify info and messages if
// needed.
//
// Returns the handler that can be executed to serve the
// requests defined by the data of this endpoint.
func (cre *CreateResourceEndpoint) ServeRoute(log logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("/%s", cre.route)

		// Extract data from the input request to perform the
		// creation of data in the DB.
		data, err := extractRouteData(route, cre.key, r)
		if err != nil {
			log.Trace(logger.Error, cre.module, fmt.Sprintf("Could not fetch data from request for route \"%s\" (err: %v)", cre.route, err))

			panic(err)
		}

		var resNames []string

		func() {
			if cre.lock != nil {
				cre.lock.Lock()
				defer cre.lock.Unlock()
			}

			resNames, err = cre.creator(data)
		}()

		if err != nil {
			log.Trace(logger.Error, cre.module, fmt.Sprintf("Could not create resource from route \"%s\" (err: %v)", cre.route, err))

			http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
			return
		}

		// We need to return a valid status code and the address of
		// the created resource, as described in the following post:
		// https://stackoverflow.com/questions/1829875/is-it-ok-by-rest-to-return-content-after-post
		// To do so we will transform the resources to include the
		// name of the route and then marshal everything in an array
		// that will be returned to the client.
		resources := make([]string, len(resNames))

		for id, resource := range resNames {
			res := fmt.Sprintf("/%s", resource)

			if cre.prefixRes {
				res = fmt.Sprintf("/%s/%s", cre.route, resource)
			}

			resources[id] = res
		}

		bts, err := json.Marshal(&resources)
		if err != nil {
			log.Trace(logger.Error, cre.module, fmt.Sprintf("Could not marshal %d created resource(s) (err; %v)", len(resNames), err))

			panic(ErrInvalidData)
		}

		notifyCreation(string(bts), w)
	}
}

// notifyCreation :
// Used to setup the input response writer to indicate that the
// resource defined by the input string has successfully been
// created and can be accessed through the url.
//
// The `resource` represent a path to access the created object.
//
// The `w` response writer will be used to indicate the status
// to the client.
func notifyCreation(resource string, w http.ResponseWriter) {
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resource))
}
