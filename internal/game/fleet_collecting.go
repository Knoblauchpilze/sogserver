package game

import (
	"fmt"
	"math"
	"oglike_server/internal/model"
)

// collectingProps :
// Used as a convenience way to represent the data
// needed to perform the collecting of resources in
// a resources pool. The two main cases are when a
// fleet arrives at a debris field or to a planet
// to perform the plunder.
//
// The `capacity` defines the total capacity of the
// transporting ships in the fleet. This capacity
// may already be used by conventional resources
// transported by the fleet.
//
// The `available` defines how much of the total
// collecting capacity is still available. It is
// at most equal to the `capacity`.
//
// The `collected` defines the resources that were
// collected by the fleet defining the properties.
// This slice is `nil` until the `collect` method
// has been called in which case it is set to the
// resources collected.
type collectingProps struct {
	capacity  int
	available float32
	collected []model.ResourceAmount
}

// newHarvestingProps :
// Used to create a new collecting properties obj
// from the input fleet. The data related to ships
// will be analyzed. The capacity will be computed
// by only considering the recyclers in the fleet.
//
// The `f` defines the input fleet from which the
// collecting props should be built.
//
// The `ships` helps gathering information on the
// ships composing the fleet.
//
// Returns the output collecting props along with
// any error.
func newHarvestingProps(f *Fleet, ships *model.ShipsModule) (collectingProps, error) {
	cp := collectingProps{}

	// Compute the harvesting capacity of the fleet.
	// To do so we will first compute the proportion
	// of the total cargo space used by conventional
	// resources.
	// This will help deducing the available space
	// for harvesting.
	totalCargoSpace := float32(0.0)

	for _, s := range f.Ships {
		sd, err := ships.GetShipFromID(s.ID)
		if err != nil {
			return cp, ErrUnableToSimulateFleet
		}

		totalCargoSpace += float32(s.Count * sd.Cargo)

		if sd.Name == "recycler" {
			cp.capacity += s.Count * sd.Cargo
		}
	}

	usedCargoSpace := f.usedCargoSpace()

	// We consider that the harvesting capacity
	// is taken last. So we first compute the
	// remaining cargo space available (through
	// `conventionalCargoSpace - usedCargoSpace`.
	// The `conventionalCargoSpace` is the diff
	// between the harvestgin cargo space and
	// the rest of the cargo which serves no
	// particular purpose but cannot be used to
	// harvest some resources.
	// This value is:
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
	conventionalCargoSpace := totalCargoSpace - float32(cp.capacity)

	availableCargoSpace := conventionalCargoSpace - usedCargoSpace
	cp.available = availableCargoSpace + float32(cp.capacity)

	if cp.available > float32(cp.capacity) {
		cp.available = float32(cp.capacity)
	}

	return cp, nil
}

// newPillagingProps :
// Used in a similar way to `newHarvestingProps`
// but to build collecting properties allowing
// to pillage the resources of a planet or moon.
// Unlike the harvesting case, all ships can be
// used to pillage resources.
//
// The `f` defines the input fleet from which
// the collecting props should be built.
//
// The `ships` helps gatjering information on
// the ships composing the fleet.
//
// Returns the output collecting props along
// with any error.
func newPillagingProps(f *Fleet, ships *model.ShipsModule) (collectingProps, error) {
	cp := collectingProps{}

	// The total cargo space is the sum of all
	// the cargo space of ships belonging to
	// the fleet.
	totalCargoSpace := float32(0.0)

	for _, s := range f.Ships {
		sd, err := ships.GetShipFromID(s.ID)
		if err != nil {
			return cp, ErrUnableToSimulateFleet
		}

		totalCargoSpace += float32(s.Count * sd.Cargo)

		cp.capacity += s.Count * sd.Cargo
	}

	usedCargoSpace := f.usedCargoSpace()

	// We have to subtract the resources already
	// carried by the fleet from the total. This
	// will leave only the available space to be
	// filled with plundered resources.
	cp.available = float32(cp.capacity) - usedCargoSpace

	if cp.available < 0 {
		cp.available = 0
	}

	return cp, nil
}

