package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/game"
)

// registerFunc :
// Defines a convenience typedef allowing to register
// some actions defined by the input string. It should
// be unmarshalled into whatever data type is needed
// to perform the registration and the corresponding
// identifier should be returned in case the creation
// is successful.
// Allows to mutualize the common parts of registering
// an upgrade action and only define specific parts of
// the process.
// In addition to the input that should be processed
// by the function, the calling framework also provide
// information about the route context through some
// tokens which usually represent the path that was
// used to contact the route. We will typically use
// these tokens to extract some info about the caller
// for whicht he action should be registered (so for
// a player in the case of technology, a planet in the
// case of buildings, ships and defenses).
type registerFunc func(input string, routeTokens []string) (string, error)

// registerUpgradeAction :
// Used to perform the creation of a handler allowing to server
// the requests to create generic upgrade actions. The precise
// creation process will be configured through values provided
// as input.
//
// The `f` defines the registration function to apply when the
// upgrade action should be created.
//
// Returns the handler to execute to perform said requests.
func (s *Server) registerUpgradeAction(f registerFunc) http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewCreateResourceEndpoint("players")

	// Configure the endpoint.
	ed.WithDataKey("action-data").WithModule("actions")
	ed.WithCreationFunc(
		func(input RouteData) ([]string, error) {
			// We need to iterate over the data retrieved from the route and
			// create actions from it.
			resources := make([]string, 0)

			// Prevent request with no data.
			if len(input.Data) == 0 {
				return resources, ErrNoData
			}

			for _, rawData := range input.Data {
				// Unmarshal and perform the creation using the provided
				// registration function.
				res, err := f(rawData, input.ExtraElems)
				if err != nil {
					return resources, err
				}

				// Successfully created the action.
				resources = append(resources, res)
			}

			// Return the path to the resources created during the process.
			return resources, nil
		},
	)

	return ed.ServeRoute(s.log)
}

// registerBuildingAction :
// Used to perform the creation of a handler allowing to server
// the requests to create building upgrade actions.
//
// Returns the handler to execute to perform said requests.
func (s *Server) registerBuildingAction() http.HandlerFunc {
	return s.registerUpgradeAction(
		func(input string, routeTokens []string) (string, error) {
			// Unmarshal the input data into a building upgrade action
			// and perform the registration through the dedicated func.
			var action game.BuildingAction

			err := json.Unmarshal([]byte(input), &action)
			if err != nil {
				return "", ErrInvalidData
			}

			// The `routeTokens` should provide the planet's id
			// so we can override any value provided by the act
			// itself to maintain consistency.
			if len(routeTokens) > 0 {
				action.Player = routeTokens[0]
			}
			if len(routeTokens) > 3 {
				action.Planet = routeTokens[2]
			}

			// Create the upgrade action.
			_, err = s.actions.CreateBuildingAction(action)

			// Build the path to access to the resource: we need to
			// include the player and planet's identifier in the route.
			res := fmt.Sprintf("%s/planets/%s", action.Player, action.Planet)

			return res, err
		},
	)
}

// registerTechnologyAction :
// Used to perform the creation of a handler allowing to server
// the requests to create technology upgrade actions.
//
// Returns the handler to execute to perform said requests.
func (s *Server) registerTechnologyAction() http.HandlerFunc {
	return s.registerUpgradeAction(
		func(input string, routeTokens []string) (string, error) {
			// Unmarshal the input data into a technology upgrade action
			// and perform the registration through the dedicated func.
			var action game.TechnologyAction

			err := json.Unmarshal([]byte(input), &action)
			if err != nil {
				return "", ErrInvalidData
			}

			// The `routeTokens` should provide both the player and
			// the planet's id so we can override any value provided
			// in the upgrade action.
			if len(routeTokens) > 0 {
				action.Player = routeTokens[0]
			}
			if len(routeTokens) > 3 {
				action.Planet = routeTokens[2]
			}

			// Create the upgrade action.
			_, err = s.actions.CreateTechnologyAction(action)

			// Build the path to access to the resource: we need to
			// include the player and planet's identifier in the route.
			res := fmt.Sprintf("%s/planets/%s", action.Player, action.Planet)

			return res, err
		},
	)
}

// registerShipAction :
// Used to perform the creation of a handler allowing to server
// the requests to create ship upgrade actions.
//
// Returns the handler to execute to perform said requests.
func (s *Server) registerShipAction() http.HandlerFunc {
	return s.registerUpgradeAction(
		func(input string, routeTokens []string) (string, error) {
			// Unmarshal the input data into a ship upgrade action
			// and perform the registration through the dedicated
			// function.
			var action game.ShipAction

			err := json.Unmarshal([]byte(input), &action)
			if err != nil {
				return "", ErrInvalidData
			}

			// The `routeTokens` should provide the planet's id
			// so we can override any value provided by the act
			// itself to maintain consistency.
			if len(routeTokens) > 0 {
				action.Player = routeTokens[0]
			}
			if len(routeTokens) > 3 {
				action.Planet = routeTokens[2]
			}

			// Create the upgrade action.
			_, err = s.actions.CreateShipAction(action)

			// Build the path to access to the resource: we need to
			// include the player and planet's identifier in the route.
			res := fmt.Sprintf("%s/planets/%s", action.Player, action.Planet)

			return res, err
		},
	)
}

// registerDefenseAction :
// Used to perform the creation of a handler allowing to server
// the requests to create defense upgrade actions.
//
// Returns the handler to execute to perform said requests.
func (s *Server) registerDefenseAction() http.HandlerFunc {
	return s.registerUpgradeAction(
		func(input string, routeTokens []string) (string, error) {
			// Unmarshal the input data into a defense upgrade
			// action and perform the registration through the
			// dedicated function.
			var action game.DefenseAction

			err := json.Unmarshal([]byte(input), &action)
			if err != nil {
				return "", ErrInvalidData
			}

			// The `routeTokens` should provide the planet's id
			// so we can override any value provided by the act
			// itself to maintain consistency.
			if len(routeTokens) > 0 {
				action.Player = routeTokens[0]
			}
			if len(routeTokens) > 3 {
				action.Planet = routeTokens[2]
			}

			// Create the upgrade action.
			_, err = s.actions.CreateDefenseAction(action)

			// Build the path to access to the resource: we need to
			// include the player and planet's identifier in the route.
			res := fmt.Sprintf("%s/planets/%s", action.Player, action.Planet)

			return res, err
		},
	)
}
