package data

import (
	"fmt"
	"oglike_server/internal/game"
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
var ErrMismatchBetweenTargetCoordsAndID = fmt.Errorf("target's coordinates not consistent with ID for fleet")

// ErrPlayerDoesNotOwnSource : Indicates that the player for a fleet does not own the source.
var ErrPlayerDoesNotOwnSource = fmt.Errorf("player does not own source of a fleet")

// ErrUniverseMismatchForFleet : Indicates that one ore more element of the fleet do not
// belong to the fleet's universe.
var ErrUniverseMismatchForFleet = fmt.Errorf("universe specified for fleet does not match the components")

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
func NewFleetProxy(data game.Instance, log logger.Logger) FleetProxy {
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
// The `filters` define some filtering properties that
// can be applied to the SQL query to only select part
// of all the fleets available. Each one is appende
// `as-is` to the query.
//
// Returns the list of fleets registered in the DB and
// matching the input list of filters. In case the error
// is not `nil` the value of the array should be ignored.
func (p *FleetProxy) Fleets(filters []db.Filter) ([]game.Fleet, error) {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"f.id",
		},
		Table:   "fleets f left join fleets_acs_components fac on f.id = fac.fleet",
		Filters: filters,
	}

	dbRes, err := p.data.Proxy.FetchFromDB(query)

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not query DB to fetch fleets (err: %v)", err))
		return []game.Fleet{}, err
	}
	defer dbRes.Close()

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

// ACSFleets :
// Return a list of ACS fleets registered so far in
// the DB. The input filters allow to only retrieve
// parts of the total fleets.
//
// The `filters` define some filtering properties to
// apply when querying the fleets.
//
// Returns the list of ACS fleets registered in the
// DB and matching the filters along with any error.
func (p *FleetProxy) ACSFleets(filters []db.Filter) ([]game.ACSFleet, error) {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table:   "fleets_acs",
		Filters: filters,
	}

	dbRes, err := p.data.Proxy.FetchFromDB(query)

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not query DB to fetch ACS fleets (err: %v)", err))
		return []game.ACSFleet{}, err
	}
	defer dbRes.Close()

	if dbRes.Err != nil {
		p.trace(logger.Error, fmt.Sprintf("Invalid query to fetch ACS fleets (err: %v)", dbRes.Err))
		return []game.ACSFleet{}, dbRes.Err
	}

	// Build objects from their ID.
	var ID string
	IDs := make([]string, 0)

	for dbRes.Next() {
		err = dbRes.Scan(&ID)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Error while fetching ACS fleet ID (err: %v)", err))
			continue
		}

		IDs = append(IDs, ID)
	}

	fleets := make([]game.ACSFleet, 0)

	for _, ID = range IDs {
		acs, err := game.NewACSFleetFromDB(ID, p.data)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Unable to fetch ACS fleet \"%s\" data from DB (err: %v)", ID, err))
			continue
		}

		fleets = append(fleets, acs)
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
	// Validate the fleet's data.
	_, err := p.validateFleet(&fleet)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Validation for fleet failed (err: %v)", err))
		return fleet.ID, err
	}

	// Import the fleet to the DB.
	err = fleet.SaveToDB(p.data.Proxy)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not create fleet for \"%s\" for \"%s\" (err: %v)", fleet.ID, fleet.Player, err))
		return fleet.ID, err
	}

	p.trace(logger.Notice, fmt.Sprintf("Created new fleet \"%s\" for \"%s\"", fleet.ID, fleet.Player))

	return fleet.ID, nil
}

// CreateACSFleet :
// Used to perform the creation of the input fleet
// along with its associated ACS. The only diff with
// the previous method is that we expect the fleet
// to be part of an ACS fleet.
// The objective and target will be derived from the
// information described by the fleet.
//
// The `fleet` defines the first component of the
// ACS fleet to create.
//
// Returns any error in case the ACS fleet cannot
// be created. Returns the identifier of the ACS
// fleet that was created.
func (p *FleetProxy) CreateACSFleet(fleet game.Fleet) (string, error) {
	// Using this endpoint can be done with two main
	// intents:
	//  - either create an ACS fleet from scratch.
	//  - or register a new component for an existing
	//    ACS fleet.
	// We will make the distinction between both by
	// analyzing the `ACS` field of the input `fleet`.
	acs, err := p.fetchOrCreateACS(&fleet)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not fetch ACS operation for fleet (err: %v)", err))
		return acs.ID, err
	}

	source, err := p.validateFleet(&fleet)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Validation for ACS fleet failed (err: %v)", err))
		return acs.ID, err
	}

	// Validate the arrival time expected by the
	// new component of the fleet (if any).
	err = acs.ValidateFleet(&fleet, source, p.data)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Failed to validate ACS fleet's data (err: %v)", err))
		return acs.ID, err
	}

	err = acs.SaveToDB(&fleet, p.data.Proxy)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not create ACS fleet for \"%s\" (err: %v)", fleet.Player, err))
		return acs.ID, err
	}

	p.trace(logger.Notice, fmt.Sprintf("Created new fleet \"%s\" for \"%s\" in ACS \"%s\"", fleet.ID, fleet.Player, acs.ID))

	return acs.ID, nil
}

