package model

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// DefensesModule :
// Refines the concept of `fixedCostsModule` for the case
// of defenses. Defense systems are elements that can be
// built on planet to oppose any attacking ship. The role
// of this module is to fetch information about systems
// that are defined in the game to make them available in
// the rest of the server.
// Just like any other game element a defense system has
// some dependencies and also some armament capabilities.
//
// The `characteristics` define the properties of each
// defense system fetched from the DB.
type DefensesModule struct {
	fixedCostsModule

	characteristics map[string]defenseProps
}

// defenseProps :
// Defines the properties defining a defense system as
// fetched from the DB. Common properties define the
// weapon value associated to the defense along with a
// shielding capability.
//
// The `shield` defines the base shielding value for
// this ship.
//
// The `weapon` is analogous to the `shield` to set
// the base attack value of the ship.
type defenseProps struct {
	shield int
	weapon int
}

// DefenseDesc :
// Defines the properties associated to a defense system
// as described in the DB. Information about the attack
// capabilities of the system are provided along with a
// list of costs required to build it.
//
// The `Shield` defines the shielding value for this
// defense system.
//
// The `Weapon` defines the weapon value (i.e. attack
// value) for this defense system.
//
// The `Cost` defines how much of each resource needs
// to be available in a place to build a copy of this
// defense system.
type DefenseDesc struct {
	UpgradableDesc

	Shield int       `json:"shield,omitempty"`
	Weapon int       `json:"weapon,omitempty"`
	Cost   FixedCost `json:"cost,omitempty"`
}

// NewDefensesModule :
// Creates a new module allowing to handle defense
// systems in the game. The associated properties
// are fetched from the DB upon calling the `Init`
// method.
//
// The `log` defines the logging layer to forward
// to the base `fixedCostsModule` element.
func NewDefensesModule(log logger.Logger) *DefensesModule {
	return &DefensesModule{
		fixedCostsModule: *newFixedCostsModule(log, Defense, "defenses"),
		characteristics:  nil,
	}
}

// valid :
// Refinement of the base `fixedCostsModule` method in
// order to perform some checks on the effects that are
// loaded in this module.
//
// Returns `true` if the fixed costs module is valid
// and the internal resources as well.
func (dm *DefensesModule) valid() bool {
	return dm.fixedCostsModule.valid() && len(dm.characteristics) > 0
}

// Init :
// Provide some more implementation to retrieve data from
// the DB by fetching the defense systems' identifiers and
// display names. This will constitute the base from which
// the upgradable module can attach the costs and various
// properties to the systems.
//
// The `proxy` represents the main data source to use
// to initialize the defense systems data.
//
// The `force` allows to erase any existing information
// and reload everything from the DB.
//
// Returns any error.
func (dm *DefensesModule) Init(proxy db.Proxy, force bool) error {
	// Prevent reload if not needed.
	if dm.valid() && !force {
		return nil
	}

	// Initialize internal values.
	dm.characteristics = make(map[string]defenseProps)

	// Load the names and base information for each defense
	// system. This operation is performed first so that the
	// rest of the data can be checked against the list of
	// systems that are defined in the game.
	err := dm.initProps(proxy)
	if err != nil {
		dm.trace(logger.Error, fmt.Sprintf("Could not initialize module (err: %v)", err))
		return err
	}

	// Perform the initialization of costs.
	err = dm.fixedCostsModule.Init(proxy, force)
	if err != nil {
		dm.trace(logger.Error, fmt.Sprintf("Failed to initialize base module (err: %v)", err))
		return err
	}

	return nil
}

// initProps :
// Used to perform the fetching of the definition of the
// defense systems from the input proxy. It will only get
// some basic info about each one such as their names and
// identifiers in order to populate the minimum info to
// fact-check the rest of the data.
//
// The `proxy` defines a convenient way to access to the
// DB.
//
// Returns any error.
func (dm *DefensesModule) initProps(proxy db.Proxy) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
			"name",
			"shield",
			"weapon",
		},
		Table:   "defenses",
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		dm.trace(logger.Error, fmt.Sprintf("Unable to initialize defenses (err: %v)", err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		dm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize defenses (err: %v)", rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var ID, name string
	var props defenseProps

	override := false
	inconsistent := false

	for rows.Next() {
		err := rows.Scan(
			&ID,
			&name,
			&props.shield,
			&props.weapon,
		)

		if err != nil {
			dm.trace(logger.Error, fmt.Sprintf("Failed to initialize defense from row (err: %v)", err))
			continue
		}

		// Check whether a technology with this identifier exists.
		if dm.existsID(ID) {
			dm.trace(logger.Error, fmt.Sprintf("Prevented override of defense \"%s\"", ID))
			override = true

			continue
		}

		// Register this technology in the association table.
		err = dm.registerAssociation(ID, name)
		if err != nil {
			dm.trace(logger.Error, fmt.Sprintf("Cannot register defense \"%s\" (id: \"%s\") (err: %v)", name, ID, err))
			inconsistent = true
		}

		dm.characteristics[ID] = props
	}

	if override || inconsistent {
		return ErrInconsistentDB
	}

	return nil
}

// Defenses :
// Used to retrieve the defenses matching the input filters
// from the data model. Note that if the DB has not yet been
// polled to retrieve data, we will return an error.
// The process will consist in first fetching the identifiers
// of the defenses matching the filters, and then build the
// rest of the data from the already fetched values.
//
// The `proxy` defines the DB to use to fetch the defenses
// description.
//
// The `filters` represent the list of filters to apply to
// the data fecthing. This will select only part of all the
// available defenses.
//
// Returns the list of defenses matching the filters along
// with any error.
func (dm *DefensesModule) Defenses(proxy db.Proxy, filters []db.Filter) ([]DefenseDesc, error) {
	// Try to initialize the module if it is not yet valid.
	if !dm.valid() {
		err := dm.Init(proxy, true)
		if err != nil {
			return []DefenseDesc{}, err
		}
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table:   "defenses",
		Filters: filters,
	}

	IDs, err := dm.fetchIDs(query, proxy)
	if err != nil {
		dm.trace(logger.Error, fmt.Sprintf("Unable to fetch defenses (err: %v)", err))
		return []DefenseDesc{}, err
	}

	// Now build the data from the fetched identifiers.
	descs := make([]DefenseDesc, 0)
	for _, ID := range IDs {
		desc, err := dm.getDefenseFromID(ID)

		if err != nil {
			dm.trace(logger.Error, fmt.Sprintf("Unable to fetch defense \"%s\" (err: %v)", ID, err))
			continue
		}

		descs = append(descs, desc)
	}

	return descs, nil
}

// getDefenseFromID :
// Used to retrieve a single defense by its identifier. It
// is similar to calling the `Defenses` method but is quite
// faster as we don't request the DB at all.
//
// The `ID` defines the identifier of the defense to fetch.
//
// Returns the description of the defense corresponding to
// the input identifier along with any error.
func (dm *DefensesModule) getDefenseFromID(ID string) (DefenseDesc, error) {
	// Attempt to retrieve the defense from its identifier.
	upgradable, err := dm.getDependencyFromID(ID)

	if err != nil {
		return DefenseDesc{}, ErrInvalidID
	}

	desc := DefenseDesc{
		UpgradableDesc: upgradable,
	}

	cost, ok := dm.costs[ID]
	if !ok {
		return desc, ErrInvalidID
	}
	desc.Cost = cost

	props, ok := dm.characteristics[ID]
	if !ok {
		return desc, ErrInvalidID
	}
	desc.Shield = props.shield
	desc.Weapon = props.weapon

	return desc, nil
}
