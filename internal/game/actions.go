package game

import (
	"fmt"
	"math"
	"oglike_server/internal/model"
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
// Note that it can refer to either a planet or a moon.
//
// The `Player` defines the player owning the planet on
// which this action is performed.
//
// The `Element` defines the identifier of the element to
// link to the action. It can either be the identifier of
// the building that is built, the identifier of the ship
// that is produced etc.
//
// The `Costs` defines the total cost of this action as
// an array of resources and quantities. This is used to
// actually remove the cost of the action from the res
// available on the planet where this action is started.
type Action struct {
	ID      string `json:"id"`
	Planet  string `json:"planet"`
	Player  string `json:"player"`
	Element string `json:"element"`
	Costs   []Cost `json:"-"`
}

// Cost :
// Convenience structure allowing to define a cost for
// an element. It regroups the resource identifier and
// the actual amount needed.
//
// The `Resource` represents the identifier of the res
// that this cost represents.
//
// The `Cost` defines the amount of resource that is
// needed.
type Cost struct {
	Resource string  `json:"resource"`
	Cost     float32 `json:"cost"`
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

// ErrLevelIncorrect :
// Used to indicate that the level provided for the action
// is not consistent with the level of the element on the
// parent planet.
var ErrLevelIncorrect = fmt.Errorf("Invalid level compared to planet for action")

// ErrNoFieldsLeft :
// Used to indicate that the action requires at least one
// free field on the planet and that it's not the case.
var ErrNoFieldsLeft = fmt.Errorf("No remaining fields left for action")

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
		validUUID(a.Player) &&
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
func newProgressActionFromDB(ID string, data model.Instance, table string) (ProgressAction, error) {
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
		return a, ErrInvalidAction
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
		return a, ErrDuplicatedAction
	}

	return a, err
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
func newFixedActionFromDB(ID string, data model.Instance, table string) (FixedAction, error) {
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
			"t.planet",
			"t.element",
			"t.amount",
			"t.remaining",
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
		return a, ErrInvalidAction
	}

	var t time.Duration

	err = dbRes.Scan(
		&a.Planet,
		&a.Element,
		&a.Amount,
		&a.Remaining,
		&t,
		&a.Player,
	)

	a.CompletionTime = duration.Duration{t}

	// Make sure that it's the only action.
	if dbRes.Next() {
		return a, ErrDuplicatedAction
	}

	return a, err
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
// The `p` defines the planet attached to this action.
// It should be fetched beforehand to make concurrency
// handling easier.
//
// Returns any error.
func (a *FixedAction) computeCompletionTime(data model.Instance, cost model.FixedCost, p *Planet) error {
	// Consistency.
	if a.Planet != p.ID {
		return ErrInvalidPlanet
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

	shipyard, err := p.GetBuilding(shipyardID)
	if err != nil {
		return err
	}
	nanite, err := p.GetBuilding(naniteID)
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

	Production []ProductionEffect `json:"production_effects,omitempty"`
	Storage    []StorageEffect    `json:"storage_effects,omitempty"`
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
	Resource string  `json:"resource"`
	Storage  float32 `json:"storage_capacity_change"`
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
func NewBuildingActionFromDB(ID string, data model.Instance) (BuildingAction, error) {
	// Create the return value and fetch the base
	// data for this action.
	a := BuildingAction{}
	a.ID = ID

	// Create the action using the base handler.
	var err error
	a.ProgressAction, err = newProgressActionFromDB(ID, data, "construction_actions_buildings")

	if err != nil {
		return a, err
	}

	err = a.fetchProductionEffects(data)
	if err != nil {
		return a, err
	}

	err = a.fetchStorageEffects(data)
	if err != nil {
		return a, err
	}

	// Update the cost for this action. We will fetch
	// the building related to the action and compute
	// how many resources are needed to build it.
	sd, err := data.Buildings.GetBuildingFromID(a.Element)
	if err != nil {
		return a, err
	}

	costs := sd.Cost.ComputeCost(a.CurrentLevel)
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

// fetchProductionEffects :
// Used to fetch the effects related to this action
// regarding production capacities from the DB.
//
// The `data` provide a way to access to the DB.
//
// Returns any error.
func (a *BuildingAction) fetchProductionEffects(data model.Instance) error {
	// Consistency.
	if a.ID == "" {
		return ErrInvalidAction
	}

	a.Production = make([]ProductionEffect, 0)

	query := db.QueryDesc{
		Props: []string{
			"resource",
			"production_change",
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
	if dbRes.Err != nil {
		return dbRes.Err
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
func (a *BuildingAction) fetchStorageEffects(data model.Instance) error {
	// Consistency.
	if a.ID == "" {
		return ErrInvalidAction
	}

	a.Storage = make([]StorageEffect, 0)

	query := db.QueryDesc{
		Props: []string{
			"resource",
			"storage_capacity_change",
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
	if dbRes.Err != nil {
		return dbRes.Err
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

// ConsolidateEffects :
// Used to make sure that the production and storage
// effects for this action are consistent with the
// desired level of the building described by it. It
// uses the input `data` model to access to needed
// information.
// Note that the effects define the difference from
// the existing level and not the absolute value of
// the output state of the action.
//
// The `data` defines a way to access to the effects
// for buildings.
//
// The `p` defines the parent planet where the action
// is meant to be performed. It should be passed to
// this function in order to make locking the resource
// more easily.
//
// Returns any error.
func (a *BuildingAction) ConsolidateEffects(data model.Instance, p *Planet) error {
	// Consistency.
	if a.Planet != p.ID {
		return ErrInvalidPlanet
	}

	// We need to retrieve the building related to this
	// action.
	bd, err := data.Buildings.GetBuildingFromID(a.Element)
	if err != nil {
		return err
	}

	// Update production effects.
	a.Production = make([]ProductionEffect, 0)

	for _, rule := range bd.Production {
		curProd := rule.ComputeProduction(a.CurrentLevel, p.AverageTemperature())
		desiredProd := rule.ComputeProduction(a.DesiredLevel, p.AverageTemperature())

		e := ProductionEffect{
			Resource:   rule.Resource,
			Production: desiredProd - curProd,
		}

		a.Production = append(a.Production, e)
	}

	// And storage effects.
	a.Storage = make([]StorageEffect, 0)

	for _, rule := range bd.Storage {
		curStorage := rule.ComputeStorage(a.CurrentLevel)
		desiredStorage := rule.ComputeStorage(a.DesiredLevel)

		e := StorageEffect{
			Resource: rule.Resource,
			Storage:  float32(desiredStorage - curStorage),
		}

		a.Storage = append(a.Storage, e)
	}

	return nil
}

// consolidateCompletionTime :
// Used to update the completion time required for this
// action to complete based on the amount of resources
// needed by the next level of the building level.
//
// The `data` allows to get information on the data
// that will be used to compute the completion time.
//
// The `p` planet defines the associated planet to this
// action in order to prevent dead lock. We assume that
// it should be fetched before validating the action.
//
// Returns any error.
func (a *BuildingAction) consolidateCompletionTime(data model.Instance, p *Planet) error {
	// Consistency.
	if a.Planet != p.ID {
		return ErrInvalidPlanet
	}

	// First, we need to determine the cost for each of
	// the individual unit to produce.
	bd, err := data.Buildings.GetBuildingFromID(a.Element)
	if err != nil {
		return err
	}

	costs := bd.Cost.ComputeCost(a.CurrentLevel)

	// Populate the cost.
	a.Costs = make([]Cost, 0)

	for res, amount := range costs {
		c := Cost{
			Resource: res,
			Cost:     float32(amount),
		}

		a.Costs = append(a.Costs, c)
	}

	// Retrieve the level of the robotics factory and the
	// nanite factory: these are the two buildings having
	// an influence on the completion time.
	roboticsID, err := data.Buildings.GetIDFromName("robotics factory")
	if err != nil {
		return err
	}
	naniteID, err := data.Buildings.GetIDFromName("nanite factory")
	if err != nil {
		return err
	}

	robotics, err := p.GetBuilding(roboticsID)
	if err != nil {
		return err
	}
	nanite, err := p.GetBuilding(naniteID)
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
func (a *BuildingAction) Validate(data model.Instance, p *Planet) error {
	// Consistency.
	if a.Planet != p.ID {
		return ErrInvalidPlanet
	}

	// Update completion time and costs.
	err := a.consolidateCompletionTime(data, p)
	if err != nil {
		return err
	}

	// Make sure that the current level of the building is
	// consistent with what's desired.
	bd, err := data.Buildings.GetBuildingFromID(a.Element)
	if err != nil {
		return err
	}

	bi, err := p.GetBuilding(bd.ID)
	if err != nil {
		return err
	}

	if bi.Level != a.CurrentLevel {
		return ErrLevelIncorrect
	}

	// Make sure that if the action requires to use one
	// more field there is at least one available.
	if p.RemainingFields() == 0 && a.DesiredLevel > a.CurrentLevel {
		return ErrNoFieldsLeft
	}

	// Validate against planet's data.
	costs := bd.Cost.ComputeCost(a.CurrentLevel)

	return p.validateAction(costs, bd.UpgradableDesc, data)
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
func NewTechnologyActionFromDB(ID string, data model.Instance) (TechnologyAction, error) {
	// Create the return value and fetch the base
	// data for this action.
	a := TechnologyAction{}
	a.ID = ID

	// Create the action using the base handler.
	var err error
	a.ProgressAction, err = newProgressActionFromDB(ID, data, "construction_actions_technologies")
	if err != nil {
		return a, err
	}

	// Update the cost for this action. We will fetch
	// the tech related to the action and compute how
	// many resources are needed to build it.
	sd, err := data.Technologies.GetTechnologyFromID(a.Element)
	if err != nil {
		return a, err
	}

	costs := sd.Cost.ComputeCost(a.CurrentLevel)
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
// action to complete based on the amount of resources
// needed by the next level of the technology level. It
// also uses the level of research labs for the player
// researching the technology.
//
// The `data` allows to get information on the data
// that will be used to compute the completion time.
//
// The `p` argument defines the planet onto which the
// action should be performed. Note that we assume it
// corresponds to the actual planet defined by this
// action. It is used in order not to dead lock with
// the planet that has probably already been acquired
// by the action creation process.
//
// Returns any error.
func (a *TechnologyAction) consolidateCompletionTime(data model.Instance, p *Planet) error {
	// Consistency.
	if a.Planet != p.ID {
		return ErrInvalidPlanet
	}

	// First, we need to determine the cost for each of
	// the individual unit to produce.
	td, err := data.Technologies.GetTechnologyFromID(a.Element)
	if err != nil {
		return err
	}

	costs := td.Cost.ComputeCost(a.CurrentLevel)

	// Populate the cost.
	a.Costs = make([]Cost, 0)

	for res, amount := range costs {
		c := Cost{
			Resource: res,
			Cost:     float32(amount),
		}

		a.Costs = append(a.Costs, c)
	}

	// Fetch the total research power available for this
	// action. It will not account for the current planet
	// research lab so we still have to use it.
	power, err := a.fetchResearchPower(data, p)
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

	hours := float64(m+c) / (1000.0 * (1.0 + float64(power)))

	t, err := time.ParseDuration(fmt.Sprintf("%fh", hours))
	if err != nil {
		return ErrInvalidDuration
	}

	a.CompletionTime = time.Now().Add(t)

	return nil
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
func (a *TechnologyAction) Validate(data model.Instance, p *Planet) error {
	// Consistency.
	if a.Planet != p.ID || a.Player != p.Player {
		return ErrInvalidPlanet
	}

	// Update completion time and costs.
	err := a.consolidateCompletionTime(data, p)
	if err != nil {
		return err
	}

	// Make sure that the current level of the technology
	// is consistent with what's desired.
	td, err := data.Technologies.GetTechnologyFromID(a.Element)
	if err != nil {
		return err
	}

	tLevel, ok := p.technologies[td.ID]
	if !ok && a.CurrentLevel > 0 {
		return ErrLevelIncorrect
	}
	if tLevel != a.CurrentLevel {
		return ErrLevelIncorrect
	}

	// Validate against planet's data.
	costs := td.Cost.ComputeCost(a.CurrentLevel)

	return p.validateAction(costs, td.UpgradableDesc, data)
}

// fetchResearchPower :
// Used to fetch the research power available for the input
// planet. It will query the list of research labs on all
// planets of the player and select the required amount as
// defined by the level of the galactic research network.
// It *will* include the level of the planet linked to this
// action.
//
// The `data` allows to access to the DB.
//
// The `planet` defines the planet for which the research
// power should be computed.
//
// Returns the research power available including the
// power brought by this planet along with any error.
func (a *TechnologyAction) fetchResearchPower(data model.Instance, planet *Planet) (int, error) {
	// First, fetch the level of the research lab on the
	// planet associated to this action: this will be the
	// base of the research.
	labID, err := data.Buildings.GetIDFromName("research lab")
	if err != nil {
		return 0, err
	}
	lab, err := planet.GetBuilding(labID)
	if err != nil {
		return 0, err
	}

	// Get the level of the intergalactic research network
	// technology reached by the player owning this planet.
	// It will indicate how many elements we should keep
	// in the list of reserch labs.
	igrn, err := data.Technologies.GetIDFromName("intergalactic research network")
	if err != nil {
		return lab.Level, err
	}

	labCount := planet.technologies[igrn]

	// Perform the query to get the levels of the labs on
	// each planet owned by this player.
	query := db.QueryDesc{
		Props: []string{
			"pb.planet",
			"pb.level",
		},
		Table: "planets_buildings pb inner join planets p on pb.planet=p.id",
		Filters: []db.Filter{
			{
				Key:    "p.player",
				Values: []string{planet.Player},
			},
		},
		// Note that we add `1` to the number of research labs in order
		// to account for the lab doing the research. Level `1` actually
		// tells that 1 lab can researching the same techno at the same
		// time than the one launching the research.
		Ordering: fmt.Sprintf("order by level desc limit %d", labCount+1),
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return lab.Level, err
	}
	if dbRes.Err != nil {
		return lab.Level, dbRes.Err
	}

	var pID string
	var labLevel int
	power := 0
	processedLabs := 0
	planetBelongsToTopLabs := false

	for dbRes.Next() && ((!planetBelongsToTopLabs && processedLabs < labCount) || planetBelongsToTopLabs) {
		err = dbRes.Scan(
			&pID,
			&labLevel,
		)

		if err != nil {
			return lab.Level, err
		}

		if pID == planet.ID {
			planetBelongsToTopLabs = true
		} else {
			power += labLevel
		}

		processedLabs++
	}

	return lab.Level + power, nil
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
func NewShipActionFromDB(ID string, data model.Instance) (ShipAction, error) {
	// Create the action.
	a := ShipAction{}
	a.ID = ID

	// Create the action using the base handler.
	var err error
	a.FixedAction, err = newFixedActionFromDB(ID, data, "construction_actions_ships")
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
		return ErrInvalidPlanet
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
func NewDefenseActionFromDB(ID string, data model.Instance) (DefenseAction, error) {
	// Create the action.
	a := DefenseAction{}
	a.ID = ID

	// Create the action using the base handler.
	var err error
	a.FixedAction, err = newFixedActionFromDB(ID, data, "construction_actions_defenses")
	if err != nil {
		return a, err
	}

	// Update the cost for this action. We will fetch
	// the defense system related to the action and
	// compute how many resources are needed to build
	// all the defenses required by the action.
	sd, err := data.Defenses.GetDefenseFromID(a.Element)
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
// action to complete. It uses internally the base handler
// which allow to handle the actual completion of the time.
// This wrapper is there to fetch the cost associate to
// the ship to build.
//
// The `data` allows to get information from the DB.
//
// The `p` defines the planet attached to this action and
// should be provided as argument to make handling of the
// concurrency easier.
//
// Returns any error.
func (a *DefenseAction) consolidateCompletionTime(data model.Instance, p *Planet) error {
	// First, we need to determine the cost for each of
	// the individual unit to produce.
	dd, err := data.Defenses.GetDefenseFromID(a.Element)
	if err != nil {
		return err
	}
	// Use the base handler.
	return a.computeCompletionTime(data, dd.Cost, p)
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
func (a *DefenseAction) Validate(data model.Instance, p *Planet) error {
	// Consistency.
	if a.Planet != p.ID {
		return ErrInvalidPlanet
	}

	// Update completion time and costs.
	err := a.consolidateCompletionTime(data, p)
	if err != nil {
		return err
	}

	// Compute the total cost of this action and validate
	// against planet's data.
	dd, err := data.Defenses.GetDefenseFromID(a.Element)
	if err != nil {
		return err
	}

	costs := dd.Cost.ComputeCost(a.Remaining)

	return p.validateAction(costs, dd.UpgradableDesc, data)
}
