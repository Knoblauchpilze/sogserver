package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// listFleets :
// Used to perform the creation of a handler allowing to serve
// the requests on fleets.
//
// Returns the handler that can be executed to serve said reqs.
func (s *Server) listFleets() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("fleets")

	allowed := map[string]string{
		"fleet_id":     "id",
		"fleet_name":   "name",
		"galaxy":       "target_galaxy",
		"solar_system": "target_solar_system",
		"position":     "target_position",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id").WithModule("fleets")
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.fleets.Fleets(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// createFleet :
// Used to perform the creation of a handler allowing to server
// the requests to create fleets.
//
// Returns the handler to execute to perform said requests.
func (s *server) createFleet() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewCreateResourceEndpoint("fleets")

	// Configure the endpoint.
	ed.WithDataKey("fleet-data")
	ed.WithCreationFunc(
		func(input RouteData) ([]string, error) {
			// We need to iterate over the data retrieved from the route and
			// create fleets from it.
			resources := make([]string, 0)

			// Prevent request with no data.
			if len(input.Data) == 0 {
				return resources, fmt.Errorf("Could not perform creation of fleet with no data")
			}

			// Iterate over the provided data to create the corresponding
			// fleets in the main DB.
			for _, rawData := range input.Data {
				res, err := createFleet(rawData, s.fleets)

				if err != nil {
					s.log.Trace(logger.Error, fmt.Sprintf("Caught error while creating fleet (err: %v)", err))
					continue
				}

				// Successfully created a fleet.
				s.log.Trace(logger.Notice, fmt.Sprintf("Created new fleet \"%s\"", res))
				resources = append(resources, res)
			}

			// Return the path to the resources created during the process.
			return resources, nil
		},
	)

	return ed.ServeRoute(s.log)
}

// createFleetComponent :
// Used to perform the creation of a handler allowing to server
// the requests to create fleet components.
//
// Returns the handler to execute to perform said requests.
func (s *server) createFleetComponent() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewCreateResourceEndpoint("fleets")

	// Configure the endpoint.
	ed.WithDataKey("fleet-data")
	ed.WithCreationFunc(
		func(input RouteData) ([]string, error) {
			// We need to iterate over the data retrieved from the route and
			// create fleets from it.
			resources := make([]string, 0)

			// Prevent request with no data.
			if len(input.Data) == 0 {
				return resources, fmt.Errorf("Could not perform creation of fleet component with no data")
			}

			// Make sure that we can retrieve the identifier of the
			// fleet for which the component should be created from
			// the route's data.
			if len(input.RouteElems) != 2 || input.RouteElems[1] != "components" {
				return resources, fmt.Errorf("Could not perform creation of fleet component, invalid input request")
			}

			fleetID := input.RouteElems[0]

			for _, rawData := range input.Data {
				res, err := createFleetComponent(rawData, fleetID, s.fleets)

				if err != nil {
					s.log.Trace(logger.Error, fmt.Sprintf("Caught error while creating fleet component (err: %v)", err))
					continue
				}

				// Successfully created a fleet component: we should prefix
				// the resource by a `components/` string in order to have
				// consistency with the input route. We should also prefix
				// with the fleet's identifier.
				fullRes := fmt.Sprintf("%s/components/%s", fleetID, res)

				s.log.Trace(logger.Notice, fmt.Sprintf("Created new fleet component \"%s\" for \"%s\"", res, fleetID))
				resources = append(resources, fullRes)
			}

			// Return the path to the resources created during the process.
			return resources, nil
		},
	)

	return ed.ServeRoute(s.log)
}

// createFleet :
// Used to perform the creation of a fleet from the data described
// in input. We will unmarshal the input data into a fleet and then
// call the dedicated handler on the `proxy` to perform the creation
// of the fleet.
// In case the creation cannot be performed an empty string is used
// as return value.
//
// The `raw` represents the data assumed to be a fleet. We will try
// to unmarshal it into a `Fleet` structure and perform the insertion
// in the DB.
//
// The `proxy` defines the fleets proxy to use to request the DB to
// insert the data.
//
// Returns the identifier of the fleet that was created along with
// any errors.
func createFleet(raw string, proxy data.FleetProxy) (string, error) {
	// Try to unmarshal the data into a valid `Fleet` struct.
	var fleet data.Fleet

	err := json.Unmarshal([]byte(raw), &fleet)
	if err != nil {
		return "", fmt.Errorf("Could not create fleet from data \"%s\" (err: %v)", raw, err)
	}

	// Create the fleet.
	err = proxy.Create(&fleet)
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
// The `proxy` defines the fleets proxy to use to request the DB to
// insert the data.
//
// Returns the identifier of the fleet component that was created
// along with any errors.
func createFleetComponent(raw string, fleetID string, proxy data.FleetProxy) (string, error) {
	// Try to unmarshal the data into a valid `FleetComponent` struct.
	var comp data.FleetComponent

	err := json.Unmarshal([]byte(raw), &comp)
	if err != nil {
		return "", fmt.Errorf("Could not create fleet component from data \"%s\" (err: %v)", raw, err)
	}

	// Force the fleet's identifier.
	comp.FleetID = fleetID

	// Create the fleet component.
	err = proxy.CreateComponent(&comp)
	if err != nil {
		return "", fmt.Errorf("Could not register fleet component from data \"%s\" (err: %v)", raw, err)
	}

	return comp.ID, nil
}
