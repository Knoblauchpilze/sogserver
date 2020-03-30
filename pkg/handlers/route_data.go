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
// `extractRouteVars` method.
//
// The `RouteElems` represents the extra path added to the route
// as it was provided to target the server. Typically if the server
// receives a request on `/universes`, the `RouteElems` will be set
// to the empty slice. On the other hand `/universes/oberon` will
// yield a single element `oberon` in the `RouteElems` slice..
//
// The `Params` define the query parameters associated to the input
// request. Note that in some case no parameters are provided.
type RouteVars struct {
	RouteElems []string
	Params     map[string]Values
}
