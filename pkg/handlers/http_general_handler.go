package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/pkg/logger"
)

// Filter :
// Generic filter that can be used to restrain the number
// of results returned by a query. At our level we only
// define the filter as a key combined with a set of some
// values (so we allow multiple values in the query) and
// we don't make assumptions as to how this should be used
// to actually query the database or anything else.
//
// The `Key` describes the value that should be filtered.
// It usually corresponds to the name of a column in the
// database.
//
// The `Options` represents the specific instances of the
// key that should be kept. Anything that is not part of
// the list of values will be ignored.
type Filter struct {
	Key     string
	Options Values
}

// EndpointDesc :
// Defines the information to describe a endpoint. Providing all
// the method of the interface will allow to easily use the below
// handler without any additional code.
// This allow to know what to do with the input request, to fetch
// properties that can be used to filter the answers and finally
// fetch the data from the DB.
//
// The `Route` method defines the raw string that should be served
// by the handler. It does not have to start by a '/' character
// (it will be stripped if this is the case) and will be the main
// entry point to serve.
//
// The `ParseFilters` allows to build a list of filters from the
// input route variables. This method allows the underlying implem
// to actually choose how to interpret the information retrieved
// from the route in its own way.
//
// The `Data` is used once the filters have been successfully
// parsed to actually retrieve the data to send back to the
// user. The data is returned through an interface along with
// any error. If an error is returned the handler will return
// an error to indicate the failure.
// In any other case it will marshal the data and send it back
// to the client.
type EndpointDesc interface {
	Route() string
	ParseFilters(vars RouteVars) []Filter
	Data(filters []Filter) (interface{}, error)
}

// InternalServerErrorString :
// Used to provide a unique string that can be used in case an
// error occurs while serving a client request and we need to
// provide an answer.
//
// Returns a common string to indicate an error.
func InternalServerErrorString() string {
	return "Unexpected server error"
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

// ServeRoute :
// Handles the request on the endpoint specified by the input route
// and let the user specify how filters should be extracted from the
// route and query parameters to then apply it to a DB query.
// In case an unexpected error happens during the setting up of the
// route variables a panic is issued so be sure to wrap the handler
// returned by this method with the adequate protections.
//
// The `endpoint` provide a description as defined in the interface
// provided by this package to allow to fully describe the behavior
// desired against filters and the fetching of the data.
// It will be used throughout the progression of the request and used
// whenever needed.
//
// The `log` allows to notify errors and warnings to the user in
// case it is needed while parsing the request.
//
// Returns the handler that can be executed to serve such requests.
func ServeRoute(endpoint EndpointDesc, log logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		routeName := sanitizeRoute(endpoint.Route())
		route := fmt.Sprintf("/%s", routeName)

		// We want to allow queries with the following syntax:
		//  - `/route/props`
		//  - `/route?props`
		// The syntax in both cases is similar in the sense that similar
		// arguments will be passed.

		// First extract the route variables: this include both the path
		// and the raw query parameters
		vars, err := extractRouteVars(route, r)
		if err != nil {
			panic(fmt.Errorf("Error while serving route \"%s\" (err: %v)", routeName, err))
		}

		// Parse the filters from the route variables.
		filters := endpoint.ParseFilters(vars)

		// Retrieve the data using the provided filters.
		data, err := endpoint.Data(filters)
		if err != nil {
			log.Trace(logger.Error, fmt.Sprintf("Unexpected error while fetching data for route \"%s\" (err: %v)", routeName, err))
			http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

			return
		}

		// Marshal the content of the data.
		err = marshalAndSend(data, w)
		if err != nil {
			log.Trace(logger.Error, fmt.Sprintf("Error while serving route \"%s\" (err: %v)", routeName, err))
		}
	}
}
