package data

import (
	"fmt"
	"oglike_server/internal/game"
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

// ErrMismatchBetweenTargetCoordsAndID : Indicates that the coordinates of the target is
// not consistent with the target's identifier.
var ErrMismatchBetweenTargetCoordsAndID = fmt.Errorf("Target's coordinates not consistent with ID for fleet")

// ErrPlayerDoesNotOwnSource : Indicates that the player for a fleet does not own the source.
var ErrPlayerDoesNotOwnSource = fmt.Errorf("Player does not own source of a fleet")

// ErrUniverseMismatchForFleet : Indicates that one ore more element of the fleet do not
// belong to the fleet's universe.
var ErrUniverseMismatchForFleet = fmt.Errorf("Universe specified for fleet does not match the components")

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
// fleets is returned..
//
// The `filters` define some filtering property that can
// be applied to the SQL query to only select part of all
// the fleets available. Each one is appended `as-is` to
// the query.
//
// Returns the list of fleets registered in the DB and
// matching the input list of filters. In case the error
// is not `nil` the value of the array should be ignored.
func (p *FleetProxy) Fleets(filters []db.Filter) ([]game.Fleet, error) {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table:   "fleets",
		Filters: filters,
	}

	dbRes, err := p.data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not query DB to fetch fleets (err: %v)", err))
		return []game.Fleet{}, err
	}
	if dbRes.Err != nil {
		p.trace(logger.Error, fmt.Sprintf("Invalid query to fetch fleets (err: %v)", dbRes.Err))
		return []game.Fleet{}, dbRes.Err
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

	fleets := make([]game.Fleet, 0)

	for _, ID = range IDs {
		f, err := game.NewFleetFromDB(ID, p.data)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Unable to fetch fleet \"%s\" data from DB (err: %v)", ID, err))
			continue
		}

		fleets = append(fleets, f)
	}

	return fleets, nil
}

// CreateFleet :
// Used to perform the creation of a new fleet for a
// player. The input data should describe the player
// willing to create a new fleetalong with some info
// about the starting planet and the ships involved.
// We will make sure that the player belongs to a uni
// consistent with the desired target.
//
// The `fleet` defines the fleet fleet to create.
//
// Returns any error in case the fleet cannot be
// created for some reasons. Returns the identifier
// of the fleet that was created.
func (p *FleetProxy) CreateFleet(fleet game.Fleet) (string, error) {
	// Assign a valid identifier if this is not already the case.
	if fleet.ID == "" {
		fleet.ID = uuid.New().String()
	}

	// In order to make sure that the fleet is valid we
	// have to check the coordinates against the parent
	// universe. This means that we should first assign
	// the universe before calling the `fleet.Valid()`.
	// This might fail in case the universe's ID in the
	// fleet is not valid. We will report this as an
	// error of the fleet in this case.
	uni, err := game.NewUniverseFromDB(fleet.Universe, p.data)
	if err == game.ErrInvalidElementID {
		return fleet.ID, game.ErrInvalidUniverseForFleet
	}

	// Check validity of the input fleet.
	if err := fleet.Valid(uni); err != nil {
		p.trace(logger.Error, fmt.Sprintf("Failed to validate fleet's data (err: %v)", err))
		return fleet.ID, err
	}

	// In order to validate and import the fleet's data
	// to the DB we need to retrieve the source planet
	// or moon and the target element (if it exists).
	source, err := game.NewPlanetFromDB(fleet.Source, p.data)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Unable to fetch source for fleet (err: %v)", err))
		return fleet.ID, game.ErrInvalidSourceForFleet
	}

	var target *game.Planet
	if fleet.Target != "" {
		vTarget, err := game.NewPlanetFromDB(fleet.Target, p.data)
		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Unable to fetch target for fleet (err: %v)", err))
			return fleet.ID, game.ErrInvalidTargetForFleet
		}

		target = &vTarget
	}

	// Check that the source and the target are distinct and
	// that the target's coordinates are consistent with the
	// actual values defined in the input data.
	if target != nil && fleet.TargetCoords != target.Coordinates {
		return fleet.ID, ErrMismatchBetweenTargetCoordsAndID
	}
	if source.Player != fleet.Player {
		return fleet.ID, ErrPlayerDoesNotOwnSource
	}

	// Consolidate the universe's identifier from the fleet's
	// data. We will force the fleet's universe to match its
	// source's player and check that the target is consistent
	// with it.
	player, err := game.NewPlayerFromDB(fleet.Player, p.data)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Unable to fetch player for fleet (err: %v)", err))
		return fleet.ID, game.ErrInvalidPlayerForFleet
	}

	if player.Universe != fleet.Universe {
		return fleet.ID, ErrUniverseMismatchForFleet
	}

	if target != nil {
		tPlayer, err := game.NewPlayerFromDB(target.Player, p.data)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Unable to fetch target player for fleet (err: %v)", err))
			return fleet.ID, game.ErrInvalidTargetForFleet
		}

		if tPlayer.Universe != fleet.Universe {
			p.trace(logger.Error, fmt.Sprintf("Target player of fleet belongs to universe \"%s\" not consistent with fleet's \"%s\"", tPlayer.Universe, fleet.Universe))
			return fleet.ID, game.ErrInvalidTargetForFleet
		}
	}

	// Consolidate the arrival time for this fleet.
	err = fleet.ConsolidateArrivalTime(p.data, &source)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not consolidate arrival time for fleet (err: %v)", err))
		return "", err
	}

	// Validate the fleet against planet's data.
	err = fleet.Validate(p.data, &source, target)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Unable to validate fleet's data for \"%s\" (err: %v)", fleet.Player, err))
		return fleet.ID, err
	}

	// The fleet seems valid, proceed to inserting
	// the data in the DB.
	query := db.InsertReq{
		Script: "create_fleet",
		Args: []interface{}{
			&fleet,
			fleet.Ships,
			fleet.Cargo,
			fleet.Consumption,
		},
	}

	err = p.data.Proxy.InsertToDB(query)

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not create fleet for \"%s\" for \"%s\" (err: %v)", fleet.ID, fleet.Player, err))
		return fleet.ID, err
	}

	p.trace(logger.Notice, fmt.Sprintf("Created new fleet \"%s\" for \"%s\"", fleet.ID, fleet.Player))

	return fleet.ID, nil
}
