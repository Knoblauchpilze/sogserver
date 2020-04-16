package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"

	"github.com/google/uuid"
)

// ActionProxy :
// Intended as a wrapper to access properties of upgrade
// actions and register some new ones on various elements
// of the game. Internally uses the common proxy defined
// in this package. Additional information is needed to
// perform verification when upgrade actions are submitted
// for creation. Indeed we want to make sure that it is
// actually possible to request the action on the planet
// or for the player it is intended to.
//
// The `pProxy` allows to fetch information about planets
// when an action concerning a planet is received.
//
// The `pRules` allows to create the needed production
// effects when an upgrade action for a building is set.
//
// The `sRules` allows to create the storage effects when
// a storage buildings (typically a hangar) is requested.
//
// The `pCosts` defines the construction costs for an
// upgradable element (e.g.g a building or a technology)
// which is used to verify that an upgrade action for an
// element is possible given the resources on a planet.
// TODO: We should maybe extend the lock period to the
// whole verification process and not just the update of
// the upgrade actions.
//
// The `fCosts` should be defines the construction costs
// for a unit like element (typically a ship or a def) so
// that it can be checked against the planet's resources
// when the creation is requested.
type ActionProxy struct {
	pProxy PlanetProxy
	pRules map[string][]ProductionRule
	sRules map[string][]StorageRule
	pCosts map[string]ConstructionCost
	fCosts map[string]FixedCost

	commonProxy
}

// NewActionProxy :
// Create a new proxy allowing to serve the requests
// related to upgrade actions.
//
// The `dbase` represents the database to use to fetch
// data related to upgrade actions.
//
// The `log` allows to notify errors and information.
//
// The `planets` provides a way to access to planets
// from the main DB.
//
// Returns the created proxy.
func NewActionProxy(dbase *db.DB, log logger.Logger, planets PlanetProxy) ActionProxy {
	proxy := ActionProxy{
		planets,
		make(map[string][]ProductionRule),
		make(map[string][]StorageRule),
		make(map[string]ConstructionCost),
		make(map[string]FixedCost),

		newCommonProxy(dbase, log),
	}

	err := proxy.init()
	if err != nil {
		log.Trace(logger.Error, fmt.Sprintf("Could not fetch technologies costs from DB (err: %v)", err))
	}

	return proxy
}

// init :
// Used to perform the initialziation of the needed
// DB variables for this proxy. This typically means
// fetching technologies costs from the DB.
//
// Returns `nil` if the technologies could be fetched
// from the DB successfully.
func (p ActionProxy) init() error {
	// Fetch from DB and aggregate resources as needed.
	var err error

	if len(p.pRules) == 0 {
		p.pRules, err = initBuildingsProductionRulesFromDB(p.dbase, p.log)
		if err != nil {
			return fmt.Errorf("Could not fetch buildings production rules from DB (err: %v)", err)
		}
	}

	if len(p.sRules) == 0 {
		p.sRules, err = initBuildingsStorageRulesFromDB(p.dbase, p.log)
		if err != nil {
			return fmt.Errorf("Could not fetch buildings storage rules from DB (err: %v)", err)
		}
	}

	// Fetch the building, technologies, ships and defenses
	// costs from the DB.
	if len(p.pCosts) == 0 {
		bCosts, err := initProgressCostsFromDB(
			p.dbase,
			p.log,
			"building",
			"buildings_costs_progress",
			"buildings_costs",
		)
		if err != nil {
			return fmt.Errorf("Could not fetch buildings construction costs from DB (err: %v)", err)
		}

		tCosts, err := initProgressCostsFromDB(
			p.dbase,
			p.log,
			"technology",
			"technologies_costs_progress",
			"technologies_costs",
		)
		if err != nil {
			return fmt.Errorf("Could not fetch technologies construction costs from DB (err: %v)", err)
		}

		for k, v := range tCosts {
			if _, ok := bCosts[k]; ok {
				p.log.Trace(logger.Error, fmt.Sprintf("Overriding progress costs for element \"%s\"", k))
			}

			bCosts[k] = v
		}

		p.pCosts = bCosts
	}

	if len(p.fCosts) == 0 {
		sCosts, err := initFixedCostsFromDB(
			p.dbase,
			p.log,
			"ship",
			"ships_costs",
		)
		if err != nil {
			return fmt.Errorf("Could not fetch ships construction costs from DB (err: %v)", err)
		}

		dCosts, err := initFixedCostsFromDB(
			p.dbase,
			p.log,
			"defense",
			"defenses_costs",
		)
		if err != nil {
			return fmt.Errorf("Could not fetch defenses construction costs from DB (err: %v)", err)
		}

		for k, v := range dCosts {
			if _, ok := sCosts[k]; ok {
				p.log.Trace(logger.Error, fmt.Sprintf("Overriding fixed costs for element \"%s\"", k))
			}

			sCosts[k] = v
		}

		p.fCosts = sCosts
	}

	return nil
}

