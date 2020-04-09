package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/handlers"
	"oglike_server/pkg/logger"
)

// mode:
// Describes the possible fleet creation modes. We either
// want to create a fleet or a fleet component.
type mode string

const (
	fleetMode          mode = "fleet"
	fleetComponentMode mode = "fleet component"
)

// listFleets :
// Used to perform the creation of a handler allowing to serve
// the requests on fleets.
//
// Returns the handler that can be executed to serve said reqs.
func (s *server) listFleets() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewEndpointDesc("fleets")

	allowed := map[string]string{
		"fleet_id":     "id",
		"fleet_name":   "name",
		"galaxy":       "target_galaxy",
		"solar_system": "target_solar_system",
		"position":     "target_position",
	}

	// TODO: For now we don't server fleet components anymore:
	// This was checked before in the `ParseFilters` method as
	// the following test:
	// `len(vars.RouteElems) == 2 && vars.RouteElems[1] == "components"`
	// It should now be restored in the `generic_resource_getter`
	// in the `extractFilters` method. *Or* we could implement
	// the regular expression routes and detect that the route
	// defined by `/fleets/fleet-id/components` better matches
	// the request than the `/fleets` route.

	// Configure the endpoint.
	ed.WithFilters(allowed).WithIDFilter("id")
	ed.WithDataFunc(
		func(filters []data.DBFilter) (interface{}, error) {
			// TODO: Should restore the:
			// `return fa.proxy.FleetComponents(dbFilters)`
			// method when we restore the fleet components
			// route.
			return s.fleets.Fleets(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// fleetCreator :
// Implements the interface requested by the creation handler in
// the `handlers` package. The main functions are describing the
// interface to create a new fleet from the data fetched from the
// input request.
//
// The `proxy` defines the proxy to use to interact with the DB
// when creating the data.
//
// The `log` allows to notify problems and information during a
// fleet's creation.
type fleetCreator struct {
	proxy data.FleetProxy
	log   logger.Logger
}

// Route :
// Implementation of the method to get the route name to create some
// new fleets.
// Returns the name of the route.
func (fc *fleetCreator) Route() string {
	return "fleets"
}

// DataKey :
// Implementation of the method to get the name of the key used to
// pass data to the server.
// Returns the name of the key.
func (fc *fleetCreator) DataKey() string {
	return "fleet-data"
}

// createFleet :
// Used to perform the creation of a fleet from the data described
// in input. We will unmarshal the input data into a fleet and then
// call the dedicated handler on the `fleetProxy` to perform the
// creation of the fleet.
// In case the creation cannot be performed an empty string is used
// as return value.
//
// The `raw` represents the data assumed to be a fleet. We will try
// to unmarshal it into a `Fleet` structure and perform the insertion
// in the DB.
//
// Returns the identifier of the fleet that was created along with
// any errors.
func (fc *fleetCreator) createFleet(raw string) (string, error) {
	// Try to unmarshal the data into a valid `Fleet` struct.
	var fleet data.Fleet

	err := json.Unmarshal([]byte(raw), &fleet)
	if err != nil {
		return "", fmt.Errorf("Could not create fleet from data \"%s\" (err: %v)", raw, err)
	}

	// Create the fleet.
	err = fc.proxy.Create(&fleet)
	if err != nil {
		return "", fmt.Errorf("Could not register fleet from data \"%s\" (err: %v)", raw, err)
	}

	return fleet.ID, nil
}

// createFleetComponent :
// Similar to the `createFleet` but used to create a fleet component
// rather than a complete fleet. The component should refer to an
// existing fleet which will be verified before inserting the input
// data into the DB.
// The user should provide the fleet identifier linked to this comp
// so as to force it in the input data.
//
// The `raw` represents the data assumed to be a fleet component. We
// will try to unmarshal it into the relevant structure and perform
// the insertion in the DB.
//
// The `fleetID` defines the identifier of the fleet for which this
// component should be created. This will be forced in the data to
// retrieve from the route.
//
// Returns the identifier of the fleet component that was created
// along with any errors.
func (fc *fleetCreator) createFleetComponent(raw string, fleetID string) (string, error) {
	// Try to unmarshal the data into a valid `FleetComponent` struct.
	var comp data.FleetComponent

	err := json.Unmarshal([]byte(raw), &comp)
	if err != nil {
		return "", fmt.Errorf("Could not create fleet component from data \"%s\" (err: %v)", raw, err)
	}

	// Force the fleet's identifier.
	comp.FleetID = fleetID

	// Create the fleet component.
	err = fc.proxy.CreateComponent(&comp)
	if err != nil {
		return "", fmt.Errorf("Could not register fleet component from data \"%s\" (err: %v)", raw, err)
	}

	return comp.ID, nil
}

// Create :
// Implementation of the method to perform the creation of the data
// related to the new fleets. We will use the internal proxy to
// request the DB to create a new fleet.
//
// The `input` represent the data fetched from the input request and
// should contain the properties of the fleets to create.
//
// Return the targets of the created resources along with any error.
func (fc *fleetCreator) Create(input handlers.RouteData) ([]string, error) {
	// We need to iterate over the data retrieved from the route and
	// create fleets from it.
	resources := make([]string, 0)

	// Prevent request with no data.
	if len(input.Data) == 0 {
		return resources, fmt.Errorf("Could not perform creation of fleet with no data")
	}

	// We can have two main scenarios: either the input request asks
	// to create a new fleet or to add a component to an existing
	// fleet. The second case is specified by the fact that the route
	// will include the identifier of the fleet to which the component
	// should be added.
	creationMode := fleetMode
	var fleetID string

	if len(input.RouteElems) == 2 && input.RouteElems[1] == "component" {
		// Update the fleet creation mode and update the identifier of
		// the fleet to make sure that we're creating the component for
		// the right fleet.
		creationMode = fleetComponentMode

		fleetID = input.RouteElems[0]
	}

	var res string
	var err error

	for _, rawData := range input.Data {
		// Perform the creation of either the fleet or the fleet comp
		// depending on what is requested from the input route.
		switch creationMode {
		case fleetMode:
			res, err = fc.createFleet(rawData)
		case fleetComponentMode:
			res, err = fc.createFleetComponent(rawData, fleetID)
		}

		if err != nil {
			fc.log.Trace(logger.Error, fmt.Sprintf("Caught error while creating %s (err: %v)", creationMode, err))
			continue
		}

		// Successfully created an fleet.
		fc.log.Trace(logger.Notice, fmt.Sprintf("Created new %s \"%s\"", creationMode, res))
		resources = append(resources, res)
	}

	// Return the path to the resources created during the process.
	return resources, nil
}

// createFleet :
// Creates a handler allowing to server requests to create new
// fleets in the main DB. This rely on the handler structure
// provided by the `handlers` package which allows to mutualize
// the extraction of the data from the input request and the
// general flow to perform the creation.
//
// Returns the handler which can be executed to perform such
// requests.
func (s *server) createFleet() http.HandlerFunc {
	return handlers.ServeCreationRoute(
		&fleetCreator{
			s.fleets,
			s.log,
		},
		s.log,
	)
}
