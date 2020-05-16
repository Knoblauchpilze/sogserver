package game

import (
	"math"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
)

// harvestingProps :
// Used as a convenience way to represent the data
// needed to perform the harvesting of resources in
// a debris field.
//
// The `ships` defines how many recyclers exist in
// the fleet.
//
// The `capacity` defines the total capacity of the
// recyclers in the fleet. This capacity may already
// be used by conventional resoruces transported by
// the fleet.
//
// The `available` defines how much of the total
// harvesting capacity is still available. It is
// at most equal to the `capacity`.
//
// The `collected` defines the resources that were
// collected by the fleet defining the properties.
// This slice is `nil` until the `collect` method
// has been called in which case it is set to the
// resources collected.
type harvestingProps struct {
	ships     int
	capacity  int
	available float32
	collected []model.ResourceAmount
}

// newHarvestingProps :
// Used to create a new harvesting properties obj
// from the input fleet. The data related to ships
// will be analyzed.
//
// The `f` defines the input fleet from which the
// harvesting props should be built.
//
// The `ships` helps gathering information on the
// ships composing the fleet.
//
// Returns the output harvesting props along with
// any error.
func newHarvestingProps(f *Fleet, ships *model.ShipsModule) (harvestingProps, error) {
	hp := harvestingProps{}

	// Compute the harvesting capacity of the fleet.
	// To do so we will first compute the proportion
	// of the total cargo space used by conventional
	// resources.
	// This will help deducing the available space
	// for harvesting.
	totalCargoSpace := float32(0.0)
	usedCargoSpace := float32(0.0)

	for _, s := range f.Ships {
		sd, err := ships.GetShipFromID(s.ID)
		if err != nil {
			return hp, ErrUnableToSimulateFleet
		}

		totalCargoSpace += float32(s.Count * sd.Cargo)

		if sd.Name == "recycler" {
			hp.capacity += s.Count * sd.Cargo
		}
	}

	for _, c := range f.Cargo {
		usedCargoSpace += c.Amount
	}

	// We consider that the harvesting capacity
	// is taken last. So we first compute the
	// remaining cargo space available (through
	// `totalCargoSpace - usedCargoSpace`. This
	// value is:
	//   - either positive in case none of the
	//     harvesting space is used (so the cargo
	//     carried by the fleet is well within
	//     the capacity).
	//   - null if all the harvesting cargo is
	//     available.
	//   - negative if some cargo space is used
	//     already by regular resources.
	// By adding the total harvesting capacity
	// we can deduce the available space left
	// to harvest the debris.
	availableCargoSpace := totalCargoSpace - usedCargoSpace
	hp.available = availableCargoSpace + float32(hp.capacity)

	if hp.available > float32(hp.capacity) {
		hp.available = float32(hp.capacity)
	}

	return hp, nil
}

// harvest :
// Used to perform the harvesting of the input
// debris field given the properties of the
// harvester fleet describe by these properties.
// The list of harvested resources is saved in
// the `collected` field of the props.
//
// The `df` defines the debris field to harvest.
func (hp *harvestingProps) harvest(df DebrisField) {
	leftToHarvest := df.amountDispersed()
	resourcesTypesToCarry := len(df.Resources)

	harvested := make(map[string]float32)

	for hp.available > 0.1 && leftToHarvest > 0.1 {
		// The amount of resources harvested in this
		// pass is a fair share of all the resource
		// types available.
		toHarvest := hp.available / float32(resourcesTypesToCarry)

		// Collect each resource: the actual amount
		// harvested might be smaller in case there's
		// not enough resources in the field.
		for id := range df.Resources {
			dfRes := &df.Resources[id]

			carried := harvested[dfRes.Resource]

			// Cannot collect more that what's available.
			collectedAmount := float32(math.Min(float64(dfRes.Amount), float64(toHarvest)))

			hp.available -= collectedAmount
			dfRes.Amount -= collectedAmount
			leftToHarvest -= collectedAmount

			// If one resource is depleted, decrease the
			// available resource type: this will allow
			// to speed up the repartition of the other
			// resource types.
			if dfRes.Amount <= 0.0 {
				resourcesTypesToCarry--
			}

			// Collect the resources.
			carried += collectedAmount
			harvested[dfRes.Resource] = carried
		}
	}

	// Convert the map of resources collected to a
	// slice.
	hp.collected = make([]model.ResourceAmount, 0)

	for res, amount := range harvested {
		ra := model.ResourceAmount{
			Resource: res,
			Amount:   amount,
		}

		hp.collected = append(hp.collected, ra)
	}
}

// harvest :
// Used to perform the harvesting of resources that
// are dispersed in the debris field specified by
// the target of the fleet.
//
// The `data` allows to access to the DB.
//
// Return any error along with the name of the
// script to execute to finalize the execution of
// the fleet.
func (f *Fleet) harvest(data Instance) (string, error) {
	// The harvesting operation requires to harvest a
	// part or all the resources that are dispersed in
	// a given debris field. The capacity available to
	// harvest is given by the cargo space of all the
	// recyclers belonging to the fleet.
	// We will assume that this function is only called
	// when the fleet is actually ready to process the
	// debris field and we won't check whether it is
	// the case.

	// If the fleet is not returning yet, process the
	// harvesting operation.
	if !f.returning {
		hp, err := newHarvestingProps(f, data.Ships)
		if err != nil {
			return "", ErrUnableToSimulateFleet
		}

		// Retrieve the description of the debris
		// field from the DB.
		df, err := NewDebrisFieldFromDB(f.TargetCoords, f.Universe, data)
		if err != nil {
			return "", ErrUnableToSimulateFleet
		}

		// Harvest this field.
		hp.harvest(df)

		// Now perform the update of resources to
		// the DB given what has been harvested.
		dispersed, err := data.Resources.Description(df.Resources)
		if err != nil {
			return "", ErrUnableToSimulateFleet
		}

		harvested, err := data.Resources.Description(hp.collected)
		if err != nil {
			return "", ErrUnableToSimulateFleet
		}

		query := db.InsertReq{
			Script: "fleet_harvesting_success",
			Args: []interface{}{
				f.ID,
				df.ID,
				hp.collected,
				dispersed,
				harvested,
			},
		}

		err = data.Proxy.InsertToDB(query)
		if err != nil {
			return "", err
		}
	}

	// The `fleet_return_to_base` script is actually
	// safe to call even in the case of a fleet that
	// should not yet return to its base. So we will
	// abuse this fact.
	return "fleet_return_to_base", nil
}
