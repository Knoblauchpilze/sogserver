package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/handlers"
	"oglike_server/pkg/logger"
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
		"player_id":   "id",
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
			if filters[id].Key == "id" {
				found = true
				filters[id].Options = append(filters[id].Options, uni)
			}
		}

		if !found {
			filters = append(
				filters,
				handlers.Filter{
					Key:     "id",
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
// interface to create a new player from the data fetched from
// the input request.
//
// The `playerProxy` defines the proxy to use to interact with
// the DB when creating the player's entry in the table.
//
// The `planetProxy` defines the proxy to use to interact with
// the DB when creating a planet for a new player.
//
// The `log` allows to notify problems and information during a
// universe's creation.
type playerCreator struct {
	playerProxy data.PlayersProxy
	planetProxy data.PlanetProxy
	log         logger.Logger
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
// The `input` represent the data fetched from the input request and
// should contain the properties of the players to create.
//
// Return the targets of the created resources along with any error.
func (pc *playerCreator) Create(input handlers.RouteData) ([]string, error) {
	// We need to iterate over the data retrieved from the route and
	// create players from it.
	var player data.Player
	resources := make([]string, 0)

	// Prevent request with no data.
	if len(input.Data) == 0 {
		return resources, fmt.Errorf("Could not perform creation of player with no data")
	}

	for _, rawData := range input.Data {
		// Try to unmarshal the data into a valid `Player` struct.
		err := json.Unmarshal([]byte(rawData), &player)
		if err != nil {
			pc.log.Trace(logger.Error, fmt.Sprintf("Could not create player from data \"%s\" (err: %v)", rawData, err))
			continue
		}

		// Create the player.
		err = pc.playerProxy.Create(&player)
		if err != nil {
			pc.log.Trace(logger.Error, fmt.Sprintf("Could not register player from data \"%s\" (err: %v)", rawData, err))
			continue
		}

		// Choose a homeworld for this account and create it.
		err = pc.planetProxy.CreateFor(player, nil)
		if err != nil {
			// Indicate that we could not create the planet for the playe. It
			// is not ideal because we should probably delete the player entry
			// (because there's no associated homeworld anyway so the player
			// will not be able to play correctly). For now though we consider
			// that it's rare enough to not handle it.
			// If it's a problem we can still handle it later. For example by
			// creating a `deletePlayer` method which will be needed anyways
			// at some point.
			pc.log.Trace(logger.Error, fmt.Sprintf("Could not create homeworld for player \"%s\" (err: %v)", player.ID, err))
		}

		// Successfully created an player.
		pc.log.Trace(logger.Notice, fmt.Sprintf("Created new player \"%s\" with id \"%s\"", player.Name, player.ID))
		resources = append(resources, player.ID)
	}

	// Return the path to the resources created during the process.
	return resources, nil
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
			s.planets,
			s.log,
		},
		s.log,
	)
}
