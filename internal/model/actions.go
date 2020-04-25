package model

import (
	"fmt"
	"math"
	"oglike_server/pkg/db"
	"oglike_server/pkg/duration"
	"time"
)

// Action :
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
//
// The `Element` defines the identifier of the element to
// link to the action. It can either be the identifier of
// the building that is built, the identifier of the ship
// that is produced etc.
type Action struct {
	ID      string `json:"id"`
	Planet  string `json:"planet"`
	Element string `json:"element"`
}

// ErrInvalidAction :
// Used to indicate that the action provided in input is
// not valid.
var ErrInvalidAction = fmt.Errorf("Invalid action with no identifier")

// ErrDuplicatedAction :
// Used to indicate that the action's identifier provided
// input is not unique in the DB.
var ErrDuplicatedAction = fmt.Errorf("Invalid not unique action")

// ErrInvalidDuration :
// Used indicate that the duration for the completion time
// of an action is not valid.
var ErrInvalidDuration = fmt.Errorf("Cannot convert completion time to duration")

// Valid :
// Used to deterimne whether this action is obviously
// not valid or not. This allows to prevent some basic
// cases where it's not even worth to try to register
// an action in the DB.
//
// Returns `true` if the action is valid.
func (a *Action) Valid() bool {
	return validUUID(a.ID) &&
		validUUID(a.Planet) &&
		validUUID(a.Element)
}

// ProgressAction :
// Specialization of the `Action` to handle the case
// of actions related to an element to upgrade to some
// upper level. It typically applies to buildings and
// technologies. Compared to the base upgrade action
// this type of element has two levels (the current
// one and the desired one) and a way to compute the
// total cost needed for the upgrade.
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
	Action

	CurrentLevel   int       `json:"current_level"`
	DesiredLevel   int       `json:"desired_level"`
	CompletionTime time.Time `json:"completion_time"`
}

