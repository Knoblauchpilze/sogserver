package data

import (
	"regexp"

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
		u.GalaxySize > 0 &&
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
	return a.Name != "" &&
		exp.MatchString(a.Mail) &&
		a.Password != ""
}

// averageTemp :
// Returns the average temperature for a planet as a float
// value.
//
// Returns the average temperature.
func (p *Planet) averageTemp() float32 {
	return float32((p.MinTemp + p.MaxTemp) / 2)
}

// remainingFields :
// Returns the number of remaining fields on the planet
// given the current buildings on it. Note that it does
// not include the potential upgrade actions.
func (p *Planet) remainingFields() int {
	// Accumulate the total used fields.
	used := 0

	for _, b := range p.Buildings {
		used += b.Level
	}

	return p.Fields - used
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

// valid :
// Used to determine whether the parameters defined for this
// fleet component are not obviously wrong. This method checks
// that the identifier provided for individual ships aren't
// empty or ill-formed and that the amount of each one is at
// least strictly positive.
//
// The `uni` represents the universe which should be attached
// to the fleet and will be used to verify that the starting
// position of the fleet component is consistent with possible
// coordinates in the universe.
//
// Returns `true` if the fleet component is valid.
func (fc *FleetComponent) valid(uni Universe) bool {
	// Check own identifier.
	if !validUUID(fc.ID) {
		return false
	}

	// Check the identifier of the player and parent fleet.
	if !validUUID(fc.FleetID) {
		return false
	}
	if !validUUID(fc.PlayerID) {
		return false
	}

	// Check the coordinates against the universe.
	if fc.Galaxy < 0 || fc.Galaxy >= uni.GalaxiesCount {
		return false
	}
	if fc.System < 0 || fc.System >= uni.GalaxySize {
		return false
	}
	if fc.Position < 0 || fc.Position >= uni.SolarSystemSize {
		return false
	}

	// Check the speed.
	if fc.Speed < 0 || fc.Speed > 1 {
		return false
	}

	// Now check individual ships.
	if len(fc.Ships) == 0 {
		return false
	}

	for _, ship := range fc.Ships {
		if !validUUID(ship.ShipID) || ship.Amount <= 0 {
			return false
		}
	}

	return true
}

// valid :
func (f *Fleet) valid(uni Universe) bool {
	// Check own identifier.
	if !validUUID(f.ID) {
		return false
	}

	// Check that the target is valid given the universe
	// into which the fleet is supposed to reside.
	if f.Galaxy < 0 || f.Galaxy >= uni.GalaxiesCount {
		return false
	}
	if f.System < 0 || f.System >= uni.GalaxySize {
		return false
	}
	if f.Position < 0 || f.Position >= uni.SolarSystemSize {
		return false
	}

	return true
}
