package data

import (
	"fmt"
	"math"
	"oglike_server/pkg/duration"
	"time"
)

// validationTools :
// Provides a convenience structure regrouping the needed
// information to perform the validation of an upgrade
// action. It contains some information about the costs
// of each element in the game along with some dependencies
// that need to be met for each element.
//
// The `pCosts` represents the list of construction costs
// for elements that have a notion of progression (like
// buildings or technologies).
//
// The `fCosts` defines the list of production cost for
// unit-like elements (typically ships and defenses).
//
// The `techTree` defines the list of dependencies of any
// element of the game to any other.
//
// The `available` represents the amount of each resource
// available at the place of the validation (typically on
// a given planet).
//
// The `buildings` defines the list of buildings that are
// available at the place of validation.
//
// The `technologies` defines the list of all the techs
// that the player already researched.
//
// The `fields` defines the remaining fields count at
// the place of validation.
type validationTools struct {
	pCosts       map[string]ConstructionCost
	fCosts       map[string]FixedCost
	techTree     map[string][]TechDependency
	available    map[string]float32
	buildings    []Building
	technologies []Technology
	fields       int
}

// enoughRes :
// Used to determine whether the input vector of resources
// can be satisfied with what is provided by this element.
//
// The `needed` represents the array of resources that are
// needed for some process.
//
// Returns `true` if the resources can be satisfied.
func (vt validationTools) enoughRes(needed []ResourceAmount) bool {
	for _, res := range needed {
		avail, ok := vt.available[res.Resource]

		if !ok {
			// Not enough of the resource (as it does not
			// even exist).
			return false
		}

		if res.Amount > avail {
			// Not enough of the resource available.
			return false
		}
	}

	return true
}

// meetTechCriteria :
// USed to determine whether the planet registered for this
// validation tools can satisfy the various dependencies of
// the `element`. We will go through what is required and
// see whether it is provided by the planet.
//
// The `element` to analyze. If this element cannot be found
// in the list of dependencies available in this object an
// error will be returned.
func (vt validationTools) meetTechCriteria(element string) (bool, error) {
	// Fetch the dependencies for the input element.
	deps, ok := vt.techTree[element]

	if !ok {
		return false, fmt.Errorf("Could not find tech tree for \"%s\"", element)
	}

	for _, dep := range deps {
		// Search the planet for an element of this nature.
		found := false

		for id := 0; id < len(vt.buildings) && !found; id++ {
			b := vt.buildings[id]

			// Check whether the dependency can be met.
			if b.ID == element {
				found = true

				if b.Level < dep.Level {
					return false, nil
				}
			}
		}

		if !found {
			for id := 0; id < len(vt.technologies) && !found; id++ {
				t := vt.technologies[id]

				// Check whether the dependency can be met.
				if t.ID == element {
					found = true

					if t.Level < dep.Level {
						return false, nil
					}
				}
			}
		}

		// If we didn't found the dependency it means that it
		// does not exist in the planet and thus is not met.
		if !found {
			return false, nil
		}
	}

	return true, nil
}

// UpgradeAction :
// Generic interface describing an upgrade action to perform
// on a planet. This can concern any kind of data but it is
// required to define at least these methods in order to be
// able to correctly be checked against the planet data. It
// mostly consists into evaluating the cost of the action so
// that we can compare it with the resources existing on the
// planet and also providing some way to verify that needed
// buildings/technologies criteria are also met.
type UpgradeAction interface {
	Validate(tools validationTools) (bool, error)
	GetPlanet() string
	UpdateCompletionTime(bm buildingModule) error
}

