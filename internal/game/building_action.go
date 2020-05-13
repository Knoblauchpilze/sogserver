package game

import (
	"fmt"
	"math"
	"oglike_server/pkg/db"
	"time"
)

// BuildingAction :
// Used as a way to refine the `ProgressAction` for
// the specific case of buildings. It mostly add the
// info to compute the completion time for a building.
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

// ErrNoFieldsLeft : Indicates that there are not fields left to perform the action.
var ErrNoFieldsLeft = fmt.Errorf("No remaining fields left for action")

// valid :
// Determines whether this action is valid. By valid we
// only mean obvious syntax errors.
//
// Returns any error or `nil` if the action seems valid.
func (a *BuildingAction) valid() error {
	if err := a.ProgressAction.valid(); err != nil {
		return err
	}

	if math.Abs(float64(a.DesiredLevel)-float64(a.CurrentLevel)) != 1 {
		return ErrInvalidLevelForAction
	}

	return nil
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
	a := BuildingAction{}

	// Create the action using the base handler.
	var err error
	a.ProgressAction, err = newProgressActionFromDB(ID, data, "construction_actions_buildings")

	// Consistency.
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
func (a *BuildingAction) fetchProductionEffects(data Instance) error {
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
				Values: []interface{}{a.ID},
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
func (a *BuildingAction) fetchStorageEffects(data Instance) error {
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
				Values: []interface{}{a.ID},
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

// SaveToDB :
// Used to save the content of this action to
// the DB. In case an error is raised during
// the operation a comprehensive error is
// returned.
//
// The `proxy` allows to access to the DB.
//
// Returns any error.
func (a *BuildingAction) SaveToDB(proxy db.Proxy) error {
	// Check consistency.
	if err := a.valid(); err != nil {
		return err
	}

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_building_upgrade_action",
		Args: []interface{}{
			a,
			a.Costs,
			a.Production,
			a.Storage,
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

	dee, ok := dbe.Err.(db.DuplicatedElementError)
	if ok {
		switch dee.Constraint {
		case "construction_actions_buildings_planet_key":
			return ErrOnlyOneActionAuthorized
		}

		return dee
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
func (a *BuildingAction) Convert() interface{} {
	return struct {
		ID             string    `json:"id"`
		Planet         string    `json:"planet"`
		Element        string    `json:"element"`
		CurrentLevel   int       `json:"current_level"`
		DesiredLevel   int       `json:"desired_level"`
		CompletionTime time.Time `json:"completion_time"`
		CreatedAt      time.Time `json:"created_at"`
	}{
		ID:             a.ID,
		Planet:         a.Planet,
		Element:        a.Element,
		CurrentLevel:   a.CurrentLevel,
		DesiredLevel:   a.DesiredLevel,
		CompletionTime: a.CompletionTime,
		CreatedAt:      a.creationTime,
	}
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
func (a *BuildingAction) ConsolidateEffects(data Instance, p *Planet) error {
	// Consistency.
	if a.Planet != p.ID {
		return ErrMismatchInVerification
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
func (a *BuildingAction) consolidateCompletionTime(data Instance, p *Planet) error {
	// Consistency.
	if a.Planet != p.ID {
		return ErrMismatchInVerification
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

	robotics := p.Buildings[roboticsID]
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

	hours := float64(m+c) / (2500.0 * (1.0 + float64(robotics.Level)) * math.Pow(2.0, float64(nanite.Level)))

	t, err := time.ParseDuration(fmt.Sprintf("%fh", hours))
	if err != nil {
		return ErrInvalidDuration
	}

	a.creationTime = time.Now()
	a.CompletionTime = a.creationTime.Add(t)

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
func (a *BuildingAction) Validate(data Instance, p *Planet) error {
	// Consistency.
	if a.Planet != p.ID {
		return ErrMismatchInVerification
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

	bi, ok := p.Buildings[bd.ID]
	if !ok || bi.Level != a.CurrentLevel {
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
