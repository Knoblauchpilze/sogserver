package model

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// ShipsModule :
// Refines the concept of `fixedCostsModule` for the
// particular case of ships. A ship is the main vector
// of interaction between planets as it can move from
// one planet to another for various reasons. A ship
// just like any upgradabe element has a list of techs
// and buildings that should be researched/built to be
// able to build the ship, and has some base properties
// of its own.
// A ship always cost the same amount no matter the
// number already built.
// Contrary to defenses, ships implement a mechanism
// called rapid fire: it indicates a certain probability
// for the ship to fire again when it hit a certain item.
// This is really interesting as it allows to make more
// shots in a single round. Rapid fires can be defined
// against other ships or against defense systems.
//
// The `characteristics` defines the properties of the
// ships such as their base weapon, shield and armour
// values or base speed.
//
// The `rfVSShips` defines a map where the rapid fires
// of ships against other ships can be loaded from DB.
//
// The `rfVSDefenses` defines the rapid fires of ships
// against defense systems.
//
// The `propulsion` defines the propulsion how much a
// single level of a propulsion technology increases
// the speed of the ships using this tech.
type ShipsModule struct {
	fixedCostsModule

	characteristics map[string]shipProps
	rfVSShips       map[string][]RapidFire
	rfVSDefenses    map[string][]RapidFire
	propulsion      map[string]int
}

// shipProps :
// Defines the properties defining a ship as fetched
// from the DB. Common properties define the weapons
// value associated to the ship by default, along with
// the shielding and armour values.
// It also defines the cost of the propulsion system
// for this ship and the technology that influences
// it.
//
// The `cargo` defines the available space to hold
// any resource. The larger this value the more of
// resources can be carried.
//
// The `shield` defines the base shielding value for
// this ship.
//
// The `weapon` is analogous to the `shield` to set
// the base attack value of the ship.
//
// The `speed` defines the base speed for this ship.
// The actual speed of the ship is then computed by
// adding the boost provided by the level of research
// for the propulsion technology used by the ship.
//
// The `propulsion` defines the identifier of the
// technology used as propulsion system for the ship.
// It should reference a valid technology.
//
// The `consumption` is an association table that
// defines for each resource the consumption of the
// ship to move a single space distance. This value
// is then multiplied by the total length of the
// journey to get the total consumption. It is not
// linked to the propulsion technology but rather to
// the intrinsic properties of the ship.
type shipProps struct {
	cargo       int
	shield      int
	weapon      int
	speed       int
	propulsion  string
	consumption map[string]int
}

// ShipDesc :
// Defines the abstract representation of a ship with
// its name and unique identifier. It can also include
// a short summary of its purpose retrieved from the
// database.
//
// The `ID` defines the unique ID for this ship.
//
// The `Name` defines a display name for the ship.
//
// The `BuildingDeps` defines a list of ID which can
// be scanned to find the buildings (and their level)
// that are needed to be available on a planet to be
// allowed to build the ship. It is some sort of way
// to represent the tech-tree.
//
// The `TechnologiesDeps` fills a similar purpose but
// register dependencies on technologies.
//
// The `RFVSShips` defines the rapid fire this ship
// has against other ships.
//
// The `RFVSDefenses` defines the rapid fire this ship
// has against defenses.
//
// The `Cost` defines how much of each resource is
// needed to build one copy of this ship.
type ShipDesc struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	BuildingsDeps    []Dependency `json:"buildings_dependencies"`
	TechnologiesDeps []Dependency `json:"technologies_dependencies"`
	RFVSShips        []RapidFire  `json:"rf_against_ships"`
	RFVSDefenses     []RapidFire  `json:"rf_against_defenses"`
	Cost             FixedCost    `json:"cost"`
}

// RapidFire :
// Describes a rapid fire from a unit on another. It is
// defined by both the identifier of the element that is
// subject to the rapid fire along with a value which
// describes the actual effect.
//
// The `Receiver` defines the element that is subject
// to a rapid fire from the provider.
//
// The `RF` defines the actual value of the rapid fire.
type RapidFire struct {
	Receiver string `json:"receiver"`
	RF       int    `json:"rf"`
}

// NewShipsModule :
// Creates a module allowing to handle ships defined
// in the game. It will fetch the necessary data to
// describe ships and their propulsion systems and
// armament before calling the base handler which is
// meant to fetch the costs associated to each ship.
//
// The `log` defines the logging layer to forward to the
// base `progressCostsModule` element.
func NewShipsModule(log logger.Logger) *ShipsModule {
	return &ShipsModule{
		fixedCostsModule: *newFixedCostsModule(log, Ship, "ships"),
		characteristics:  nil,
		rfVSShips:        nil,
		rfVSDefenses:     nil,
		propulsion:       nil,
	}
}

