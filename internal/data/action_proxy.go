package data

import (
	"encoding/json"
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"

	"github.com/google/uuid"
)

// ActionProxy :
// Intended as a wrapper to access and register new props
// for actions that aim at upgrading buildings, technos
// or building new ships or defenses on planets.
// The goal of this proxy is to hide the real layout of
// the DB to the exterior world and provide a sanitized
// environment to perform these interactions.
//
// The `dbase` is the database that is wrapped by this
// object. It is checked for consistency upon building the
// wrapper.
//
// The `log` allows to perform display to the user so as
// to inform of potential issues and debug information to
// the outside world.
//
// The `prodRules` defines the production rules defines
// for various buildings. This is used when creating a
// building upgrade action to measure the consequences
// of upgrading a building.
//
// The `storageRules` defines something similar to the
// `prodRules` but for storage associated to each item.
//
// The `progressCosts` regroups the costs of all the unit
// in the game indifferently to whether it is a building
// or a technology. It is used to make sure that there
// are actually enough resources on the planet before
// performing the registration of any upgrade action.
// These costs only include the elements following a rule
// of progression which depends on the level of the elem.
//
// The `fixedCosts` are similar to the `progressCosts`
// but are used for elements that are unit-like in the
// sense that they have a fixed costs and they cannot
// be upgraded (hence no progression).
//
// The `planetProxy` defines the proxy allowing to query
// a planet from an identifier. It is used in conjunction
// to the production rules to allow the computation of
// the precise production on a particular planet.
type ActionProxy struct {
	dbase         *db.DB
	log           logger.Logger
	prodRules     map[string][]ProductionRule
	storageRules  map[string][]StorageRule
	progressCosts map[string]ConstructionCost
	fixedCosts    map[string]FixedCost
	planetProxy   PlanetProxy
}

