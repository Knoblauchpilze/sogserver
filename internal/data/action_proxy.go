package data

import (
	"fmt"
	"oglike_server/internal/model"
	"oglike_server/pkg/logger"

	"github.com/google/uuid"
)

// ActionProxy :
// Provides a way to handle the upgrade actions for the
// server. Upgrade actions can be attached to buildings
// technologies, ships or defenses. They are also linked
// to a planet where the action is actually performed.
// This proxy intends at providing the necessary checks
// to make sure that a request to register a new action
// is possible given the resources and infrastructure
// currently existing on the planet.
type ActionProxy struct {
	commonProxy
}

// NewActionProxy :
// Create a new proxy allowing to serve the requests
// related to actions.
//
// The `data` defines the data model to use to fetch
// information and verify requests.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewActionProxy(data model.Instance, log logger.Logger) ActionProxy {
	return ActionProxy{
		commonProxy: newCommonProxy(data, log, "actions"),
	}
}

// createAction :
// Used to mutualize the creation of upgrade action by
// considering that what differs between an action that
// aims at upgrading a building, a technology, a ship
// or a defense is mostly the action itself and the
// script that will be used to perform the insertion.
//
// The `action` defines the upgrade action itself that
// should be inserted in the DB. It will be marshalled
// and passed on to the insertion script.
//
// The `script` represents the name of the script to
// use to perform the insertion of the input upgrade
// action into the DB.
//
// Returns any error that occurred during the insertion
// of the upgrade action in the DB.
func (p *ActionProxy) createAction(action UpgradeAction, script string) error {
	// Make sure that the input data describe a valid action.
	err := p.verifyAction(action)
	if err != nil {
		return fmt.Errorf("Could not create upgrade action (err: %v)", err)
	}

	// Create the query and execute it.
	query := insertReq{
		script: script,
		args:   []interface{}{action},
	}

	err = p.insertToDB(query)

	// Check for errors.
	if err != nil {
		return fmt.Errorf("Could not import upgrade action %s (err: %s)", action, err)
	}

	return nil
}

// verifyAction :
// Used internally to make sure that the action can
// be executed given the resources existing on the
// planet where it's needed and given the dependencies
// that might have to be met.
//
// The `a` defines the element that should be verified
// for consistency.
//
// Returns an error if the action cannot be performed
// for some reason and `nil` if the action is possible.
func (p *ActionProxy) verifyAction(a UpgradeAction) error {
	// We need to make sure that both the resources on the
	// planet where the action should be performed are at
	// a sufficient level to allow the action and also that
	// the dependencies (both buildings and technologies)
	// to allow the construction of the element described
	// by the action are met.

	// Fetch the planet related to the upgrade action.
	planet, err := p.fetchPlanet(a.GetPlanet())
	if err != nil {
		return fmt.Errorf("Cannot retrieve planet \"%s\" to verify action (err: %v)", a.GetPlanet(), err)
	}

	// Fetch the player related to this planet.
	player, err := p.fetchPlayer(planet.PlayerID)
	if err != nil {
		return fmt.Errorf("Cannot retrieve player \"%s\" to verify action (err: %v)", planet.PlayerID, err)
	}

	// Convert the resources into usable data.
	availableResources := make(map[string]float32)

	for _, res := range planet.Resources {
		existing, ok := availableResources[res.ID]

		if ok {
			return fmt.Errorf("Overriding resource \"%s\" amount in planet \"%s\" from %f to %f", res.ID, planet.ID, existing, res.Amount)
		}

		availableResources[res.ID] = res.Amount
	}

	// Create the building module of the planet.
	bm := buildingModule{
		buildings: p.buildings,
		planet:    planet,
	}

	// Populate the validation tool.
	vt := validationTools{
		pCosts:       p.pCosts,
		fCosts:       p.fCosts,
		techTree:     p.techTree,
		available:    availableResources,
		buildings:    planet.Buildings,
		technologies: player.Technologies,
		fields:       planet.remainingFields(),
	}

	// Perform the validation.
	valid, err := a.Validate(vt)
	if err != nil {
		return fmt.Errorf("Could not validate action on \"%s\" (err: %v)", planet.ID, err)
	}

	if !valid {
		return fmt.Errorf("Action cannot be performed on planet \"%s\"", a.GetPlanet())
	}

	// The action is valid, compute the completion time
	// from the data existing on the planet.
	err = a.UpdateCompletionTime(bm)
	if err != nil {
		return fmt.Errorf("Could not update completion time for action on \"%s\" (err: %v)", a.GetPlanet(), err)
	}

	return nil
}

