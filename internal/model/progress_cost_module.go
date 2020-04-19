package model

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// progressCostsModule :
// Refine the interface provided by the upgradables module
// to handle cases where the upgradable defines some sort
// of progression rule that is applied to compute the cost
// of a specific level. It typically relates to the notion
// of upgrading an element to the next level.
// This typically applies for elements like buildings or
// technologies.
//
// The `costs` defines for each element handled by this
// module the corresponding cost. The costs refers to the
// identifiers of resources and associate a single integer
// value to each resource.
type progressCostsModule struct {
	upgradablesModule

	costs map[string]ProgressCost
}

// ProgressCost :
// Defines a set of costs for the resources to build
// a progress element. The costs include both all the
// resources' identifiers needed for the first level
// along with the progression rule which allows to
// compute the cost for the other levels.
// The progression formula to compute the costs for
// a new level is like the following:
// `cost(n) = cost(0) * progressionRule ^ n`.
// The larger the `progressionRule` the quicker the
// costs will rise with the level.
//
// The `InitCosts` represents a map where the keys are
// resources' identifiers and the values are the init
// cost of the element for the corresponding resource.
// If a resource does not have its identifier in this
// map, it means that the element does not require any
// quantity of it.
//
// The `ProgressionRule` defines a value that should
// be used to multiply the initial costfor a particular
// resource to obtain the cost at any level.
type ProgressCost struct {
	InitCosts       map[string]int
	ProgressionRule float32
}

// newProgressCost :
// Creates a new progress cost structure with an empty
// but valid map.
//
// Returns the created progress cost.
func newProgressCost() ProgressCost {
	return ProgressCost{
		InitCosts:       make(map[string]int),
		ProgressionRule: 1.0,
	}
}

// newProgressCostsModule :
// Creates a new module allowing to handle elements that
// have a progressive costs with a notion of level.
// This module can have a specific type as several game
// elements use this approach.
// A similar behavior to the `upgradablesModule` exists.
//
// The `log` defines the logging layer to forward to the
// base `upgradablesModule` element.
//
// The `kind` defines the type of upgradable associated to
// this module. It will help to determine which tables are
// relevant when fetching data from the DB.
//
// The `module` defines the string identifying the module
// to forward to the logging layer.
func newProgressCostsModule(log logger.Logger, kind upgradable, module string) *progressCostsModule {
	return &progressCostsModule{
		upgradablesModule: *newUpgradablesModule(log, kind, module),
		costs:             make(map[string]ProgressCost),
	}
}

// valid :
// Refinement of the base `upgradablesModule` valid method
// in order to perform some checks on the costs that are
// loaded in this module.
//
// Returns `true` if the upgradables module is valid and
// the internal resources as well.
func (pcm *progressCostsModule) valid() bool {
	return pcm.upgradablesModule.valid() && len(pcm.costs) > 0
}

// Init :
// Refinement of the base `upgradablesModule` behavior in
// order to add the fetching of the costs of each element
// associated to this module. This will typically fetch a
// new table in the DB where such costs are defined.
//
// The `dbase` represents the main data source to use
// to initialize the resources data.
//
// The `force` allows to erase any existing information
// and reload everything from the DB.
//
// Returns any error.
func (pcm *progressCostsModule) Init(dbase *db.DB, force bool) error {
	if dbase == nil {
		pcm.trace(logger.Error, fmt.Sprintf("Unable to initialize progress costs module from nil DB"))
		return db.ErrInvalidDB
	}

	// Prevent reload if not needed.
	if pcm.valid() && !force {
		return nil
	}

	// Initialize internal values.
	pcm.costs = make(map[string]ProgressCost)

	proxy := db.NewProxy(dbase)

	// We need to perform two queries: first to retrieve
	// the progression rule and then the initial cost of
	// the first level. We will first proceed to fetching
	// the progression rule as there should be a single
	// rule for all the elements.
	// After that we can retrieve the costs and associate
	// each one to the set of buildings we already fetched
	// in the first step. If a cost does not correspond to
	// any known building we found an inconsistency.
	err := pcm.initProgressionRules(proxy)
	if err != nil {
		pcm.trace(logger.Error, fmt.Sprintf("Failed to initialize progression rules for %s (err: %v)", pcm.uType, err))
		return err
	}

	err = pcm.initCosts(proxy)
	if err != nil {
		pcm.trace(logger.Error, fmt.Sprintf("Failed to initialize init costs for %s (err: %v)", pcm.uType, err))
		return err
	}

	return nil
}

// initProgressionRules :
// Used internally when fetching data from the DB to
// initialize the progression rules for elements to
// handle by this module.
//
// The `proxy` allows to access data from the DB in
// a convenient way.
//
// Returns any error.
func (pcm *progressCostsModule) initProgressionRules(proxy db.Proxy) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"element",
			"progress",
		},
		Table:   fmt.Sprintf("%s_costs_progress", pcm.uType),
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		pcm.trace(logger.Error, fmt.Sprintf("Unable to initialize %s progression costs rules (err: %v)", pcm.uType, err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		pcm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize %s progression costs rules (err: %v)", pcm.uType, rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var elem string
	var progress float32

	override := false

	for rows.Next() {
		err := rows.Scan(
			&elem,
			&progress,
		)

		if err != nil {
			pcm.trace(logger.Error, fmt.Sprintf("Failed to initialize progression costs rules from row (err: %v)", err))
			continue
		}

		// Register this value.
		costs, ok := pcm.costs[elem]

		// Check for overrides.
		if ok {
			pcm.trace(logger.Error, fmt.Sprintf("Overriding progress cost rule for \"%s\" (%f to %f)", elem, costs.ProgressionRule, progress))
			override = true
		}

		if !ok {
			costs = newProgressCost()
		}

		costs.ProgressionRule = progress

		pcm.costs[elem] = costs
	}

	if override {
		return ErrInconsistentDB
	}

	return nil
}

// initCosts :
// Used to perform the initialization of the inital costs
// for elements handled by this module. It will refine the
// data fetched from the progression rules so that the cost
// for any level can be computed.
//
// The `proxy` allows to easily access data from the DB.
//
// Returns any error.
func (pcm *progressCostsModule) initCosts(proxy db.Proxy) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"element",
			"res",
			"cost",
		},
		Table:   fmt.Sprintf("%s_costs", pcm.uType),
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		pcm.trace(logger.Error, fmt.Sprintf("Unable to initialize %s progression costs (err: %v)", pcm.uType, err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		pcm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize %s progression costs (err: %v)", pcm.uType, rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var elem, res string
	var cost int

	override := false
	inconsistent := false

	for rows.Next() {
		err := rows.Scan(
			&elem,
			&res,
			&cost,
		)

		if err != nil {
			pcm.trace(logger.Error, fmt.Sprintf("Failed to initialize progression costs from row (err: %v)", err))
			continue
		}

		// Register this value.
		costs, ok := pcm.costs[elem]

		// Check for overrides.
		if !ok {
			pcm.trace(logger.Error, fmt.Sprintf("Cannot interpret costs for \"%s\" with no associated progression rule", elem))
			inconsistent = true

			continue
		} else {
			e, ok := costs.InitCosts[res]

			if ok {
				pcm.trace(logger.Error, fmt.Sprintf("Overriding progress cost for resource \"%s\" for \"%s\" (%d to %d)", res, elem, e, cost))
				override = true
			}
		}

		costs.InitCosts[res] = cost

		pcm.costs[elem] = costs
	}

	if override || inconsistent {
		return ErrInconsistentDB
	}

	return nil
}
