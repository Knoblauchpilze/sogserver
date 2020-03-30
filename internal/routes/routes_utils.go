package routes

import (
	"net/http"
	"strings"
)

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
