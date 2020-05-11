package game

import (
	"fmt"
	"math"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"oglike_server/pkg/duration"
	"time"
)

// FixedAction :
// Specialization of the `action` to describe an action
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
	action

	Amount         int               `json:"amount"`
	Remaining      int               `json:"remaining"`
	CompletionTime duration.Duration `json:"completion_time"`
}

// ErrInvalidAmountForAction : Indicates that the action has an invalid amount.
var ErrInvalidAmountForAction = fmt.Errorf("Invalid amount provided for action")

// valid :
// Determines whether this action is valid. By valid we
// only mean obvious syntax errors.
//
// Returns any error or `nil` if the action seems valid.
func (a *FixedAction) valid() error {
	if err := a.action.valid(); err != nil {
		return err
	}

	if a.Amount <= 0 {
		return ErrInvalidAmountForAction
	}
	if a.Remaining < 0 {
		return ErrInvalidAmountForAction
	}
	if a.Remaining > a.Amount {
		return ErrInvalidAmountForAction
	}

	return nil
}

// newFixedActionFromDB :
// Similar to the `newProgressActionFromDB` but it
// is used to initialize the fields defined by a
// `FixedAction` data structure.
//
// The `ID` defines the identifier of the action to
// fetch from the DB.
//
// The `data` allows to actually access to the data
// in the DB.
//
// The `table` defines the name of the table to be
// queried for this action.
//
// Returns the progress action along with any error.
func newFixedActionFromDB(ID string, data model.Instance, table string) (FixedAction, error) {
	// Create the action.
	a := FixedAction{}

	var err error
	a.action, err = newAction(ID)

	// Consistency.
	if err != nil {
		return a, err
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"planet",
			"element",
			"amount",
			"remaining",
			"completion_time",
		},
		Table: table,
		Filters: []db.Filter{
			{
				Key:    "id",
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

	var t time.Duration

	err = dbRes.Scan(
		&a.Planet,
		&a.Element,
		&a.Amount,
		&a.Remaining,
		&t,
	)

	a.CompletionTime = duration.Duration{t}

	// Make sure that it's the only action.
	if dbRes.Next() {
		return a, ErrDuplicatedElement
	}

	return a, err
}

// computeCompletionTime :
// Used to update the completion time required for this
// action to complete based on the amount of units to be
// produced.
//
// The `data` allows to get information on the elements
// that will be used to compute the completion time: it
// is usually some buildings existing on the planet that
// is linked to the action.
//
// The `costs` define the amount of resources needed to
// build a single unit.
//
// The `p` defines the planet attached to this action.
// It should be fetched beforehand to make concurrency
// handling easier.
//
// Returns any error.
func (a *FixedAction) computeCompletionTime(data model.Instance, cost model.FixedCost, p *Planet) error {
	// Consistency.
	if a.Planet != p.ID {
		return ErrMismatchInVerification
	}

	costs := cost.ComputeCost(1)

	// Populate the cost of the whole action.
	totCosts := cost.ComputeCost(a.Amount)
	a.Costs = make([]Cost, 0)

	for res, amount := range totCosts {
		c := Cost{
			Resource: res,
			Cost:     float32(amount),
		}

		a.Costs = append(a.Costs, c)
	}

	// Retrieve the level of the shipyard and the nanite
	// factory: these are the two buildings that have an
	// influence on the completion time.
	shipyardID, err := data.Buildings.GetIDFromName("shipyard")
	if err != nil {
		return err
	}
	naniteID, err := data.Buildings.GetIDFromName("nanite factory")
	if err != nil {
		return err
	}

	shipyard := p.Buildings[shipyardID]
	nanite := p.Buildings[naniteID]

	// Retrieve the cost in metal and crystal as it is
	// the only costs that matters.
	metalDesc, err := data.Resources.GetResourceFromName("metal")
	if err != nil {
		return err
	}
	crystalDesc, err := data.Resources.GetResourceFromName("crystal")
	if err != nil {
		return err
	}

	m := costs[metalDesc.ID]
	c := costs[crystalDesc.ID]

	hours := float64(m+c) / (2500.0 * (1.0 + float64(shipyard.Level)) * math.Pow(2.0, float64(nanite.Level)))

	t, err := time.ParseDuration(fmt.Sprintf("%fh", hours))
	if err != nil {
		return ErrInvalidDuration
	}

	a.creationTime = time.Now()
	a.CompletionTime = duration.Duration{t}

	return nil
}
