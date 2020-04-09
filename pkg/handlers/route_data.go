package handlers

// Values :
// A convenience define to allow for easy manipulation of a list
// of strings as a single element. This is mostly used to be able
// to interpret multiple values for a single query parameter in
// an easy way.
type Values []string

// RouteVars :
// Define common information to be passed in the route to contact
// the server. We handle extra path that can be added to the route
// (typically to refine the behavior expected from the base route)
// and some query parameters.
// An object of this type is extracted for each single request to
// put some sort of format when contacting the server. This is
// then passed to the underlying interface implementation to set
// and generate some filters from these variables. These values
// can be extracted from the input request through calling the
// `ExtractRouteVars` method.
//
// The `RouteElems` represents the extra path added to the route
// as it was provided to target the server. Typically if the server
// receives a request on `/universes`, the `RouteElems` will be set
// to the empty slice. On the other hand `/universes/oberon` will
// yield a single element `oberon` in the `RouteElems` slice.
//
// The `Params` define the query parameters associated to the input
// request. Note that in some case no parameters are provided.
type RouteVars struct {
	RouteElems []string
	Params     map[string]Values
}

// RouteData :
// Used to define the data that can be passed to a route based on
// its name. It is constructed by appending the "-data" suffix to
// the route and getting the corresponding value.
// It is useful to group common behavior of all the interfaces on
// this server.
//
// The `RouteElems` represents the extra path added to the route
// as it was provided to target the server. Typically if the server
// receives a request on `/universe`, the `RouteElems` will be set
// to the empty slice. On the other hand `/universe/oberon` will
// yield a single element `oberon` in the `RouteElems` slice.
//
// The `Data` represents the data extracted from the route itself.
// It is represented as an array of raw strings which are usually
// unmarshalled into meaningful structures by the data creation
// process.
type RouteData struct {
	RouteElems []string
	Data       Values
}
