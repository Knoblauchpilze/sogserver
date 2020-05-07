package model

import (
	"fmt"
	"math"
	"strconv"
)

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
//
// The `Type` defines which part of the coordinate should be
// accessed. Indeed in the game at a specified position one
// can find a planet and possibly a moon and a debris fields.
// This element allows to define precisely which location is
// to be accessed.
type Coordinate struct {
	Galaxy   int      `json:"galaxy"`
	System   int      `json:"system"`
	Position int      `json:"position"`
	Type     Location `json:"location"`
}

// Location :
// Describes the possible locations of a coordinate. It usually
// indicates whether the coordinate refers to the planet, moon
// or debris fields of a specific location.
type Location string

// Define the possible location for a coordinate.
const (
	World  Location = "planet"
	Moon   Location = "moon"
	Debris Location = "debris"
)

// ErrInvalidCoordinateType :
// Used to indicate that a coordinate was constructed with an
// invalid location identifier.
var ErrInvalidCoordinateType = fmt.Errorf("Unknown coordinate location type")

// existsLocation :
// Used to make sure that the input location is a valid
// coordinate type.
//
// The `loc` defines the type of the location that is to
// be checked.
//
// Returns `true` if the input location is a valid item
// for a location.
func existsLocation(loc Location) bool {
	return loc == World || loc == Moon || loc == Debris
}

// NewPlanetCoordinate :
// Used to create a new coordinate object from the input data.
// No controls are performed to verify that the input coords
// are actually consistent with anything.
// The coordinate will refer to the planet location at the
// specified coordinates.
//
// The `galaxy` represents the index of the galaxy of the coords.
//
// The `system` represents the solar system index.
//
// The `position` defines the position of the planet within its
// parent solar system.
//
// Returns the created coordinate object.
func NewPlanetCoordinate(galaxy int, system int, position int) Coordinate {
	return Coordinate{
		Galaxy:   galaxy,
		System:   system,
		Position: position,
		Type:     World,
	}
}

// NewMoonCoordinate :
// Similar to the `NewPlanetCoordinate` but defines some
// coordinates allowing to access the moon location.
//
// The `galaxy` defines the index of the galaxy for the
// coordinates.
//
// The `system` defines the solar system.
//
// Tje `position` defines the position of the element in
// the solar system.
//
// Returns the created coordinates.
func NewMoonCoordinate(galaxy int, system int, position int) Coordinate {
	return Coordinate{
		Galaxy:   galaxy,
		System:   system,
		Position: position,
		Type:     Moon,
	}
}

// NewDebrisCoordinate :
// Similar to the `NewPlanetCoordinate` but defines some
// coordinates allowing to access the debris fields item
// at a specified location.
//
// The `galaxy` defines the index of the galaxy for the
// coordinates.
//
// The `system` defines the solar system.
//
// Tje `position` defines the position of the element in
// the solar system.
//
// Returns the created coordinates.
func NewDebrisCoordinate(galaxy int, system int, position int) Coordinate {
	return Coordinate{
		Galaxy:   galaxy,
		System:   system,
		Position: position,
		Type:     Debris,
	}
}

// newCoordinate :
// Creates a coordinate with the specified galaxy, etc.
// and location.
//
// The `galaxy` defines the index of the galaxy.
//
// The `system` defines the index of the solar system
// in the galaxy.
//
// The `position` defines the index of the element in
// the system.
//
// The `loc` defines the location of the element from
// the input coordinate.
//
// Returns the created coordinate with any error in
// case the location is not valid.
func newCoordinate(galaxy int, system int, position int, loc Location) (Coordinate, error) {
	if !existsLocation(loc) {
		return Coordinate{}, ErrInvalidCoordinateType
	}

	c := Coordinate{
		Galaxy:   galaxy,
		System:   system,
		Position: position,
		Type:     loc,
	}

	return c, nil
}

// String :
// Implementation of the stringer interface for a coord.
// Helps printing this data structure to a stream or to
// visually see it in the logs.
//
// Returns the string representing the coordinates.
func (c Coordinate) String() string {
	return fmt.Sprintf("[G: %d, S: %d, P: %d %s]", c.Galaxy, c.System, c.Position, c.Type)
}

// Linearize :
// Used as a simple way to extract a single integer from a
// coordinates object. We use the input counts to create an
// integer which regroup the galaxy, solar system and pos
// of the planet into a single integer.
//
// The `galaxySize` defines the number of solar system that
// can be found in a single galaxy. We will extract digits
// count for this number to know how to linearize the coord.
//
// The `solarSystemSize` defines a similar value for the
// number of planets that can be found in a solar system.
//
// Return an integer representing the coordinates within
// the provided coordinates system.
func (c Coordinate) Linearize(galaxySize int, solarSystemSize int) int {
	// Compute the number of digits needed to express each
	// part of the coordinate.
	sDigits := len(strconv.Itoa(galaxySize))
	pDigits := len(strconv.Itoa(solarSystemSize))

	sOffset := int(math.Pow10(sDigits))
	pOffset := int(math.Pow10(pDigits))

	return c.Position + c.System*pOffset + c.Galaxy*pOffset*sOffset
}

// generateSeed :
// Used to generate a valid seed from the coordinates defined
// by this object. We use this as a semi-procedural way to use
// information about the position to compute a single integer
// value. This is similar to the linearization except we don't
// need to preserve the readability of any of the individual
// components of the coordinate.
//
// Returns the generated seed.
func (c Coordinate) generateSeed() int {
	// We will use the Cantor's pairing function as defined in
	// the following article twice in a row:
	// https://en.wikipedia.org/wiki/Pairing_function
	k1 := (c.Position+c.System)*(c.Position+c.System+1)/2 + c.System
	return (k1+c.Galaxy)*(k1+c.Galaxy+1)/2 + c.Galaxy
}

// valid :
// Used to determine whether this set of coordinates is valid
// given the input bounds for each element. Note that we also
// assume that a negative coordinate is not valid.
//
// The `galaxyCount` defines the maximum number of galaxies.
//
// The `galaxySize` defines how many solar system exists in
// each galaxy.
//
// The `solarSystemSize` defines how many planet can be found
// in each solar system.
//
// Returns `true` if the coordinate is valid.
func (c Coordinate) valid(galaxyCount int, galaxySize int, solarSystemSize int) bool {
	return c.Galaxy >= 0 && c.Galaxy < galaxyCount &&
		c.System >= 0 && c.System < galaxySize &&
		c.Position >= 0 && c.Position < solarSystemSize &&
		existsLocation(c.Type)
}

// distanceTo :
// Used to compute the distance from this position to the
// other provided as input. Note that the concept of a
// distance is specific to og and has few real meaning.
//
// The `other` defines the other coordinates for which a
// distance should be computed.
//
// Returns the distance between the two coordinates.
func (c Coordinate) distanceTo(other Coordinate) int {
	// Most of the information is extracted from there:
	// https://ogame.fandom.com/wiki/Talk:Fuel_Consumption

	// Case where galaxies are different.
	if c.Galaxy != other.Galaxy {
		dGal := float64(c.Galaxy - other.Galaxy)
		return 20000 * int(math.Abs(dGal))
	}

	// Case where systems are different.
	if c.System != other.System {
		dSys := float64(c.System - other.System)
		return 2700 + (95 * int(math.Abs(dSys)))
	}

	// Case where positions are different.
	if c.Position != other.Position {
		dPos := float64(c.Position - other.Position)
		return 1000 + (5 * int(math.Abs(dPos)))
	}

	// Within same position the cost is always identical.
	return 5
}