// computeCost :
// Used to compute the construction cost of the action
// based on the level it aims at reaching and the total
// cost of various elements defined in the input table.
//
// The `costs` defines the initial costs and the rules
// to make the progress for various in-game elements.
// The map is indexed by ID key (so one of them should
// match the `a.ElementID` value).
//
// Returns a slice containing for each resource that
// is needed for this action the amount necessary. In
// case the input map does not define anything for the
// action an error is returned.
func (a *ProgressAction) computeCost(costs map[string]ConstructionCost) ([]ResourceAmount, error) {
	// Find this action in the input table.
	cost, ok := costs[a.ElementID]

	if !ok {
		return []ResourceAmount{}, fmt.Errorf("Cannot compute cost for action \"%s\" defining unknown element \"%s\"", a.ID, a.ElementID)
	}

	needed := cost.ComputeCosts(a.DesiredLevel)

	return needed, nil
}

// Validate :
// Implementation of the `UpgradeAction` interface to
// perform the validation of the data contained in the
// current action against the information provided by
// the game framework. We will check that each element
// required by the validation tools allow the action
// to be performed.
//
// The `tools` allow to define the technological deps
// between elements and some resources that are present
// on the place where the action should be launched.
//
// Returns `true` if the action can be launched given
// the information provided in input.
func (a *ProgressAction) Validate(tools validationTools) (bool, error) {
	// We need to make sure that there are enough resources
	// available given the cost of this action.
	needed, err := a.computeCost(tools.pCosts)
	if err != nil {
		return false, fmt.Errorf("Unable to determine cost for action on \"%s\" (err: %v)", a.ElementID, err)
	}

	if !tools.enoughRes(needed) {
		return false, nil
	}

	// We need to make sure that the technologies and the
	// buildings needed to compute the action are also
	// existing on the planet.
	meet, err := tools.meetTechCriteria(a.ElementID)
	if err != nil {
		return false, fmt.Errorf("Unable to determine whether \"%s\" meets the tech criteria (err: %v)", a.ElementID, err)
	}

	return meet, nil
}

// Validate :
// Refinement of the `ProgressAction` method in order
// to add the verification that the number of fields
// of the planet where the building should be built
// still allows to build one more level of anything.
//
// The `tools` allow to define the technological deps
// between elements and some resources that are present
// on the place where the action should be launched.
//
// Returns `true` if the action can be launched given
// the information provided in input.
func (a *BuildingAction) Validate(tools validationTools) (bool, error) {
	// Make sure that there are some remaining fields on
	// the planet.
	if a.CurrentLevel < a.DesiredLevel && tools.fields == 0 {
		// No more fields available.
		return false, nil
	}

	return a.ProgressAction.Validate(tools)
}

// UpdateCompletionTime :
// Implementation of the `UpgradeAction` to allow the
// computation of a valid completion time for the item
// represented by this upgrade action.
//
// The `bm` defines the building module containing info
// about the levels of some buildings (like the robotics
// factory) which are used to determine the completion
// time.
func (a *BuildingAction) UpdateCompletionTime(bm buildingModule) error {
	// Retrieve the level of the robotics factory and the
	// nanite factory: both have an influence on the time
	// to complete the action.
	robotics, err := bm.getLevelOf("robotics factory")
	if err != nil {
		return fmt.Errorf("Cannot update building action completion time (err: %v)", err)
	}

	nanite, err := bm.getLevelOf("nanite factory")
	if err != nil {
		return fmt.Errorf("Cannot update building action completion time (err: %v)", err)
	}

	// TODO: Get the amount of metal and crystal required.
	m := 1000.0
	c := 1000.0

	hours := (m + c) / (2500.0 * (1.0 + float64(robotics)) * math.Pow(2.0, float64(nanite)))

	t, err := time.ParseDuration(fmt.Sprintf("%fh", hours))
	if err != nil {
		return fmt.Errorf("Cannot convert completion time for building action (err: %v)", err)
	}

	a.CompletionTime = time.Now().Add(t)

	return nil
}

