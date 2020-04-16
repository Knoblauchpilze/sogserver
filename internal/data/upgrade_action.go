package data

import (
	"fmt"
	"math"
	"oglike_server/pkg/duration"
	"time"
)

// validationTools :
// Provides a convenience structure regrouping the needed
// information to perform the validation of an upgrade
// action. It contains some information about the costs
// of each element in the game along with some dependencies
// that need to be met for each element.
type validationTools struct {
	pCosts   map[string]ConstructionCost
	fCosts   map[string]FixedCost
	techTree map[string]TechDependency
}

// UpgradeAction :
// Generic interface describing an upgrade action to perform
// on a planet. This can concern any kind of data but it is
// required to define at least these methods in order to be
// able to correctly be checked against the planet data. It
// mostly consists into evaluating the cost of the action so
// that we can compare it with the resources existing on the
// planet and also providing some way to verify that needed
// buildings/technologies criteria are also met.
type UpgradeAction interface {
	Validate(tools validationTools) (bool, error)
}

// BaseUpgradeAction :
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
type BaseUpgradeAction struct {
	ID        string `json:"id"`
	PlanetID  string `json:"planet"`
	ElementID string `json:"element"`
}

// valid :
// Allows to make sure that the upgrade action is valid by
// checking that all the internal fields have values that
// are at least not obviously wrong.
//
// Returns `true` if the action seems valid.
func (a BaseUpgradeAction) valid() bool {
	return validUUID(a.ID) &&
		validUUID(a.PlanetID) &&
		validUUID(a.ElementID)
}

// String :
// Implementation of the `Stringer` interface to allow to
// easily display this action if needed.
//
// Returns the strig describing this action.
func (a BaseUpgradeAction) String() string {
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

	BaseUpgradeAction
}

// valid :
// Used to refine the behavior of the base upgrade action
// to make sure that the levels provided for this action
// are correct.
//
// Returns `true` if this action is not obviously wrong.
func (a ProgressAction) valid() bool {
	return a.BaseUpgradeAction.valid() &&
		a.CurrentLevel >= 0 &&
		a.DesiredLevel >= 0 &&
		((a.IsStrictlyUpgradable && a.DesiredLevel == a.CurrentLevel+1) ||
			(!a.IsStrictlyUpgradable && math.Abs(float64(a.DesiredLevel)-float64(a.CurrentLevel)) == 1))
}

// computeCost :
// Used to compute the construction cost of the action
// based on the level it aims at reaching and the total
// cost of various elements defined in the input table.
//
// The `costs` defines the initial costs and the rules
// to make the progress for various in-game elements.
// The map is indexed by ID key (so one of them should
// match the `a.ElementID` value).
//
// Returns a slice containing for each resource that
// is needed for this action the amount necessary. In
// case the input map does not define anything for the
// action an error is returned.
func (a ProgressAction) computeCost(costs map[string]ConstructionCost) ([]ResourceAmount, error) {
	// Find this action in the input table.
	cost, ok := costs[a.ElementID]

	if !ok {
		return []ResourceAmount{}, fmt.Errorf("Cannot compute cost for action \"%s\" defining unknown element \"%s\"", a.ID, a.ElementID)
	}

	needed := cost.ComputeCosts(a.DesiredLevel)

	return needed, nil
}

// Validate :
// Implementation of the `UpgradeAction` interface to
// perform the validation of the data contained in the
// current action against the information provided by
// the game framework. We will check that each element
// required by the validation tools allow the action
// to be performed.
//
// The `tools` allow to define the technological deps
// between elements and some resources that are present
// on the place where the action should be launched.
//
// Returns `true` if the action can be launched given
// the information provided in input.
func (a ProgressAction) Validate(tools validationTools) (bool, error) {
	// TODO: Should add the resources on a planet to be able
	// to validate the costs computed for this action against
	// a certain amount of resources.
	return false, fmt.Errorf("Not implemented")
}

// FixedAction :
// Specialization of the `UpgradeAction` to provide an
// action that concerns a unit-like element. This type
// of element cannot be upgraded and is rather built
// in a certain amount on a planet.
//
// The `Amount` defines the number of the unit to be
// produced by this action.
//
// The `Remaining` defines how many elements are still
// to be built at the moment of the analysis.
//
// The `CompletionTime`  defines the time it takes to
// complete the construction of a single unit of this
// element. The remaining time is thus given by the
// following: `Remaining * CompletionTime`. Note that
// it is a bit different to what is provided by the
// `ProgressAction` where the completion time is some
// absolute time at which the action is finished.
type FixedAction struct {
	Amount         int               `json:"amount"`
	Remaining      int               `json:"remaining"`
	CompletionTime duration.Duration `json:"completion_time"`

	BaseUpgradeAction
}

// valid :
// Used to refine the behavior of the base upgrade action
// to make sure that the amounts provided for this action
// are correct.
//
// Returns `true` if this action is not obviously wrong.
func (a FixedAction) valid() bool {
	return a.BaseUpgradeAction.valid() &&
		a.Amount > 0 &&
		a.Remaining >= 0 &&
		a.Remaining <= a.Amount
}

// computeTotalCost :
// Used to compute the construction cost of the action
// based on the total number of unit described by it.
// It uses the provided table to retrieve the actual
// cost of a single unit.
//
// The `costs` defines the initial costs of a single
// unit. The map is indexed by ID key (so one of them
// should match the `a.ElementID` value).
//
// Returns a slice containing for each resource that
// is needed for this action the total amount that is
// still needed given the `a.Remaining` number to be
// built. In case the input map does not define anything
// for the action an error is returned.
func (a FixedAction) computeTotalCost(costs map[string]FixedCost) ([]ResourceAmount, error) {
	// Find this action in the input table.
	cost, ok := costs[a.ElementID]

	if !ok {
		return []ResourceAmount{}, fmt.Errorf("Cannot compute cost for action \"%s\" defining unknown element \"%s\"", a.ID, a.ElementID)
	}

	needed := cost.ComputeCosts(a.Remaining)

	return needed, nil
}

// Validate :
// Similar to the equivalent method in the `ProgressAction`
// method: required to implement the interface defined by
// the `UpgradeAction`.
//
// The `tools` allow to define the technological deps
// between elements and some resources that are present
// on the place where the action should be launched.
//
// Returns `true` if the action can be launched given
// the information provided in input.
func (a FixedAction) Validate(tools validationTools) (bool, error) {
	// TODO: Should add the resources on a planet to be able
	// to validate the costs computed for this action against
	// a certain amount of resources.
	return false, fmt.Errorf("Not implemented")
}