// valid :
// Refinement of the base `fixedCostsModule` method in
// order to perform some checks on the effects that are
// loaded in this module.
//
// Returns `true` if the fixed costs module is valid
// and the internal resources as well.
func (sm *ShipsModule) valid() bool {
	return sm.fixedCostsModule.valid() &&
		len(sm.characteristics) > 0 &&
		len(sm.rfVSShips) > 0 &&
		len(sm.rfVSDefenses) > 0 &&
		len(sm.propulsion) > 0
}

// Init :
// Provide some more implementation to retrieve data from
// the DB by fetching the armament and general properties
// of ships from the DB before handing over to the base
// fixed costs module.
// The role of this method is to populate the base list
// of ships so that base implementation can know to which
// element the costs should be binded.
//
// The `dbase` represents the main data source to use
// to initialize the buildings data.
//
// The `force` allows to erase any existing information
// and reload everything from the DB.
//
// Returns any error.
func (sm *ShipsModule) Init(dbase *db.DB, force bool) error {
	if dbase == nil {
		sm.trace(logger.Error, fmt.Sprintf("Unable to initialize module from nil DB"))
		return db.ErrInvalidDB
	}

	// Prevent reload if not needed.
	if sm.valid() && !force {
		return nil
	}

	// Initialize internal values.
	sm.characteristics = make(map[string]shipProps)
	sm.rfVSShips = make(map[string][]RapidFire)
	sm.rfVSDefenses = make(map[string][]RapidFire)
	sm.propulsion = make(map[string]int)

	proxy := db.NewProxy(dbase)

	// Load the names and base information for each ship.
	// This operation is performed first so that the rest
	// of the data can be checked against the actual list
	// of ships that are defined in the game.
	err := sm.initCharacteristics(proxy)
	if err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Could not initialize module (err: %v)", err))
		return err
	}

	// Perform the initialization of the fixed costs, and
	// various data from the base handlers.
	err = sm.fixedCostsModule.Init(dbase, force)
	if err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Failed to initialize base module (err: %v)", err))
		return err
	}

	// Update the rapid fires both for ships and for the
	// defense systems.
	err = sm.initRapidFires(proxy, "ships_rapid_fire", &sm.rfVSShips)
	if err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Unable to initialize rapid fires (err: %v)", err))
		return err
	}

	err = sm.initRapidFires(proxy, "ships_rapid_fire_defenses", &sm.rfVSDefenses)
	if err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Unable to initialize rapid fires (err: %v)", err))
		return err
	}

	// And finally the propulsion upgrade systems.
	err = sm.initPropulsions(proxy)
	if err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Unable to initialize propulsion systems (err: %v)", err))
		return err
	}

	return nil
}

// initCharacteristics :
// Used to perform the fetching of the definition of ships
// from the input proxy. It will only get some basic info
// about the ships such as their names and identifiers in
// order to populate the minimum information to fact-check
// the rest of the data (like the rapid fires, etc.).
//
// The `proxy` defines a convenient way to access to the
// DB.
//
// Returns any error.
func (sm *ShipsModule) initCharacteristics(proxy db.Proxy) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
			"name",
			"propulsion",
			"speed",
			"cargo",
			"shield",
			"weapon",
		},
		Table:   "ships",
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Unable to initialize ships (err: %v)", err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize ships (err: %v)", rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var ID, name string
	var props shipProps

	override := false
	inconsistent := false

	for rows.Next() {
		err := rows.Scan(
			&ID,
			&name,
			&props.propulsion,
			&props.speed,
			&props.cargo,
			&props.shield,
			&props.weapon,
		)

		if err != nil {
			sm.trace(logger.Error, fmt.Sprintf("Failed to initialize ship from row (err: %v)", err))
			continue
		}

		// Check whether a ship with this identifier exists.
		if sm.existsID(ID) {
			sm.trace(logger.Error, fmt.Sprintf("Prevented override of ship \"%s\"", ID))
			override = true

			continue
		}

		// Register this ship in the association table.
		err = sm.registerAssociation(ID, name)
		if err != nil {
			sm.trace(logger.Error, fmt.Sprintf("Cannot register ship \"%s\" (id: \"%s\") (err: %v)", name, ID, err))
			inconsistent = true

			continue
		}

		sm.characteristics[ID] = props
	}

	if override || inconsistent {
		return ErrInconsistentDB
	}

	rows.Close()

	// Now update the consumption of the ships through a query
	// on the dedicated table.
	query.Props = []string{
		"ship",
		"res",
		"amount",
	}
	query.Table = "ships_propulsion_cost"

	rows, err = proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Unable to initialize ships (err: %v)", err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize ships (err: %v)", rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var res string
	var consumption int

	sanity := make(map[string]map[string]int)

	for rows.Next() {
		err := rows.Scan(
			&ID,
			&res,
			&consumption,
		)

		if err != nil {
			sm.trace(logger.Error, fmt.Sprintf("Failed to initialize ship consumption from row (err: %v)", err))
			continue
		}

		// Check whether a ship with this identifier exists.
		if !sm.existsID(ID) {
			sm.trace(logger.Error, fmt.Sprintf("Cannot register propulsion consumption for \"%s\" not defined in DB", ID))
			inconsistent = true

			continue
		}

		// Check for overrides.
		eCons, ok := sanity[ID]
		if !ok {
			eCons = make(map[string]int)
		} else {
			c, ok := eCons[res]

			if ok {
				sm.trace(logger.Error, fmt.Sprintf("Overriding propulsion consumption for resource \"%s\" of \"%s\" from %d to %d", res, ID, c, consumption))
				override = true
			}
		}

		eCons[res] = consumption
		sanity[ID] = eCons

		// Register this value: note that we know that the props
		// exist because we checked that `sm.existsID` before.
		props := sm.characteristics[ID]
		props.consumption[res] = consumption
		sm.characteristics[ID] = props
	}

	if override || inconsistent {
		return ErrInconsistentDB
	}

	return nil
}