// Buildings :
// Allows to fetch the list of upgrade action currently
// registered in the DB given the filters parameters. It
// can be used to get an idea of the actions pending for
// a planet regarding the buildings.
// The user can choose to filter parts of the buildings
// actions using an array of filters that will be applied
// to the SQL query.
// No controls is enforced on the filters so one should
// make sure that it's consistent with the underlying
// table.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// upgrade actions available. Each one is appended `as-is`
// to the query.
//
// Returns the list of building upgrade actions along with
// any errors. Note that in case the error is not `nil` the
// returned list is to be ignored.
func (p *ActionProxy) Buildings(filters []DBFilter) ([]BuildingUpgradeAction, error) {
	// Create the query and execute it.
	query := queryDesc{
		props: []string{
			"id",
			"planet",
			"building",
			"current_level",
			"desired_level",
			"completion_time",
		},
		table:   "construction_actions_buildings",
		filters: filters,
	}

	// Create the query and execute it.
	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch buildings upgrade actions (err: %v)", err)
	}

	// Populate the return value.
	actions := make([]BuildingUpgradeAction, 0)
	var act BuildingUpgradeAction

	for res.next() {
		err = res.scan(
			&act.ID,
			&act.PlanetID,
			&act.BuildingID,
			&act.CurrentLevel,
			&act.DesiredLevel,
			&act.CompletionTime,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for building upgrade action (err: %v)", err))
			continue
		}

		actions = append(actions, act)
	}

	return actions, nil
}

// Technologies :
// Similar to the `Buildings` feature but can be used to get
// the list of technology upgrade actions. Instead of fetching
// by planet (which would not make much sense) the result is
// fetched at a player level.
// The input filters describe some additional criteria that
// should be matched by the upgrade actions.
//
// The `filters` define some filtering properties to apply
// to the selected upgrade actions. Each one is used directly
// agains the columns of the related table.
//
// Returns the list of technology upgrade actions matching
// the input filters. This list should be ignored if the
// error is not `nil`.
func (p *ActionProxy) Technologies(filters []DBFilter) ([]TechnologyUpgradeAction, error) {
	// Create the query and execute it.
	query := queryDesc{
		props: []string{
			"id",
			"player",
			"technology",
			"planet",
			"current_level",
			"desired_level",
			"completion_time",
		},
		table:   "construction_actions_technologies",
		filters: filters,
	}

	// Create the query and execute it.
	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch technologies upgrade actions (err: %v)", err)
	}

	// Populate the return value.
	actions := make([]TechnologyUpgradeAction, 0)
	var act TechnologyUpgradeAction

	for res.next() {
		err = res.scan(
			&act.ID,
			&act.PlayerID,
			&act.TechnologyID,
			&act.PlanetID,
			&act.CurrentLevel,
			&act.DesiredLevel,
			&act.CompletionTime,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for technology upgrade action (err: %v)", err))
			continue
		}

		actions = append(actions, act)
	}

	return actions, nil
}