// CreateBuildingAction :
// Used to perform the creation of the building upgrade
// action described by the input data to the DB. In case
// the creation can not be performed an error is sent
// back to the client.
//
// The `action` describes the element to create in DB.
// It corresponds to the desired upgrade action requested
// by the user.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`. It also provides the identifier of the action
// that was created by this method.
func (p *ActionProxy) CreateBuildingAction(action BuildingAction) (string, error) {
	// Assign a valid identifier if this is not already the case.
	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	// Check whether the action is valid.
	err := p.verifyAction(&action)
	if err != nil {
		return fmt.Errorf("Could not create upgrade action (err: %v)", err)
	}

	// We need to create the data related to the production and storage
	// upgrade that will be brought by this building if it finishes. It
	// will be registered in the DB alongside the upgrade action and we
	// need it for the import script.
	prodEffects, err := p.fetchBuildingProductionEffects(&action)
	if err != nil {
		return fmt.Errorf("Could not import upgrade action for %s (err: %v)", &action, err)
	}

	storageEffects, err := p.fetchBuildingStorageEffects(&action)
	if err != nil {
		return fmt.Errorf("Could not import upgrade action for %s (err: %v)", &action, err)
	}

	// Marshal the input building action to pass it to the import script.
	query := insertReq{
		script: "create_building_upgrade_action",
		args: []interface{}{
			action,
			prodEffects,
			storageEffects,
		},
	}

	err = p.insertToDB(query)

	// Check for errors.
	if err != nil {
		return fmt.Errorf("Could not import upgrade action for \"%s\" (err: %s)", action.PlanetID, err)
	}

	p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to upgrade \"%s\" to level %d on \"%s\"", action.ElementID, action.DesiredLevel, action.PlanetID))

	// All is well.
	return nil
}

// fetchBuildingProductionEffects :
// Used to fetch the production effetcs that increasing the
// input building as described in the action will have. It
// is returned as a marshallable array that can be used to
// provide to the upgrade action script.
//
// The `building` describes the upgrade action for which the
// production effects should be created.
//
// Returns the production effects along with any error.
func (p *ActionProxy) fetchBuildingProductionEffects(action *BuildingAction) ([]ProductionEffect, error) {
	// Make sure that the action is valid.
	if action == nil || !action.valid() {
		return []ProductionEffect{}, fmt.Errorf("Cannot fetch building upgrade action production effects for invalid action")
	}

	// Search for production rules of the building described
	// by the action: if some exists, we need to create the
	// corresponding effect.
	prodEffects := make([]ProductionEffect, 0)

	rules, ok := p.pRules[action.ElementID]

	if !ok {
		// No production rule, nothing to add.
		return prodEffects, nil
	}

	// As we will need the production, we need to fetch the
	// planet onto which the building will be built.
	planet, err := p.fetchPlanet(action.PlanetID)
	if err != nil {
		return []ProductionEffect{}, fmt.Errorf("Cannot fetch planet \"%s\" for building upgrade action production effects (err: %v)", action.PlanetID, err)
	}

	for _, rule := range rules {
		// The `Effect` should reference the difference from the
		// current situation. So we need to compute the difference
		// between the current production and the next level.
		curProd := rule.ComputeProduction(action.CurrentLevel, planet.averageTemp())
		desiredProd := rule.ComputeProduction(action.DesiredLevel, planet.averageTemp())

		effect := ProductionEffect{
			Action:   action.ID,
			Resource: rule.Resource,
			Effect:   desiredProd.Amount - curProd.Amount,
		}

		prodEffects = append(prodEffects, effect)
	}

	return prodEffects, nil
}

