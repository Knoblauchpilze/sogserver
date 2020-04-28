package model

import (
	"fmt"
	"oglike_server/pkg/db"
	"time"
)

// Fleet :
// Defines a fleet in the OG context. A fleet is composed of
// several ships, that can be grouped in distinct components.
// Several players can join a fleet meaning that the starting
// point of the fleet cannot be expressed as a single coord.
// However a fleet always has a single destination which is
// reached by all the components at the same time.
//
// The `ID` represents a way to uniquely identify the fleet.
//
// The `Name` defines the name that the user provided when
// the fleet was created. It might be empty in case no name
// was provided.
//
// The `Universe` defines the identifier of the universe this
// fleet belongs to. Indeed a fleet is linked to some coords
// which are linked to a universe. It also is used to make
// sure that only players of this universe can participate in
// the fleet.
//
// The `Objective` is a string defining the action intended
// for this fleet. It is a way to determine which purpose
// the fleet serves. This string represents an identifier
// in the objectives description table.
//
// The `Target` defines the destination coordinates of the
// fleet. Note that depending on the objective of the fleet
// it might not always refer to an existing planet.
//
// The `ArrivalTime` describes the time at which the fleet
// is meant to reach its destination without taking into
// account the potential delays.
//
// The `Comps` defines the list of components defining the
// fleet. Each component correspond to some ships starting
// from a single location and travelling as a single unit.
type Fleet struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Universe    string     `json:"universe"`
	Objective   string     `json:"objective"`
	Target      Coordinate `json:"target"`
	ArrivalTime time.Time  `json:"arrival_time"`
	Comps       Components `json:"components"`
}

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
// The `Origin` defines the coordinates from which the
// component was launched.
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
type Component struct {
	ID       string       `json:"id"`
	Player   string       `json:"player"`
	Origin   Coordinate   `json:"coordinates"`
	Speed    float32      `json:"speed"`
	JoinedAt time.Time    `json:"joined_at"`
	Ships    ShipsInFleet `json:"ships"`
	Fleet    string       `json:"-"`
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

// ErrInvalidFleet :
// Used to indicate that the fleet provided in input is
// not valid.
var ErrInvalidFleet = fmt.Errorf("Invalid fleet with no identifier")

// ErrDuplicatedFleet :
// Used to indicate that the fleet's identifier provided
// input is not unique in the DB.
var ErrDuplicatedFleet = fmt.Errorf("Invalid not unique fleet")

// ErrInvalidFleetComponent :
// Used to indicate that the fleet component's identifier
// provided in input is not valid.
var ErrInvalidFleetComponent = fmt.Errorf("Invalid fleet component with no identifier")

// ErrDuplicatedFleetComponent :
// Used to indicate that the fleet component's ID provided
// input is not unique in the DB.
var ErrDuplicatedFleetComponent = fmt.Errorf("Invalid not unique fleet component")

// Valid :
// Defines whether this fleet is valid given the bounds
// for the coordinates that are admissible in the uni
// this fleet evolves in.
//
// Returns `true` if the fleet is valid.
func (f *Fleet) Valid(uni Universe) bool {
	return validUUID(f.ID) &&
		validUUID(f.Universe) &&
		validUUID(f.Objective) &&
		f.Target.valid(uni.GalaxiesCount, uni.GalaxySize, uni.SolarSystemSize) &&
		f.Comps.valid(uni)
}

// String :
// Implementation of the `Stringer` interface to make
// sure displaying this fleet is easy.
//
// Returns the corresponding string.
func (f Fleet) String() string {
	return fmt.Sprintf("[id: %s, universe: %s, target: %s]", f.ID, f.Universe, f.Target)
}

// Valid :
// Used to determine whether the fleet component defined
// by this element is valid or not. We will check that
// the starting coordinate are valid and the each ship
// packet is also valid.
//
// The `uni` defines the universe of the fleet which is
// used to verify the starting position of this element.
//
// Returns `true` if the component is valid.
func (fc Component) Valid(uni Universe) bool {
	return validUUID(fc.ID) &&
		validUUID(fc.Player) &&
		fc.Origin.valid(uni.GalaxiesCount, uni.GalaxySize, uni.SolarSystemSize) &&
		fc.Speed >= 0.0 && fc.Speed <= 1.0 &&
		fc.Ships.valid() &&
		validUUID(fc.Fleet)
}

// String :
// Implementation of the `Stringer` interface to make
// sure displaying this fleet component is easy.
//
// Returns the corresponding string.
func (fc Component) String() string {
	return fmt.Sprintf("[id: %s, player: %s, origin: %s]", fc.ID, fc.Player, fc.Origin)
}

// valid :
// Used to perform a chain validation on all the elems
// for this slice.
//
// The `uni` defines the universe to perform the check
// of coordinates for components.
//
// Returns `true` if all individual components are
// valid.
func (fcs Components) valid(uni Universe) bool {
	for _, comp := range fcs {
		if !comp.Valid(uni) {
			return false
		}
	}

	return true
}

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

// NewFleetFromDB :
// Used to fetch the content of the fleet from
// the input DB and populate all internal fields
// from it. In case the DB cannot be fetched or
// some errors are encoutered, the return value
// will include a description of the error.
//
// The `ID` defines the ID of the fleet to get.
// It is fetched from the DB and should refer
// to an existing fleet.
//
// The `data` allows to actually perform the DB
// requests to fetch the fleet's data.
//
// Returns the fleet as fetched from the DB along
// with any errors.
func NewFleetFromDB(ID string, data Instance) (Fleet, error) {
	// Create the fleet.
	f := Fleet{
		ID: ID,
	}

	// Consistency.
	if f.ID == "" {
		return f, ErrInvalidFleet
	}

	// Fetch the fleet's content.
	err := f.fetchGeneralInfo(data)
	if err != nil {
		return f, err
	}

	err = f.fetchComponents(data)
	if err != nil {
		return f, err
	}

	return f, nil
}

// fetchGeneralInfo :
// Used internally when building a fleet from the
// DB to retrieve general information such as the
// objective and target of the fleet.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (f *Fleet) fetchGeneralInfo(data Instance) error {
	// Consistency.
	if f.ID == "" {
		return ErrInvalidFleet
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"name",
			"uni",
			"objective",
			"arrival_time",
			"target_galaxy",
			"target_solar_system",
			"target_position",
		},
		Table: "fleets",
		Filters: []db.Filter{
			{
				Key:    "id",
				Values: []string{f.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}

	// Scan the fleet's data.
	var g, s, p int

	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return ErrInvalidFleet
	}

	err = dbRes.Scan(
		&f.Name,
		&f.Universe,
		&f.Objective,
		&f.ArrivalTime,
		&g,
		&s,
		&p,
	)

	f.Target = NewCoordinate(g, s, p)

	// Make sure that it's the only fleet.
	if dbRes.Next() {
		return ErrDuplicatedFleet
	}

	return err
}

// fetchComponents :
// Used to fetch data related to a fleet: this is
// meant to represent all the individual components
// of the fleest meaning the various waves of ships
// that have been created by the players that joined
// the fleet.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (f *Fleet) fetchComponents(data Instance) error {
	// Consistency.
	if f.ID == "" {
		return ErrInvalidFleet
	}

	f.Comps = make([]Component, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
			"player",
			"start_galaxy",
			"start_solar_system",
			"start_position",
			"speed",
			"joined_at",
		},
		Table: "fleet_elements",
		Filters: []db.Filter{
			{
				Key:    "fleet",
				Values: []string{f.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}

	// Populate the return value.
	var comp Component
	var g, s, p int

	for dbRes.Next() {
		err = dbRes.Scan(
			&comp.ID,
			&comp.Player,
			&g,
			&s,
			&p,
			&comp.Speed,
			&comp.JoinedAt,
		)

		if err != nil {
			return err
		}

		comp.Origin = NewCoordinate(g, s, p)
		comp.Fleet = f.ID

		err = comp.fetchShips(data)
		if err != nil {
			return err
		}

		f.Comps = append(f.Comps, comp)
	}

	return nil
}

// Convert :
// Implementation of the `db.Convertible` interface
// from the DB package in order to only include fields
// that need to be marshalled in the fleet's creation.
//
// Returns the converted version of the planet which
// only includes relevant fields.
func (f *Fleet) Convert() interface{} {
	return struct {
		ID          string    `json:"id"`
		Name        string    `json:"name"`
		Universe    string    `json:"uni"`
		Objective   string    `json:"objective"`
		Galaxy      int       `json:"target_galaxy"`
		System      int       `json:"target_solar_system"`
		Position    int       `json:"target_position"`
		ArrivalTime time.Time `json:"arrival_time"`
	}{
		ID:          f.ID,
		Name:        f.Name,
		Universe:    f.Universe,
		Objective:   f.Objective,
		Galaxy:      f.Target.Galaxy,
		System:      f.Target.System,
		Position:    f.Target.Position,
		ArrivalTime: f.ArrivalTime,
	}
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
		Galaxy   int       `json:"start_galaxy"`
		System   int       `json:"start_solar_system"`
		Position int       `json:"start_position"`
		Speed    float32   `json:"speed"`
		JoinedAt time.Time `json:"joined_at"`
	}{
		ID:       fc.ID,
		Fleet:    fc.Fleet,
		Player:   fc.Player,
		Galaxy:   fc.Origin.Galaxy,
		System:   fc.Origin.System,
		Position: fc.Origin.Position,
		Speed:    fc.Speed,
		JoinedAt: fc.JoinedAt,
	}
}
