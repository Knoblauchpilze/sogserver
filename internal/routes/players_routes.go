package routes

import (
	"encoding/json"
	"net/http"
	"oglike_server/internal/game"
	"oglike_server/pkg/db"
)

// listPlayers :
// Used to perform the creation of a handler allowing to serve
// the requests on players.
//
// Returns the handler that can be executed to serve said reqs.
func (s *Server) listPlayers() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("players")

	allowed := map[string]string{
		"player":   "id",
		"account":  "account",
		"universe": "universe",
		"name":     "name",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id").WithModule("players")
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.players.Players(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// createPlayer :
// Used to perform the creation of a handler allowing to server
// the requests to create players.
//
// Returns the handler to execute to perform said requests.
func (s *Server) createPlayer() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewCreateResourceEndpoint("players")

	// Configure the endpoint.
	ed.WithDataKey("player-data").WithModule("players")
	ed.WithCreationFunc(
		func(input RouteData) ([]string, error) {
			// We need to iterate over the data retrieved from the route and
			// create players from it.
			var player game.Player
			resources := make([]string, 0)

			// Prevent request with no data.
			if len(input.Data) == 0 {
				return resources, ErrNoData
			}

			for _, rawData := range input.Data {
				// Try to unmarshal the data into a valid `Player` struct.
				err := json.Unmarshal([]byte(rawData), &player)
				if err != nil {
					return resources, ErrInvalidData
				}

				// Create the player.
				res, err := s.players.Create(player)
				if err != nil {
					return resources, err
				}

				// Update the player's identifier.
				player.ID = res

				// Choose a homeworld for this account and create it.
				_, err = s.planets.CreateFor(player)

				if err != nil {
					// Indicate that we could not create the planet for the player. It
					// is not ideal because we should probably delete the player entry
					// (because there's no associated homeworld anyway so the player
					// will not be able to play correctly). For now though we consider
					// that it's rare enough to not handle it.
					// If it's a problem we can still handle it later. For example by
					// creating a `deletePlayer` method which will be needed anyways
					// at some point.
					return resources, err
				}

				// Successfully created a player.
				resources = append(resources, res)
			}

			// Return the path to the resources created during the process.
			return resources, nil
		},
	)

	return ed.ServeRoute(s.log)
}
