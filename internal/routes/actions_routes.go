package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/handlers"
	"oglike_server/pkg/logger"
)

// actionCreator :
// Implements the interface requested by the creation handler in
// the `handlers` package. The main functions are describing the
// interface to create a new action from the data fetched from
// the input request.
// As the upgrade action are all very similar in behavior this
// component can be parameterized in order to define the type of
// actions that it can handled. That's why some additional props
// are defined such as the route to serve.
//
// The `route` defines the route that should be served by this
// handler. It should reference to one of the known upgrade
// action types.
//
// The `accessRoute` defines the access route to access facets
// to fetch the action that are produced by this element.
//
// The `proxy` defines the proxy to use to interact with the DB
// when creating the data.
//
// The `log` allows to notify problems and information during a
// universe's creation.
type actionCreator struct {
	route       string
	accessRoute string
	proxy       data.ActionProxy
	log         logger.Logger
}

// Route :
// Implementation of the method to get the route name to create some
// new accounts.
// Returns the name of the route.
func (ac *actionCreator) Route() string {
	return ac.route
}

// AccessRoute :
// Implementation of the method to get the route name to access to
// the data created by this handler. This is basically the `account`
// route.
func (ac *actionCreator) AccessRoute() string {
	return ac.accessRoute
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
	// TODO: Should try to unmarshal in the correct format based on the route.
	var action data.BuildingUpgradeAction
	resources := make([]string, 0)

	// Prevent request with no data.
	if len(input.Data) == 0 {
		return resources, fmt.Errorf("Could not perform creation of building upgrade action with no data")
	}

	for _, rawData := range input.Data {
		// Try to unmarshal the data into a valid `BuildingUpgradeAction` struct.
		err := json.Unmarshal([]byte(rawData), &action)
		if err != nil {
			ac.log.Trace(logger.Error, fmt.Sprintf("Could not create building upgrade action from data \"%s\" (err: %v)", rawData, err))
			continue
		}

		// Create the building action.
		err = ac.proxy.CreateBuildingAction(&action)
		if err != nil {
			ac.log.Trace(logger.Error, fmt.Sprintf("Could not register building upgrade action from data \"%s\" (err: %v)", rawData, err))
			continue
		}

		// Successfully created the action.
		ac.log.Trace(logger.Notice, fmt.Sprintf("Registered action to upgrade \"%s\" to level %d on \"%s\"", action.BuildingID, action.Level, action.PlanetID))
		resources = append(resources, action.ID)
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
			"action/building",
			"actions/buildings",
			s.upgradeAction,
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
			"action/technology",
			"actions/technologies",
			s.upgradeAction,
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
			"action/ship",
			"actions/ships",
			s.upgradeAction,
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
			"action/defense",
			"actions/defenses",
			s.upgradeAction,
			s.log,
		},
		s.log,
	)
}