// UpdateCompletionTime :
// Implementation of the `UpgradeAction` to allow the
// computation of a valid completion time for the tech
// represented by this upgrade action.
//
// The `bm` defines the building module containing info
// about the levels of some buildings (like the research
// lab) which are used to determine the completion time.
func (a *TechnologyAction) UpdateCompletionTime(bm buildingModule) error {
	// Retrieve the level of the research lab: it is the
	// only building that has an influence on the total
	// time to complete the action.
	research, err := bm.getLevelOf("research lab")
	if err != nil {
		return fmt.Errorf("Cannot update technology action completion time (err: %v)", err)
	}

	// TODO: Get the amount of metal and crystal required.
	m := 1000.0
	c := 1000.0

	hours := (m + c) / (1000.0 * (1.0 + float64(research)))

	t, err := time.ParseDuration(fmt.Sprintf("%fh", hours))
	if err != nil {
		return fmt.Errorf("Cannot convert completion time for technology action (err: %v)", err)
	}

	a.CompletionTime = time.Now().Add(t)

	return nil
}

// computeTotalCost :
// Used to compute the construction cost of the action
// based on the total number of unit described by it.
// It uses the provided table to retrieve the actual
// cost of a single unit.
//
// The `costs` defines the initial costs of a single
// unit. The map is indexed by ID key (so one of them
// should match the `a.ElementID` value).
//
// Returns a slice containing for each resource that
// is needed for this action the total amount that is
// still needed given the `a.Remaining` number to be
// built. In case the input map does not define anything
// for the action an error is returned.
func (a *FixedAction) computeTotalCost(costs map[string]FixedCost) ([]ResourceAmount, error) {
	// Find this action in the input table.
	cost, ok := costs[a.ElementID]

	if !ok {
		return []ResourceAmount{}, fmt.Errorf("Cannot compute cost for action \"%s\" defining unknown element \"%s\"", a.ID, a.ElementID)
	}

	needed := cost.ComputeCosts(a.Remaining)

	return needed, nil
}

// Validate :
// Similar to the equivalent method in the `ProgressAction`
// method: required to implement the interface defined by
// the `UpgradeAction`.
//
// The `tools` allow to define the technological deps
// between elements and some resources that are present
// on the place where the action should be launched.
//
// Returns `true` if the action can be launched given
// the information provided in input.
func (a *FixedAction) Validate(tools validationTools) (bool, error) {
	// We need to make sure that there are enough resources
	// available given the cost of this action.
	needed, err := a.computeTotalCost(tools.fCosts)
	if err != nil {
		return false, fmt.Errorf("Unable to determine cost for action on \"%s\" (err: %v)", a.ElementID, err)
	}

	if !tools.enoughRes(needed) {
		return false, nil
	}

	// We need to make sure that the technologies and the
	// buildings needed to compute the action are also
	// existing on the planet.
	meet, err := tools.meetTechCriteria(a.ElementID)
	if err != nil {
		return false, fmt.Errorf("Unable to determine whether \"%s\" meets the tech criteria (err: %v)", a.ElementID, err)
	}

	return meet, nil
}

// UpdateCompletionTime :
// Implementation of the `UpgradeAction` to allow the
// computation of a valid completion time for each item
// of this fixed action. This uses the information of the
// building module.
//
// The `bm` defines the building module containing info
// about the levels of some buildings (like the shipyard)
// which are used to determine the completion time.
func (a *FixedAction) UpdateCompletionTime(bm buildingModule) error {
	// Retrieve the level of the shipyard and the nanite
	// factory: these are the two buildings that have an
	// influence on the completion time.
	shipyard, err := bm.getLevelOf("shipyard")
	if err != nil {
		return fmt.Errorf("Cannot update fixed action completion time (err: %v)", err)
	}

	nanite, err := bm.getLevelOf("nanite factory")
	if err != nil {
		return fmt.Errorf("Cannot update fixed action completion time (err: %v)", err)
	}

	// TODO: Get the amount of metal and crystal required.
	m := 1000.0
	c := 1000.0

	hours := (m + c) / (2500.0 * (1.0 + float64(shipyard)) * math.Pow(2.0, float64(nanite)))

	t, err := time.ParseDuration(fmt.Sprintf("%fh", hours))
	if err != nil {
		return fmt.Errorf("Cannot convert completion time for fixed action (err: %v)", err)
	}

	a.CompletionTime = duration.Duration{t}

	return nil
}
