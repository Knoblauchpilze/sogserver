package game

import "oglike_server/internal/model"

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
	sd, err := data.Ships.GetShipFromID(a.Element)
	if err != nil {
		return err
	}

	costs := sd.Cost.ComputeCost(a.Remaining)

	return p.validateAction(costs, sd.UpgradableDesc, data)
}