// newProgressActionFromDB :
// Used to query the progress action referred by the
// input identifier assuming it is contained in the
// provided table.
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
func newProgressActionFromDB(ID string, data Instance, table string) (ProgressAction, error) {
	// Create the action.
	a := ProgressAction{
		Action: Action{
			ID: ID,
		},
	}

	// Consistency.
	if a.ID == "" {
		return a, ErrInvalidAction
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"planet",
			"element",
			"current_level",
			"desired_level",
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

	// Scan the action's data.
	err = dbRes.Scan(
		&a.Planet,
		&a.Element,
		&a.CurrentLevel,
		&a.DesiredLevel,
		&a.CompletionTime,
	)

	// Make sure that it's the only action.
	if dbRes.Next() {
		return a, ErrDuplicatedAction
	}

	return a, nil
}

// Valid :
// Specialize the behavior provided by the `Action`
// base element in order to include the verification
// that both levels are at least positive.
//
// Returns `true` if this action is valid.
func (a *ProgressAction) valid() bool {
	return a.Action.Valid() &&
		a.CurrentLevel >= 0 &&
		a.DesiredLevel >= 0
}

// FixedAction :
// Specialization of the `Action` to describe an action
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
	Action

	Amount         int               `json:"amount"`
	Remaining      int               `json:"remaining"`
	CompletionTime duration.Duration `json:"completion_time"`
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
func newFixedActionFromDB(ID string, data Instance, table string) (FixedAction, error) {
	// Create the action.
	a := FixedAction{
		Action: Action{
			ID: ID,
		},
	}

	// Consistency.
	if a.ID == "" {
		return a, ErrInvalidAction
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

	// Scan the action's data.
	err = dbRes.Scan(
		&a.Planet,
		&a.Element,
		&a.Amount,
		&a.Remaining,
		&a.CompletionTime,
	)

	// Make sure that it's the only action.
	if dbRes.Next() {
		return a, ErrDuplicatedAction
	}

	return a, nil
}

// Valid :
// Refines the behavior provided by the base `Action`
// to make sure that the amount and remaining quantities
// for this action are valid.
//
// Returns `true` if this action is not obviously wrong.
func (a *FixedAction) Valid() bool {
	return a.Action.Valid() &&
		a.Amount > 0 &&
		a.Remaining >= 0 &&
		a.Remaining <= a.Amount
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
// Returns any error.
func (a *FixedAction) computeCompletionTime(data Instance, cost FixedCost) error {
	// Retrieve the planet associated to this action.
	planet, err := NewPlanetFromDB(a.Planet, data)
	if err != nil {
		return err
	}

	costs := cost.ComputeCost(a.Remaining)

	// Retrieve the level of the shipyard and the nanite
	// factory: these are the two buildings that have an
	// influence on the completion time.
	shipyardID, err := data.Buildings.getIDFromName("shipyard")
	if err != nil {
		return err
	}
	naniteID, err := data.Buildings.getIDFromName("nanite factory")
	if err != nil {
		return err
	}

	shipyard, err := planet.GetBuilding(shipyardID)
	if err != nil {
		return err
	}
	nanite, err := planet.GetBuilding(naniteID)
	if err != nil {
		return err
	}

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

	a.CompletionTime = duration.Duration{t}

	return nil
}

// BuildingAction :
// Used as a way to refine the `ProgressAction` for the
// specific case of buildings. It mostly add the info
// to compute the completion time for a building.
//
// The `ProdEffects` describes the production changes
// to apply in case this action completes. It will used
// to add this value to the production of said resource
// on the planet where this action is performed.
//
// The `StorageEffects` are similar to the production
// effects except it applies to the storage capacities
// of a resource on a planet.
type BuildingAction struct {
	ProgressAction

	Production []ProductionEffect `json:"production_effects"`
	Storage    []StorageEffect    `json:"storage_effects"`
}

// ProductionEffect :
// Defines a production effect that a building upgrade
// action can have on the production of a planet. It is
// used to regroup the resource and the value of the
// change brought by the building upgrade action.
//
// The `Action` defines the identifier of the action to
// which this effect is linked.
//
// The `Resource` defines the resource which is changed
// by the building upgrade action.
//
// The `Production` defines the actual effect of the
// upgrade action. This value should be added to the
// existing production for the resource on the planet
// in case the action completes.
type ProductionEffect struct {
	Resource   string  `json:"resource"`
	Production float32 `json:"production_change"`
}

// StorageEffect :
// Defines a storage effect that a building upgrade
// action can have on the capacity of a resource that
// can be stored on a planet. It is used to regroup
// the resource and the value of the change brought
// by the building upgrade action.
//
// The `Resource` defines the resource which is changed
// by the building upgrade action.
//
// The `Storage` defines the actual effect of the
// upgrade action. This value should be added to the
// existing storage capacity for the resource if the
// upgrade action completes.
type StorageEffect struct {
	Resource string  `json:"res"`
	Storage  float32 `json:"capacity_change"`
}

// Valid :
// Used to refine the behavior of the progress action to
// make sure that the levels provided for this action are
// correct.
//
// Returns `true` if this action is not obviously wrong.
func (a *BuildingAction) Valid() bool {
	return a.ProgressAction.valid() &&
		math.Abs(float64(a.DesiredLevel)-float64(a.CurrentLevel)) == 1
}

// NewBuildingActionFromDB :
// Used to query the building action referred by the
// input identifier from the DB. It assumes that the
// action already exists under the specified ID.
//
// The `ID` defines the identifier of the action to
// fetch from the DB.
//
// The `data` allows to actually access to the data
// in the DB.
//
// Returns the corresponding building action along
// with any error.
func NewBuildingActionFromDB(ID string, data Instance) (BuildingAction, error) {
	// Create the return value and fetch the base
	// data for this action.
	ba := BuildingAction{}
	ba.ID = ID

	var err error
	ba.ProgressAction, err = newProgressActionFromDB(ID, data, "construction_actions_buildings")

	if err != nil {
		return ba, err
	}

	err = ba.fetchProductionEffects(data)
	if err != nil {
		return ba, err
	}

	err = ba.fetchStorageEffects(data)

	return ba, err
}

// fetchProductionEffects :
// Used to fetch the effects related to this action
// regarding production capacities from the DB.
//
// The `data` provide a way to access to the DB.
//
// Returns any error.
func (a *BuildingAction) fetchProductionEffects(data Instance) error {
	// Consistency.
	if a.ID == "" {
		return ErrInvalidAction
	}

	a.Production = make([]ProductionEffect, 0)

	query := db.QueryDesc{
		Props: []string{
			"res",
			"new_production",
		},
		Table: "construction_actions_buildings_production_effects",
		Filters: []db.Filter{
			{
				Key:    "action",
				Values: []string{a.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}

	// Populate the return value.
	var pe ProductionEffect

	for dbRes.Next() {
		err = dbRes.Scan(
			&pe.Resource,
			&pe.Production,
		)

		if err != nil {
			return err
		}

		a.Production = append(a.Production, pe)
	}

	return nil
}

// fetchStorageEffects :
// Used to fetch the effects related to this action
// regarding storage capacities from the DB.
//
// The `data` provide a way to access to the DB.
//
// Returns any error.
func (a *BuildingAction) fetchStorageEffects(data Instance) error {
	// Consistency.
	if a.ID == "" {
		return ErrInvalidAction
	}

	a.Storage = make([]StorageEffect, 0)

	query := db.QueryDesc{
		Props: []string{
			"res",
			"new_storage_capacity",
		},
		Table: "construction_actions_buildings_storage_effects",
		Filters: []db.Filter{
			{
				Key:    "action",
				Values: []string{a.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}

	// Populate the return value.
	var se StorageEffect

	for dbRes.Next() {
		err = dbRes.Scan(
			&se.Resource,
			&se.Storage,
		)

		if err != nil {
			return err
		}

		a.Storage = append(a.Storage, se)
	}

	return nil
}

// ConsolidateCompletionTime :
// Used to update the completion time required for this
// action to complete based on the amount of resources
// needed by the next level of the building level.
//
// The `data` allows to get information on the data
// that will be used to compute the completion time.
//
// Returns any error.
func (a *BuildingAction) ConsolidateCompletionTime(data Instance) error {
	// First, we need to determine the cost for each of
	// the individual unit to produce.
	bd, err := data.Buildings.getBuildingFromID(a.Element)
	if err != nil {
		return err
	}

	// Retrieve the planet associated to this action.
	planet, err := NewPlanetFromDB(a.Planet, data)
	if err != nil {
		return err
	}

	costs := bd.Cost.ComputeCost(a.DesiredLevel)

	// Retrieve the level of the robotics factory and the
	// nanite factory: these are the two buildings having
	// an influence on the completion time.
	roboticsID, err := data.Buildings.getIDFromName("robotics factory")
	if err != nil {
		return err
	}
	naniteID, err := data.Buildings.getIDFromName("nanite factory")
	if err != nil {
		return err
	}

	robotics, err := planet.GetBuilding(roboticsID)
	if err != nil {
		return err
	}
	nanite, err := planet.GetBuilding(naniteID)
	if err != nil {
		return err
	}

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

	hours := float64(m+c) / (2500.0 * (1.0 + float64(robotics.Level)) * math.Pow(2.0, float64(nanite.Level)))

	t, err := time.ParseDuration(fmt.Sprintf("%fh", hours))
	if err != nil {
		return ErrInvalidDuration
	}

	a.CompletionTime = time.Now().Add(t)

	return nil
}

// TechnologyAction :
// Used as a way to refine the `ProgressAction` for the
// specific case of technologies. It mostly add the info
// to compute the completion time for a technology.
type TechnologyAction struct {
	ProgressAction
}

// Valid :
// Used to refine the behavior of the progress action to
// make sure that the levels provided for this action are
// correct.
//
// Returns `true` if this action is not obviously wrong.
func (a *TechnologyAction) Valid() bool {
	return a.ProgressAction.Valid() &&
		a.DesiredLevel == a.CurrentLevel+1
}

// NewTechnologyActionFromDB :
// Used similarly to the `NewBuildingActionFromDB`
// element but to fetch the actions related to the
// research of a new technology by a player on a
// given planet.
//
// The `ID` defines the identifier of the action to
// fetch from the DB.
//
// The `data` allows to actually access to the
// data in the DB.
//
// Returns the corresponding technology action
// along with any error.
func NewTechnologyActionFromDB(ID string, data Instance) (TechnologyAction, error) {
	// Create the return value and fetch the base
	// data for this action.
	ta := TechnologyAction{}
	ta.ID = ID

	var err error
	ta.ProgressAction, err = newProgressActionFromDB(ID, data, "construction_actions_technologies")

	return ta, err
}

// ConsolidateCompletionTime :
// Used to update the completion time required for this
// action to complete based on the amount of resources
// needed by the next level of the technology level. It
// also uses the level of research labs for the player
// researching the technology.
//
// The `data` allows to get information on the data
// that will be used to compute the completion time.
//
// Returns any error.
func (a *TechnologyAction) ConsolidateCompletionTime(data Instance) error {
	// First, we need to determine the cost for each of
	// the individual unit to produce.
	td, err := data.Technologies.getTechnologyFromID(a.Element)
	if err != nil {
		return err
	}

	// Retrieve the planet associated to this action.
	planet, err := NewPlanetFromDB(a.Planet, data)
	if err != nil {
		return err
	}

	costs := td.Cost.ComputeCost(a.DesiredLevel)

	// Retrieve the level of the research lab.
	// TODO: We should aggregate that with the intergalactic
	// research network at some point.
	labID, err := data.Buildings.getIDFromName("research lab")
	if err != nil {
		return err
	}
	lab, err := planet.GetBuilding(labID)
	if err != nil {
		return err
	}

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

	hours := float64(m+c) / (1000.0 * (1.0 + float64(lab.Level)))

	t, err := time.ParseDuration(fmt.Sprintf("%fh", hours))
	if err != nil {
		return ErrInvalidDuration
	}

	a.CompletionTime = time.Now().Add(t)

	return nil
}

// ShipAction :
// Used as a convenience define to refer to the action
// of creating one or several ships on a planet.
//
type ShipAction struct {
	FixedAction
}

// DefenseAction :
// Used as a convenience define to refer to the action
// of creating one or more defense systems on a planet.
type DefenseAction struct {
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
func NewShipActionFromDB(ID string, data Instance) (ShipAction, error) {
	// Create the action.
	a := ShipAction{}
	a.ID = ID

	var err error
	a.FixedAction, err = newFixedActionFromDB(ID, data, "construction_actions_ships")

	return a, err
}

// ConsolidateCompletionTime :
// Used to update the completion time required for this
// action to complete based on the amount of units to be
// produced.
//
// The `data` allows to get information on the buildings
// that will be used to compute the completion time.
//
// Returns any error.
func (a *ShipAction) ConsolidateCompletionTime(data Instance) error {
	// First, we need to determine the cost for each of
	// the individual unit to produce.
	sd, err := data.Ships.getShipFromID(a.Element)
	if err != nil {
		return err
	}

	// Use the base handler.
	return a.computeCompletionTime(data, sd.Cost)
}

// NewDefenseActionFromDB :
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
// Returns the corresponding defense action along
// with any error.
func NewDefenseActionFromDB(ID string, data Instance) (DefenseAction, error) {
	// Create the action.
	a := DefenseAction{}
	a.ID = ID

	var err error
	a.FixedAction, err = newFixedActionFromDB(ID, data, "construction_actions_defenses")

	return a, err
}

// ConsolidateCompletionTime :
// Used to update the completion time required for this
// action to complete. It uses internally the base handler
// which allow to handle the actual completion of the time.
// This wrapper is there to fetch the cost associate to
// the ship to build.
//
// The `data` allows to get information from the DB.
//
// Returns any error.
func (a *DefenseAction) ConsolidateCompletionTime(data Instance) error {
	// First, we need to determine the cost for each of
	// the individual unit to produce.
	dd, err := data.Defenses.getDefenseFromID(a.Element)
	if err != nil {
		return err
	}
	// Use the base handler.
	return a.computeCompletionTime(data, dd.Cost)
}
