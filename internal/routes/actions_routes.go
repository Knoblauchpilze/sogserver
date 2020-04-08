package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/handlers"
	"oglike_server/pkg/logger"
)

// creatorFunc :
// Convenience define which allows to refer to the creation
// process in an abstract way. Indeed we can mutualize lots
// of the creation process except for the actual call to the
// DB where the data is created.
// Using this type allows to provide a function that handles
// the conversion of the input `data` to a valid type and
// calls the adequate handler in the DB.
type creatorFunc func(rawAction []byte) (string, error)

// actionCreator :
// Implements the interface requested by the creation handler in
// the `handlers` package. The main functions are describing the
// interface to create a new action from the data fetched from
// the input request.
// As the upgrade action are all very similar in behavior this
// component can be parameterized in order to define the type of
// actions that it can handled. That's why some additional props
// are defined such as the route to serve or the actual data to
// try to insert in the DB.
//
// The `route` defines the route that should be served by this
// handler. It should reference to one of the known upgrade
// action types.
//
// The `creator` represents the function that is called when
// handling the creation of the `data` in the DB. Indeed all
// the upgrade actions only differ by the fact that they use
// a different method of the `ActionProxy` provided for this
// creator so it makes sense to try to mutualize as much as
// possible the common process and provide ways to configure
// the specific part.
//
// The `log` allows to notify problems and information during
// an action's creation.
type actionCreator struct {
	route   string
	creator creatorFunc
	log     logger.Logger
}

// Route :
// Implementation of the method to get the route name to create some
// new accounts.
// Returns the name of the route.
func (ac *actionCreator) Route() string {
	return ac.route
}

// DataKey :
// Implementation of the method to get the name of the key used to
// pass data to the server.
// Returns the name of the key.
func (ac *actionCreator) DataKey() string {
	return "action-data"
}

// Create :
// Implementation of the method to perform the creation of the data
// related to the new actions. We will use the internal proxy to
// request the DB to create a new action.
//
// The `input` represent the data fetched from the input request and
// should contain the properties of the actions to create.
//
// Return the targets of the created resources along with any error.
func (ac *actionCreator) Create(input handlers.RouteData) ([]string, error) {
	// We need to iterate over the data retrieved from the route and
	// create actions from it.
	resources := make([]string, 0)

	// Prevent request with no data.
	if len(input.Data) == 0 {
		return resources, fmt.Errorf("Could not perform creation of upgrade action with no data")
	}

	for _, rawData := range input.Data {
		// Try to unmarshal the data into a valid struct as needed by the
		// upgrade action and then perform the insertion in the DB. We do
		// request an error indicating if the insertion was successful and
		// the string representing the inserted resource.
		resource, err := ac.creator([]byte(rawData))
		if err != nil {
			ac.log.Trace(logger.Error, fmt.Sprintf("Could not register upgrade action from data \"%s\" (err: %v)", rawData, err))
			continue
		}

		// Successfully created the action.
		resources = append(resources, resource)
	}

	// Return the path to the resources created during the process.
	return resources, nil
}

// registerBuildingAction :
// Creates a handler allowing to server requests to create new
// actions to request an upgrade of a new building on a planet.
// This rely on the handler structure provided by the `handlers`
// package which allows to mutualize the extraction of the data
// from the input request and the general flow to perform the
// creation.
//
// Returns the handler which can be executed to perform such
// requests.
func (s *server) registerBuildingAction() http.HandlerFunc {
	return handlers.ServeCreationRoute(
		&actionCreator{
			"actions/buildings",
			func(rawAction []byte) (string, error) {
				// First unmarshal the input data into a valid struct
				// that can be used to register a building upgrade.
				var ua data.BuildingUpgradeAction

				err := json.Unmarshal([]byte(rawAction), &ua)
				if err != nil {
					return "", fmt.Errorf("Could not create building upgrade action from data \"%s\" (err: %v)", rawAction, err)
				}

				// Create the upgrade action.
				err = s.upgradeAction.CreateBuildingAction(&ua)

				if err != nil {
					return "", fmt.Errorf("Could not register building upgrade action from data \"%v\" (err: %v)", ua, err)
				}

				return ua.ID, nil
			},
			s.log,
		},
		s.log,
	)
}

