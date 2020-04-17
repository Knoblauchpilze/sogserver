package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"

	"github.com/google/uuid"
)

// shipInFleetForDB :
// Specialization of the `ShipInFleet` data structure which
// allows to append the missing info to be easily inserted
// into the DB.
// We basically add an identifier for the fleet component
// of this ship and an identifier.
//
// The `ID` defines the identifier of this fleet component
// ship.
//
// The `FleetCompID` defines the identifier of the fleet
// component describing this ship.
type shipInFleetForDB struct {
	ID          string `json:"id"`
	FleetCompID string `json:"fleet_element"`
	ShipInFleet
}

// FleetProxy :
// Intended as a wrapper to access properties of fleets and
// their components and retrieve data from the database. It
// uses the common proxy defined in this package.
// In addition to the base properties the fleet proxy also
// needs to have access to the universes (so as to verify
// that fleets are not going outside of the bounds provided
// by a universe) and the players (to make sure that owners
// actually exist).
//
// The `uProxy` represents a reference to the proxy allowing
// to access to universes.
//
// The `pProxy` represents a reference to the proxy allowing
// to access to players.
type FleetProxy struct {
	uProxy UniverseProxy
	pProxy PlayerProxy

	commonProxy
}

// NewFleetProxy :
// Create a new proxy allowing to serve the requests
// related to fleets. It uses two other proxies: one
// to access to the universes and the other one to
// access the players. This is used to make sure that
// creating a fleet is consistent with the properties
// of each element.
//
// The `dbase` represents the database to use to fetch
// data related to fleets.
//
// The `log` allows to notify errors and information.
//
// The `unis` defines a way to access to universes as
// registered in the DB.
//
// The `players` defines a way to access to players
// as defined in the DB.
//
// Returns the created proxy.
func NewFleetProxy(dbase *db.DB, log logger.Logger, unis UniverseProxy, players PlayerProxy) FleetProxy {
	return FleetProxy{
		unis,
		players,

		newCommonProxy(dbase, log),
	}
}

// Fleets :
// Allows to fetch the list of fleets currently registered in
// the server
// The user can choose to filter parts of the fleets using an
// array of filters that will be applied to the SQL query.
// No controls is enforced on the filters so one should make
// sure that it's consistent with the underlying table.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// fleets available. Each one is appended `as-is` to the SQL
// query.
//
// Returns the list of fleets along with any errors. Note that
// in case the error is not `nil` the returned list is to be
// ignored.
func (p *FleetProxy) Fleets(filters []DBFilter) ([]Fleet, error) {
	// Create the query and execute it.
	query := queryDesc{
		props: []string{
			"id",
			"name",
			"uni",
			"objective",
			"arrival_time",
			"target_galaxy",
			"target_solar_system",
			"target_position",
		},
		table:   "fleets",
		filters: filters,
	}

	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch fleets (err: %v)", err)
	}

	// Populate the return value.
	fleets := make([]Fleet, 0)
	var fleet Fleet

	for res.next() {
		err = res.scan(
			&fleet.ID,
			&fleet.Name,
			&fleet.UniverseID,
			&fleet.Objective,
			&fleet.ArrivalTime,
			&fleet.Galaxy,
			&fleet.System,
			&fleet.Position,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for fleet (err: %v)", err))
			continue
		}

		fleets = append(fleets, fleet)
	}

	return fleets, nil
}

// FleetComponents :
// Used to fetch data related to a fleet: this includes the
// individual components of the fleet (which is mostly used
// in the case of group attacks).
//
// The `filters` define some properties that are used when
// fetching the parent fleet for which components should be
// retrieved. It is assumed that these filteres allow to
// fetch a single fleet. If this is not the case an error
// is returned.
//
// Returns the components associated to the fleet along with
// any error.
func (p *FleetProxy) FleetComponents(filters []DBFilter) ([]FleetComponent, error) {
	// Fetch the fleet from the filters.
	fleets, err := p.Fleets(filters)
	if err != nil {
		return nil, fmt.Errorf("Could not fetch components for fleet (err: %v)", err)
	}

	// Check that we only found a single fleet matching the
	// input filters: this will be the fleet for which the
	// components should be retrieved.
	if len(fleets) != 1 {
		return nil, fmt.Errorf("Found %d fleet(s) matching filters, cannot fetch components", len(fleets))
	}

	fleet := fleets[0]

	// Create the query and execute it.
	query := queryDesc{
		props: []string{
			"id",
			"player",
			"start_galaxy",
			"start_solar_system",
			"start_position",
			"speed",
			"joined_at",
		},
		table: "fleet_elements",
		filters: []DBFilter{
			{
				Key:    "fleet",
				Values: []string{fleet.ID},
			},
		},
	}

	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch fleet \"%s\" details (err: %v)", fleet.ID, err)
	}

	// Populate the return value.
	components := make([]FleetComponent, 0)
	comp := FleetComponent{
		FleetID: fleet.ID,
	}

	for res.next() {
		err = res.scan(
			&comp.ID,
			&comp.PlayerID,
			&comp.Galaxy,
			&comp.System,
			&comp.Position,
			&comp.Speed,
			&comp.JoinedAt,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for fleet \"%s\" component (err: %v)", fleet.ID, err))
			continue
		}

		err = p.fetchFleetComponentData(&comp)
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch data for fleet \"%s\" (component \"%s\" failed, err: %v)", fleet.ID, comp.PlayerID, err))
			continue
		}

		components = append(components, comp)
	}

	return components, nil
}

