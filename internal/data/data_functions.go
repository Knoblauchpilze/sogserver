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
