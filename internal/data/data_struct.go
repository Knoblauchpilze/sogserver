package data

import (
	"oglike_server/pkg/duration"
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

// ConstructionCost :
// Defines for a single element the associated costs
// for the first level along with a progression rule
// followed to compute the cost at any level.
//
// The `InitCosts` rerpesents a map where the keys are
// resources' identifiers and the values are the init
// cost of the element for the corresponding resource.
// If a resource does not have its identifier in this
// map, it means that the element does not require any
// quantity of it.
//
// The `ProgressionRule` defines a value that should
// be used to multiply the initial cost to obtain the
// cost at any level. All the elements follow a rule
// where the cost at level `N` is something along the
// line of `InitCost * ProgressionRule ^ N`.
// The larger this value, the quicker the costs will
// rise with the level.
type ConstructionCost struct {
	InitCosts       map[string]int
	ProgressionRule float32
}

// ProductionRule :
// Used to define the rule to produce some quantity
// of a resource for an element (usually a building).
// Some of the in-game buildings are able to generate
// resources meaning that it provides a certain amount
// or a resource at each step of time (usually using
// the second as time unit).
// The higher the level of the building, the more of
// the resource will be produced using a progression
// rule similar to `Base * Level * Exp ^ Level`.
//
// The `Resource` defines an identifier of the res
// that is produced by the element. It should be
// linked to an actual existing resource in the
// `resources` table.
//
// The `InitProd` defines the base production at the
// `0`-th level for this resource. This is the base
// from where the production gains are scaled.
//
// The `ProgressionRule` defines the base associated
// to the exponential growth in production for the
// resource. The larger this value the quicker the
// production will rise with each level.
//
// The `TemperatureCoeff` defines a coefficient to
// apply to the production which depends on the temp
// of the planet where the production rule is applied.
// If this value is positive it means that the hotter
// the planet is, the more production for this res is
// to be expected, while the effect is reversed if the
// coefficient is negative.
// A coefficient of `0` means that the temperature of
// the planet does not have any impact on the resource
// prod.
//
// The `TemperatureOffset` is used in conjunction with
// the `TemperatureCoeff` and allows to provide some
// bost in the computation.
// Typically the temperature dependent part of the
// production of a resource looks something like this:
// `TemperatureCoeff * T + TemperatureOffset`.
type ProductionRule struct {
	Resource          string
	InitProd          int
	ProgressionRule   float32
	TemperatureCoeff  float32
	TemperatureOffset float32
}

// ResourceDesc :
// Defines the abstract representation of a resource which
// is bascially an identifier and the actual name of the
// resource.
//
// The `ID` defines the identifier of the resource in the
// table gathering all the in-game resources. This is used
// in most of the other operations referencing resources.
//
// The `Name` defines the human-readable name of the res
// as displayed to the user.
//
// The `BaseProd` defines the production without modifiers
// for this resource on each planet. It represents a way
// for the user to get resources without building anything
// else.
//
// The `BaseStorage` defines the base capacity to store
// the resource without any modifiers (usually hangars).
//
// The `BaseAmount` defines the base amount for this res
// that can be found on any new planet in the game.
type ResourceDesc struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	BaseProd    int    `json:"base_production"`
	BaseStorage int    `json:"base_storage"`
	BaseAmount  int    `json:"base_amount"`
}

// Resource :
// Defines a substance that can be produced or consumed by
// any element of the game. It is typically used to build
// some new elements (buildings, ships, etc.) or consumed
// to produce some other resource. A resource is always
// linked to a planet where it is actually held (waiting
// to be used).
//
// The `ID` defines the identifier of the resource. This
// can be used to uniquely refer to the resource and to
// communicate with the server.
//
// The `Planet` defines the identifier of the planet on
// which the resources are currently held.
//
// The `Amount` defines the quantity of said resources.
// Depending on the context it can either be a quantity
// available or needed.
//
// The `Production` defines the production level for this
// resource for each second. This value takes into account
// all the modifiers which might impact it.
//
// The `Storage` defines how many units of the resource
// can be stored in the location where it is placed.
type Resource struct {
	ID         string  `json:"res"`
	Planet     string  `json:"planet"`
	Amount     float32 `json:"amount"`
	Production float32 `json:"production"`
	Storage    float32 `json:"storage_capacity"`
}

// TechDependency :
// Defines a dependency between two elements. We assume that the
// element for which the dependency is created is linked to this
// object in another way.
// A dependency indicates that a piece of the game isn't available
// until the following criteria is met.
//
// The `ID` defines a unique identifier for which the object is
// dependent upon. This can be an identifier on any kind of item
// (technology, building, etc.) and the context allows to choose
// which kind it is.
//
// The `Level` defines the minimum level at which the dependency
// is met: if the item described by its `ID` does not have at
// least the following level the dependency is unmet and the item
// should be made unavailable.
type TechDependency struct {
	ID    string `json:"id"`
	Level int    `json:"level"`
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
	Resource string `json:"resource"`
	Amount   int    `json:"amount"`
}

