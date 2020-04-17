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
// The `plProxy` allows to fetch information about players
// when data concerning the owner of a planet is needed.
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
//
// The `techTree` defines the dependencies for any item
// in the game using a map where the key is the id of
// the item and the values represent a combination of
// the level of the dependency and the identifier of
// the element.
type ActionProxy struct {
	pRules   map[string][]ProductionRule
	sRules   map[string][]StorageRule
	pCosts   map[string]ConstructionCost
	fCosts   map[string]FixedCost
	techTree map[string][]TechDependency

	planetsDependentProxy
	playersDependentProxy
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
// The `players` provides a way to access to players
// from the main DB.
//
// Returns the created proxy.
func NewActionProxy(dbase *db.DB, log logger.Logger, planets PlanetProxy, players PlayerProxy) ActionProxy {
	proxy := ActionProxy{
		make(map[string][]ProductionRule),
		make(map[string][]StorageRule),
		make(map[string]ConstructionCost),
		make(map[string]FixedCost),
		make(map[string][]TechDependency),

		newPlanetsDependentProxy(planets),
		newPlayersDependentProxy(players),
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
func (p *ActionProxy) init() error {
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

	// Fetch tech tree dependencies.
	if len(p.techTree) == 0 {
		techTree, err := p.initTechTree()
		if err != nil || techTree == nil {
			return fmt.Errorf("Could not fetch tech tree from DB (err: %v)", err)
		}

		p.techTree = techTree
	}

	return nil
}

// initTechTree :
// Used to fetch the tech tree defined in the DB for the list
// of elements of the game. All buildings, technologies, ships
// and defenses will be aggregated in a single map so that it
// is easier to manipulate.
//
// The `dbase` defines the source from where the data of the
// tech dependencies should be fetched.
//
// The `log` allows to notify errors and info to the user in
// case of any failure.
//
// Returns a map representing all the data of the tech tree
// of the game along with any error.
func (p *ActionProxy) initTechTree() (map[string][]TechDependency, error) {
	// We need to scan all the dependencies tables so namely:
	//   - tech_tree_buildings_dependencies
	//   - tech_tree_technologies_dependencies
	//   - tech_tree_buildings_vs_technologies
	//   - tech_tree_technologies_vs_buildings
	//   - tech_tree_ships_vs_buildings
	//   - tech_tree_ships_vs_technologies
	//   - tech_tree_defenses_vs_buildings
	//   - tech_tree_defenses_vs_technologies
	// Each of these tables as a similar structure like:
	//   - element uuid
	//   - requirement uuid
	//   - level integer
	// Where the `element` can be any of `building`, `technology`
	// `ship` or `defense`.
	// We will select everything from these tables and then use
	// the data to populate the `techTree` map.
	techTree := make(map[string][]TechDependency)
	sanity := make(map[string]map[string]int)

	// Create the query and execute it.
	query := queryDesc{
		props: []string{
			"",
			"requirement",
			"level",
		},
		table:   "",
		filters: []DBFilter{},
	}

	// Fetch buildings dependencies on buildings.
	query.props[0] = "building"
	query.table = "tech_tree_buildings_dependencies"

	err := p.populateTechTree(query, &techTree, &sanity)
	if err != nil {
		return nil, err
	}

	// Fetch buildings dependencies on technologies.
	query.props[0] = "building"
	query.table = "tech_tree_buildings_vs_technologies"

	err = p.populateTechTree(query, &techTree, &sanity)
	if err != nil {
		return nil, err
	}

	// Fetch technologies dependencies on buildings.
	query.props[0] = "technology"
	query.table = "tech_tree_technologies_vs_buildings"

	err = p.populateTechTree(query, &techTree, &sanity)
	if err != nil {
		return nil, err
	}

	// Fetch technologies dependencies on technologies.
	query.props[0] = "technology"
	query.table = "tech_tree_technologies_dependencies"

	err = p.populateTechTree(query, &techTree, &sanity)
	if err != nil {
		return nil, err
	}

	// Fetch ships dependencies on buildings.
	query.props[0] = "ship"
	query.table = "tech_tree_ships_vs_buildings"

	err = p.populateTechTree(query, &techTree, &sanity)
	if err != nil {
		return nil, err
	}

	// Fetch ships dependencies on technologies.
	query.props[0] = "ship"
	query.table = "tech_tree_ships_vs_technologies"

	err = p.populateTechTree(query, &techTree, &sanity)
	if err != nil {
		return nil, err
	}

	// Fetch defenses dependencies on buildings.
	query.props[0] = "defense"
	query.table = "tech_tree_defenses_vs_buildings"

	err = p.populateTechTree(query, &techTree, &sanity)
	if err != nil {
		return nil, err
	}

	// Fetch defenses dependencies on technologies.
	query.props[0] = "defense"
	query.table = "tech_tree_defenses_vs_technologies"

	err = p.populateTechTree(query, &techTree, &sanity)
	if err != nil {
		return nil, err
	}

	return techTree, nil
}

// populateTechTree :
// Used to analyze the query in input and populate the tech
// tree from its result.
//
// The `query` defines the query that should be performed
// and which will return information about the tech tree.
//
// The `techTree` defines the existing tech tree.
//
// The `sanity` is a map registering all the already found
// dependency along with their level. It prevents any case
// where a dependency would be overriden by another one.
// An error message is still displayed in this case.
//
// Returns any error (mainly in case the DB has not been
// queried properly).
func (p *ActionProxy) populateTechTree(query queryDesc, techTree *map[string][]TechDependency, sanity *map[string]map[string]int) error {
	// Execute the query.
	res, err := p.fetchDB(query)
	defer res.Close()

	if err != nil {
		return fmt.Errorf("Cannot fetch buildings tech tree (err: %v)", err)
	}

	// Analyze the result and build the tech tree.
	var elem, dep string
	var req int

	for res.next() {
		err = res.scan(
			&elem,
			&dep,
			&req,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve dependency info while building tech tree (err: %v)", err))
			continue
		}

		// Check that we don't override anything.
		override := false

		e, ok := (*sanity)[elem]
		if !ok {
			e = make(map[string]int)
			e[dep] = req
		} else {
			d, ok := e[dep]

			if ok {
				p.log.Trace(logger.Error, fmt.Sprintf("Prevented override of dependency on \"%s\" for \"%s\" (existing: %d, new: %d)", dep, elem, d, req))
				override = true
			}

			e[dep] = req
		}

		(*sanity)[elem] = e

		// Register the dependency for this element.
		if !override {
			deps, ok := (*techTree)[elem]
			if !ok {
				deps = make([]TechDependency, 0)
			}

			deps = append(deps, TechDependency{dep, req})

			(*techTree)[elem] = deps
		}
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
func (p *ActionProxy) Buildings(filters []DBFilter) ([]ProgressAction, error) {
	// Update the upgrade actions for the planet described by
	// the input filters: this will prune already completed
	// upgrade actions and also update the remaining ones. It
	// is needed to make sure that we only retrieve actions
	// still valid at the moment of the request.
	err := p.updateActionsFromFilters(filters)
	if err != nil {
		return []ProgressAction{}, fmt.Errorf("Could not update upgrade action before retrieving building ones (err: %v)", err)
	}

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

	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return []ProgressAction{}, fmt.Errorf("Could not query DB to fetch buildings upgrade actions (err: %v)", err)
	}

	// Populate the return value.
	actions := make([]ProgressAction, 0)
	var act ProgressAction

	for res.next() {
		err = res.scan(
			&act.ID,
			&act.PlanetID,
			&act.ElementID,
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
func (p *ActionProxy) Technologies(filters []DBFilter) ([]ProgressAction, error) {
	// Similarly to the `Buildings`, we want to update the upgrade
	// actions so that only valid ones are fetched by this method.
	err := p.updateActionsFromFilters(filters)
	if err != nil {
		return []ProgressAction{}, fmt.Errorf("Could not update upgrade action before retrieving technology ones (err: %v)", err)
	}

	// Create the query and execute it.
	query := queryDesc{
		props: []string{
			"id",
			"technology",
			"planet",
			"current_level",
			"desired_level",
			"completion_time",
		},
		table:   "construction_actions_technologies",
		filters: filters,
	}

	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return []ProgressAction{}, fmt.Errorf("Could not query DB to fetch technologies upgrade actions (err: %v)", err)
	}

	// Populate the return value.
	actions := make([]ProgressAction, 0)
	var act ProgressAction

	for res.next() {
		err = res.scan(
			&act.ID,
			&act.ElementID,
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
func (p *ActionProxy) Ships(filters []DBFilter) ([]FixedAction, error) {
	// Similarly to the `Buildings`, we want to update the upgrade
	// actions so that only valid ones are fetched by this method.
	err := p.updateActionsFromFilters(filters)
	if err != nil {
		return []FixedAction{}, fmt.Errorf("Could not update upgrade action before retrieving ship ones (err: %v)", err)
	}

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

	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return []FixedAction{}, fmt.Errorf("Could not query DB to fetch ships construction actions (err: %v)", err)
	}

	// Populate the return value.
	actions := make([]FixedAction, 0)
	var act FixedAction

	for res.next() {
		err = res.scan(
			&act.ID,
			&act.PlanetID,
			&act.ElementID,
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
func (p *ActionProxy) Defenses(filters []DBFilter) ([]FixedAction, error) {
	// Similarly to the `Buildings`, we want to update the upgrade
	// actions so that only valid ones are fetched by this method.
	err := p.updateActionsFromFilters(filters)
	if err != nil {
		return []FixedAction{}, fmt.Errorf("Could not update upgrade action before retrieving defense ones (err: %v)", err)
	}

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

	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return []FixedAction{}, fmt.Errorf("Could not query DB to fetch defenses construction actions (err: %v)", err)
	}

	// Populate the return value.
	actions := make([]FixedAction, 0)
	var act FixedAction

	for res.next() {
		err = res.scan(
			&act.ID,
			&act.PlanetID,
			&act.ElementID,
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

// updateActionsFromFilters :
// Used to analyze the filters provided as input to find a
// planet for which actions should be upgraded. In case a
// planet's identifier cannot be determined from the input
// filters an error is returned.
//
// The `filters` defines the list of elements which should
// be scanned in order to find a planet's identifier.
//
// Return any error.
func (p *ActionProxy) updateActionsFromFilters(filters []DBFilter) error {
	// Traverse the filters to find one named `planet`.
	planetID := ""
	found := false

	for id := 0; id < len(filters) && !found; id++ {
		if filters[id].Key == "planet" {
			// Detect invalid filters.
			if len(filters[id].Values) != 1 {
				return fmt.Errorf("Cannot determine unique planet identifier from %d filters", len(filters[id].Values))
			}

			found = true
			planetID = filters[id].Values[0]
		}
	}

	if !found || len(planetID) == 0 {
		return fmt.Errorf("Could not find planet identifier from %d input filter(s)", len(filters))
	}

	// Fetch the planet (which will update the construction
	// actions) so that we can update the technologies that
	// may have an upgrade action running for the player's
	// owning the planet. The update of the technologies is
	// also automatically done when fetching the player so
	// it's not needed to actually do it (we just have to
	// fetch the player).
	pla, err := p.fetchPlanet(planetID)
	if err != nil {
		return fmt.Errorf("Could not find planet from identifier \"%s\"", planetID)
	}

	_, err = p.fetchPlayer(pla.PlayerID)
	if err != nil {
		return fmt.Errorf("Could not find player \"%s\" from planet \"%s\"", pla.PlayerID, planetID)
	}

	return nil
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

	// Populate the validation tool.
	vt := validationTools{
		pCosts:       p.pCosts,
		fCosts:       p.fCosts,
		techTree:     p.techTree,
		available:    availableResources,
		buildings:    planet.Buildings,
		technologies: player.Technologies,
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
// `nil`.
func (p *ActionProxy) CreateBuildingAction(action ProgressAction) error {
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
func (p *ActionProxy) fetchBuildingProductionEffects(action *ProgressAction) ([]ProductionEffect, error) {
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
func (p *ActionProxy) fetchBuildingStorageEffects(action *ProgressAction) ([]StorageEffect, error) {
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
// `nil`.
func (p *ActionProxy) CreateTechnologyAction(action ProgressAction) error {
	// Assign a valid identifier if this is not already the case.
	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	// Perform the creation of the action through the
	// dedicated handler.
	err := p.createAction(action, "create_technology_upgrade_action")

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
// `nil`.
func (p *ActionProxy) CreateShipAction(action FixedAction) error {
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
// `nil`.
func (p *ActionProxy) CreateDefenseAction(action FixedAction) error {
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
		p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to build \"%s\" on \"%s\"", action.ElementID, action.PlanetID))
	}

	return err
}
