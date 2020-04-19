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

// Coordinate :
// Defines what is a coordinate in the og server context. It
// allows to locate a planet within its universe, galaxy, and
// finally its solar system.
//
// The `Galaxy` defines the position of the planet within the
// universe. Its value is consistent with the maximum number
// of galaxies in the universe.
//
// The `System` defines the position of the solar system that
// contains the planet within the galaxy. It should be valid
// according to the number of systems defined per galaxy.
//
// The `Position` defines the position of the planet within
// its solar system.
type Coordinate struct {
	Galaxy   int `json:"galaxy"`
	System   int `json:"system"`
	Position int `json:"position"`
}

// ResourceAmount :
// Defines a certain amount of a resource in the game. It is
// basically an association between a resource and an amount
// (defining how much of the resource is needed).
//
// The `Resource` defines the identifier of the resource that
// this association describes.
//
// The `Amount` defines how much of the resource is needed.
type ResourceAmount struct {
	Resource string  `json:"resource"`
	Amount   float32 `json:"amount"`
}

// Building :
// Defines a building that can occupy a slot on a planet. Each
// building has a name and a level (default being `0`).
//
// The `ID` defines the index of the building as defined in the
// internal database and allows to uniquely refer to this type
// of building.
//
// The `Level` defines the current level of the building that
// is built on this planet.
//
// The `Cost` define how much it will cost to upgrade the
// building to the next level.
//
// The `Production` defines how much of each resource the
// current level of the building produces. Note that this
// might be completely empty in case the building is not
// producing anything or negative if the building actually
// consumes a resource to work.
//
// The `ProductionIncrease` defines how much the next level
// will bring to the production level of each resource.
// The production of the next level is given by adding both
// the `Production` and the `ProductionIncrease` for each
// resource.
type Building struct {
	ID                 string           `json:"id"`
	Level              int              `json:"level"`
	Cost               []ResourceAmount `json:"costs"`
	Production         []ResourceAmount `json:"production"`
	ProductionIncrease []ResourceAmount `json:"production_increase"`
}

// Technology :
// Defines a technology in the og context. It defines the
// identifier of the technology which allows to access the
// description of the technology and other information.
//
// The `ID` defines the identifier of the technology.
//
// The `Level` defines the current technology level of this
// technology on the account of a player.
//
// The `Cost` define the amount of each resource needed to
// research the next level of this technology.
type Technology struct {
	ID    string           `json:"id"`
	Level int              `json:"level"`
	Cost  []ResourceAmount `json:"cost"`
}

// Ship :
// Defines a ship within the OG context. Such a ship is defined
// through its identifier which can be used to fetch additional
// information about it (name, speed, etc.) and a count, which
// defines the number of ships currently available.
//
// The `ID` defines the identifier to use to refer to this ship
// and possibly fetch more info about it.
//
// The `Count` defines how many ships of this type are currently
// available on a given planet.
type Ship struct {
	ID    string `json:"id"`
	Count int    `json:"count"`
}

// Defense :
// Defines a defense system that can be built on a planet. The
// default count is `0` and the user could fetch more info on
// this system using the provided unique identifier.
//
// The `ID` defines the unique identifier for this defense
// system. It can be used to fetch more info about it.
//
// The `Count` defines the number of defense of this type that
// are currently available on the planet.
type Defense struct {
	ID    string `json:"id"`
	Count int    `json:"count"`
}

// Planet :
// Define a planet which is an object within a certain universe
// and associated to a certain player. The planet is described
// only has its structure and not its exact content (like ships,
// defenses, etc.).
//
// The `PlayerID` defines the identifier of the player which owns
// this planet. It is relative to an account and a universe.
//
// The `ID` defines the identifier of the planet within all the
// planets registered in og.
//
// The `Galaxy` defines the parent galaxy of the planet. It is
// used as a simple way to marshal the planet and be able to
// import this structure directly into the DB rather than using
// the `Coordinate` type.
//
// The `System` completes the `Galaxy` in the determination of
// the coordinates of the planet.
//
// The `Position` defines the position of the planet within
// its parent solar system.
//
// The `Name` of the planet as defined by the user.
//
// The `Fields` define the number of available fields in the
// planet. The number of used fields is computed from the
// infrastructure built on the planet but is not returned here.
//
// The `MinTemp` defines the minimum temperature of the planet
// in degrees.
//
// The `MaxTemp` defines the maximum temperatue of the planet
// in degrees.
//
// The `Diameter` defines the diameter of the planet expressed
// in kilometers.
//
// The `Resources` define the resources currently stored on the
// planet. This is basically the quantity available to produce
// some buildings, ships, etc.
//
// The `Buildings` defines the list of buildings currently built
// on the planet. Note that it does not provide information on
// buildings *being* built.
//
// The `Ships` defines the list of ships currently deployed on
// this planet. It does not include ships currently moving from
// or towards the planet.
//
// The `Defense` defines the list of defenses currently built
// on the planet. This does not include defenses *being* built.
type Planet struct {
	PlayerID  string     `json:"player"`
	ID        string     `json:"id"`
	Galaxy    int        `json:"galaxy"`
	System    int        `json:"solar_system"`
	Position  int        `json:"position"`
	Name      string     `json:"name"`
	Fields    int        `json:"fields"`
	MinTemp   int        `json:"min_temperature"`
	MaxTemp   int        `json:"max_temperature"`
	Diameter  int        `json:"diameter"`
	Resources []Resource `json:"resources"`
	Buildings []Building `json:"buildings"`
	Ships     []Ship     `json:"ships"`
	Defenses  []Defense  `json:"defenses"`
}

// Player :
// Define a player which is basically a name in a universe.
// We also provide both the identifier of this player along
// with its account index.
//
// The `AccountID` represents the identifier of the accounts
// associated with this player. An account can be registered
// on any number of universes (with a limit of `1` pseudo
// per universe).
//
// The `UniverseID` is the identifier of the universe in which
// this player is registered. This determines where it can
// perform actions.
//
// The `ID` represents the identifier of the player's current
// instance in this universe.
//
// The `Name` represents the in-game display for this player.
// It is distinct from the account's name.
//
// The `Technologies` defines the level of each in-game tech
// already researched by the player. Note that technologies
// with a level of `0` are not included in the output list.
type Player struct {
	AccountID    string       `json:"account"`
	UniverseID   string       `json:"uni"`
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Technologies []Technology `json:"technologies"`
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