// BuildingDesc :
// Defines the abstract representation of a building with its
// name and unique identifier. It might also include a short
// summary of its role retrieved from the database.
//
// The `ID` defines the unique identifier for this building.
//
// The `Name` defines a human readable name for the building.
//
// The `Description` defines a short text describing the role
// of the building and its purpose.
//
// The `BuildingDeps` defines a list of identifiers which
// represent the buildings (and their associated level) which
// need to be available for this building to be built. It is
// some sort of representation of the tech-tree.
//
// The `TechnologiesDeps` fills a similar purpose but register
// dependencies on technologies and not buildings.
type BuildingDesc struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Desc             string           `json:"desc"`
	BuildingsDeps    []TechDependency `json:"buildings_dependencies"`
	TechnologiesDeps []TechDependency `json:"technologies_dependencies"`
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

// TechnologyDesc :
// Defines the abstract representation of a technology with
// its name and unique identifier. It might also include a
// short summary of its purpose retrieved from the database.
//
// The `ID` defines the unique identifier for this technology.
//
// The `Name` defines a human readable name for the technology.
//
// The `Description` defines a short text describing the role
// of the technology and its applications.
//
// The `BuildingDeps` defines a list of identifiers which
// represent the buildings (and their associated level) which
// need to be available for this technology to be built. It is
// some sort of representation of the tech-tree.
//
// The `TechnologiesDeps` fills a similar purpose but register
// dependencies on technologies and not buildings.
type TechnologyDesc struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Desc             string           `json:"desc"`
	BuildingsDeps    []TechDependency `json:"buildings_dependencies"`
	TechnologiesDeps []TechDependency `json:"technologies_dependencies"`
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

// RapidFire :
// Describes a rapid fire from a unit on another. It is
// defined by both the identifier of the element that is
// subject to the rapid fire along with a value which
// describes the actual effect.
//
// The `Receiver` defines the element that is subject
// to a rapid fire from the provider.
//
// The `RF` defines the actual value of the rapid fire.
type RapidFire struct {
	Receiver string `json:"receiver"`
	RF       int    `json:"rf"`
}

// ShipDesc :
// Defines the abstract representation of a ship with
// its name and unique identifier. It can also include
// a short summary of its purpose retrieved from the
// database.
//
// The `ID` defines the unique identifier for this ship.
//
// The `Name` defines a human readable name for the ship.
//
// The `Description` defines a short text describing the
// role of the ship and its capabilities.
//
// The `BuildingDeps` defines a list of identifiers which
// represent the buildings (and their associated level)
// which need to be available for this ship to be built.
// It is some sort of representation of the tech-tree.
//
// The `TechnologiesDeps` fills a similar purpose but
// register dependencies on technologies and not buildings.
//
// The `RFVSShips` defines the rapid fire this ship has
// against other ships.
//
// The `RFVSDefenses` defines the rapid fire this ship
// has against defenses.
//
// The `Cost` defines how much of each resource is needed
// to build example of this ship.
type ShipDesc struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Desc             string           `json:"desc"`
	BuildingsDeps    []TechDependency `json:"buildings_dependencies"`
	TechnologiesDeps []TechDependency `json:"technologies_dependencies"`
	RFVSShips        []RapidFire      `json:"rf_against_ships"`
	RFVSDefenses     []RapidFire      `json:"rf_against_defenses"`
	Cost             []ResourceAmount `json:"cost"`
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

// DefenseDesc :
// Defines the abstract representation of a defense
// with its name and unique identifier. It can also
// include a short summary of its purpose retrieved
// from the database.
//
// The `ID` defines the unique identifier for this
// defense.
//
// The `Name` defines a human readable name for the
// defense.
//
// The `Description` defines a short text describing
// the role of the defense and its principle.
//
// The `BuildingDeps` defines a list of identifiers
// which represent the buildings (and their associated
// level) which need to be available for this defense
// to be built. It is some sort of representation of
// the tech-tree.
//
// The `TechnologiesDeps` fills a similar purpose but
// register dependencies on technologies and not on
// buildings.
//
// The `Cost` defines how much of each resource is needed
// to build example of this ship.
type DefenseDesc struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Desc             string           `json:"desc"`
	BuildingsDeps    []TechDependency `json:"buildings_dependencies"`
	TechnologiesDeps []TechDependency `json:"technologies_dependencies"`
	Cost             []ResourceAmount `json:"cost"`
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

// UpgradeAction :
// General interface for an upgrade action that can be
// applied to buildings, technologies, ships and also
// defenses.
// Such an interface allows to easily mutualize the
// insertion process of such actions in the DB.
type UpgradeAction interface {
	GetID() string
	valid() bool
	String() string
}

