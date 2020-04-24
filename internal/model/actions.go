package model

import (
	"fmt"
	"math"
	"oglike_server/pkg/duration"
	"time"
)

// Action :
// Provide the base building block for an action in the game.
// Basically each time a player wants to start building a new
// element on a planet, or research a technology we use this
// action mechanism to register the intent and keep track of
// the pending operations.
// This approach allows to only update the actions when it is
// needed (typically when the planet where it is registered
// is accessed again) and thus put minimum pressure on the
// server.
// The action can refer to a building, a technology or some
// construction actions such as creating a new ship or a new
// defense system. The common properties are grouped in this
// element.
//
// The `ID` defines the identifier of the action.
//
// The `Planet` defines the planet linked to this action.
// All the action require a parent planet to be scheduled.
//
// The `Element` defines the identifier of the element to
// link to the action. It can either be the identifier of
// the building that is built, the identifier of the ship
// that is produced etc.
type Action struct {
	ID      string `json:"id"`
	Planet  string `json:"planet"`
	Element string `json:"element"`
}

// Valid :
// Used to deterimne whether this action is obviously
// not valid or not. This allows to prevent some basic
// cases where it's not even worth to try to register
// an action in the DB.
//
// Returns `true` if the action is valid.
func (a Action) Valid() bool {
	return validUUID(a.ID) &&
		validUUID(a.Planet) &&
		validUUID(a.Element)
}

// ProgressAction :
// Specialization of the `Action` to handle the case
// of actions related to an element to upgrade to some
// upper level. It typically applies to buildings and
// technologies. Compared to the base upgrade action
// this type of element has two levels (the current
// one and the desired one) and a way to compute the
// total cost needed for the upgrade.
//
// The `CurrentLevel` represents the current level
// of the element to upgrade.
//
// The `DesiredLevel` represents the desired level
// of the element after the upgrade action is done.
//
// The `CompletionTime` will be computed from the
// cost of the action and the facilities existing
// on the planet where the action is triggered.
type ProgressAction struct {
	Action

	CurrentLevel   int       `json:"current_level"`
	DesiredLevel   int       `json:"desired_level"`
	CompletionTime time.Time `json:"completion_time"`
}

// Valid :
// Specialize the behavior provided by the `Action`
// base element in order to include the verification
// that both levels are at least positive.
//
// Returns `true` if this action is valid.
func (a ProgressAction) valid() bool {
	return a.Action.Valid() &&
		a.CurrentLevel >= 0 &&
		a.DesiredLevel >= 0
}

// FixedAction :
// Specialization of the `Action` to describe an action
// that concerns a unit-like element. These elements are
// not upgradable but rather built in a certain amount
// on a planet.
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
	Action

	Amount         int               `json:"amount"`
	Remaining      int               `json:"remaining"`
	CompletionTime duration.Duration `json:"completion_time"`
}

// Valid :
// Refines the behavior provided by the base `Action`
// to make sure that the amount and remaining quantities
// for this action are valid.
//
// Returns `true` if this action is not obviously wrong.
func (a FixedAction) Valid() bool {
	return a.Action.Valid() &&
		a.Amount > 0 &&
		a.Remaining >= 0 &&
		a.Remaining <= a.Amount
}

// BuildingAction :
// Used as a way to refine the `ProgressAction` for the
// specific case of buildings. It mostly add the info
// to compute the completion time for a building.
type BuildingAction struct {
	ProgressAction
}

// Valid :
// Used to refine the behavior of the progress action to
// make sure that the levels provided for this action are
// correct.
//
// Returns `true` if this action is not obviously wrong.
func (a BuildingAction) Valid() bool {
	return a.ProgressAction.valid() &&
		math.Abs(float64(a.DesiredLevel)-float64(a.CurrentLevel)) == 1
}

// TechnologyAction :
// Used as a way to refine the `ProgressAction` for the
// specific case of technologies. It mostly add the info
// to compute the completion time for a technology.
type TechnologyAction struct {
	ProgressAction
}

// Valid :
// Used to refine the behavior of the progress action to
// make sure that the levels provided for this action are
// correct.
//
// Returns `true` if this action is not obviously wrong.
func (a TechnologyAction) Valid() bool {
	return a.ProgressAction.Valid() &&
		a.DesiredLevel == a.CurrentLevel+1
}

// ShipAction :
// Used as a convenience define to refer to the action
// of creating one or several ships on a planet.
//
type ShipAction struct {
	FixedAction
}

// DefenseAction :
// Used as a convenience define to refer to the action
// of creating one or more defense systems on a planet.
type DefenseAction struct {
	FixedAction
}

// NewBuildingActionFromDB :
// Used to query the building action referred by the
// input identifier from the DB. It assumes that the
// action already exists under the specified ID.
//
// The `ID` defines the identifier of the action to
// fetch from the DB.
//
// The `data` allows to actually access to the data
// in the DD.
//
// Returns the corresponding building action along
// with any error.
func NewBuildingActionFromDB(ID string, data Instance) (BuildingAction, error) {
	// TODO: Handle this.
	return BuildingAction{}, fmt.Errorf("Not implemented")
}

// NewTechnologyActionFromDB :
// Used similarly to the `NewBuildingActionFromDB`
// element but to fetch the actions related to the
// research of a new technology by a player on a
// given planet.
//
// The `ID` defines the identifier of the action to
// fetch from the DB.
//
// The `data` allows to actually access to the
// data in the DB.
//
// Returns the corresponding technology action
// along with any error.
func NewTechnologyActionFromDB(ID string, data Instance) (TechnologyAction, error) {
	// TODO: Handle this.
	return TechnologyAction{}, fmt.Errorf("Not implemented")
}

// NewShipActionFromDB :
// Used similarly to the `NewBuildingActionFromDB`
// element but to fetch the actions related to the
// construction of new defense systems on a planet.
//
// The `ID` defines the identifier of the action to
// fetch from the DB.
//
// The `data` allows to actually access to the data
// in the DB.
//
// Returns the corresponding ship action along with
// any error.
func NewShipActionFromDB(ID string, data Instance) (ShipAction, error) {
	// TODO: Handle this.
	return ShipAction{}, fmt.Errorf("Not implemented")
}

// NewDefenseActionFromDB :
// Used similarly to the `NewBuildingActionFromDB`
// element but to fetch the actions related to the
// construction of new defense systems on a planet.
//
// The `ID` defines the identifier of the action to
// fetch from the DB.
//
// The `data` allows to actually access to the data
// in the DB.
//
// Returns the corresponding defense action along
// with any error.
func NewDefenseActionFromDB(ID string, data Instance) (DefenseAction, error) {
	// TODO: Handle this.
	return DefenseAction{}, fmt.Errorf("Not implemented")
}
