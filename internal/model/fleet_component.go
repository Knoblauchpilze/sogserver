package model

import (
	"fmt"
	"oglike_server/pkg/db"
	"time"
)

// Component :
// Defines a single element participating to a fleet. This
// is usually defining a single wave of ship as it was set
// and joined the fleet. This component is composed of a
// certain number of ships and travels at a single speed.
// The most basic fleet has a single component (and must
// have at least one) but large fleets can benefit from
// being split into several parts as they might be able
// to better use their firepower.
//
// The `ID` represents the ID of the fleet component as
// defined in the DB. It allows to uniquely identify it.
//
// The `Player` defines the identifier of the player that
// launched this fleet component. It means that all the
// ships in this component belongs to the player.
//
// The `Planet` defines the identifier of the starting
// planet of this component. Unlike the destination of
// the fleet which might not be a planet yet a fleet
// must start from an existing location.
//
// The `Speed` defines the travel speed of this fleet
// component. It is used to precisely determine how
// much this component impacts the final arrival time
// of the fleet and also for the consumption of fuel.
// The value is in the range `[0, 1]` representing a
// percentage of the maximum available speed.
//
// The `JoinedAt` defines the time at which this player
// has joined the main fleet and created this component.
//
// The `Ships` define the actual ships involved in this
// fleet component.
//
// The `Fleet` is used as an internal value allowing
// to determine to which fleet this component is linked.
//
// The `Consumption` defines the amount of resources
// that is needed for this component to move from its
// starting location to its destination.
//
// The `Cargo` holds the resources moved around by the
// fleet component. It accounts for all the cargo that
// is available on all ships.
//
// The `ArrivalTime` defines the expected arrival time
// of the component to its destination. It should be
// consistent with what's expected by the parent fleet
// and allows to slightly offset the arrival time if
// needed.
type Component struct {
	ID          string           `json:"id"`
	Player      string           `json:"-"`
	Planet      string           `json:"planet"`
	Speed       float32          `json:"speed"`
	JoinedAt    time.Time        `json:"joined_at"`
	Ships       ShipsInFleet     `json:"ships"`
	Fleet       string           `json:"-"`
	Consumption []Consumption    `json:"-"`
	Cargo       []ResourceAmount `json:"-"`
	ArrivalTime time.Time        `json:"-"`
}

// Components :
// Convenience define to refer to a list of fleet
// components.
type Components []Component

// ShipInFleet :
// Defines a single ship involved in a fleet component.
// This is basically the building blocks of fleet. This
// element defines a set of similar ships that joined a
// fleet from a single location (and thus belong to a
// single player).
//
// The `ID` defines the identifier of the ship that is
// involved in the fleet component.
//
// The `Count` defines how many ships of the specified
// type are involved.
type ShipInFleet struct {
	ID    string `json:"ship"`
	Count int    `json:"count"`
}

// ShipsInFleet :
// Convenience define to refer to a list of ships
// belonging to a fleet component. Allows to define
// some methods on this type to ease the consistency
// checks.
type ShipsInFleet []ShipInFleet

// ErrInvalidFleetComponent :
// Used to indicate that the fleet component's identifier
// provided in input is not valid.
var ErrInvalidFleetComponent = fmt.Errorf("Invalid fleet component with no identifier")

// ErrDuplicatedFleetComponent :
// Used to indicate that the fleet component's ID provided
// input is not unique in the DB.
var ErrDuplicatedFleetComponent = fmt.Errorf("Invalid not unique fleet component")

// ErrInsufficientCargo :
// Used to indicate that the fleet component is supposed
// to transport more resources than the available cargo
// space available.
var ErrInsufficientCargo = fmt.Errorf("Insufficient cargo space to hold resources")

// ErrArrivalTimeMismatch :
// Used to indicate that the arrival time computed for a
// given fleet component is incompatible with the time its
// parent fleet should arrive.
var ErrArrivalTimeMismatch = fmt.Errorf("Fleet and component arrival times mismatch")

// valid :
// Used to verify that the ship assigned to a component
// are valid. It should contain a valid ship's ID and a
// non zero amount of ship.
//
// Returns `true` if the ship is valid.
func (sif ShipInFleet) valid() bool {
	return validUUID(sif.ID) && sif.Count > 0
}

// valid :
// Used to perform a chain validation on all the ships
// sets defined in the slice.
//
// Returns `true` if all individual components are
// valid.
func (sifs ShipsInFleet) valid() bool {
	for _, sif := range sifs {
		if !sif.valid() {
			return false
		}
	}

	return len(sifs) > 0
}

// valid :
// Used to perform a chain validation on all the elems
// for this slice.
//
// Returns `true` if all individual components are
// valid.
func (fcs Components) valid() bool {
	for _, comp := range fcs {
		if !comp.Valid() {
			return false
		}
	}

	return true
}