// BuildingUpgradeAction :
// Defines the upgrade action associated to a building. It
// describes the building to which it is associated along
// with the planet where the building is built.
// Indication about the duration of the action and a level
// that will be reached by the building is provided.
//
// The `ID` defines the identifier of this building upgrade
// action as registered in the DB. It allows to uniquely
// identify it if needed (for example to fetch it later).
//
// The `PlanetID` defines an identifier of the planet where
// the building is built.
//
// The `BuildingID` defines the identifier of the building
// to which this action is related.
//
// The `CurrentLevel` defines the current level reached
// for this building. Should always be smaller than the
// `DesiredLevel`.
//
// The `DesiredLevel` defines the desired level for the
// building.
//
// The `CompletionTime` defines the time at which this
// action will be completed.
type BuildingUpgradeAction struct {
	ID             string    `json:"id"`
	PlanetID       string    `json:"planet"`
	BuildingID     string    `json:"building"`
	CurrentLevel   int       `json:"current_level"`
	DesiredLevel   int       `json:"desired_level"`
	CompletionTime time.Time `json:"completion_time"`
}

// TechnologyUpgradeAction :
// Defines the upgrade action associated to a technology.
// It describes the technology to which it is associated
// along with the player who is researching it. It also
// indicates the duration of the action and a level that
// will be reached by the technology at the end of the
// action.
//
// The `ID` defines the identifier of this technology
// upgrade action as registered in the DB. It allows to
// uniquely identify it if needed (for example to fetch
// it later).
//
// The `PlayerID` defines an identifier of the player who
// is researching the technology.
//
// The `TechnologyID` defines the identifier of the tech
// that is being researched.
//
// The `PlanetID` allows to figure where the technology
// is being researched. This allows to know where the
// resources should be restored in case the player does
// cancel the research.
//
// The `CurrentLevel` defines the current level reached
// for this technology. Should always be smaller than
// the `DesiredLevel`.
//
// The `DesiredLevel` defines the desired level for the
// technology.
//
// The `CompletionTime` defines the time at which this
// action will be completed.
type TechnologyUpgradeAction struct {
	ID             string    `json:"id"`
	PlayerID       string    `json:"player"`
	TechnologyID   string    `json:"technology"`
	PlanetID       string    `json:"planet"`
	CurrentLevel   int       `json:"current_level"`
	DesiredLevel   int       `json:"desired_level"`
	CompletionTime time.Time `json:"completion_time"`
}

// ShipUpgradeAction :
// Defines the upgrade action associated to a ship. It is
// used to start the construction of a new ship on one of
// the planet of a player and describes the ship that is
// to be built and the planet onto which the building is
// performed.
//
// The `ID` defines the identifier of this ship upgrade
// action as registered in the DB. It allows to uniquely
// identify it if needed (for example to fetch it later).
//
// The `PlanetID` defines an identifier of the planet on
// which the ship is built.
//
// The `ShipID` defines the identifier of the ship that
// is being built.
//
// The `Amount` defines how many ships should be built
// by this upgrade action.
//
// The `Remaining` defines how many ships are yet to be
// built by this upgrade action.
//
// The `CompletionTime` defines the duration of a single
// element of the action. It might represent the duration
// of the total action in case the `Amount` is set to 1.
type ShipUpgradeAction struct {
	ID             string            `json:"id"`
	PlanetID       string            `json:"planet"`
	ShipID         string            `json:"ship"`
	Amount         int               `json:"amount"`
	Remaining      int               `json:"remaining"`
	CompletionTime duration.Duration `json:"completion_time"`
}

// DefenseUpgradeAction :
// Defines the upgrade action associated to a defense. It
// is used to start the construction of a new defense onto
// one of the planet of a player and describes the defense
// that is to be built and the planet onto which the
// construction should take place.
//
// The `ID` defines the identifier of this defense upgrade
// action as registered in the DB. It allows to uniquely
// identify it if needed (for example to fetch it later).
//
// The `PlanetID` defines an identifier of the planet on
// which the ship is built.
//
// The `DefenseID` defines the identifier of the defense
// that is being built.
//
// The `Amount` defines how many ships should be built
// by this upgrade action.
//
// The `Remaining` defines how many defeneses are yet to
// be built by this upgrade action.
//
// The `CompletionTime` defines the duration of a single
// element of the action. It might represent the duration
// of the total action in case the `Amount` is set to 1.
type DefenseUpgradeAction struct {
	ID             string            `json:"id"`
	PlanetID       string            `json:"planet"`
	DefenseID      string            `json:"defense"`
	Amount         int               `json:"amount"`
	Remaining      int               `json:"remaining"`
	CompletionTime duration.Duration `json:"completion_time"`
}
