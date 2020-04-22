package data

import (
	"time"
)

// ShipInFleet :
// Defines a single ship involved in a fleet component which
// is an identifier referencing the ship and the amount that
// is directly included in the fleet component.
//
// The `ShipID` defines the identifier of the ship that is
// involved in the fleet component.
//
// The `Amount` defines how many ships of the specified type
// are involved.
type ShipInFleet struct {
	ShipID string `json:"ship"`
	Amount int    `json:"amount"`
}

// FleetComponent :
// Defines a single element participating to a fleet. This is
// the most basic element that can take part into a fleet: it
// is composed of some ships belonging to a single player. We
// also provide the information about the starting position
// of the ships and the player that launched the fleet.
//
// The `ID` represents the identifier of the fleet component
// as defined in the DB. It allows to uniquely identify it.
//
// The `FleetID` defines the identifier of the parent fleet
// this component is attached to.
//
// The `PlayerID` defines the identifier of the account that
// launched this fleet component.
//
// The `Galaxy` defines the start coordinate of this fleet
// component. This *must* refer to an actual planet or moon
// and is kept as a single value in order to allow easy
// integration with the DB.
//
// The `System` refines the starting coordinates of the fleet
// component.
//
// The `Position` defines the position within the parent
// system this fleet componentn started from.
//
// The `Speed` defines the travel speed of this fleet
// component. It is used to precisely determine how much
// this component impacts the final arrival time of the
// fleet and also for the consumption of fuel.
//
// The `JoinedAt` defines the time at which this player has
// joined the main fleet and created this fleet component.
//
// The `Ships` define the actual ships involved in this
// fleet component.
type FleetComponent struct {
	ID       string        `json:"id"`
	FleetID  string        `json:"fleet"`
	PlayerID string        `json:"player"`
	Galaxy   int           `json:"start_galaxy"`
	System   int           `json:"start_solar_system"`
	Position int           `json:"start_position"`
	Speed    float32       `json:"speed"`
	JoinedAt time.Time     `json:"joined_at"`
	Ships    []ShipInFleet `json:"ships"`
}

// Fleet :
// Defines a fleet with its objective and coordinates. It also
// defines the possible name of the fleet.
//
// The `ID` represents a way to uniquely identify the fleet.
//
// The `Name` defines the name that the user provided when the
// fleet was created. It might be empty in case no name was
// provided.
//
// The `UniverseID` defines the identifier of the universe
// this fleet belongs to. Indeed a fleet is linked to some
// coordinates which are linked to a universe. It also is
// used to make sure that only players of this universe can
// participate in the fleet.
//
// The `Objective` is a string defining the action intended
// for this fleet. It is a way to determine which purpose the
// fleet serves.
//
// The `Gakaxy` defines the galaxy of the target this fleet
// is directed towards. It is kept as a single element and
// not transformed into a `Coordinate` object in order to
// allow easy integration with the DB.
// Note that this does not necessarily reference a planet.
//
// The `System` completes the information of the `Galaxy`
// to refine the destination of the fleet.
//
// The `Position` defines the position in the destination
// solar system this fleet is directed towards.
//
// The `ArrivalTime` describes the time at which the fleet
// is meant to reach its destination without taking into
// account the potential delays.
type Fleet struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	UniverseID  string    `json:"uni"`
	Objective   string    `json:"objective"`
	Galaxy      int       `json:"target_galaxy"`
	System      int       `json:"target_solar_system"`
	Position    int       `json:"target_position"`
	ArrivalTime time.Time `json:"arrival_time"`
}

// ProductionEffect :
// Defines a production effect that a building upgrade
// action can have on the production of a planet. It is
// used to regroup the resource and the value of the
// change brought by the building upgrade action.
//
// The `Action` defines the identifier of the action to
// which this effect is linked.
//
// The `Resource` defines the resource which is changed
// by the building upgrade action.
//
// The `Effect` defines the actual effect of the upgrade
// action. This value should be substituted to the planet
// production if the upgrade action completes.
type ProductionEffect struct {
	Action   string  `json:"action"`
	Resource string  `json:"res"`
	Effect   float32 `json:"new_production"`
}

// StorageEffect :
// Defines a storage effect that a building upgrade
// action can have on the capacity of a resource that
// can be stored on a planet. It is used to regroup
// the resource and the value of the change brought
// by the building upgrade action.
//
// The `Action` defines the identifier of the action
// to which this effect is linked.
//
// The `Resource` defines the resource which is changed
// by the building upgrade action.
//
// The `Effect` defines the actual effect of the upgrade
// action. This value should be substituted to the planet
// storage capacity for the resource if the upgrade
// action completes.
type StorageEffect struct {
	Action   string  `json:"action"`
	Resource string  `json:"res"`
	Effect   float32 `json:"new_storage_capacity"`
}
