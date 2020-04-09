package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/handlers"
	"oglike_server/pkg/logger"
)

// listPlayers :
// Used to perform the creation of a handler allowing to serve
// the requests on players.
//
// Returns the handler that can be executed to serve said reqs.
func (s *server) listPlayers() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewEndpointDesc("players")

	allowed := map[string]string{
		"player_id":   "id",
		"account_id":  "account",
		"universe_id": "uni",
		"player_name": "name",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithIDFilter("id")
	ed.WithDataFunc(
		func(filters []data.DBFilter) (interface{}, error) {
			return s.players.Players(filters)
		},
	)

	return ed.ServeRoute(s.log)
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
// player's creation.
type playerCreator struct {
	playerProxy data.PlayerProxy
	planetProxy data.PlanetProxy
	log         logger.Logger
}

// Route :
// Implementation of the method to get the route name to create some
// new players.
// Returns the name of the route.
func (pc *playerCreator) Route() string {
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
			// Indicate that we could not create the planet for the player. It
			// is not ideal because we should probably delete the player entry
			// (because there's no associated homeworld anyway so the player
			// will not be able to play correctly). For now though we consider
			// that it's rare enough to not handle it.
			// If it's a problem we can still handle it later. For example by
			// creating a `deletePlayer` method which will be needed anyways
			// at some point.
			pc.log.Trace(logger.Error, fmt.Sprintf("Could not create homeworld for player \"%s\" (err: %v)", player.ID, err))
		} else {
			// Successfully created an player.
			pc.log.Trace(logger.Notice, fmt.Sprintf("Created new player \"%s\" with id \"%s\"", player.Name, player.ID))
			resources = append(resources, player.ID)
		}
	}

	// Return the path to the resources created during the process.
	return resources, nil
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
