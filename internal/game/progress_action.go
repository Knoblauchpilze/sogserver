package game

import (
	"fmt"
	"oglike_server/pkg/db"
	"time"
)

// ProgressAction :
// Specialization of the `action` to handle the case
// of actions related to an element to upgrade to
// some higher level. It typically applies to the
// case of buildings and technologies. Compared to
// the base upgrade action this type of element has
// two levels (the current one and the desired one)
// and a way to compute the total cost needed for
// the upgrade.
type ProgressAction struct {
	// Reuses the notion of a simple action.
	action

	// The `CurrentLevel` represents the current level
	// of the element to upgrade.
	CurrentLevel int `json:"current_level"`

	// The `DesiredLevel` represents the desired level
	// of the element after the upgrade action is done.
	DesiredLevel int `json:"desired_level"`

	// Points defined how many points will this level
	// of the progress action will bring to the player
	// successfully completing it.
	Points float32 `json:"points"`

	// The `CompletionTime` will be computed from the
	// cost of the action and the facilities existing
	// on the planet where the action is triggered.
	CompletionTime time.Time `json:"completion_time"`
}

// ErrInvalidLevelForAction : Indicates that the action has an invalid level.
var ErrInvalidLevelForAction = fmt.Errorf("invalid level provided for action")

// ErrLevelIncorrect : The level provided for the action is not consistent
// what's available in the verification data.
var ErrLevelIncorrect = fmt.Errorf("invalid level compared to planet for action")

// ErrOnlyOneActionAuthorized : Indicates that another action of the same kind is already running.
var ErrOnlyOneActionAuthorized = fmt.Errorf("only a single action of that kind allowed")

// valid :
// Determines whether this action is valid. By valid we
// only mean obvious syntax errors.
//
// Returns any error or `nil` if the action seems valid.
func (a *ProgressAction) valid() error {
	if err := a.action.valid(); err != nil {
		return err
	}

	if a.CurrentLevel < 0 {
		return ErrInvalidLevelForAction
	}
	if a.DesiredLevel < 0 {
		return ErrInvalidLevelForAction
	}

	return nil
}

// newProgressFromDB :
// Used to query the progress action referred by
// the input identifier assuming it is contained
// in the provided table.
//
// The `ID` defines the identifier of the action
// to fetch from the DB.
//
// The `data` allows to actually access to the
// data in the DB.
//
// The `table` defines the name of the table to
// be queried for this action.
//
// The `moon` argument defines whether this action
// links to a moon or a planet.
//
// Returns the progress action along with any
// error.
func newProgressActionFromDB(ID string, data Instance, table string, moon bool) (ProgressAction, error) {
	// Create the action.
	a := ProgressAction{}

	var err error
	a.action, err = newAction(ID, moon)

	// Consistency.
	if err != nil {
		return a, err
	}

	body := "planet"
	if moon {
		body = "moon"
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			body,
			"element",
			"current_level",
			"desired_level",
			"points",
			"completion_time",
			"created_at",
		},
		Table: table,
		Filters: []db.Filter{
			{
				Key:    "id",
				Values: []interface{}{a.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)

	// Check for errors.
	if err != nil {
		return a, err
	}
	defer dbRes.Close()

	if dbRes.Err != nil {
		return a, dbRes.Err
	}

	// Scan the action's data.
	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return a, ErrElementNotFound
	}

	err = dbRes.Scan(
		&a.Planet,
		&a.Element,
		&a.CurrentLevel,
		&a.DesiredLevel,
		&a.Points,
		&a.CompletionTime,
		&a.creationTime,
	)

	// Make sure that it's the only action.
	if dbRes.Next() {
		return a, ErrDuplicatedElement
	}

	return a, err
}