// fetchFleetComponentData :
// Used to fetch data related to a fleet component. It includes
// the actual ship involved in the component and their actual
// amount.
// individual components of the fleet (which is mostly used
// in the case of group attacks).
//
// The `com` references the fleet component for which the data
// should be fetched. We assume that the internal fields (and
// more specifically the identifier) are already populated.
//
// Returns any error.
func (p *FleetProxy) fetchFleetComponentData(comp *FleetComponent) error {
	// Check whether the fleet component has an identifier assigned.
	if comp.ID == "" {
		return fmt.Errorf("Unable to fetch data from fleet component with invalid identifier")
	}

	// Create the query and execute it.
	query := queryDesc{
		props: []string{
			"ship",
			"amount",
		},
		table: "fleet_ships",
		filters: []DBFilter{
			{
				Key:    "fleet_element",
				Values: []string{comp.ID},
			},
		},
	}

	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return fmt.Errorf("Could not query DB to fetch fleet component \"%s\" details (err: %v)", comp.ID, err)
	}

	// Populate the return value.
	comp.Ships = make([]ShipInFleet, 0)
	var ship ShipInFleet

	for res.next() {
		err = res.scan(
			&ship.ShipID,
			&ship.Amount,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for fleet component \"%s\" (err: %v)", comp.ID, err))
			continue
		}

		comp.Ships = append(comp.Ships, ship)
	}

	return nil
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

// fetchUniverse :
// Used to fetch the universe from the DB. The input identifier
// is meant to represent a universe registered in the DB. We
// will make sure that it can be fetched and that a single item
// is matching in the DB.
// In case no universe can be found an error is returned.
//
// The `id` defines the index of the universe to fetch.
//
// Returns the universe corresponding to the input identifier
// along with any errors.
func (p *FleetProxy) fetchUniverse(id string) (Universe, error) {
	// Create the db filters from the input identifier.
	// TODO: Maybe regroup these methods (this one and `fetchPlanet`, etc.)
	// in the `commonProxy`.
	filters := make([]DBFilter, 1)

	filters[0] = DBFilter{
		"id",
		[]string{id},
	}

	unis, err := p.uProxy.Universes(filters)

	// Check for errors and cases where we retrieve several
	// universes.
	if err != nil {
		return Universe{}, err
	}
	if len(unis) != 1 {
		return Universe{}, fmt.Errorf("Retrieved %d universes for id \"%s\" (expected 1)", len(unis), id)
	}

	return unis[0], nil
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

// fetchFleet :
// Used to fetch the fleet described by the input identifier in
// the DB. It is used when a fleet component should be added to
// an existing fleet. We will check that there is one and only
// one fleet matching the input identifier to return it.
//
// The `id` defines the identifier of the fleet to fetch.
//
// Returns the fleet corresponding to the identifier along with
// any error.
func (p *FleetProxy) fetchFleet(id string) (Fleet, error) {
	// Create the db filters from the input identifier.
	filters := make([]DBFilter, 1)

	filters[0] = DBFilter{
		"id",
		[]string{id},
	}

	fleets, err := p.Fleets(filters)

	// Check for errors and cases where we retrieve several
	// fleets.
	if err != nil {
		return Fleet{}, err
	}
	if len(fleets) != 1 {
		return Fleet{}, fmt.Errorf("Retrieved %d fleets for id \"%s\" (expected 1)", len(fleets), id)
	}

	return fleets[0], nil
}

// fetchPlayer :
// Used to fetch the player described by the input identifier in
// the internal DB. It is mostly used to check consistency when
// performing the creation of a new fleet component for a fleet.
//
// The `id` defines the identifier of the player to fetch.
//
// Returns the player corresponding to the identifier along with
// any error.
func (p *FleetProxy) fetchPlayer(id string) (Player, error) {
	// Create the db filters from the input identifier.
	filters := make([]DBFilter, 1)

	filters[0] = DBFilter{
		"id",
		[]string{id},
	}

	players, err := p.pProxy.Players(filters)

	// Check for errors and cases where we retrieve several
	// players.
	if err != nil {
		return Player{}, err
	}
	if len(players) != 1 {
		return Player{}, fmt.Errorf("Retrieved %d players for id \"%s\" (expected 1)", len(players), id)
	}

	return players[0], nil
}