// initRapidFires :
// Used to initialzie the rapid fire tables for each ship
// against other ships and defense systems. Assumes that
// a list of ships is already existing so that we can be
// sure that the rapid fires are actually valid.
//
// The `proxy` defines a convenient way to access to the
// data from the DB.
//
// The `name` of the table that should be fetched to
// get the rapid fires.
//
// The `out` provides the map that should be populated
// by rapid fires fetched by this module.
//
// Returns any error.
func (sm *ShipsModule) initRapidFires(proxy db.Proxy, table string, out *map[string][]RapidFire) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"ship",
			"target",
			"rapid_fire",
		},
		Table:   table,
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Unable to initialize rapid fires (err: %v)", err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize rapid fires (err: %v)", rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var ID string
	var rf RapidFire

	override := false
	inconsistent := false

	sanity := make(map[string]map[string]int)

	for rows.Next() {
		err := rows.Scan(
			&ID,
			&rf.Receiver,
			&rf.RF,
		)

		if err != nil {
			sm.trace(logger.Error, fmt.Sprintf("Failed to initialize rapid fire from row (err: %v)", err))
			continue
		}

		// Check whether a ship with this identifier exists
		// (both for the provider and receiver of the RF).
		if !sm.existsID(ID) {
			sm.trace(logger.Error, fmt.Sprintf("Cannot register rapid fire for \"%s\" not defined in DB", ID))
			inconsistent = true

			continue
		}

		// Check for overrides.
		eRFs, ok := sanity[ID]
		if !ok {
			eRFs = make(map[string]int)
			eRFs[rf.Receiver] = rf.RF
		} else {
			_, ok := eRFs[rf.Receiver]

			if ok {
				sm.trace(logger.Error, fmt.Sprintf("Prevented override of rapid fire for ship \"%s\" on \"%s\"", ID, rf.Receiver))
				override = true

				continue
			}

			eRFs[rf.Receiver] = rf.RF
		}

		sanity[ID] = eRFs

		// Register this value.
		rfs, ok := (*out)[ID]

		if !ok {
			rfs = make([]RapidFire, 0)
		}

		rfs = append(rfs, rf)
		(*out)[ID] = rfs
	}

	if override || inconsistent {
		return ErrInconsistentDB
	}

	return nil
}

// initPropulsions :
// Used to initialize the propulsion systems used in the
// game. Basically each propulsion system increases the
// speed of a ship by a certain amount for each additional
// level of the technology. This information is retrieved
// from this function.
//
// The `proxy` defines a convenience way to access the DB.
//
// Returns any error.
func (sm *ShipsModule) initPropulsions(proxy db.Proxy) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"propulsion",
			"increase",
		},
		Table:   "ships_propulsion_increase",
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Unable to initialize propulsion systems (err: %v)", err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize propulsion systems (err: %v)", rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var ID string
	var increase int

	override := false

	for rows.Next() {
		err := rows.Scan(
			&ID,
			&increase,
		)

		if err != nil {
			sm.trace(logger.Error, fmt.Sprintf("Failed to initialize propulsion system from row (err: %v)", err))
			continue
		}

		// Check for overrides.
		i, ok := sm.propulsion[ID]
		if ok {
			sm.trace(logger.Error, fmt.Sprintf("Overriding propulsion increase for \"%s\" from %d to %d", ID, i, increase))
		}

		sm.propulsion[ID] = increase
	}

	if override {
		return ErrInconsistentDB
	}

	return nil
}
