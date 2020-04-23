package data

import (
	"fmt"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"

	"github.com/google/uuid"
)

// FleetProxy :
// Intended as a wrapper to access properties of the fleets
// registered in the DB. Usually the user only wants to see
// some specific fleets and thus we provide ways to filter
// the results to only select some of the fleets.
// A fleet's main interest is its destination: it might be
// the case that a fleet does not have any planet as target
// in case of a colony mission for example. In this case it
// makes more sense to refer to a fleet through its target
// rather than a planet.
// As a fleet can also be composed of ships of various
// players it also don't make much sense to link it to a
// specific player.
type FleetProxy struct {
	commonProxy
}

// NewFleetProxy :
// Create a new proxy allowing to serve the requests
// related to fleets.
//
// The `data` defines the data model to use to fetch
// information and verify actions.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewFleetProxy(data model.Instance, log logger.Logger) FleetProxy {
	return FleetProxy{
		commonProxy: newCommonProxy(data, log, "fleets"),
	}
}

// Fleets :
// Return a list of fleets registered so far in the DB.
// The returned list take into account the filters that
// are provided as input to only include the fleets
// matching all the criteria. A full description of the
// fleets is returned, including all its components and
// the ships associated to each one.
//
// The `filters` define some filtering property that can
// be applied to the SQL query to only select part of all
// the fleets available. Each one is appended `as-is` to
// the query.
//
// Returns the list of fleets registered in the DB and
// matching the input list of filters. In case the error
// is not `nil` the value of the array should be ignored.
func (p *FleetProxy) Fleets(filters []db.Filter) ([]model.Fleet, error) {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table:   "fleets",
		Filters: filters,
	}

	res, err := p.proxy.FetchFromDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not query DB to fetch fleets (err: %v)", err))
		return []model.Fleet{}, err
	}

	// We now need to retrieve all the identifiers that matched
	// the input filters and then build the corresponding unis
	// object for each one of them.
	var ID string
	IDs := make([]string, 0)

	for res.Next() {
		err = res.Scan(&ID)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Error while fetching fleet ID (err: %v)", err))
			continue
		}

		IDs = append(IDs, ID)
	}

	fleets := make([]model.Fleet, 0)

	for _, ID = range IDs {
		uni, err := model.NewFleetFromDB(ID, p.data)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Unable to fetch fleet \"%s\" data from DB (err: %v)", ID, err))
			continue
		}

		fleets = append(fleets, uni)
	}

	return fleets, nil
}

// Create :
// Used to perform the creation of the fleet described
// by the input data to the DB. In case the creation can
// not be performed an error is returned.
//
// The `fleet` describes the element to create in DB. Its
// value may be modified by the function mainly to update
// the identifier of the fleet if none have been set.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`.
func (p *FleetProxy) Create(fleet *Fleet) error {
	// Assign a valid identifier if this is not already the case.
	if fleet.ID == "" {
		fleet.ID = uuid.New().String()
	}

	// Fetch the universe related to the fleet to create.
	uni, err := p.fetchUniverse(fleet.UniverseID)
	if err != nil {
		return fmt.Errorf("Could not create fleet \"%s\", unable to fetch universe (err: %v)", fleet.ID, err)
	}

	// Validate that the input data describe a valid universe.
	if !fleet.valid(uni) {
		return fmt.Errorf("Could not create fleet \"%s\", some properties are invalid", fleet.ID)
	}

	// Create the query and execute it.
	query := insertReq{
		script: "create_fleet",
		args:   []interface{}{*fleet},
	}

	err = p.insertToDB(query)

	// Check for errors.
	if err != nil {
		return fmt.Errorf("Could not import universe \"%s\" (err: %v)", uni.Name, err)
	}

	return nil
}

// CreateComponent :
// Used to perform the creation of a new component for a fleet.
// The component should describe the player willing to join the
// fleet along with some information about the starting position
// and the ships involved.
// We will make sure that the player belongs to the rigth uni,
// that the starting position is valid compared to the actual
// dimensions of the universe and that the fleet exists.
//
// The `comp` defines the fleet component to create.
//
// Returns any error in case the component cannot be added to
// the fleet for some reasons.
func (p *FleetProxy) CreateComponent(comp *FleetComponent) error {
	// Assign a valid identifier if this is not already the case.
	if comp.ID == "" {
		comp.ID = uuid.New().String()
	}

	// Fetch the fleet related to this component.
	fleet, err := p.fetchFleet(comp.FleetID)
	if err != nil {
		return fmt.Errorf("Could not create fleet component for fleet \"%s\" (err: %v)", comp.FleetID, err)
	}

	// Fetch the player related to the fleet component.
	player, err := p.fetchPlayer(comp.PlayerID)
	if err != nil {
		return fmt.Errorf("Could not create fleet component for fleet \"%s\" (err: %v)", comp.FleetID, err)
	}

	// Fetch the universe related to the fleet to create.
	uni, err := p.fetchUniverse(fleet.UniverseID)
	if err != nil {
		return fmt.Errorf("Could not create fleet component for fleet \"%s\" (err: %v)", comp.ID, err)
	}

	// Check validity of the input fleet component.
	if !comp.valid(uni) {
		return fmt.Errorf("Could not create fleet component for fleet \"%s\", some properties are invalid", comp.FleetID)
	}

	// In case the player does not belong to the same universe
	// as the fleet, this is a problem.
	if uni.ID != player.UniverseID {
		return fmt.Errorf("Could not create fleet component for fleet \"%s\", player belongs to \"%s\" but fleet is in \"%s\"", comp.FleetID, player.UniverseID, uni.ID)
	}

	// Convert the input ships to something that can be directly
	// inserted into the DB. We need to manually create the ID
	// and assign the identifier of the parent fleet component.
	shipsForDB := make([]shipInFleetForDB, len(comp.Ships))

	for id, ship := range comp.Ships {
		shipsForDB[id] = shipInFleetForDB{
			ID:          uuid.New().String(),
			FleetCompID: comp.ID,
			ShipInFleet: ship,
		}
	}

	// Create the query and execute it.
	query := insertReq{
		script: "create_fleet_component",
		args: []interface{}{
			comp,
			shipsForDB,
		},
	}

	err = p.insertToDB(query)

	// Check for errors.
	if err != nil {
		return fmt.Errorf("Could not import fleet component for fleet \"%s\" (err: %s)", comp.FleetID, err)
	}

	return nil
}
