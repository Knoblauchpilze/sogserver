package data

import (
	"fmt"
	"math"
	"time"
)

// UpgradeAction :
// Provide the base building block for an action in the game.
// Such an action always has a planet associated to it where
// it will take place along with some way of identifying it.
//
// The `ID` defines an identifier for this action. It is used
// to populate the `ID` field when inserting the action in
// the DB.
//
// The `PlanetID` defines an identifier for the planet this
// action is related to. Basically any construction action
// should be started from somewhere and this is defined by
// this attribute.
//
// The `ElementID` defines the element on which this action
// is meant to have an effect. Typically it can refer to the
// ID of an in-game building, technology, etc. which needs
// to be upgraded. Depending on the precise type of this
// element the related DB table will vary.
//
// The `IsUnitLike` defines whether the action is related to
// the construction of something that resembles a unit (i.e.
// it has a fixed cost and no possible upgrades) or something
// that can be upgraded (like a building or a technology for
// example).
// Depending on this status the way to compute the total
// cost required to launch the action will be different.
type UpgradeAction struct {
	ID         string `json:"id"`
	PlanetID   string `json:"planet"`
	ElementID  string `json:"element"`
	IsUnitLike bool
}

// valid :
// Allows to make sure that the upgrade action is valid by
// checking that all the internal fields have values that
// are at least not obviously wrong.
//
// Returns `true` if the action seems valid.
func (a UpgradeAction) valid() bool {
	return validUUID(a.ID) &&
		validUUID(a.PlanetID) &&
		validUUID(a.ElementID)
}

// String :
// Implementation of the `Stringer` interface to allow to
// easily display this action if needed.
//
// Returns the strig describing this action.
func (a UpgradeAction) String() string {
	return fmt.Sprintf("\"%s\"", a.PlanetID)
}

// ProgressAction :
// Specialization of the `UpgradeAction` to handle the case
// of action related to an element that can be upgraded. It
// typically applies to buildings and technologies. Compared
// to the base upgrade action this type of element has two
// levels (the current one and the desired one) and a way to
// compute the total cost needed for the upgrade.
//
// The `CurrentLevel` represents the current level of the
// element to upgrade.
//
// The `DesiredLevel` represents the desired level of the
// element after the upgrade action is complete.
//
// The `CompletionTime` will be computed from the cost of
// the action and the facilities existing on the planet
// where the action is triggered.
//
// The `IsStrictlyUpgradable` defines whether the progress
// action can reference requests to downgrade the element
// referred to by this action (so typically allowing the
// `CurrentLevel` to be greater than the `DesiredLevel`).
type ProgressAction struct {
	CurrentLevel         int       `json:"current_level"`
	DesiredLevel         int       `json:"desired_level"`
	CompletionTime       time.Time `json:"completion_time"`
	IsStrictlyUpgradable bool

	UpgradeAction
}

// valid :
// Used to refine the behavior of the base upgrade action
// to make sure that the levels provided for this action
// are correct.
//
// Returns `true` if this action is not obviously wrong.
func (a ProgressAction) valid() bool {
	return a.UpgradeAction.valid() &&
		a.CurrentLevel >= 0 &&
		a.DesiredLevel >= 0 &&
		((a.IsStrictlyUpgradable && a.DesiredLevel == a.CurrentLevel+1) ||
			(!a.IsStrictlyUpgradable && math.Abs(float64(a.DesiredLevel)-float64(a.CurrentLevel)) == 1))
}