// registerTechnologyAction :
// Creates a handler allowing to server requests to create new
// actions to request an upgrade of a technology on a player.
// This rely on the handler structure provided by the `handlers`
// package which allows to mutualize the extraction of the data
// from the input request and the general flow to perform the
// creation.
//
// Returns the handler which can be executed to perform such
// requests.
func (s *server) registerTechnologyAction() http.HandlerFunc {
	return handlers.ServeCreationRoute(
		&actionCreator{
			"actions/technologies",
			func(rawAction []byte) (string, error) {
				// First unmarshal the input data into a valid struct
				// that can be used to register a technology upgrade.
				var ua data.TechnologyUpgradeAction

				err := json.Unmarshal([]byte(rawAction), &ua)
				if err != nil {
					return "", fmt.Errorf("Could not create technology upgrade action from data \"%s\" (err: %v)", rawAction, err)
				}

				// Create the upgrade action.
				err = s.upgradeAction.CreateTechnologyAction(&ua)

				if err != nil {
					return "", fmt.Errorf("Could not register technology upgrade action from data \"%v\" (err: %v)", ua, err)
				}

				return ua.ID, nil
			},
			s.log,
		},
		s.log,
	)
}

// registerShipAction :
// Creates a handler allowing to server requests to create new
// actions to request the construction of a new ship on one of
// the planets linked to a player.
// This rely on the handler structure provided by the `handlers`
// package which allows to mutualize the extraction of the data
// from the input request and the general flow to perform the
// creation.
//
// Returns the handler which can be executed to perform such
// requests.
func (s *server) registerShipAction() http.HandlerFunc {
	return handlers.ServeCreationRoute(
		&actionCreator{
			"actions/ships",
			func(rawAction []byte) (string, error) {
				// First unmarshal the input data into a valid struct
				// that can be used to register a ship upgrade.
				var ua data.ShipUpgradeAction

				err := json.Unmarshal([]byte(rawAction), &ua)
				if err != nil {
					return "", fmt.Errorf("Could not create ship upgrade action from data \"%s\" (err: %v)", rawAction, err)
				}

				// Create the upgrade action.
				err = s.upgradeAction.CreateShipAction(&ua)

				if err != nil {
					return "", fmt.Errorf("Could not register ship upgrade action from data \"%v\" (err: %v)", ua, err)
				}

				return ua.ID, nil
			},
			s.log,
		},
		s.log,
	)
}

// registerDefenseAction :
// Creates a handler allowing to server requests to create new
// actions to request the construction of a defense on a planet
// for a given player.
// This rely on the handler structure provided by the `handlers`
// package which allows to mutualize the extraction of the data
// from the input request and the general flow to perform the
// creation.
//
// Returns the handler which can be executed to perform such
// requests.
func (s *server) registerDefenseAction() http.HandlerFunc {
	return handlers.ServeCreationRoute(
		&actionCreator{
			"action/defenses",
			func(rawAction []byte) (string, error) {
				// First unmarshal the input data into a valid struct
				// that can be used to register a defense upgrade.
				var ua data.DefenseUpgradeAction

				err := json.Unmarshal([]byte(rawAction), &ua)
				if err != nil {
					return "", fmt.Errorf("Could not create defense upgrade action from data \"%s\" (err: %v)", rawAction, err)
				}

				// Create the upgrade action.
				err = s.upgradeAction.CreateDefenseAction(&ua)

				if err != nil {
					return "", fmt.Errorf("Could not register defense upgrade action from data \"%v\" (err: %v)", ua, err)
				}

				return ua.ID, nil
			},
			s.log,
		},
		s.log,
	)
}
