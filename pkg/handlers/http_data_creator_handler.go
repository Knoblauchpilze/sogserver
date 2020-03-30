package handlers

import (
	"fmt"
	"net/http"
	"oglike_server/pkg/logger"
)

// CreationEndpointDesc :
// Defines the information to describe a endpoint which can be
// used to handle data creation. This interface allows a handler
// to perform general configuration such as fetching the route
// variables and data and format it in a way that can be easily
// interpreted by this object which can then perform the actual
// data creation.
//
// The `Route` method defines the raw string that should be served
// by the handler. It does not have to start by a '/' character
// (it will be stripped if this is the case) and will be the main
// entry point to serve.
//
// The `DataKey` allows to determine which key should be scanned
// to retrieve the data to use for the creation of the resource.
//
// The `Create` facet is used once the filters have been parsed
// successfully to actually perform the creation of the data in
// the DB or any other data model linked to the description obj.
// The interface expects to be provided an error (if any) and a
// path to access the created resources.
// It will automatically be sent back to the client to conform
// to the REST API architecture.
type CreationEndpointDesc interface {
	Route() string
	DataKey() string
	Create(data RouteData) (string, error)
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
	// Notify the status.
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(resource))
}

// ServeCreationRoute :
// Used to create a general handler allowing to retrieve data to create
// an object. This method will parse the content provided for the route
// it is binded to and call the appropriate handler to perform the real
// object's creation.
// Note that in the case of an error happening while parsing the route
// data a panic will be issued: make sure to wrap this handler with the
// adequate protections.
//
// The `log` allows to notify errors and warnings to the user in case it
// is needed while parsing the request.
//
// Returns the handler that can be executed to serve such requests.
func ServeCreationRoute(endpoint CreationEndpointDesc, log logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		routeName := sanitizeRoute(endpoint.Route())

		// We want to allow queries where the data is provided with a key
		// provided by the endpoint itself. Perform the extraction of the
		// data from the request.
		data, err := extractRouteData(endpoint.DataKey(), r)
		if err != nil {
			panic(fmt.Errorf("Could not fetch data from request for route \"%s\" (err: %v)", routeName, err))
		}

		resName, err := endpoint.Create(data)
		if err != nil {
			log.Trace(logger.Error, fmt.Sprintf("Could not create resource from route \"%s\" (err: %v)", routeName, err))
			http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

			return
		}

		// We need to return a valid status code and the address of
		// the created resource, as described in the following post:
		// https://stackoverflow.com/questions/1829875/is-it-ok-by-rest-to-return-content-after-post
		resource := fmt.Sprintf("/%s/%s", routeName, resName)
		notifyCreation(resource, w)
	}
}
