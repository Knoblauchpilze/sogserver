package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"

	"github.com/google/uuid"
)

// DataFunc :
// Convenience define which allows to refer to the process
// to fetch data from the DB given a set of filters. It is
// used as a generic handler which will be provided to the
// resource getter and used when the data from the input
// request has been parsed. The return values includes both
// the data itself and any error.
// This wrapper is used in an `EndpointDesc` to mutualize
// even more the basic functionalities to fetch data of
// different kind from the main DB.
type DataFunc func(filters []db.Filter) (interface{}, error)

// GetResourceEndpoint :
// Defines the information to describe a endpoint. This allows to
// mutualize most of the processing to actually serve the `GET`
// requests on the server as we can rely on the specific behavior
// from the DB proxies to specialize the data.
// The rest of the process is very similar where one needs to
// extract the information from the route, build some filters to
// apply when fetching from the DB and then serialize the data
// before sending it back to the user.
//
// The `route` defines a string that should be contained in the
// input string: it helps separating the routes from the query
// parameters that can be provided to filter out some results in
// the output result.
//
// The `fetcher` references the function to use to retrieve
// data from the DB once the filters have been retrieved from
// the input request and converted into allowed filters.
//
// The `filters` defines some association table between the
// input query parameters and their association as DB filters
// that can be used internally. We provide some sort of mapping
// in order to hide the potential complexity of the filters to
// apply.
//
// The `idFilter` filter corresponds to a special filter that
// can be extracted from the route itself. Typically it allows
// the typical REST syntax of `/resource/resource-id` where a
// resource can be fetched from its identifier appended to the
// general resource's route.
// If this value is empty (default case) it will be ignored.
// Note that the path of the route will be checked so that it
// is at least a valid syntax for a `uuid`.
//
// The `resFilter` filter corresponds to the generci semantic
// of the REST syntax where a collection can be accessed through
// `/path/to/collection` while a specific resource from this
// collection can be accessed with `/path/to/collection/res-id`.
// This filter will be in charge of collecting the last token of
// the route and register it as a filter.
// Note that it will only activate in case the route defines
// some extra elements *in addition* to its registered path.
// Note that the path of the route will be checked so that it
// is at least a valid syntax for a `uuid`.
//
// The `module` defines a string that can be used to make the
// logs display more explicit by specifying this module's id.
// This string should be unique across the application and is
// used as a mean to easily distinguish between the different
// services composing the server.
type GetResourceEndpoint struct {
	route     string
	fetcher   DataFunc
	filters   map[string]string
	idFilter  string
	resFilter string
	module    string
}

// NewGetResourceEndpoint :
// Creates a new empty endpoint description with the provided
// route. The fetcher func is defined as an empty element to
// avoid fetching anything and no filters are provided.
//
// The `route` will be sanitized and represent the path that
// is associated to this endpoint.
//
// Returns the created end point description.
func NewGetResourceEndpoint(route string) *GetResourceEndpoint {
	return &GetResourceEndpoint{
		route: sanitizeRoute(route),
	}
}

// WithFilters :
// Provide a way to assign some filters for a given endpoint
// which means associating some query parameters with their
// DB description.
//
// The `filters` define the association table describing the
// filters to attach to this endpoint.
//
// Returns the endpoint to allow chain calling.
func (gre *GetResourceEndpoint) WithFilters(filters map[string]string) *GetResourceEndpoint {
	gre.filters = filters
	return gre
}

// WithIDFilter :
// Defines that this endpoint is able to handle filtering of
// the resources through an identifier provided in the route.
// The second part of the route (if it exists) will be used
// as a filter with the specified name.
//
// The `id` defines the string representing the identifier
// filter to apply on the DB.
//
// Returns the endpoint to allow chain calling.
func (gre *GetResourceEndpoint) WithIDFilter(id string) *GetResourceEndpoint {
	gre.idFilter = id
	return gre
}

// WithResourceFilter :
// Defines that this endpoint is able to handle filtering of
// the resources through an identifier provided in the route.
// It is similar to the `ID` filter but it applies to the
// last token of the route. Typically with the following ex:
// `/id/some/path/to/a/resource/res-id`
// The `WithIDFilter` will be able to fetch the `id` part
// and make a filter of it, while the resource filter will
// catch the `res-id` part and make a filter out of it.
// Note that in case the route path has length `1`, this
// filter will not be triggered if the `WithIDFilter` method
// is active on the endpoint (in order to prevent conflicts).
//
// The `id` defines the string representing the identifier
// to apply on the DB.
//
// Returns the endpoint to allow chain calling.
func (gre *GetResourceEndpoint) WithResourceFilter(id string) *GetResourceEndpoint {
	gre.resFilter = id
	return gre
}

// WithDataFunc :
// Assigns the input data function as the main way to query
// data for this endpoint.
//
// The `f` represents the data function that should be used
// by this endpoint to fetch data.
//
// Returns this endpoint to allow chain calling.
func (gre *GetResourceEndpoint) WithDataFunc(f DataFunc) *GetResourceEndpoint {
	gre.fetcher = f
	return gre
}