// fetchBuildingStorageEffects :
// Very similar to the `fetchBuildingProductionEffects` but
// is related to the storage effects of a building upgrade
// action.
// The returned value can be used alongside the action in
// the import script.
//
// The `building` describes the upgrade action for which the
// storage effects should be created.
//
// Returns the storage effects along with any error.
func (p *ActionProxy) fetchBuildingStorageEffects(action *BuildingAction) ([]StorageEffect, error) {
	// Make sure that the action is valid.
	if action == nil || !action.valid() {
		return []StorageEffect{}, fmt.Errorf("Cannot fetch building upgrade action storage effects for invalid action")
	}

	// Search for srotage effects for the building described
	// by the action: if some exists, we need to create the
	// corresponding effect.
	storageEffects := make([]StorageEffect, 0)

	rules, ok := p.sRules[action.ElementID]

	if !ok {
		// No storage rule, nothing to add.
		return storageEffects, nil
	}

	for _, rule := range rules {
		// The `Effect` should reference the difference from the
		// current situation. So we need to compute the storage
		// increase (or decrease) from the current level.
		curStorage := rule.ComputeStorage(action.CurrentLevel)
		desiredStorage := rule.ComputeStorage(action.DesiredLevel)

		effect := StorageEffect{
			Action:   action.ID,
			Resource: rule.Resource,
			Effect:   desiredStorage.Amount - curStorage.Amount,
		}

		storageEffects = append(storageEffects, effect)
	}

	return storageEffects, nil
}

// CreateTechnologyAction :
// Used to perform the creation of the technology upgrade
// action described by the input data to the DB. In case
// the creation can not be performed an error is returned
// to the client.
//
// The `action` describes the element to create in DB.
// It corresponds to the desired upgrade action requested
// by the user.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`. It also indicates the identifier of the action
// that was created.
func (p *ActionProxy) CreateTechnologyAction(action TechnologyAction) (string, error) {
	// Assign a valid identifier if this is not already the case.
	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	// Perform the creation of the action through the
	// dedicated handler.
	err := p.createAction(&action, "create_technology_upgrade_action")

	if err == nil {
		p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to upgrade \"%s\" to level %d for \"%s\"", action.ElementID, action.DesiredLevel, action.PlanetID))
	}

	return err
}

// CreateShipAction :
// Used to perform the creation of the ship upgrade action
// described by the input data to the DB. In case the
// creation can not be performed an error is returned to
// the client.
//
// The `action` describes the element to create in DB.
// It corresponds to the desired upgrade action requested
// by the user.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`. It also indicates the identifier of the action
// that was created.
func (p *ActionProxy) CreateShipAction(action FixedAction) (string, error) {
	// Assign a valid identifier if this is not already the case.
	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	// Make sure that the `remaining` count is identical to
	// the initial amount.
	action.Remaining = action.Amount

	// Perform the creation of the action through the
	// dedicated handler.
	err := p.createAction(&action, "create_ship_upgrade_action")

	if err == nil {
		p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to build \"%s\" on \"%s\"", action.ElementID, action.PlanetID))
	}

	return err
}

// CreateDefenseAction :
// Used to perform the creation of the defense upgrade
// action described by the input data to the DB. In case
// the creation can not be performed an error is returned
// to the client.
//
// The `action` describes the element to create in DB.
// It corresponds to the desired upgrade action requested
// by the user.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`. It also indicates the identifier of the action
// that was created.
func (p *ActionProxy) CreateDefenseAction(action FixedAction) (string, error) {
	// Assign a valid identifier if this is not already the case.
	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	// Make sure that the `remaining` count is identical to
	// the initial amount.
	action.Remaining = action.Amount

	// Perform the creation of the action through the
	// dedicated handler.
	err := p.createAction(&action, "create_defense_upgrade_action")

	if err == nil {
		p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to build \"%s\" on \"%s\"", action.ElementID, action.PlanetID))
	}

	return err
}
