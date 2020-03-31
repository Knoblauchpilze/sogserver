package routes

import (
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/handlers"
)

// playerAdapter :
// Implements the interface requested by the general handler in
// the `handlers` package. The main functions are describing the
// interface to retrieve information about the players from a
// database.
//
// The `proxy` defines the proxy to use to interact with the DB
// when fetching the data.
type playerAdapter struct {
	proxy data.PlayersProxy
}

// Route :
// Implementation of the method to get the route name to access
// the players' data.
// Returns the name of the route.
func (pa *playerAdapter) Route() string {
	return "players"
}

// ParseFilters :
// Implementation of the method to get the filters from the route
// variables. This will allow to fetch players through a precise
// identifier.
//
// The `vars` are the variables (both route element and the query
// parameters) retrieved from the input request.
//
// Returns the list of filters extracted from the input info.
func (pa *playerAdapter) ParseFilters(vars handlers.RouteVars) []handlers.Filter {
	filters := make([]handlers.Filter, 0)

	// Traverse the input parameters and select only the ones relevant
	// to players querying.
	allowed := map[string]string{
		"player_id":   pa.proxy.GetIdentifierDBColumnName(),
		"account_id":  "account",
		"universe_id": "uni",
		"player_name": "name",
	}

	for key, values := range vars.Params {
		// Check whether this key corresponds to an player filter. If
		// this is the case we can register it.
		filterName, ok := allowed[key]

		if ok && len(values) > 0 {
			// None of the filters associated to the player are numeric
			// for now. If it's the case later on it would have to be
			// modified.
			filter := handlers.Filter{
				Key:     filterName,
				Options: values,
			}

			filters = append(filters, filter)
		}
	}

	// We also need to fetch parts of the route that can be used to
	// provide a filter on the player's identifier. More precisely
	// the route can define something like:
	// `/players/player-id` which we will interpret as a filter on
	// the player's identifier.
	// Note that we assume that if the route contains more than `1`
	// element it *always* contains an identifier as second token.
	if len(vars.RouteElems) > 0 {
		uni := vars.RouteElems[0]

		// Append the identifier filter to the existing list.
		found := false
		for id := range filters {
			if filters[id].Key == pa.proxy.GetIdentifierDBColumnName() {
				found = true
				filters[id].Options = append(filters[id].Options, uni)
			}
		}

		if !found {
			filters = append(
				filters,
				handlers.Filter{
					Key:     pa.proxy.GetIdentifierDBColumnName(),
					Options: []string{uni},
				},
			)
		}
	}

	return filters
}

// Data :
// Implementation of the method to get the data related to players
// from the internal DB. We will use the internal DB proxy to get
// the info while still applying the filters.
//
// The `filters` represent the filters extracted from the route and
// as provided by the `ParseFilters` method. We need to convert it
// into a semantic that can be interpreted by the DB.
//
// Returns the data related to the players along with any errors.
func (pa *playerAdapter) Data(filters []handlers.Filter) (interface{}, error) {
	// Convert the input request filters into DB filters.
	dbFilters := make([]data.DBFilter, 0)
	for _, filter := range filters {
		dbFilters = append(
			dbFilters,
			data.DBFilter{
				Key:    filter.Key,
				Values: filter.Options,
			},
		)
	}

	// Use the DB proxy to fetch the data.
	return pa.proxy.Players(dbFilters)
}

// playerCreator :
// Implements the interface requested by the creation handler in
// the `handlers` package. The main functions are describing the
// interface to retrieve information about the players from a
// database.
//
// The `proxy` defines the proxy to use to interact with the DB
// when fetching the data.
type playerCreator struct {
	proxy data.PlayersProxy
}

// Route :
// Implementation of the method to get the route name to create some
// new players.
// Returns the name of the route.
func (pc *playerCreator) Route() string {
	return "player"
}

// AccessRoute :
// Implementation of the method to get the route name to access to
// the data created by this handler. This is basically the `players`
// route.
func (pc *playerCreator) AccessRoute() string {
	return "players"
}

// DataKey :
// Implementation of the method to get the name of the key used to
// pass data to the server.
// Returns the name of the key.
func (pc *playerCreator) DataKey() string {
	return "player-data"
}

// Create :
// Implementation of the method to perform the creation of the data
// related to the new players. We will use the internal proxy to
// request the DB to create a new player.
//
// The `data` represent the data fetched from the input request and
// should contain the properties of the players to create.
//
// Return the targets of the created resources along with any error.
func (pc *playerCreator) Create(data handlers.RouteData) ([]string, error) {
	// TODO: Implement this.
	return []string{}, fmt.Errorf("Not implemented")
}

// listPlayers :
// Creates a handler allowing to serve requests on the players
// by interrogating the main DB. We uses the handler structure in
// the `handlers` package and provide the needed endpoint desc as
// requested.
//
// Returns the handler that can be executed to serve such requests.
func (s *server) listPlayers() http.HandlerFunc {
	return handlers.ServeRoute(
		&playerAdapter{
			s.players,
		},
		s.log,
	)
}

// createPlayer :
// Creates a handler allowing to server requests to create new
// players in the main DB. This rely on the handler structure
// provided by the `handlers` package which allows to mutualize
// the extraction of the data from the input request and the
// general flow to perform the creation.
//
// Returns the handler which can be executed to perform such
// requests.
func (s *server) createPlayer() http.HandlerFunc {
	return handlers.ServeCreationRoute(
		&playerCreator{
			s.players,
		},
		s.log,
	)
}
