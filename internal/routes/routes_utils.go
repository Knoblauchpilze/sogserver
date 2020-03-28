package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// InternalServerErrorString :
// Used to provide a unique string that can be used in case an
// error occurs while serving a client request and we need to
// provide an answer.
//
// Returns a common string to indicate an error.
func InternalServerErrorString() string {
	return "Unexpected server error"
}

// extractRoute :
// Convenience method allowing to strip the input prefix from the
// route defined in an input request to keep only the part that is
// specific to a server behavior.
//
// The `r` argument represents the request from which the route
// should be extracted. An error is raised in case this requets is
// not valid.
//
// The `prefix` represents the prefix to be stripped from the input
// request. If the prefix does not exist in the route an error is
// returned as well.
//
// Returns either the route stripped from the prefix or an error if
// something went wrong.
func extractRoute(r *http.Request, prefix string) (string, error) {
	if r == nil {
		return "", fmt.Errorf("Cannot strip prefix \"%s\" from invalid route", prefix)
	}

	route := r.URL.String()

	if !strings.HasPrefix(route, prefix) {
		return "", fmt.Errorf("Cannot strip prefix \"%s\" from route \"%s\"", prefix, route)
	}

	return strings.TrimPrefix(route, prefix), nil
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