// Valid :
// Used to determine whether the fleet component defined
// by this element is valid or not. We will check that
// the starting coordinate are valid and the each ship
// packet is also valid.
//
// Returns `true` if the component is valid.
func (fc *Component) Valid() bool {
	return validUUID(fc.ID) &&
		validUUID(fc.Player) &&
		validUUID(fc.Planet) &&
		fc.Speed >= 0.0 && fc.Speed <= 1.0 &&
		fc.Ships.valid() &&
		// Allow for either a valid fleet identifier of
		// no identifier at all in case the fleet for
		// this component does not exist yet.
		(fc.Fleet == "" || validUUID(fc.Fleet))
}

// String :
// Implementation of the `Stringer` interface to make
// sure displaying this fleet component is easy.
//
// Returns the corresponding string.
func (fc Component) String() string {
	return fmt.Sprintf("[id: %s, player: %s, planet: %s]", fc.ID, fc.Player, fc.Planet)
}

// consolidateConsumption :
// Used to perform the consolidation of the consumption
// required for this component to take off. It does not
// handle the fact that the parent planet actually has
// the needed fuel but only to compute it.
// The result of the computations will be saved in the
// component itself.
//
// The `data` allows to get information from the DB on
// the consumption of ships.
//
// Returns any error.
func (fc *Component) consolidateConsumption(data Instance) error {
	// TODO: Implement this.
	return fmt.Errorf("Not implemented")
}

// ConsolidateArrivalTime :
// Used to perform the update of the arrival time for
// this fleet component based on the technologies of
// the planet it starts from.
//
// The `data` allows to get information from the DB
// related to the propulsion used by each ships and
// their consumption.
//
// The `p` defines the planet from which this comp
// should start and will be used to update the values
// of the techs that should be used for computing a
// speed for each ship.
//
// Returns any error.
func (fc *Component) ConsolidateArrivalTime(data Instance, p *Planet) error {
	// TODO: Implement this.
	return fmt.Errorf("Not implemented")
}

// Validate :
// Used to make sure that the component can be created
// from the input planet. The check of the number of
// ships required by the component against the actual
// ships count on the planet will be verified along
// with the fuel consumption and resources loading.
//
// The `data` allows to access to the DB if needed.
//
// The `p` defines the planet attached to this action:
// it needs to be provided as input so that resource
// locking is easier.
//
// The `f` defines the parent fleet for this component.
//
// Returns any error.
func (fc *Component) Validate(data Instance, p *Planet, f *Fleet) error {
	// Consistency.
	if fc.Planet != p.ID {
		return ErrInvalidPlanet
	}

	// Update consumption.
	err := fc.consolidateConsumption(data)
	if err != nil {
		return err
	}

	// Make sure that the cargo defined for this fleet
	// component can be stored in the ships.
	totCargo := 0

	for _, ship := range fc.Ships {
		sd, err := data.Ships.getShipFromID(ship.ID)

		if err != nil {
			return err
		}

		totCargo += sd.Cargo
	}

	var totNeeded float32
	for _, res := range fc.Cargo {
		totNeeded += res.Amount
	}

	if totNeeded > float32(totCargo) {
		return ErrInsufficientCargo
	}

	// Check that the arrival time for this component
	// is consistent with what's expected by the fleet.
	// TODO: Relax the constraint to allow fleet offset.
	if fc.ArrivalTime != f.ArrivalTime {
		return ErrArrivalTimeMismatch
	}

	// Validate the amount of fuel available on the
	// planet compared to the amount required and
	// that there are enough resources to be taken
	// from the planet.
	return p.validateComponent(fc.Consumption, fc.Cargo, data)
}

// fetchShips :
// Used to populate the internal data of the fleet
// component with info from the DB. It will assume
// the identifier for this fleet component is at
// least provided.
//
// The `data` allows to access to the DB.
//
// Returns any error.
func (fc *Component) fetchShips(data Instance) error {
	// Check whether the fleet component has an identifier assigned.
	if fc.ID == "" {
		return ErrInvalidFleetComponent
	}

	fc.Ships = make([]ShipInFleet, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"ship",
			"count",
		},
		Table: "fleet_ships",
		Filters: []db.Filter{
			{
				Key:    "fleet_element",
				Values: []string{fc.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}
	if dbRes.Err != nil {
		return dbRes.Err
	}

	// Populate the return value.
	var sif ShipInFleet

	for dbRes.Next() {
		err = dbRes.Scan(
			&sif.ID,
			&sif.Count,
		)

		if err != nil {
			return err
		}

		fc.Ships = append(fc.Ships, sif)
	}

	return nil
}

// Convert :
// Implementation of the `db.Convertible` interface
// from the DB package in order to only include fields
// that need to be marshalled in the fleet component's
// creation.
//
// Returns the converted version of the planet which
// only includes relevant fields.
func (fc *Component) Convert() interface{} {
	return struct {
		ID       string    `json:"id"`
		Fleet    string    `json:"fleet"`
		Player   string    `json:"player"`
		Planet   string    `json:"planet"`
		Speed    float32   `json:"speed"`
		JoinedAt time.Time `json:"joined_at"`
	}{
		ID:       fc.ID,
		Fleet:    fc.Fleet,
		Player:   fc.Player,
		Planet:   fc.Planet,
		Speed:    fc.Speed,
		JoinedAt: fc.JoinedAt,
	}
}
