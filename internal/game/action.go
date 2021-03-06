package game

import (
	"fmt"
	"time"
)

// action :
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
// Note that it can refer to either a planet or a moon.
//
// The `Element` defines the identifier of the element to
// link to the action. It can either be the identifier of
// the building that is built, the identifier of the ship
// that is produced etc.
//
// The `Costs` defines the total cost of this action as
// an array of resources and quantities. This is used to
// actually remove the cost of the action from the res
// available on the planet where this action is started.
//
// The `creationTime` defines the date at which the action
// is set to start.
//
// Used to define whether this action is related to a
// planet or a moon. This will condition the tables to
// use to fetch or save information.
type action struct {
	ID           string `json:"id"`
	Planet       string `json:"planet,omitempty"`
	Element      string `json:"element"`
	Costs        []Cost `json:"-"`
	creationTime time.Time
	moon         bool
}

// Cost :
// Convenience structure allowing to define a cost for
// an element. It regroups the resource identifier and
// the actual amount needed.
//
// The `Resource` represents the identifier of the res
// that this cost represents.
//
// The `Cost` defines the amount of resource that is
// needed.
type Cost struct {
	Resource string  `json:"resource"`
	Cost     float32 `json:"cost"`
}

// ErrInvalidPlanetForAction : The planet associated to an action is not valid.
var ErrInvalidPlanetForAction = fmt.Errorf("invalid planet associated to action")

// ErrInvalidPlayerForAction : The player associated to an action is not valid.
var ErrInvalidPlayerForAction = fmt.Errorf("invalid player associated to action")

// ErrInvalidElementForAction : The element associated to an action is not valid.
var ErrInvalidElementForAction = fmt.Errorf("invalid element associated to action")

// ErrMismatchInVerification : Indicates that the element provided to verify the
// action mismatched the expected values.
var ErrMismatchInVerification = fmt.Errorf("mismatch in verification data for action")

// ErrInvalidDuration : Indicates that the duration of an action could not be validated.
var ErrInvalidDuration = fmt.Errorf("cannot convert completion time to duration for action")

// ErrNonExistingPlanet : Indicates that the planet for the action does not exist.
var ErrNonExistingPlanet = fmt.Errorf("parent planet does not exist")

// ErrNonExistingMoon : Indicates that the moon for the action does not exist.
var ErrNonExistingMoon = fmt.Errorf("parent moon does not exist")

// ErrNonExistingElement : Indicates that the element for the action does not exist.
var ErrNonExistingElement = fmt.Errorf("invalid upgraded element for action")

// valid :
// Determines whether the action is valid. By valid we only
// mean obvious syntax errors.
//
// Returns any error or `nil` if the action seems valid.
func (a *action) valid() error {
	if !validUUID(a.ID) {
		return ErrInvalidElementID
	}
	if !validUUID(a.Planet) {
		return ErrInvalidPlanetForAction
	}
	if !validUUID(a.Element) {
		return ErrInvalidElementForAction
	}

	return nil
}

// newAction :
// Used to create an empty base action from the input
// identifier.
//
// The `ID` defines the identifier for this action.
//
// The `moon` defines whether this action is linked
// to a moon (or a planet).
//
// Returns the created action along with any error.
func newAction(ID string, moon bool) (action, error) {
	// Create the action.
	a := action{
		ID:   ID,
		moon: moon,
	}

	// Consistency.
	if !validUUID(a.ID) {
		return a, ErrInvalidElementID
	}

	return a, nil
}
