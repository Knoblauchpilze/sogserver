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

// shipInFleetForDB :
// Specialization of the `model.ShipInFleet` data structure
// which allows to append the missing info to perform the
// insertion of the data to the DB. We basically need to
// add the identifier of the component and the identifier
// of the fleet element.
//
// The `ID` defines the identifier of this fleet component
// ship.
//
// The `FleetCompID` defines the identifier of the fleet
// component describing this ship.
type shipInFleetForDB struct {
	ID          string `json:"id"`
	FleetCompID string `json:"fleet_element"`

	model.ShipInFleet
}

// resourceInFleetForDB :
// Similar to the `shipInFleetForDB` but holds a resource
// and an amount that is carried by a fleet component.
//
// The `FleetCompID` defines the identifier of the fleet
// component to which this resources is associated.
type resourceInFleetForDB struct {
	FleetCompID string `json:"fleet_element"`

	model.ResourceAmount
}

// fleetDesc :
// Convenience structure allowing to regroup information
// fetched for a fleet from a component's description. It
// defines the target of the fleet, the actual fleet and
// whether or not this fleet has been created specifically
// for the component or was already existing.
//
// The `fleet` defines the actual fleet object.
//
// The `target` defines a planet representing the target
// of the fleet. Can either be fetched from the existing
// fleet or created from the component's target.
//
// The `created` boolean defines whether this fleet was
// existing in the DB or was created specifically for the
// component.
type fleetDesc struct {
	fleet   model.Fleet
	target  *model.Planet
	created bool
}

// ErrInvalidFleet :
// Used to indicate that the fleet component provided
// in input could not be analyzed in some way: it is
// usually because fetching some related data failed
// but can also indicate that the input values were
// not correct.
var ErrInvalidFleet = fmt.Errorf("Unable to analyze invalid action")

// ErrImpossibleFleet :
// Used to indicate that the creation of the fleet
// component cannot be performed on the planet due
// to a lack of ships, fuel or resources.
var ErrImpossibleFleet = fmt.Errorf("Cannot perform creation of fleet component on planet")

// ErrPlayerNotInUniverse :
// Used to indicate that the player's identifier associated
// to a fleet component is not consistent with the universe
// associated to the fleet.
var ErrPlayerNotInUniverse = fmt.Errorf("Invalid player identifier compared to universe")

// ErrComponentAtDestination :
// Used to indicate that a fleet component is actually set
// to start from the destination of the fleet it is meant
// to join. This is not possible.
var ErrComponentAtDestination = fmt.Errorf("Fleet component cannot join fleet directed towards starting position")

// NewFleetProxy :
// Create a new proxy allowing to serve the requests
// related to fleets.
//
// The `data` defines the data model to use to fetch
// information and verify requests.
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

	dbRes, err := p.proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not query DB to fetch fleets (err: %v)", err))
		return []model.Fleet{}, err
	}
	if dbRes.Err != nil {
		p.trace(logger.Error, fmt.Sprintf("Invalid query to fetch fleets (err: %v)", dbRes.Err))
		return []model.Fleet{}, dbRes.Err
	}

	// We now need to retrieve all the identifiers that matched
	// the input filters and then build the corresponding item
	// object for each one of them.
	var ID string
	IDs := make([]string, 0)

	for dbRes.Next() {
		err = dbRes.Scan(&ID)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Error while fetching fleet ID (err: %v)", err))
			continue
		}

		IDs = append(IDs, ID)
	}

	fleets := make([]model.Fleet, 0)

	for _, ID = range IDs {
		uni, err := model.NewReadOnlyFleet(ID, p.data)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Unable to fetch fleet \"%s\" data from DB (err: %v)", ID, err))
			continue
		}

		fleets = append(fleets, uni)
	}

	return fleets, nil
}

