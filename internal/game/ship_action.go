package game

import (
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"oglike_server/pkg/duration"
	"time"
)

// ShipAction :
// Used as a convenience define to refer to the action
// of creating one or several ships on a planet.
//
type ShipAction struct {
	FixedAction
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
func NewShipActionFromDB(ID string, data model.Instance) (ShipAction, error) {
	// Create the action.
	a := ShipAction{}

	// Create the action using the base handler.
	var err error
	a.FixedAction, err = newFixedActionFromDB(ID, data, "construction_actions_ships")

	// Consistency.
	if err != nil {
		return a, err
	}

	// Update the cost for this action. We will fetch
	// the ship related to the action and compute how
	// many resources are needed to build all the ships
	// required by the action.
	sd, err := data.Ships.GetShipFromID(a.Element)
	if err != nil {
		return a, err
	}

	costs := sd.Cost.ComputeCost(a.Remaining)
	a.Costs = make([]Cost, 0)

	for res, amount := range costs {
		c := Cost{
			Resource: res,
			Cost:     float32(amount),
		}

		a.Costs = append(a.Costs, c)
	}

	return a, nil
}

// SaveToDB :
// Used to save the content of this action to
// the DB. In case an error is raised during
// the operation a comprehensive error is
// returned.
//
// The `proxy` allows to access to the DB.
//
// Returns any error.
func (a *ShipAction) SaveToDB(proxy db.Proxy) error {
	// Check consistency.
	if err := a.valid(); err != nil {
		return err
	}

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_ship_upgrade_action",
		Args: []interface{}{
			a,
			a.Costs,
			"planet",
		},
	}

	err := proxy.InsertToDB(query)

	// Analyze the error in order to provide some
	// comprehensive message.
	dbe, ok := err.(db.Error)
	if !ok {
		return err
	}

	fkve, ok := dbe.Err.(db.ForeignKeyViolationError)
	if ok {
		switch fkve.ForeignKey {
		case "planet":
			return ErrNonExistingPlanet
		case "element":
			return ErrNonExistingElement
		}

		return fkve
	}

	return dbe
}

// Convert :
// Implementation of the `db.Convertible` interface
// from the DB package in order to only include fields
// that need to be marshalled in the fleet's creation.
//
// Returns the converted version of this action which
// only includes relevant fields.
func (a *ShipAction) Convert() interface{} {
	return struct {
		ID             string            `json:"id"`
		Planet         string            `json:"planet"`
		Element        string            `json:"element"`
		Amount         int               `json:"amount"`
		Remaining      int               `json:"remaining"`
		CompletionTime duration.Duration `json:"completion_time"`
		CreatedAt      time.Time         `json:"created_at"`
	}{
		ID:             a.ID,
		Planet:         a.Planet,
		Element:        a.Element,
		Amount:         a.Amount,
		Remaining:      a.Remaining,
		CompletionTime: a.CompletionTime,
		CreatedAt:      a.creationTime,
	}
}

// consolidateCompletionTime :
// Used to update the completion time required for this
// action to complete based on the amount of units to be
// produced.
//
// The `data` allows to get information on the buildings
// that will be used to compute the completion time.
//
// The `p` defines the planet attached to this action and
// should be provided as argument to make handling of the
// concurrency easier.
//
// Returns any error.
func (a *ShipAction) consolidateCompletionTime(data model.Instance, p *Planet) error {
	// First, we need to determine the cost for each of
	// the individual unit to produce.
	sd, err := data.Ships.GetShipFromID(a.Element)
	if err != nil {
		return err
	}

	// Use the base handler.
	return a.computeCompletionTime(data, sd.Cost, p)
}

// Validate :
// Used to make sure that the action can be performed on
// the planet it is linked to. This will check that the
// tech tree is consistent with what's expected from the
// ship, that resources are available etc.
//
// The `data` allows to access to the DB if needed.
//
// The `p` defines the planet attached to this action:
// it needs to be provided as input so that resource
// locking is easier.
//
// Returns any error.
func (a *ShipAction) Validate(data model.Instance, p *Planet) error {
	// Consistency.
	if a.Planet != p.ID {
		return ErrMismatchInVerification
	}

	// Update completion time and costs.
	err := a.consolidateCompletionTime(data, p)
	if err != nil {
		return err
	}

	// Compute the total cost of this action and validate
	// against planet's data.
	_, err = data.Ships.GetShipFromID(a.Element)
	if err != nil {
		return err
	}

	// costs := sd.Cost.ComputeCost(a.Remaining)

	// return p.validateAction(costs, sd.UpgradableDesc, data)
	return nil
}
