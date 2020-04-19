package model

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// fixedCostsModule :
// Refine the interface provided by the upgradables module
// to handle cases where the upgradable has a fixed cost
// (meaning that no matter the number already built costs
// are always the same). This is typically used for items
// that are unit-like (ships and defenses for example).
//
// The `costs` defines for each element handled by this
// module the corresponding cost. The costs refers to the
// identifiers of resources and associate a single integer
// value to each resource.
type fixedCostsModule struct {
	upgradablesModule

	costs map[string]FixedCost
}

// FixedCost :
// Defines a set of costs for several resources. The
// keys of the map correspond to resources identifier
// as defined in the DB while the values correspond
// to the cost for this resource.
//
// The `InitCosts` define the association between the
// cost for each resource.
type FixedCost struct {
	InitCosts map[string]int
}

// newFixedCost :
// Creates a new fixed cost structure with an empty
// but valid map.
//
// Returns the created fixed cost.
func newFixedCost() FixedCost {
	return FixedCost{
		InitCosts: make(map[string]int),
	}
}

// newFixedCostsModule :
// Creates a new module allowing to handle elements that
// have a fixed cost no matter the amount already built.
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
func newFixedCostsModule(log logger.Logger, kind upgradable, module string) *fixedCostsModule {
	return &fixedCostsModule{
		upgradablesModule: *newUpgradablesModule(log, kind, module),
		costs:             make(map[string]FixedCost),
	}
}

// valid :
// Refinement of the base `upgradablesModule` valid method
// in order to perform some checks on the costs that are
// loaded in this module.
//
// Returns `true` if the upgradables module is valid and
// the internal resources as well.
func (fcm *fixedCostsModule) valid() bool {
	return fcm.upgradablesModule.valid() && len(fcm.costs) > 0
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
func (fcm *fixedCostsModule) Init(dbase *db.DB, force bool) error {
	if dbase == nil {
		fcm.trace(logger.Error, fmt.Sprintf("Unable to initialize fixed costs module from nil DB"))
		return db.ErrInvalidDB
	}

	// Prevent reload if not needed.
	if fcm.valid() && !force {
		return nil
	}

	// Initialize internal values.
	fcm.costs = make(map[string]FixedCost)

	proxy := db.NewProxy(dbase)

	// Create the query to fetch the fixed costs and execute it.
	query := db.QueryDesc{
		Props: []string{
			"element",
			"res",
			"cost",
		},
		Table:   fmt.Sprintf("%s_costs", fcm.uType),
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		fcm.trace(logger.Error, fmt.Sprintf("Unable to initialize %s fixed costs (err: %v)", fcm.uType, err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		fcm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize %s fixed costs (err: %v)", fcm.uType, rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var elem, res string
	var cost int

	override := false

	for rows.Next() {
		err := rows.Scan(
			&elem,
			&res,
			&cost,
		)

		if err != nil {
			fcm.trace(logger.Error, fmt.Sprintf("Failed to initialize fixed cost from row (err: %v)", err))
			continue
		}

		// Register this value.
		costs, ok := fcm.costs[elem]
		if !ok {
			costs = newFixedCost()
		}

		// Check for overrides.
		c, ok := costs.InitCosts[res]
		if ok {
			fcm.trace(logger.Error, fmt.Sprintf("Overriding fixed cost for resource \"%s\" on \"%s\" (%d to %d)", res, elem, c, cost))
			override = true
		}

		costs.InitCosts[res] = cost

		fcm.costs[elem] = costs
	}

	if override {
		return ErrInconsistentDB
	}

	return nil
}
