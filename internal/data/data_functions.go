package data

import (
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/google/uuid"
)

// validUUID :
// Used to check whether the input string can be interpreted
// as a valid identifier.
func validUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// valid :
// Used to determine whether the parameters defined for this
// universe are consistent with what is expected. This will
// typically check that the ratios are in the range `[0; 1]`
// and some other common assumptions.
// Note that it requires that the `ID` is valid as well.
//
// Returns `true` if the universe is valid (i.e. all values
// are consistent with the expected ranges).
func (u *Universe) valid() bool {
	// First check for the identifier.
	if !validUUID(u.ID) {
		return false
	}

	// Then common properties.
	return u.Name != "" &&
		u.EcoSpeed > 0 &&
		u.FleetSpeed > 0 &&
		u.ResearchSpeed > 0 &&
		u.FleetsToRuins >= 0.0 && u.FleetsToRuins <= 1.0 &&
		u.DefensesToRuins >= 0.0 && u.DefensesToRuins <= 1.0 &&
		u.FleetConsumption >= 0.0 && u.FleetConsumption <= 1.0 &&
		u.GalaxiesCount > 0 &&
		u.SolarSystemSize > 0
}

// valid :
// Used to determine whether the parameters defined for this
// account are consistent with what is expected. It is mostly
// used to check that the name is valid and that the e-mail
// address makes sense.
func (a *Account) valid() bool {
	// First check for the identifier.
	if !validUUID(a.ID) {
		return false
	}

	// Note that we *verified* the following regular expression
	// does compile so we don't check for errors.
	exp, _ := regexp.Compile("^[a-zA-Z0-9]*[a-zA-Z0-9_.+-][a-zA-Z0-9]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$")

	// Check common properties.
	return a.Name != "" && exp.MatchString(a.Mail)
}

// valid :
// Used to determine whether the parameters defined for this
// player are consistent with what is expected. We will only
// make sure that the identifiers associated to the account
// and the universe are not blatantly wrong and that the name
// provided is not empty.
// Whether the universe and account actually exist is not
// checked here.
// Note that it requires that the `ID` is valid as well.
//
// Returns `true` if the player is valid (i.e. all values are
// consistent with the expected ranges).
func (p *Player) valid() bool {
	// First check for the identifier.
	if !validUUID(p.ID) {
		return false
	}

	// Then the other identifiers.
	if !validUUID(p.AccountID) {
		return false
	}
	if !validUUID(p.UniverseID) {
		return false
	}

	// Finally a valid name should be provided.
	return p.Name != ""
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
// coordinates object. We use the input universe to create
// an integer which regroup the galaxy, solar system and
// position of the planet into a single integer.
//
// The `uni` defines the universe to which the coordinates
// belong which is used to linearize it.
//
// Return an integer representing the coordinates within
// the provided universe.
func (c Coordinate) Linearize(uni Universe) int {
	// Compute the number of digits needed to express each
	// part of the coordinate.
	sDigits := len(strconv.Itoa(uni.GalaxySize))
	pDigits := len(strconv.Itoa(uni.SolarSystemSize))

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
