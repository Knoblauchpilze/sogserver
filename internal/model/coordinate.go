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
type Coordinate struct {
	Galaxy   int `json:"galaxy"`
	System   int `json:"system"`
	Position int `json:"position"`
}

// NewCoordinate :
// Used to create a new coordinate object from the input data.
// No controls are performed to verify that the input coords
// are actually consistent with anything.
//
// The `galaxy` represents the index of the galaxy of the coords.
//
// The `system` represents the solar system index.
//
// The `position` defines the position of the planet within its
// parent solar system.
//
// Returns the created coordinate object.
func NewCoordinate(galaxy int, system int, position int) Coordinate {
	return Coordinate{
		galaxy,
		system,
		position,
	}
}

// String :
// Implementation of the stringer interface for a coord.
// Helps printing this data structure to a stream or to
// visually see it in the logs.
//
// Returns the string representing the coordinates.
func (c Coordinate) String() string {
	return fmt.Sprintf("[G: %d, S: %d, P: %d]", c.Galaxy, c.System, c.Position)
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
