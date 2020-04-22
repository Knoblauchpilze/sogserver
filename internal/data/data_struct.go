package data

import (
	"time"
)

// Account :
// Defines a player's account within the OG context. It is
// not related to any universe and defines what could be
// called the root account for each player. It is then used
// each time the user wants to join a new universe so as to
// merge all these accounts in a single entity.
//
// The `ID` defines the identifier of the player, which is
// used to uniquely distinguish between two accounts.
//
// The `Name` describes the user provided name for this
// account. It can be duplicated among several accounts
// as we're using the identifier to guarantee uniqueness.
//
// The `Mail` defines the email address associated to the
// account. It can be used to make sure that no two accounts
// share the same address.
//
// The `Password` defines the password that the user should
// enter to grant access to the account.
type Account struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Mail     string `json:"mail"`
	Password string `json:"password"`
}

// Universe :
// Define a universe in terms of OG semantic. This is a set
// of planets gathered in a certain number of galaxies and
// a set of parameters that configure the economic, combat
// and technologies available in it.
//
// The `ID` defines the unique identifier for this universe.
//
// The `Name` defines a human-redable name for it.
//
// The `EcoSpeed` is a value in the range `[0; inf]` which
// defines a multiplication factor that is added to shorten
// the economy (i.e. building construction time, etc.).
//
// The `FleetSpeed` is similar to the `EcoSpeed` but controls
// the speed boost for fleets travel time.
//
// The `ResearchSpeed` controls how researches are shortened
// compared to the base value.
//
// The `FleetsToRuins` defines the percentage of resources
// that go into a debris fields when a ship is destroyed in
// a battle.
//
// The `DefensesToRuins` defines a similar percentage for
// defenses in the event of a battle.
//
// The `FleetConsumption` is a value in the range `[0; 1]`
// defining how the consumption is biased compared to the
// canonical value.
//
// The `GalaxiesCount` defines the number of galaxies in
// the universe.
//
// The `GalaxySize` defines the number of solar systems
// in a single galaxy.
//
// The `SolarSystemSize` defines the number of planets in
// each solar system of each galaxy.
type Universe struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	EcoSpeed         int     `json:"economic_speed"`
	FleetSpeed       int     `json:"fleet_speed"`
	ResearchSpeed    int     `json:"research_speed"`
	FleetsToRuins    float32 `json:"fleets_to_ruins_ratio"`
	DefensesToRuins  float32 `json:"defenses_to_ruins_ratio"`
	FleetConsumption float32 `json:"fleets_consumption_ratio"`
	GalaxiesCount    int     `json:"galaxies_count"`
	GalaxySize       int     `json:"galaxy_size"`
	SolarSystemSize  int     `json:"solar_system_size"`
}

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