// newACSPillagingProps :
// Used in a similar way to `newPillagingProps`
// but to build the collecting properties for
// an ACS fleet. It does not differ that much
// during the collecting phase as the fleet is
// still treated as a single entity. The actual
// split of the pillaged resources for each
// individual fleet is done afterwards.
//
// The `acs` defines the input ACS operation
// which should pillage something.
//
// The `ships` helps gathering information on
// ships composing the fleet.
//
// Returns the output collecting props along
// with any error.
func newACSPillagingProps(acs *ACSFleet, ships *model.ShipsModule) (collectingProps, error) {
	// TODO: Implement this.
	return collectingProps{}, fmt.Errorf("Not implemented")
}

// collect :
// Generic implementation of the collecting
// process on a slice of resources. We assume
// that the capacity and checks to ensure that
// the resources are actually meaningful have
// already been performed.
// The collected resources will be stored in
// the `collected` field of the `cp` object.
//
// The `resources` represent the list of res
// to collect.
func (cp *collectingProps) collect(resources []model.ResourceAmount) {
	leftToCollect := float32(0.0)
	for _, res := range resources {
		leftToCollect += res.Amount
	}

	resourcesTypesToCarry := len(resources)

	collected := make(map[string]float32)

	for cp.available > 0.1 && leftToCollect > 0.1 {
		// The amount of resources collected in this
		// pass is a fair share of all the resources
		// types available.
		toCollect := cp.available / float32(resourcesTypesToCarry)

		// Collect each resource: the actual amount
		// collected might be smaller in case there's
		// not enough resources in the field.
		for id := range resources {
			res := &resources[id]

			carried := collected[res.Resource]

			// Cannot collect more that what's available.
			collectedAmount := float32(math.Min(float64(res.Amount), float64(toCollect)))

			cp.available -= collectedAmount
			res.Amount -= collectedAmount
			leftToCollect -= collectedAmount

			// If one resource is depleted, decrease the
			// available resource type: this will allow
			// to speed up the repartition of the other
			// resource types.
			if res.Amount <= 0.0 {
				resourcesTypesToCarry--
			}

			// Collect the resources.
			carried += collectedAmount
			collected[res.Resource] = carried
		}
	}

	// Convert the map of resources collected to a
	// slice.
	cp.collected = make([]model.ResourceAmount, 0)

	for res, amount := range collected {
		ra := model.ResourceAmount{
			Resource: res,
			Amount:   amount,
		}

		cp.collected = append(cp.collected, ra)
	}
}

// harvest :
// Used to perform the collecting of the input
// debris field given the properties of the
// harvester fleet describe by these properties.
// The list of harvested resources is saved in
// the `collected` field of the props.
//
// The `df` defines the debris field to harvest.
func (cp *collectingProps) harvest(df DebrisField) {
	cp.collect(df.Resources)
}

// pillage :
// Used to perform the collecting of resources
// on the provided planet given the capacity and
// available space in the props.
// The list of collected resources is saved in
// the `collected` field of the props.

// The `p` defines the planet from which res
// shouls be collected.
//
// The `ratio` defines the collecting rate on the
// planet: it corresponds to the maximum amount
// of resources that will be pillaged from the
// stock on the planet in case the capacity does
// allow it.
//
// The `data` allows to gather properties of
// resources to pillage.
//
// Returns any error.
func (cp *collectingProps) pillage(p *Planet, ratio float32, data Instance) error {
	// Compute the amount of resources to be
	// plundered from the planet. We consider
	// only the movable resources and using
	// the provided pillage ratio.
	toPillage := make([]model.ResourceAmount, 0)

	for _, res := range p.Resources {
		rDesc, err := data.Resources.GetResourceFromID(res.Resource)
		if err != nil {
			return err
		}

		if !rDesc.Movable {
			continue
		}

		r := model.ResourceAmount{
			Resource: res.Resource,
			Amount:   res.Amount * ratio,
		}

		toPillage = append(toPillage, r)
	}

	// Collect resources.
	cp.collect(toPillage)

	return nil
}