// CreateComponent :
// Used to perform the creation of a new component for
// a fleet. The component should describe the player
// willing to join the fleet along with some info about
// the starting planet and the ships involved.
// We will make sure that the player belongs to a uni
// consistent with the desired target. We will also be
// making the necessary adjustments to create the fleet
// that will receive this component if this is not the
// case.
//
// The `comp` defines the fleet component to create.
//
// Returns any error in case the component cannot be
// added to the fleet for some reasons. Returns the
// identifier of the component that was created as
// well.
func (p *FleetProxy) CreateComponent(comp model.Component) (string, error) {
	// Assign a valid identifier if this is not already the case.
	if comp.ID == "" {
		comp.ID = uuid.New().String()
	}

	// Check validity of the input fleet component.
	if !comp.Valid() {
		p.trace(logger.Error, fmt.Sprintf("Failed to validate fleet component's data %s", comp))
		return "", ErrInvalidFleet
	}

	// Acquire the lock on the player associated to this
	// fleet component along with the planet it should
	// start from.
	player, err := model.NewReadWritePlayer(comp.Player, p.data)
	defer func() {
		err := player.Close()
		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Could not release lock on player \"%s\" (err: %v)", player.ID, err))
		}
	}()

	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not fetch player \"%s\" to create component for \"%s\" err: %v)", comp.Player, comp.Fleet, err))
		return "", err
	}

	// Fetch the planet related to this fleet and use
	// it as read write access.
	source, err := model.NewReadWritePlanet(comp.Planet, p.data)
	defer func() {
		err := source.Close()
		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Could not release lock on planet \"%s\" (err: %v)", source.ID, err))
		}
	}()

	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not fetch planet related to fleet component (err: %v)", err))
		return "", ErrInvalidFleet
	}

	// Make sure that the component is not directed towards
	// its started position.
	if comp.Target == source.Coordinates {
		p.trace(logger.Error, fmt.Sprintf("Fleet component is directed towards planet \"%s\" which is its started location at %s", source.ID, comp.Target))
		return "", ErrInvalidFleet
	}

	// Consolidate the arrival time for this component.
	err = comp.ConsolidateArrivalTime(p.data, &source)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not consolidate arrival time for component (err: %v)", err))
		return "", ErrInvalidFleet
	}

	// Fetch the fleet related to this component.
	// Note that in case the component does not
	// yet have a fleet associated to it this is
	// the moment to create it.
	fDesc, err := p.fetchFleetForComponent(&comp, player.Universe)
	defer func() {
		err := fDesc.fleet.Close()
		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Could not release lock on fleet \"%s\" (err: %v)", fDesc.fleet.ID, err))
		}

		if fDesc.target != nil {
			err = fDesc.target.Close()
			if err != nil {
				p.trace(logger.Error, fmt.Sprintf("Could not release lock on planet \"%s\" (err: %v)", fDesc.target.ID, err))
			}
		}
	}()

	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Unable to fetch fleet \"%s\" to create component for \"%s\" (err: %v)", comp.Fleet, comp.Player, err))
		return fDesc.fleet.ID, ErrInvalidFleet
	}

	// Make sure that the target of the fleet is not
	// the source of the component.
	if source.Coordinates == fDesc.fleet.Target {
		p.trace(logger.Error, fmt.Sprintf("Fleet component is starting from the fleet's destination at %s", source.Coordinates))
		return fDesc.fleet.ID, ErrComponentAtDestination
	}

	// Validate the component against planet's data.
	err = comp.Validate(p.data, &source, fDesc.target, &fDesc.fleet)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Cannot create fleet component for \"%s\" from \"%s\" (err: %v)", comp.Player, source.ID, err))
		return fDesc.fleet.ID, ErrImpossibleFleet
	}

	// We can perform the insertion of both the fleet and the
	// component now that we know that both are valid. Note
	// that the insertion of the fleet is only required in
	// case the component is not being attached to an existing
	// one.
	if fDesc.created {
		query := db.InsertReq{
			Script: "create_fleet",
			Args: []interface{}{
				&fDesc.fleet,
			},
		}

		err = p.proxy.InsertToDB(query)

		// Check for errors.
		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Could not create fleet for \"%s\" from \"%s\" (err: %v)", comp.Player, comp.Planet, err))
			return fDesc.fleet.ID, err
		}

		p.trace(logger.Notice, fmt.Sprintf("Created new fleet \"%s\" from component \"%s\"", fDesc.fleet.ID, comp.ID))
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

	// Perform a similar operation on the cargo defined for
	// this component.
	resForDB := make([]resourceInFleetForDB, len(comp.Cargo))

	for id, res := range comp.Cargo {
		resForDB[id] = resourceInFleetForDB{
			FleetCompID:    comp.ID,
			ResourceAmount: res,
		}
	}

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_fleet_component",
		Args: []interface{}{
			&comp,
			shipsForDB,
			resForDB,
			comp.Consumption,
		},
	}

	err = p.proxy.InsertToDB(query)

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not create component for \"%s\" in \"%s\" (err: %v)", comp.Player, fDesc.fleet.ID, err))
		return fDesc.fleet.ID, err
	}

	p.trace(logger.Notice, fmt.Sprintf("Created new fleet component \"%s\" for \"%s\" in \"%s\"", comp.ID, comp.Player, fDesc.fleet.ID))

	return fDesc.fleet.ID, nil
}