// Ships :
// Similar to the `Buildings` feature but can be used to get
// the list of ships construction actions. The input filters
// describe some additional criteria that should be matched
// by the construction actions (typically actions related to
// a specific ship, etc.).
//
// The `filters` define some filtering properties to apply
// to the selected upgrade actions. Each one is used directly
// agains the columns of the related table.
//
// Returns the list of ships being built that match the input
// filters. This list should be ignored if the error is not
// `nil`.
func (p *ActionProxy) Ships(filters []DBFilter) ([]ShipUpgradeAction, error) {
	// Create the query and execute it.
	query := queryDesc{
		props: []string{
			"id",
			"planet",
			"ship",
			"amount",
			"remaining",
			"completion_time",
		},
		table:   "construction_actions_ships",
		filters: filters,
	}

	// Create the query and execute it.
	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch ships construction actions (err: %v)", err)
	}

	// Populate the return value.
	actions := make([]ShipUpgradeAction, 0)
	var act ShipUpgradeAction

	for res.next() {
		err = res.scan(
			&act.ID,
			&act.PlanetID,
			&act.ShipID,
			&act.Amount,
			&act.Remaining,
			&act.CompletionTime,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for ship construction action (err: %v)", err))
			continue
		}

		actions = append(actions, act)
	}

	return actions, nil
}