// validateFleet :
// Used to perform the validation of the input fleet
// against the data existing in the DB. The fleet is
// directly modified and any error is returned if a
// property seems off.
// The goal is to compute the internal properties of
// the fleet such as its arrival time and consumption.
//
// The `fleet` represents the element to validate.
//
// Returns any error along with th esource planet of
// the fleet.
func (p *FleetProxy) validateFleet(fleet *game.Fleet) (*game.Planet, error) {
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
		return nil, game.ErrInvalidUniverseForFleet
	}

	// Check validity of the input fleet.
	if err := fleet.Valid(uni); err != nil {
		p.trace(logger.Error, fmt.Sprintf("Failed to validate fleet's data (err: %v)", err))
		return nil, err
	}

	// In order to validate and import the fleet's data
	// to the DB we need to retrieve the source planet
	// or moon and the target element (if it exists).
	source, err := game.NewPlanetFromDB(fleet.Source, p.data)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Unable to fetch source for fleet (err: %v)", err))
		return &source, game.ErrInvalidSourceForFleet
	}

	var target *game.Planet
	if fleet.Target != "" {
		vTarget, err := game.NewPlanetFromDB(fleet.Target, p.data)
		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Unable to fetch target for fleet (err: %v)", err))
			return &source, game.ErrInvalidTargetForFleet
		}

		target = &vTarget
	}

	// Check that the source and the target are distinct and
	// that the target's coordinates are consistent with the
	// actual values defined in the input data.
	if target != nil && fleet.TargetCoords != target.Coordinates {
		return &source, ErrMismatchBetweenTargetCoordsAndID
	}
	if source.Player != fleet.Player {
		return &source, ErrPlayerDoesNotOwnSource
	}

	// Consolidate the universe's identifier from the fleet's
	// data. We will force the fleet's universe to match its
	// source's player and check that the target is consistent
	// with it.
	player, err := game.NewPlayerFromDB(fleet.Player, p.data)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Unable to fetch player for fleet (err: %v)", err))
		return &source, game.ErrInvalidPlayerForFleet
	}

	if player.Universe != fleet.Universe {
		return &source, ErrUniverseMismatchForFleet
	}

	if target != nil {
		tPlayer, err := game.NewPlayerFromDB(target.Player, p.data)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Unable to fetch target player for fleet (err: %v)", err))
			return &source, game.ErrInvalidTargetForFleet
		}

		if tPlayer.Universe != fleet.Universe {
			p.trace(logger.Error, fmt.Sprintf("Target player of fleet belongs to universe \"%s\" not consistent with fleet's \"%s\"", tPlayer.Universe, fleet.Universe))
			return &source, game.ErrInvalidTargetForFleet
		}
	}

	mul, err := game.NewMultipliersFromDB(fleet.Universe, p.data)
	if err != nil {
		return &source, err
	}

	// Consolidate the arrival time for this fleet.
	err = fleet.ConsolidateArrivalTime(p.data, &source, mul.Fleet)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not consolidate arrival time for fleet (err: %v)", err))
		return &source, err
	}

	// Validate the fleet against planet's data.
	err = fleet.Validate(p.data, &source, target)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Unable to validate fleet's data for \"%s\" (err: %v)", fleet.Player, err))
		return &source, err
	}

	return &source, nil
}

// fetchOrCreateACS :
// Used to perform the creation of the ACS fleet
// associated to the input fleet if needed or to
// retrieve it if it already exists.
//
// The `fleet` defines the fleet for which the
// ACS should be fetched.
//
// Returns the ACS fleet along with any error.
func (p *FleetProxy) fetchOrCreateACS(fleet *game.Fleet) (*game.ACSFleet, error) {
	// In case the fleet does not have any info
	// of an existing ACS we will create one from
	// scratch.
	if fleet.ACS == "" {
		acs := game.NewACSFleet(fleet)

		if err := acs.Valid(); err != nil {
			return &acs, err
		}

		return &acs, nil
	}

	// Fetch the ACS from the DB.
	acs, err := game.NewACSFleetFromDB(fleet.ACS, p.data)
	return &acs, err
}