// WithModule :
// Assigns a new string as the module name for this object.
//
// The `module` defines the name of the module to assign to
// this object
//
// Returns this endpoint to allow chain calling.
func (gre *GetResourceEndpoint) WithModule(module string) *GetResourceEndpoint {
	gre.module = module
	return gre
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
func (gre *GetResourceEndpoint) ServeRoute(log logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("/%s", gre.route)

		// First extract the route variables: this include both the path
		// and the raw query parameters
		vars, err := extractRouteVars(route, r)
		if err != nil {
			panic(fmt.Errorf("Error while serving route \"%s\" (err: %v)", gre.route, err))
		}

		// Parse the filters from the route variables.
		filters := gre.extractFilters(vars)

		// Retrieve the data using the provided filters.
		if gre.fetcher == nil {
			// The fetcher is not assigned, terminate the request here.
			return
		}

		data, err := gre.fetcher(filters)
		if err != nil {
			log.Trace(logger.Error, gre.module, fmt.Sprintf("Unexpected error while fetching data for route \"%s\" (err: %v)", gre.route, err))
			http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

			return
		}

		// Marshal the content of the data.
		err = marshalAndSend(data, w)
		if err != nil {
			log.Trace(logger.Error, gre.module, fmt.Sprintf("Error while serving route \"%s\" (err: %v)", gre.route, err))
		}
	}
}

// extractFilters :
// Implementation of the method to get the filters from the route
// variables. This will allow to fetch data and still be able to
// filter out some elements based on these filters.
// The filters are translated from a query parameters semantic to
// something that can be understood by the DB through the mapping
// table defined in this endpoint.
//
// The `vars` are the variables (both route element and the query
// parameters) retrieved from the input request.
//
// Returns the list of filters extracted from the input info.
func (gre *GetResourceEndpoint) extractFilters(vars RouteVars) []db.Filter {
	filters := make([]db.Filter, 0)

	for key, values := range vars.Params {
		// Check whether this filter is allowed.
		filterName, ok := gre.filters[key]

		if ok && len(values) > 0 {
			filter := db.Filter{
				Key:    filterName,
				Values: values,
			}

			filters = append(filters, filter)
		}
	}

	// We also need to fetch parts of the route that can be used to
	// provide a filter on the identifier of the resource fetched
	// by this handler.
	// More precisely the route can be defined in a way like:
	// `/resource/resource-id` which we will interpret as a filter
	// on the resource's identifier.
	// Note that we assume that if the route contains more than `1`
	// element it *always* contains an identifier as second token.
	// This behavior is only active in case the `idFilter` internal
	// string is not empty.
	if len(gre.idFilter) > 0 && len(vars.ExtraElems) > 0 {
		filter := vars.ExtraElems[0]

		// Make sure that this filter could be used as a valid `uuid`.
		if _, err := uuid.Parse(filter); err == nil {
			// Append the identifier filter to the existing list.
			found := false
			for id := range filters {
				if filters[id].Key == gre.idFilter {
					found = true
					filters[id].Values = append(filters[id].Values, filter)
				}
			}

			if !found {
				filters = append(
					filters,
					db.Filter{
						Key:    gre.idFilter,
						Values: []string{filter},
					},
				)
			}
		}
	}

	// Finally we need to fetch the resource identifier that might
	// be provided in the route's tokens. Typically imagine a route
	// like `/resource/resource-id`, we want to detect whether the
	// last element of the route should be considered as a resource
	// identifier that can used as a filter.
	if len(gre.resFilter) > 0 && len(vars.ExtraElems) > 0 {
		filter := vars.ExtraElems[len(vars.ExtraElems)-1]

		// Make sure that this filter could be used as a valid `uuid`.
		if _, err := uuid.Parse(filter); err == nil {
			// Append the identifier filter to the existing list.
			found := false
			for id := range filters {
				if filters[id].Key == gre.resFilter {
					found = true
					filters[id].Values = append(filters[id].Values, filter)
				}
			}

			if !found {
				filters = append(
					filters,
					db.Filter{
						Key:    gre.resFilter,
						Values: []string{filter},
					},
				)
			}
		}
	}

	return filters
}

// marshalAndSend :
// Used to send the input data after marshalling it to the provided
// response writer. In case the data cannot be marshalled a `500`
// error is returned and this is indicated in the return value.
//
// The `data` represents the data to send back to the client.
//
// The `w` represents the response writer to use to send data back.
//
// Returns any error encountered either when marshalling the data
// or when sending the data.
func marshalAndSend(data interface{}, w http.ResponseWriter) error {
	// Marshal the content before sending it back.
	out, err := json.Marshal(data)
	if err != nil {
		http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

		return err
	}

	// Notify the client.
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(out)

	return err
}
