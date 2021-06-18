package game

import (
	"encoding/json"
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
type FixedAction struct {
	action

	// The `Amount` defines the number of the unit to be
	// produced by this action.
	Amount int `json:"amount"`

	// The `Remaining` defines how many elements are still
	// to be built at the moment of the analysis.
	Remaining int `json:"remaining"`

	// The `CompletionTime` defines the time it takes to
	// complete the construction of a single unit of this
	// element. The remaining time is thus given by the
	// following: `Remaining * CompletionTime`. Note that
	// it is a bit different to what is provided by the
	// `ProgressAction` where the completion time is some
	// absolute time at which the action is finished.
	CompletionTime duration.Duration `json:"completion_time"`
}

// ErrInvalidAmountForAction : Indicates that the action has an invalid amount.
var ErrInvalidAmountForAction = fmt.Errorf("invalid amount provided for action")

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
// The `moon` defines whether this action is linked
// to a moon or a planet.
//
// Returns the progress action along with any error.
func newFixedActionFromDB(ID string, data Instance, table string, moon bool) (FixedAction, error) {
	// Create the action.
	a := FixedAction{}

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
			"amount",
			"remaining",
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

	var t time.Duration

	err = dbRes.Scan(
		&a.Planet,
		&a.Element,
		&a.Amount,
		&a.Remaining,
		&t,
		&a.creationTime,
	)

	a.CompletionTime = duration.NewDuration(t)

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
// The `ratio` defines a flat multiplier to apply to
// the result of the validation and more specifically
// to the computation of the completion time. It helps
// taking into account the properties of the parent's
// universe.
//
// Returns any error.
func (a *FixedAction) computeCompletionTime(data Instance, cost model.FixedCost, p *Planet, ratio float32) error {
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

	// Retrieve the cost in metal and crystal as these are
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
	hours *= float64(ratio)

	t, err := time.ParseDuration(fmt.Sprintf("%fh", hours))
	if err != nil {
		return ErrInvalidDuration
	}

	// The creation time for this action should be as
	// soon as the last action finishes. We have all
	// the relevant information from the input planet
	// to compute it.
	// We also now that the slice fetches the ships
	// and defenses actions in ascending order which
	// means that the last action to finish will be
	// put in last position. In case no construction
	// is available, use the current time as a ref.
	// Note finally that in case the creation time
	// as defined by the last construction action is
	// in the past, we will still use the current
	// time as it probably means that the action is
	// actually some leftover not yet processed.
	a.creationTime = time.Now()
	timeToStart := time.Now()

	if len(p.ShipsConstruction) > 0 {
		s := p.ShipsConstruction[len(p.ShipsConstruction)-1]

		completionTime := s.creationTime.Add(time.Duration(s.Remaining) * s.CompletionTime.Duration)
		if completionTime.After(timeToStart) {
			timeToStart = completionTime
		}
	}
	if len(p.DefensesConstruction) > 0 {
		d := p.DefensesConstruction[len(p.DefensesConstruction)-1]

		completionTime := d.creationTime.Add(time.Duration(d.Remaining) * d.CompletionTime.Duration)
		if completionTime.After(timeToStart) {
			timeToStart = completionTime
		}
	}

	if timeToStart.After(a.creationTime) {
		a.creationTime = timeToStart
	}

	a.CompletionTime = duration.NewDuration(t)

	return nil
}

// MarshalJSON :
// Used to marshal the content defined by this fixed
// cost action in order to make it available to other
// tools.
// This implements the marshaller interface.
//
// Returns the marshalled content and an error.
func (a *FixedAction) MarshalJSON() ([]byte, error) {
	o := struct {
		ID      string `json:"id"`
		Planet  string `json:"planet,omitempty"`
		Element string `json:"element"`

		Amount         int               `json:"amount"`
		Remaining      int               `json:"remaining"`
		CompletionTime duration.Duration `json:"completion_time"`
		CreationTime   time.Time         `json:"created_at"`
	}{
		ID:      a.ID,
		Planet:  a.Planet,
		Element: a.Element,

		Amount:         a.Amount,
		Remaining:      a.Remaining,
		CompletionTime: a.CompletionTime,
		CreationTime:   a.creationTime,
	}

	return json.Marshal(o)
}