// fetchFleetForComponent :
// Used to fetch the fleet related to the input component.
// This might either mean fetch it from the DB in case it
// already exists or create a new fleet if the fleet does
// not exist yet.
//
// The `comp` defines the fleet component for which the
// fleet should be fetched.
//
// The `universe` defines the identifier of the universe
// which is associated to the player owning the input
// fleet component. It will be used in case a fleet for
// the component does not exist yet to fetch some props.
//
// Returns the fleet associated to this component along
// with the target planet (which might be `nil` in case
// the objective of the fleet is compatible with it) and
// any errors.
func (p *FleetProxy) fetchFleetForComponent(comp *model.Component, universe string) (fleetDesc, error) {
	var f fleetDesc
	var err error

	// In case the component has a fleet associated to it
	// we can fetch it through the dedicated handler.
	if comp.Fleet != "" {
		f.fleet, err = model.NewReadWriteFleet(comp.Fleet, p.data)

		if err != nil {
			return f, err
		}

		// Override the information provided by the fleet in
		// the component.
		comp.Objective = f.fleet.Objective
		comp.Target = f.fleet.Target
		comp.Name = f.fleet.Name

		// Register this component as a member of the fleet.
		f.fleet.Comps = append(f.fleet.Comps, *comp)
	}

	// Attempt to retrieve the target planet associated
	// to the fleet. This means first fetching the uni
	// that is associated to the fleet and then the
	// actual planet.
	uni, err := model.NewUniverseFromDB(universe, p.data)
	if err != nil {
		return f, model.ErrInvalidUniverse
	}

	// Retrieve the target planet if needed.
	f.target, err = uni.GetPlanetAt(comp.Target, comp.Player, p.data)
	if err != nil && err != model.ErrPlanetNotFound {
		return f, err
	}

	// In case the fleet was valid, return now.
	if f.fleet.ID != "" {
		return f, nil
	}

	// Otherwise we need to create the fleet and then
	// register it in the DB.
	f.created = true

	planetID := ""
	if err == nil {
		planetID = f.target.ID
	}

	f.fleet, err = model.NewEmptyReadWriteFleet(uuid.New().String(), p.data)
	if err != nil {
		return f, nil
	}

	f.fleet.Name = comp.Name
	f.fleet.Universe = uni.ID
	f.fleet.Objective = comp.Objective
	f.fleet.Target = comp.Target
	f.fleet.Planet = planetID
	f.fleet.ArrivalTime = comp.ArrivalTime
	f.fleet.Comps = []model.Component{
		*comp,
	}

	// Associate the component with the fleet.
	comp.Fleet = f.fleet.ID

	// Make sure the fleet is valid.
	if !f.fleet.Valid(uni) {
		p.trace(logger.Error, fmt.Sprintf("Failed to validate fleet's data %s", f.fleet))
		return f, ErrInvalidFleet
	}

	// Make sure that the objective specified for the
	// fleet is consistent with the ships existing in
	// the component.
	// Indeed as this is the first component that is
	// joining the fleet it *has* to be allowed for
	// the objective. Other elements joining later
	// will be allowed to not contain any ship that
	// can perform the action but the first one is
	// special.
	// That is why no check is performed when the
	// fleet can be found from the component's id.
	err = f.fleet.Validate(p.data)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Failed to validate fleet's data %s (err: %v)", f.fleet, err))
		return f, ErrInvalidFleet
	}

	return f, nil
}
