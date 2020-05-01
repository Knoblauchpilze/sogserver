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
		return []model.Fleet{}, err
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
		return comp.ID, ErrInvalidFleet
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
		return comp.ID, err
	}

	// Fetch the planet related to this fleet and use
	// it as read write access.
	planet, err := model.NewReadWritePlanet(comp.Planet, p.data)
	defer func() {
		err := planet.Close()
		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Could not release lock on planet \"%s\" (err: %v)", planet.ID, err))
		}
	}()

	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not fetch planet related to fleet component (err: %v)", err))
		return comp.ID, ErrInvalidFleet
	}

	// Consolidate the arrival time for this component.
	err = comp.ConsolidateArrivalTime(p.data, &planet)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not consolidate arrival time for component (err: %v)", err))
		return comp.ID, ErrInvalidFleet
	}

	// Fetch the fleet related to this component.
	// Note that in case the component does not
	// yet have a fleet associated to it this is
	// the moment to create it.
	// fleet, err := p.fetchFleetForComponent(comp)
	// defer func() {
	// 	err := fleet.Close()
	// 	if err != nil {
	// 		p.trace(logger.Error, fmt.Sprintf("Could not release lock on fleet \"%s\" (err: %v)", fleet.ID, err))
	// 	}
	// }()

	// if err != nil {
	// 	p.trace(logger.Error, fmt.Sprintf("Unable to fetch fleet \"%s\" to create component for \"%s\" (err: %v)", comp.Fleet, comp.Player, err))
	// 	return comp.ID, ErrInvalidFleet
	// }

	// Validate the component against planet's data.
	// err = comp.Validate(p.data, &planet, &fleet)
	err = comp.Validate(p.data, &planet, nil)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Cannot create fleet component for \"%s\" from \"%s\" (err: %v)", comp.Player, planet.ID, err))
		return comp.ID, ErrImpossibleFleet
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
	query := db.InsertReq{
		Script: "create_fleet_component",
		Args: []interface{}{
			&comp,
			shipsForDB,
		},
	}

	err = p.proxy.InsertToDB(query)

	// Check for errors.
	// if err != nil {
	// 	p.trace(logger.Error, fmt.Sprintf("Could not create component for \"%s\" in \"%s\" (err: %v)", comp.Player, fleet.ID, err))
	// 	return comp.ID, err
	// }

	// p.trace(logger.Notice, fmt.Sprintf("Created new fleet component \"%s\" for \"%s\" in \"%s\"", comp.ID, comp.Player, fleet.ID))

	return comp.ID, nil
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
// Returns the fleet associated to this component along
// with any errors.
func (p *FleetProxy) fetchFleetForComponent(comp model.Component) (model.Fleet, error) {
	// In case the component has a fleet associated to it
	// we can fetch it through the dedicated handler.
	if comp.Fleet != "" {
		return model.NewReadWriteFleet(comp.Fleet, p.data)
	}

	// We need to generate a new fleet from this component.
	// TODO: Handle this.
	return model.Fleet{}, fmt.Errorf("Not implemented")
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
// `nil`. Returns the identifier of the fleet that was
// created as well.
func (p *FleetProxy) Create(fleet model.Fleet) (string, error) {
	// Assign a valid identifier if this is not already the case.
	if fleet.ID == "" {
		fleet.ID = uuid.New().String()
	}

	// First we need to fetch the universe related to the
	// planet to create.
	uni, err := model.NewUniverseFromDB(fleet.Universe, p.data)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Unable to fetch universe \"%s\" to create fleet (err: %v)", fleet.Universe, err))
		return fleet.ID, err
	}

	// Check consistency.
	if !fleet.Valid(uni) {
		p.trace(logger.Error, fmt.Sprintf("Failed to validate fleet's data %s", fleet))
		return fleet.ID, model.ErrInvalidFleet
	}

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_fleet",
		Args:   []interface{}{&fleet},
	}

	err = p.proxy.InsertToDB(query)

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not create fleet in \"%s\" (err: %v)", fleet.Universe, err))
		return fleet.ID, err
	}

	p.trace(logger.Notice, fmt.Sprintf("Created new fleet \"%s\" in \"%s\" targetting \"%s\"", fleet.ID, uni.ID, fleet.Target))

	return fleet.ID, nil
}
