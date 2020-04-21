package model

import (
	"fmt"
	"oglike_server/pkg/db"
)

// Planet :
// Define a planet which is an object within a certain
// universe and associated to a certain player. This
// element is the base where all fleets can originate
// and is usually the main source of production in any
// universe. It contains a set of resources, some ships
// that are deployed on it and a set of buildings.
// The user can also start some construction actions
// that can increase the number of ships present on the
// planet for example.
// A planet is a complex object which requires a lock
// to be correctly built so that no other process can
// temper with the data related to it.
//
// The `ID` defines the identifier of the planet within
// all the planets registered in og.
//
// The `PlayerID` defines the identifier of the player
// who owns this planet. It is relative to an account
// and a universe.
//
// The `Coordinates` define the position of the planet
// in its parent universe. All needed coordinates are
// guaranteed to be valid within it.
//
// The `Name` of the planet as defined by the user.
//
// The `Fields` define the number of available fields in
// the planet. The number of used fields is computed from
// the infrastructure built on the planet but is derived
// from the actual level of each building.
//
// The `MinTemp` defines the minimum temperature of the
// planet in degrees.
//
// The `MaxTemp` defines the maximum temperature of the
// planet in degrees.
//
// The `Diameter` defines the diameter of the planet in
// kilometers.
//
// The `Resources` define the resources currently stored
// on the planet. This is basically the quantity available
// to produce some buildings, ships, etc.
//
// The `Buildings` defines the list of buildings currently
// built on the planet. Note that it does not provide info
// on buildings *being* built.
//
// The `Ships` defines the list of ships currently deployed
// on this planet. It does not include ships currently moving
// from or towards the planet.
//
// The `Defense` defines the list of defenses currently
// built on the planet. This does not include defenses
// *being* built.
type Planet struct {
	ID          string         `json:"id"`
	PlayerID    string         `json:"player"`
	Coordinates Coordinate     `json:"coordinate"`
	Name        string         `json:"name"`
	Fields      int            `json:"fields"`
	MinTemp     int            `json:"min_temperature"`
	MaxTemp     int            `json:"max_temperature"`
	Diameter    int            `json:"diameter"`
	Resources   []ResourceInfo `json:"resources"`
	Buildings   []BuildingInfo `json:"buildings"`
	Ships       []ShipInfo     `json:"ships"`
	Defenses    []DefenseInfo  `json:"defenses"`
}

// ResourceInfo :
// Defines the information needed to qualify the amount
// of resource of a certain type existing on a planet.
// In addition to the base description of the resource,
// this object defines the current storage capacity of
// the planet along with the production given the set
// of buildings (and more specifically mines) built on
// the planet.
//
// The `Resource` defines the identifier of the resource
// that this association describes.
//
// The `Amount` defines how much of the resource is
// present on this planet.
//
// The `Storage` defines the maximum amount that can
// be stored on this planet. Note that the `Amount`
// can exceed the `Storage` in which case the prod is
// stopped.
//
// The `Production` defines the maximum theoretical
// production of this resource given the buildings
// existing on the planet. Note that this value is
// always set to the maximum possible and does not
// account for the fact that the production may be
// stopped (in case the storage is full for example).
// The production is given in units per hour.
type ResourceInfo struct {
	Resource   string `json:"resource"`
	Amount     int    `json:"amount"`
	Storage    int    `json:"storage"`
	Production int    `json:"production"`
}

// BuildingInfo :
// Defines the information about a building of a
// planet. It reuses most of the base description
// for a building with the addition of a level to
// indicate the current state reached on this
// planet.
//
// The `Level` defines the level reached by this
// building on a particular planet.
type BuildingInfo struct {
	BuildingDesc

	Level int `json:"level"`
}

// ShipInfo :
// Similar to the `BuildingInfo` but defines the
// amount of ships currently deployed on a planet.
// Note that this is a snapshot at a certain time
// and does not account for ships that may be
// built currently.
//
// The `Amount` defines how many of this ship are
// currently deployed on the planet.
type ShipInfo struct {
	ShipDesc

	Amount int `json:"amount"`
}

// DefenseInfo :
// Similar to the `ShipInfo` but defines the same
// properties for a defense system.
//
// The `Amount` defines how many of this defense
// system are currently deployed on the planet.
type DefenseInfo struct {
	DefenseDesc

	Amount int `json:"amount"`
}

// getDefaultPlanetName :
// Used to retrieve a default name for a planet. The
// generated name will be different based on whether
// the planet is a homeworld or no.
//
// The `isHomeworld` is `true` if we should generate
// a name for the homeworld and `false` otherwise.
//
// Returns a string corresponding to the name of the
// planet.
func getDefaultPlanetName(isHomeWorld bool) string {
	if isHomeWorld {
		return "homeworld"
	}

	return "planet"
}

// getPlanetTemperatureAmplitude :
// Used to retrieve the default planet temperature's
// amplitude. Basically the interval between the min
// and max temperature will always be equal to this
// value.
//
// Returns the default temperature amplitude for the
// planets.
func getPlanetTemperatureAmplitude() int {
	return 50
}

// NewPlanet :
// Used to fetch the content of the planet from the
// input DB and populate all internal fields from it.
// In case the DB cannot be fetched or some errors
// are encoutered, the return value will include a
// description of the error.
//
// The `proxy` allows to actually perform the DB
// requests to fetch the planet's data.
//
// Returns the planet as fetched from the DB along
// with any errors.
func NewPlanet(proxy db.Proxy) (Planet, error) {
	// TODO: Handle this.
	return Planet{}, fmt.Errorf("Not implemented")
}

// AverageTemperature :
// Returns the average temperature for this planet.
func (p *Planet) AverageTemperature() float32 {
	return float32(p.MinTemp+p.MaxTemp) / 2.0
}

// RemainingFields :
// Returns the number of remaining fields on the planet
// given the current buildings on it. Note that it does
// not include the potential upgrade actions.
func (p *Planet) RemainingFields() int {
	// Accumulate the total used fields.
	used := 0

	for _, b := range p.Buildings {
		used += b.Level
	}

	return p.Fields - used
}
