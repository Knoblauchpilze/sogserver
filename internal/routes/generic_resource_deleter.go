package routes

import (
	"fmt"
	"net/http"
	"oglike_server/internal/game"
	"oglike_server/pkg/logger"
)

// deleteFunc :
// Convenience define which allows to refer to the process
// to delete a resource from the DB. The resources passed
// as input argument will be removed from the DB if it can
// be found. If no such resource exist an error should be
// returned.
type deleteFunc func(resource string) error

// DeleteResourceEndpoint :
// Common information to perform the deletion of an abstract
// resource from the server. This endpoint is related to a
// route which defines the actual data being removed by the
// endpoint along with a deleter function which is called to
// perform the deletion.
// Such an architecture is useful to mutualize the fetching
// of the resource to delete from the route's options and
// to only keep the specific deletion behavior for this res.
//
// The `route` defines the string of the endpoint serving
// the deletion behavior. It is usually linked to the res
// and the collection it belongs to.
//
// The `deleter` references the function to use to delete
// the resource from the server. It will be fed the ID of
// the resource extracted from the route.
//
// The `module` defines a string that can be used to make
// the logs displayed more explicit by specifying common
// string to identify the collection associated to this
// endpoint.
//
// The `lock` allows to define whether a locker should be
// applied when fetching data from the DB. This lock will
// be automatically acquired before passing on the request
// to the `deleter` function and release afterwards. If it
// is set to `nil` (default behavior) no lock is acquired.
type DeleteResourceEndpoint struct {
	route   string
	deleter deleteFunc
	module  string
	lock    *game.Instance
}

// NewDeleteResourceEndpoint :
// Creates a new empty endpoint description allowing to
// delete resources from the server for the specified
// route.
//
// The `route` will be sanitized and represent the path
// that is associated to this endpoint.
//
// Returns the created end point description.
func NewDeleteResourceEndpoint(route string) *DeleteResourceEndpoint {
	return &DeleteResourceEndpoint{
		route: sanitizeRoute(route),
	}
}

// WithDeleterFunc :
// Assigns the input deleter function as the main way
// to remove resoruces from the server for this endpoint.
//
// The `d` represents the deleter function that should
// be used by this endpoint to remove data.
//
// Returns this endpoint to allow chain calling.
func (dre *DeleteResourceEndpoint) WithDeleterFunc(d deleteFunc) *DeleteResourceEndpoint {
	dre.deleter = d
	return dre
}

// WithModule :
// Assigns a new string as the module name for this object.
//
// The `module` defines the name of the module to assign to
// this object
//
// Returns this endpoint to allow chain calling.
func (dre *DeleteResourceEndpoint) WithModule(module string) *DeleteResourceEndpoint {
	dre.module = module
	return dre
}

// WithLocker :
// Assigns a new locker element to use when deleting
// data.
//
// The `locker` defines the element to acquire before
// a request on DB data can be served.
//
// Returns this endpoint to allow chain calling.
func (dre *DeleteResourceEndpoint) WithLocker(i game.Instance) *DeleteResourceEndpoint {
	dre.lock = &i
	return dre
}

// ServeRoute :
// Returns a handler using this endpoint description to
// be able to serve requests given the configuration as
// defined by the user.
// Note that the handler is returned and still need to
// be activated to serve the behavior.
//
// The `log` will be used to notify info and messages if
// needed.
//
// Returns the handler that can be executed to serve the
// requests defined by the data of this endpoint.
func (dre *DeleteResourceEndpoint) ServeRoute(log logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("/%s", dre.route)

		// First extract the route variables: we will mostly
		// use the extra path.
		vars, err := extractRouteVars(route, r)
		if err != nil {
			log.Trace(logger.Error, dre.module, fmt.Sprintf("Error while serving route \"%s\" (err: %v)", dre.route, err))
			panic(err)
		}

		// Make sure that there's a resource to be deleted.
		if len(vars.ExtraElems) == 0 {
			// No resource is provided for deletion. This could
			// indicate that the request is meant to delete the
			// entire collection of resources but we will forbid
			// this. So as per the common practices we will use
			// a `405` error code to prevent this.
			// See here: https://www.restapitutorial.com/lessons/httpmethods.html
			http.Error(w, fmt.Sprintf("Delete of entire collection not allowed for \"%s\"", dre.route), http.StatusMethodNotAllowed)
		}

		// The resource to delete is the first extra element
		// parsed from the route.
		resource := vars.ExtraElems[0]

		// Make sure that the operational layer of this
		// endpoint is assigned.
		if dre.deleter == nil {
			return
		}

		func() {
			if dre.lock != nil {
				dre.lock.Lock()
				defer dre.lock.Unlock()
			}

			err = dre.deleter(resource)
		}()

		if err != nil {
			log.Trace(logger.Error, dre.module, fmt.Sprintf("Unexpected error while deleting data for route \"%s\" (err: %v)", dre.route, err))

			// Detect special cases of a not found element.
			if err == game.ErrElementNotFound {
				http.Error(w, fmt.Sprintf("%v", err), http.StatusNotFound)
			} else {
				http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)
			}

			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
