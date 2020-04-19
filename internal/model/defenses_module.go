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
// The `ID` defines the unique ID for this defense.
//
// The `Name` defines a display name for the defense.
//
// The `BuildingDeps` defines a list of identifiers
// which represent the buildings (and their associated
// level) which need to be available for this defense
// to be built. It is some sort of representation of
// the tech-tree.
//
// The `TechnologiesDeps` fills a similar purpose but
// register dependencies on technologies and not on
// buildings.
//
// The `Cost` defines how much of each resource needs
// to be available in a place to build a copy of this
// defense system.
type DefenseDesc struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	BuildingsDeps    []Dependency `json:"buildings_dependencies"`
	TechnologiesDeps []Dependency `json:"technologies_dependencies"`
	Cost             FixedCost    `json:"cost"`
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
// The `dbase` represents the main data source to use
// to initialize the defense systems data.
//
// The `force` allows to erase any existing information
// and reload everything from the DB.
//
// Returns any error.
func (dm *DefensesModule) Init(dbase *db.DB, force bool) error {
	if dbase == nil {
		dm.trace(logger.Error, fmt.Sprintf("Unable to initialize module from nil DB"))
		return db.ErrInvalidDB
	}

	// Prevent reload if not needed.
	if dm.valid() && !force {
		return nil
	}

	// Initialize internal values.
	dm.characteristics = make(map[string]defenseProps)

	proxy := db.NewProxy(dbase)

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
	err = dm.fixedCostsModule.Init(dbase, force)
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
