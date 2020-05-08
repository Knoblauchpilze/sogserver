package game

import (
	"fmt"
	"oglike_server/internal/model"
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
	action

	CurrentLevel   int       `json:"current_level"`
	DesiredLevel   int       `json:"desired_level"`
	CompletionTime time.Time `json:"completion_time"`
}

// ErrInvalidLevelForAction : Indicates that the action has an invalid level.
var ErrInvalidLevelForAction = fmt.Errorf("Invalid level provided for action")

// ErrLevelIncorrect : The level provided for the action is not consistent
// what's available in the verification data.
var ErrLevelIncorrect = fmt.Errorf("Invalid level compared to planet for action")

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
// Returns the progress action along with any
// error.
func newProgressActionFromDB(ID string, data model.Instance, table string) (ProgressAction, error) {
	// Create the action.
	a := ProgressAction{}

	var err error
	a.action, err = newAction(ID)

	// Consistency.
	if err != nil {
		return a, err
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"t.planet",
			"t.element",
			"t.current_level",
			"t.desired_level",
			"t.completion_time",
			"p.player",
		},
		Table: fmt.Sprintf("%s t inner join planets p on t.planet = p.id", table),
		Filters: []db.Filter{
			{
				Key:    "t.id",
				Values: []string{a.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return a, err
	}
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
		&a.CompletionTime,
		&a.Player,
	)

	// Make sure that it's the only action.
	if dbRes.Next() {
		return a, ErrDuplicatedElement
	}

	return a, err
}