// Defenses :
// Similar to the `Buildings` feature but can be used to get
// the list of defenses construction actions. The provided
// filters describe some additional criteria that should be
// matched by the construction actions (typically actions
// related to a specific defense system, etc.).
//
// The `filters` define some filtering properties to apply
// to the selected upgrade actions. Each one is used directly
// agains the columns of the related table.
//
// Returns the list of defenses being built that match the
// input filters. This list should be ignored if the error
// is not `nil`.
func (p *ActionProxy) Defenses(filters []DBFilter) ([]DefenseUpgradeAction, error) {
	// Create the query and execute it.
	query := queryDesc{
		props: []string{
			"id",
			"planet",
			"defense",
			"amount",
			"remaining",
			"completion_time",
		},
		table:   "construction_actions_defenses",
		filters: filters,
	}

	// Create the query and execute it.
	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch defenses construction actions (err: %v)", err)
	}

	// Populate the return value.
	actions := make([]DefenseUpgradeAction, 0)
	var act DefenseUpgradeAction

	for res.next() {
		err = res.scan(
			&act.ID,
			&act.PlanetID,
			&act.DefenseID,
			&act.Amount,
			&act.Remaining,
			&act.CompletionTime,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for defense construction action (err: %v)", err))
			continue
		}

		actions = append(actions, act)
	}

	return actions, nil
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
// The `action` defines the element that should be
// verified for consistency.
//
// Returns an error if the action cannot be performed
// for some reason and `nil` if the action is possible.
func (p *ActionProxy) verifyAction(action UpgradeAction) error {
	// We need to make sure that both the resources on the
	// planet where the action should be performed are at
	// a sufficient level to allow the action and also that
	// the dependencies (both buildings and technologies)
	// to allow the construction of the element described
	// by the action are met.
	// TODO: Implement the checking of all the above.
	return fmt.Errorf("Not implemented")
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
// `nil`.
func (p *ActionProxy) CreateBuildingAction(action *BuildingUpgradeAction) error {
	// Assign a valid identifier if this is not already the case.
	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	// Check whether the action is valid.
	err := p.verifyAction(action)
	if err != nil {
		return fmt.Errorf("Could not create upgrade action (err: %v)", err)
	}

	// We need to create the data related to the production and storage
	// upgrade that will be brought by this building if it finishes. It
	// will be registered in the DB alongside the upgrade action and we
	// need it for the import script.
	prodEffects, err := p.fetchBuildingProductionEffects(action)
	if err != nil {
		return fmt.Errorf("Could not import upgrade action for %s (err: %v)", *action, err)
	}

	storageEffects, err := p.fetchBuildingStorageEffects(action)
	if err != nil {
		return fmt.Errorf("Could not import upgrade action for %s (err: %v)", *action, err)
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

	p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to upgrade \"%s\" to level %d on \"%s\"", action.BuildingID, action.DesiredLevel, action.PlanetID))

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
func (p *ActionProxy) fetchBuildingProductionEffects(action *BuildingUpgradeAction) ([]ProductionEffect, error) {
	// Make sure that the action is valid.
	if action == nil || !action.valid() {
		return []ProductionEffect{}, fmt.Errorf("Cannot fetch building upgrade action production effects for invalid action")
	}

	// Search for production rules of the building described
	// by the action: if some exists, we need to create the
	// corresponding effect.
	prodEffects := make([]ProductionEffect, 0)

	rules, ok := p.pRules[action.BuildingID]

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
func (p *ActionProxy) fetchBuildingStorageEffects(action *BuildingUpgradeAction) ([]StorageEffect, error) {
	// Make sure that the action is valid.
	if action == nil || !action.valid() {
		return []StorageEffect{}, fmt.Errorf("Cannot fetch building upgrade action storage effects for invalid action")
	}

	// Search for srotage effects for the building described
	// by the action: if some exists, we need to create the
	// corresponding effect.
	storageEffects := make([]StorageEffect, 0)

	rules, ok := p.sRules[action.BuildingID]

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

// fetchPlanet :
// Used to fetch the planet described by the input identifier
// in the internal DB. It is used to compute the production
// effects in the case of a building upgrade action.
//
// The `id` defines the identifier of the planet to fetch.
//
// Returns the planet corresponding to the identifier along
// with any error.
func (p *ActionProxy) fetchPlanet(id string) (Planet, error) {
	// Create the db filters from the input identifier.
	filters := make([]DBFilter, 1)

	filters[0] = DBFilter{
		"p.id",
		[]string{id},
	}

	planets, err := p.pProxy.Planets(filters)

	// Check for errors and cases where we retrieve several
	// players.
	if err != nil {
		return Planet{}, err
	}
	if len(planets) != 1 {
		return Planet{}, fmt.Errorf("Retrieved %d planets for id \"%s\" (expected 1)", len(planets), id)
	}

	return planets[0], nil
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
// `nil`.
func (p *ActionProxy) CreateTechnologyAction(action *TechnologyUpgradeAction) error {
	// Assign a valid identifier if this is not already the case.
	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	// Perform the creation of the action through the
	// dedicated handler.
	err := p.createAction(action, "create_technology_upgrade_action")

	if err == nil {
		p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to upgrade \"%s\" to level %d for \"%s\"", action.TechnologyID, action.DesiredLevel, action.PlayerID))
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
// `nil`.
func (p *ActionProxy) CreateShipAction(action *ShipUpgradeAction) error {
	// Assign a valid identifier if this is not already the case.
	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	// Make sure that the `remaining` count is identical to
	// the initial amount.
	action.Remaining = action.Amount

	// Perform the creation of the action through the
	// dedicated handler.
	err := p.createAction(action, "create_ship_upgrade_action")

	if err == nil {
		p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to build \"%s\" on \"%s\"", action.ShipID, action.PlanetID))
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
// `nil`.
func (p *ActionProxy) CreateDefenseAction(action *DefenseUpgradeAction) error {
	// Assign a valid identifier if this is not already the case.
	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	// Make sure that the `remaining` count is identical to
	// the initial amount.
	action.Remaining = action.Amount

	// Perform the creation of the action through the
	// dedicated handler.
	err := p.createAction(action, "create_defense_upgrade_action")

	if err == nil {
		p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to build \"%s\" on \"%s\"", action.DefenseID, action.PlanetID))
	}

	return err
}
