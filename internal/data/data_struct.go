package data

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
// The `Mail` defines the email address associated to the
// account. It can be used to make sure that no two accounts
// share the same address.
type Account struct {
	ID   string `json:"id"`
	Mail string `json:"mail"`
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
type Player struct {
	AccountID  string `json:"account_id"`
	UniverseID string `json:"universe_id"`
	ID         string `json:"player_id"`
	Name       string `json:"name"`
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
// The `FleetToRuins` defines the percentage of resources
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
// The `SolarSystemSize` defines the number of planets in
// each solar system of each galaxy.
type Universe struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	EcoSpeed         int     `json:"eco_speed"`
	FleetSpeed       int     `json:"fleet_speed"`
	ResearchSpeed    int     `json:"research_speed"`
	FleetToRuins     float32 `json:"fleet_to_ruins"`
	DefensesToRuins  float32 `json:"defenses_to_ruins"`
	FleetConsumption float32 `json:"fleet_consumption"`
	GalaxiesCount    int     `json:"galaxies_count"`
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
// The `Coords` defines the coordinate of the planet within its
// parent universe. The coordinates should be consistent with
// the limits defined for the universe.
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
type Planet struct {
	PlayerID string     `json:"player_id"`
	ID       string     `json:"id"`
	Coords   Coordinate `json:"coordinates"`
	Name     string     `json:"name"`
	Fields   int        `json:"fields"`
	MinTemp  float32    `json:"min_temperature"`
	MaxTemp  float32    `json:"max_temperature"`
	Diameter int        `json:"diameter"`
}

// Research :
// Defines a research in the og context. It defines the identifier
// of the technology which allows to access the description of the
// technology and other information.
//
// The `ID` defines the identifier of the technology.
//
// The `Name` of the technology.
//
// The `Level` defines the current research level of this technology
// on the account of a player.
type Research struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Level int    `json:"level"`
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
type Building struct {
	ID    string `json:"id"`
	Level int    `json:"level"`
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

// Fleet :
// Defines a fleet with its objective and coordinates. It also
// defines the posible name of the fleet.
//
// The `ID` represents a way to uniquely identify the fleet.
//
// The `Name` is the name that was given by the user to the
// fleet upon its creation. Note that it might be empty.
//
// The `Objective` is a string defining the action intended
// for this fleet. It is a way to determine which purpose the
// fleet serves.
//
// The `Coords` define the coordinates of the target planet
// of this fleet. Note that it might not be a planet in case
// the fleet's objective allows to travel to an empty location.
type Fleet struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Objective string     `json:"objective"`
	Coords    Coordinate `json:"coordinates"`
}
