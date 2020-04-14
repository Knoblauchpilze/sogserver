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

// ComputeCosts :
// Used to perform the computation of the resources needed
// to build the `level`-th level of the element described
// by these construction costs.
// The level is clamped to be in the range `[0; +inf[` if
// this is not already the case.
//
// The `level` for which the costs should be computed. It
// is clamped to be positive.
//
// Returns a slice describing the amount needed of each
// resource needed by the item.
func (cc ConstructionCost) ComputeCosts(level int) []ResourceAmount {
	// Clamp the input level.
	fLevel := math.Max(0.0, float64(level))

	costs := make([]ResourceAmount, 0)

	for res, cost := range cc.InitCosts {
		costForRes := ResourceAmount{
			res,
			float32(float64(cost) * math.Pow(float64(cc.ProgressionRule), fLevel)),
		}

		costs = append(costs, costForRes)
	}

	return costs
}

// ComputeProduction :
// Used to perform the computation of the resources that
// are produced by the level `level` of the element that
// is described by the input production rule.
// The level is clamped to be in the range `[0; +inf[` if
// this is not already the case.
//
// The `level` for which the production should be computed.
// It is clamped to be positive.
//
// The `temperature` defines the average temperature of
// the planet where the production is evaluated. It is
// used to determine the temperature dependent part of the
// resource production.
//
// Returns the amount of resource that are produced by the
// selected rule with the specified level and temperature
// values.
func (pr ProductionRule) ComputeProduction(level int, temperature float32) ResourceAmount {
	// Clamp the input level.
	fLevel := math.Max(0.0, float64(level))
	fInitProd := float64(pr.InitProd)

	// Compute both parts of the production (temperature
	// dependent and independent).
	tempDep := float64(pr.TemperatureOffset + temperature*pr.TemperatureCoeff)
	tempIndep := fInitProd * fLevel * math.Pow(float64(pr.ProgressionRule), fLevel)

	prod := ResourceAmount{
		Resource: pr.Resource,
		Amount:   float32(tempDep * tempIndep),
	}

	return prod
}

// ComputeStorage :
// Used to perform the computation of the amount of res
// that can be held at the specified level.
//
// The `level` for which the storage capacity should be
// computed.
//
// Returns the amount of resource that can be held for
// the specified level by this storage.
func (sr StorageRule) ComputeStorage(level int) ResourceAmount {
	factor := float64(sr.Multiplier) * math.Exp(float64(sr.Progress)*float64(level))

	capacity := float32(sr.InitStorage * int(math.Floor(factor)))

	res := ResourceAmount{
		Resource: sr.Resource,
		Amount:   capacity,
	}

	return res
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

// GetID :
// Used to retrieve identifier of the building upgrade action.
// This basically returns the corresponding internal field.
//
// Returns a string corresponding to the identifier of this
// upgrade action.
func (a BuildingUpgradeAction) GetID() string {
	return a.ID
}

// valid :
// Determines whether the input action is valid or not based
// on the values of internal fields. We don't actually check
// the data against what's in the DB but check internally if
// all fields have plausible values.
//
// Returns `true` if the action is not obviously invalid.
func (a BuildingUpgradeAction) valid() bool {
	return validUUID(a.ID) &&
		validUUID(a.PlanetID) &&
		validUUID(a.BuildingID) &&
		a.CurrentLevel >= 0 &&
		a.DesiredLevel >= 0 &&
		math.Abs(float64(a.DesiredLevel)-float64(a.CurrentLevel)) == 1
}

// String :
// Implementation of the `Stringer` interface which allows
// to make easier the display of such an object.
//
// Returns a string description of the building upgrade
// action. This is mainly composed by the planet's ID as
// we want to group such failure for analysis purposes.
func (a BuildingUpgradeAction) String() string {
	return fmt.Sprintf("\"%s\"", a.PlanetID)
}

// GetID :
// Used to retrieve identifier of the technology upgrade
// action. This returns the corresponding internal field.
//
// Returns a string corresponding to the identifier of this
// upgrade action.
func (a TechnologyUpgradeAction) GetID() string {
	return a.ID
}

// valid :
// Determines whether the input action is valid or not based
// on the values of internal fields. We don't actually check
// the data against what's in the DB but check internally if
// all fields have plausible values.
//
// Returns `true` if the action is not obviously invalid.
func (a TechnologyUpgradeAction) valid() bool {
	return validUUID(a.ID) &&
		validUUID(a.PlayerID) &&
		validUUID(a.TechnologyID) &&
		a.CurrentLevel >= 0 &&
		a.DesiredLevel == a.CurrentLevel+1
}

// String :
// Implementation of the `Stringer` interface which allows
// to make easier the display of such an object.
//
// Returns a string description of the technology upgrade
// action. This is mainly composed by the player's ID as
// we want to group such failure for analysis purposes.
func (a TechnologyUpgradeAction) String() string {
	return fmt.Sprintf("\"%s\"", a.PlayerID)
}

// GetID :
// Used to retrieve identifier of the ship upgrade action.
// This basically returns the corresponding internal field.
//
// Returns a string corresponding to the identifier of this
// upgrade action.
func (a ShipUpgradeAction) GetID() string {
	return a.ID
}

// valid :
// Determines whether the input action is valid or not based
// on the values of internal fields. We don't actually check
// the data against what's in the DB but check internally if
// all fields have plausible values.
//
// Returns `true` if the action is not obviously invalid.
func (a ShipUpgradeAction) valid() bool {
	return validUUID(a.ID) &&
		validUUID(a.PlanetID) &&
		validUUID(a.ShipID) &&
		a.Amount > 0 &&
		a.Remaining >= 0
}

// String :
// Implementation of the `Stringer` interface which allows
// to make easier the display of such an object.
//
// Returns a string description of the ship upgrade action.
// This is mainly composed by the planet's ID as we want to
// group such failure for analysis purposes.
func (a ShipUpgradeAction) String() string {
	return fmt.Sprintf("\"%s\"", a.PlanetID)
}

// GetID :
// Used to retrieve identifier of the defense upgrade action.
// This basically returns the corresponding internal field.
//
// Returns a string corresponding to the identifier of this
// upgrade action.
func (a DefenseUpgradeAction) GetID() string {
	return a.ID
}

// valid :
// Determines whether the input action is valid or not based
// on the values of internal fields. We don't actually check
// the data against what's in the DB but check internally if
// all fields have plausible values.
//
// Returns `true` if the action is not obviously invalid.
func (a DefenseUpgradeAction) valid() bool {
	return validUUID(a.ID) &&
		validUUID(a.PlanetID) &&
		validUUID(a.DefenseID) &&
		a.Amount > 0 &&
		a.Remaining >= 0
}

// String :
// Implementation of the `Stringer` interface which allows
// to make easier the display of such an object.
//
// Returns a string description of the defense upgrade
// action. This is mainly composed by the planet's ID as
// we want to group such failure for analysis purposes.
func (a DefenseUpgradeAction) String() string {
	return fmt.Sprintf("\"%s\"", a.PlanetID)
}
