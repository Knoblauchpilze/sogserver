package data

// Player :
// Define a player which is basically a name in a universe.
// We also provide both the identifier of this player along
// with its account index.
//
// The `Account` represents the identifier of the accounts
// associated with this player. An account can be registered
// on any number of universes (with a limit of `1` pseudo
// per universe).
//
// The `Universe` is the identifier of the universe in which
// this player is registered. This determines where it can
// perform actions.
//
// The `ID` represents the identifier of the player's current
// instance in this universe.
//
// The `Name` represents the in-game display for this player.
// It is distinct from the account's name.
type Player struct {
	Account  string `json:"account_id"`
	Universe string `json:"universe_id"`
	ID       string `json:"player_id"`
	Name     string `json:"name"`
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
// The `Player` defines the identifier of the player which owns
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
	Player   string     `json:"player_id"`
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
