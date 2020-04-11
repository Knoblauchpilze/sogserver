package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/logger"
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
// The `route` defines the route to serve to register upgrade
// actions.
//
// The `f` defines the registration function to apply when the
// upgrade action should be created.
//
// Returns the handler to execute to perform said requests.
func (s *server) registerUpgradeAction(route string, f registerFunc) http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewCreateResourceEndpoint(route)

	// Configure the endpoint.
	ed.WithDataKey("action-data")
	ed.WithCreationFunc(
		func(input RouteData) ([]string, error) {
			// We need to iterate over the data retrieved from the route and
			// create actions from it.
			resources := make([]string, 0)

			// Prevent request with no data.
			if len(input.Data) == 0 {
				return resources, fmt.Errorf("Could not perform creation of defense upgrade action with no data")
			}

			for _, rawData := range input.Data {
				// Unmarshal and perform the creation using the provided
				// registration function.
				res, err := f(rawData, input.RouteElems)
				if err != nil {
					s.log.Trace(logger.Error, fmt.Sprintf("Could not register upgrade action from data \"%s\" (err: %v)", rawData, err))
					continue
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
func (s *server) registerBuildingAction() http.HandlerFunc {
	return s.registerUpgradeAction(
		"planets",
		func(input string, routeTokens []string) (string, error) {
			// Unmarshal the input data into a building upgrade action
			// and perform the registration through the dedicated func.
			var action data.BuildingUpgradeAction

			err := json.Unmarshal([]byte(input), &action)
			if err != nil {
				return "", err
			}

			// The `routeTokens` should also provide the planet's id
			// so we can override any value provided in the upgrade
			// action.
			if len(routeTokens) > 0 {
				action.PlanetID = routeTokens[0]
			}

			// Create the upgrade action.
			err = s.upgradeAction.CreateBuildingAction(&action)

			// Build the path to access to the resource: we need to
			// include the planet's identifier in the route.
			res := fmt.Sprintf("%s/buildings/%s", action.PlanetID, action.ID)

			return res, err
		},
	)
}

// registerTechnologyAction :
// Used to perform the creation of a handler allowing to server
// the requests to create technology upgrade actions.
//
// Returns the handler to execute to perform said requests.
func (s *server) registerTechnologyAction() http.HandlerFunc {
	return s.registerUpgradeAction(
		"players",
		func(input string, routeTokens []string) (string, error) {
			// Unmarshal the input data into a technology upgrade action
			// and perform the registration through the dedicated func.
			var action data.TechnologyUpgradeAction

			err := json.Unmarshal([]byte(input), &action)
			if err != nil {
				return "", err
			}

			// The `routeTokens` should also provide the player's id
			// so we can override any value provided in the upgrade
			// action.
			if len(routeTokens) > 0 {
				action.PlayerID = routeTokens[0]
			}

			// Create the upgrade action.
			err = s.upgradeAction.CreateTechnologyAction(&action)

			return action.ID, err
		},
	)
}

// registerShipAction :
// Used to perform the creation of a handler allowing to server
// the requests to create ship upgrade actions.
//
// Returns the handler to execute to perform said requests.
func (s *server) registerShipAction() http.HandlerFunc {
	return s.registerUpgradeAction(
		"planets",
		func(input string, routeTokens []string) (string, error) {
			// Unmarshal the input data into a ship upgrade action
			// and perform the registration through the dedicated
			// function.
			var action data.ShipUpgradeAction

			err := json.Unmarshal([]byte(input), &action)
			if err != nil {
				return "", err
			}

			// The `routeTokens` should also provide the planet's id
			// so we can override any value provided in the upgrade
			// action.
			if len(routeTokens) > 0 {
				action.PlanetID = routeTokens[0]
			}

			// Create the upgrade action.
			err = s.upgradeAction.CreateShipAction(&action)

			return action.ID, err
		},
	)
}

// registerDefenseAction :
// Used to perform the creation of a handler allowing to server
// the requests to create defense upgrade actions.
//
// Returns the handler to execute to perform said requests.
func (s *server) registerDefenseAction() http.HandlerFunc {
	return s.registerUpgradeAction(
		"planets",
		func(input string, routeTokens []string) (string, error) {
			// Unmarshal the input data into a defense upgrade
			// action and perform the registration through the
			// dedicated function.
			var action data.DefenseUpgradeAction

			err := json.Unmarshal([]byte(input), &action)
			if err != nil {
				return "", err
			}

			// The `routeTokens` should also provide the planet's id
			// so we can override any value provided in the upgrade
			// action.
			if len(routeTokens) > 0 {
				action.PlanetID = routeTokens[0]
			}

			// Create the upgrade action.
			err = s.upgradeAction.CreateDefenseAction(&action)

			return action.ID, err
		},
	)
}