// NewActionProxy :
// Create a new proxy on the input `dbase` to access the
// properties of upgrade actions registered in the DB. It
// includes all kind of actions (for buildings, ships,
// defenses or technologies).
// In case the provided DB is `nil` a panic is issued.
//
// The `dbase` represents the database to use to fetch
// data related to upgrade actions.
//
// The `log` will be used to notify information so that
// we can have an idea of the activity of this component.
// One possible example is for timing the requests.
//
// The `planets` defines a proxy that can be used to
// fetch information about the planets when creating
// the building upgrade actions.
//
// Returns the created proxy.
func NewActionProxy(dbase *db.DB, log logger.Logger, planets PlanetProxy) ActionProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create actions proxy from invalid DB"))
	}

	// Fetch the production rules for each building.
	prodRules, err := initBuildingsProductionRulesFromDB(dbase, log)
	if err != nil {
		log.Trace(logger.Error, fmt.Sprintf("Could not fetch buildings production rules from DB (err: %v)", err))
	}

	// Fetch the storage rules for each building.
	storageRules, err := initBuildingsStorageRulesFromDB(dbase, log)
	if err != nil {
		log.Trace(logger.Error, fmt.Sprintf("Could not fetch buildings storage rules from DB (err: %v)", err))
	}

	// Fetch the building, technologies, ships and defenses
	// costs from the DB.
	bCosts, err := initProgressCostsFromDB(dbase, log, "building", "buildings_costs_progress", "buildings_costs")
	if err != nil {
		log.Trace(logger.Error, fmt.Sprintf("Could not fetch buildings construction costs from DB (err: %v)", err))
	}

	tCosts, err := initProgressCostsFromDB(dbase, log, "technology", "technologies_costs_progress", "technologies_costs")
	if err != nil {
		log.Trace(logger.Error, fmt.Sprintf("Could not fetch technologies construction costs from DB (err: %v)", err))
	}

	sCosts, err := initFixedCostsFromDB(dbase, log, "ship", "ships_costs")
	if err != nil {
		log.Trace(logger.Error, fmt.Sprintf("Could not fetch ships construction costs from DB (err: %v)", err))
	}

	dCosts, err := initFixedCostsFromDB(dbase, log, "defense", "defenses_costs")
	if err != nil {
		log.Trace(logger.Error, fmt.Sprintf("Could not fetch defenses construction costs from DB (err: %v)", err))
	}

	// Group common costs in a single map.
	for k, v := range tCosts {
		if _, ok := bCosts[k]; ok {
			log.Trace(logger.Error, fmt.Sprintf("Overriding progress costs for element \"%s\"", k))
		}

		bCosts[k] = v
	}

	for k, v := range dCosts {
		if _, ok := sCosts[k]; ok {
			log.Trace(logger.Error, fmt.Sprintf("Overriding fixed costs for element \"%s\"", k))
		}

		sCosts[k] = v
	}

	return ActionProxy{
		dbase,
		log,
		prodRules,
		storageRules,
		bCosts,
		sCosts,
		planets,
	}
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
	props := []string{
		"id",
		"planet",
		"building",
		"current_level",
		"desired_level",
		"completion_time",
	}

	table := "construction_actions_buildings"

	query := fmt.Sprintf("select %s from %s", strings.Join(props, ", "), table)

	if len(filters) > 0 {
		query += " where"

		for id, filter := range filters {
			if id > 0 {
				query += " and"
			}
			query += fmt.Sprintf(" %s", filter)
		}
	}

	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch buildings upgrade actions (err: %v)", err)
	}

	// Populate the return value.
	actions := make([]BuildingUpgradeAction, 0)
	var act BuildingUpgradeAction

	for rows.Next() {
		err = rows.Scan(
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
	props := []string{
		"id",
		"player",
		"technology",
		"planet",
		"current_level",
		"desired_level",
		"completion_time",
	}

	table := "construction_actions_technologies"

	query := fmt.Sprintf("select %s from %s", strings.Join(props, ", "), table)

	if len(filters) > 0 {
		query += " where"

		for id, filter := range filters {
			if id > 0 {
				query += " and"
			}
			query += fmt.Sprintf(" %s", filter)
		}
	}

	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch technologies upgrade actions (err: %v)", err)
	}

	// Populate the return value.
	actions := make([]TechnologyUpgradeAction, 0)
	var act TechnologyUpgradeAction

	for rows.Next() {
		err = rows.Scan(
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
	props := []string{
		"id",
		"planet",
		"ship",
		"amount",
		"remaining",
		"completion_time",
	}

	table := "construction_actions_ships"

	query := fmt.Sprintf("select %s from %s", strings.Join(props, ", "), table)

	if len(filters) > 0 {
		query += " where"

		for id, filter := range filters {
			if id > 0 {
				query += " and"
			}
			query += fmt.Sprintf(" %s", filter)
		}
	}

	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch ships construction actions (err: %v)", err)
	}

	// Populate the return value.
	actions := make([]ShipUpgradeAction, 0)
	var act ShipUpgradeAction

	for rows.Next() {
		err = rows.Scan(
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
	props := []string{
		"id",
		"planet",
		"defense",
		"amount",
		"remaining",
		"completion_time",
	}

	table := "construction_actions_defenses"

	query := fmt.Sprintf("select %s from %s", strings.Join(props, ", "), table)

	if len(filters) > 0 {
		query += " where"

		for id, filter := range filters {
			if id > 0 {
				query += " and"
			}
			query += fmt.Sprintf(" %s", filter)
		}
	}

	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch defenses construction actions (err: %v)", err)
	}

	// Populate the return value.
	actions := make([]DefenseUpgradeAction, 0)
	var act DefenseUpgradeAction

	for rows.Next() {
		err = rows.Scan(
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
	// Check whether the action is valid.
	if !action.valid() {
		return fmt.Errorf("Could not create upgrade action, some properties are invalid")
	}

	// Marshal the input action to pass it to the import script.
	data, err := json.Marshal(action)
	if err != nil {
		return fmt.Errorf("Could not import upgrade action for %s (err: %v)", action, err)
	}
	jsonToSend := string(data)

	query := fmt.Sprintf("select * from %s('%s')", script, jsonToSend)
	_, err = p.dbase.DBExecute(query)

	// Check for errors during the insertion process.
	if err != nil {
		return fmt.Errorf("Could not import upgrade action %s (err: %s)", action, err)
	}

	// All is well.
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
	if !action.valid() {
		return fmt.Errorf("Could not create upgrade action, some properties are invalid")
	}

	err := p.verifyAction(action)
	if err != nil {
		return fmt.Errorf("Could not create building action (err: %v)", err)
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
	data, err := json.Marshal(action)
	if err != nil {
		return fmt.Errorf("Could not import upgrade action for %s (err: %v)", *action, err)
	}
	jsonForAction := string(data)

	data, err = json.Marshal(prodEffects)
	if err != nil {
		return fmt.Errorf("Could not import upgrade action for %s (err: %v)", *action, err)
	}
	jsonForProdEffects := string(data)

	data, err = json.Marshal(storageEffects)
	if err != nil {
		return fmt.Errorf("Could not import upgrade action for %s (err: %v)", *action, err)
	}
	jsonForStorageEffects := string(data)

	query := fmt.Sprintf("select * from %s('%s', '%s', '%s')", "create_building_upgrade_action", jsonForAction, jsonForProdEffects, jsonForStorageEffects)
	_, err = p.dbase.DBExecute(query)

	// Check for errors during the insertion process.
	if err != nil {
		return fmt.Errorf("Could not import upgrade action %s (err: %s)", *action, err)
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

	rules, ok := p.prodRules[action.BuildingID]

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

	rules, ok := p.storageRules[action.BuildingID]

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

	planets, err := p.planetProxy.Planets(filters)

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

	err := p.verifyAction(action)
	if err != nil {
		return fmt.Errorf("Could not create technology action (err: %v)", err)
	}

	err = p.createAction(action, "create_technology_upgrade_action")

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

	err := p.verifyAction(action)
	if err != nil {
		return fmt.Errorf("Could not create ship action (err: %v)", err)
	}

	err = p.createAction(action, "create_ship_upgrade_action")

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

	err := p.verifyAction(action)
	if err != nil {
		return fmt.Errorf("Could not create defense action (err: %v)", err)
	}

	err = p.createAction(action, "create_defense_upgrade_action")

	if err == nil {
		p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to build \"%s\" on \"%s\"", action.DefenseID, action.PlanetID))
	}

	return err
}
