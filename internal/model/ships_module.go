package model

import (
	"fmt"
	"math"
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
type ShipsModule struct {
	fixedCostsModule

	// The `characteristics` defines the properties of the
	// ships such as their base weapon, shield and armour
	// values or base speed.
	characteristics map[string]shipProps

	// The `rfVSShips` defines a map where the rapid fires
	// of ships against other ships can be loaded from DB.
	rfVSShips map[string][]RapidFire

	// The `rfVSDefenses` defines the rapid fires of ships
	// against defense systems.
	rfVSDefenses map[string][]RapidFire

	// The `propulsion` defines the propulsion how much a
	// single level of a propulsion technology increases
	// the speed of the ships using this tech.
	propulsion map[string]int
}

// shipProps :
// Defines the properties defining a ship as fetched
// from the DB. Common properties define the weapons
// value associated to the ship by default, along with
// the shielding and armour values.
// It also defines the cost of the propulsion system
// for this ship and the technology that influences
// it.
type shipProps struct {
	// The `cargo` defines the available space to hold
	// any resource. The larger this value the more of
	// resources can be carried.
	cargo int

	// The `shield` defines the base shielding value for
	// this ship.
	shield int

	// The `weapon` is analogous to the `shield` to set
	// the base attack value of the ship.
	weapon int

	// The `engines` defines the possible engines that
	// are associated to this ship depending on the
	// research level of the player.
	engines []Engine

	// The `consumption` is an association table that
	// defines for each resource the consumption of the
	// ship to move a single space distance. This value
	// is then multiplied by the total length of the
	// journey to get the total consumption. It is not
	// linked to the propulsion technology but rather to
	// the intrinsic properties of the ship.
	consumption []Consumption

	// The `deployment` is an association table that
	// defines for each resource the deployment cost
	// of this ship orbiting a foreign planet. Note
	// that the consumption is expressed in units
	// per hour.
	deployment []Consumption
}

// Engine :
// Describes the propulsion system used by a ship.
// The engine is composed of a propulsion system
// (which corresponds to a technology identifier)
// and a minimum level which indicates the research
// level that should be reached for this engine to
// be unlocked.
type Engine struct {
	// The `Propulsion` defines an identifier for the
	// technology used to build the engine.
	Propulsion PropulsionDesc `json:"propulsion_desc"`

	// The `MinLevel` defines the minimum level of the
	// propulsion technology in order for this engine
	// to be used.
	MinLevel int `json:"min_level"`

	// The `Speed` defines the base speed provided by
	// the engine.
	Speed int `json:"speed"`
}

// PropulsionDesc :
// Defines a propulsion system which is basically
// an aggregate of a propulsion technology and a
// factor describing the increase in speed that a
// single level increase brings.
type PropulsionDesc struct {
	// The `Propulsion` defines the identifier of the
	// propulsion technology.
	Propulsion string `json:"propulsion"`

	// The `Increase` defines the increase in speed
	// that each level of the propulsion technology
	// brings to the speed of the ship.
	Increase int `json:"increase"`
}

// Consumption :
// Alias to refer to the amount of resource that
// is needed to allow a journey.
type Consumption struct {
	ResourceAmount
}

// ShipDesc :
// Defines the abstract representation of a ship with
// its name and unique identifier. It can also include
// a short summary of its purpose retrieved from the
// database.
//
// The `Cargo`  defines the amount of cargo space on
// this ship. It can be used to store any mix of some
// resources.
//
// The `Shield` define sthe shielding value for this
// ship.
//
// The `Weapon` defines the attack value for this
// ship.
//
// The `Engines` defines the list of engines that
// are associated to this ship along with the lvl
// at which it applies.
//
// The `Consumption` defines the cost of the ship
// when it flies from a planet to another in units
// per distance.
//
// The `Deployment` defines the deployment cost
// for this ship when orbiting a foreign planet.
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
	UpgradableDesc

	Cargo        int           `json:"cargo"`
	Shield       int           `json:"shield"`
	Weapon       int           `json:"weapon"`
	Engines      []Engine      `json:"engines"`
	Consumption  []Consumption `json:"consumption"`
	Deployment   []Consumption `json:"deployment_cost"`
	RFVSShips    []RapidFire   `json:"rf_against_ships"`
	RFVSDefenses []RapidFire   `json:"rf_against_defenses"`
	Cost         FixedCost     `json:"cost"`
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

// SelectEngine :
// Used to allow the selection of the best propulsion
// system based on the input technologies (supposedly
// linked to a player).
//
// The `techs` define the technologies that have been
// researched by a player.
//
// Returns the best suited engine for this ship based
// on the input technologies.
func (s ShipDesc) SelectEngine(techs map[string]int) Engine {
	// The list of `engines` for this ship is ordered
	// by complexity. We will check which engine has
	// been unlocked and select the most complex one.
	locked := true
	id := 0

	for id = 0; id < len(s.Engines) && locked; id++ {
		eng := s.Engines[len(s.Engines)-1-id]

		// This engine is locked if the minimum level of
		// the technology required is not met.
		level := techs[eng.Propulsion.Propulsion]
		locked = (level < eng.MinLevel)
	}

	// Note: by default, we consider that the first
	// engine is always available. There are other
	// means that will guarantee that we prevent the
	// actual creation of ships in the first place
	// if the technology for the engine is not met.
	if locked {
		return s.Engines[0]
	}

	return s.Engines[len(s.Engines)-id]
}

// ComputeSpeed :
// Used to compute the speed of the engine based on
// the input technologies.
//
// The `techs` define the technologies that have
// been researched by a player.
//
// Returns the speed of this engine.
func (e Engine) ComputeSpeed(techs map[string]int) int {
	// Return the level of the technology used by
	// the engine. In case the technology is not
	// researched a `0` value will provided which
	// is fine.
	level := techs[e.Propulsion.Propulsion]

	return e.Propulsion.ComputeSpeed(e.Speed, level)
}

// ComputeSpeed :
// Used to compute the speed reached by this propulsion
// system given the base speed and the level of the tech
// researched so far.
//
// The `base` defines the base speed of the element.
//
// The `level` defines the level reached by the tech of
// the propulsion system.
//
// Returns the speed reached by the element.
func (p PropulsionDesc) ComputeSpeed(base int, level int) int {
	ratio := 1.0 + float64(level*p.Increase)/100
	fSpeed := float64(base) * ratio

	return int(math.Round(fSpeed))
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
// The `proxy` represents the main data source to use to
// initialize the buildings data.
//
// The `force` allows to erase any existing information
// and reload everything from the DB.
//
// Returns any error.
func (sm *ShipsModule) Init(proxy db.Proxy, force bool) error {
	// Prevent reload if not needed.
	if sm.valid() && !force {
		return nil
	}

	// Initialize internal values.
	sm.characteristics = make(map[string]shipProps)
	sm.rfVSShips = make(map[string][]RapidFire)
	sm.rfVSDefenses = make(map[string][]RapidFire)
	sm.propulsion = make(map[string]int)

	// And finally the propulsion upgrade systems.
	err := sm.initPropulsions(proxy)
	if err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Unable to initialize propulsion systems (err: %v)", err))
		return err
	}

	// Load the names and base information for each ship.
	// This operation is performed first so that the rest
	// of the data can be checked against the actual list
	// of ships that are defined in the game.
	err = sm.initCharacteristics(proxy)
	if err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Could not initialize module (err: %v)", err))
		return err
	}

	// Perform the initialization of the fixed costs, and
	// various data from the base handlers.
	err = sm.fixedCostsModule.Init(proxy, force)
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
			&props.cargo,
			&props.shield,
			&props.weapon,
		)

		if err != nil {
			sm.trace(logger.Error, fmt.Sprintf("Failed to initialize ship from row (err: %v)", err))
			inconsistent = true

			continue
		}

		// Check whether a ship with this identifier exists.
		if sm.existsID(ID) {
			sm.trace(logger.Error, fmt.Sprintf("Prevented override of ship \"%s\"", ID))
			override = true

			continue
		}

		// Retrieve consumption for this ship.
		props.consumption, err = sm.fetchConsumptionForShip(ID, proxy)
		if err != nil {
			sm.trace(logger.Error, fmt.Sprintf("Failure to fetch consumption for ship \"%s\" (err: %v)", ID, err))
			inconsistent = true

			continue
		}
		if len(props.consumption) == 0 {
			sm.trace(logger.Error, fmt.Sprintf("Didn't fetch any consumption for \"%s\"", ID))
			inconsistent = true

			continue
		}

		// Retrieve the list of engines used by this ship.
		props.engines, err = sm.fetchEnginesForShip(ID, proxy)
		if err != nil {
			sm.trace(logger.Error, fmt.Sprintf("Failure to fetch engines for ship \"%s\" (err: %v)", ID, err))
			inconsistent = true

			continue
		}
		if len(props.engines) == 0 {
			sm.trace(logger.Error, fmt.Sprintf("Didn't fetch any engine for \"%s\"", ID))
			inconsistent = true

			continue
		}

		// Retrieve the deployment cost for this ship.
		props.deployment, err = sm.fetchDeploymentCostForShip(ID, proxy)
		if err != nil {
			sm.trace(logger.Error, fmt.Sprintf("Failure to fetch deployment costs for ships \"%s\" (err: %v)", ID, err))
			inconsistent = true

			continue
		}
		if len(props.deployment) == 0 {
			sm.trace(logger.Error, fmt.Sprintf("Didn't fetch any deployment cost for \"%s\"", ID))
			inconsistent = true

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
	inconsistent := false

	for rows.Next() {
		err := rows.Scan(
			&ID,
			&increase,
		)

		if err != nil {
			sm.trace(logger.Error, fmt.Sprintf("Failed to initialize propulsion system from row (err: %v)", err))
			inconsistent = true

			continue
		}

		// Check for overrides.
		i, ok := sm.propulsion[ID]
		if ok {
			sm.trace(logger.Error, fmt.Sprintf("Overriding propulsion increase for \"%s\" from %d to %d", ID, i, increase))
			override = true
		}

		sm.propulsion[ID] = increase
	}

	if override || inconsistent {
		return ErrInconsistentDB
	}

	return nil
}

// fetchConsumptionForShip :
// Used to initialize the consumption for a ship provided
// as an identifier. The list of resources used by this
// ship when it flies is fetched and returned.
//
// The `ID` defines the identifier of the ship for which
// the consumption should be retrieved.
//
// The `proxy` allows to access to the DB.
//
// Returns the consumption for this ship along with any
// error.
func (sm *ShipsModule) fetchConsumptionForShip(ID string, proxy db.Proxy) ([]Consumption, error) {
	consumption := make([]Consumption, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"res",
			"amount",
		},
		Table: "ships_propulsion_cost",
		Filters: []db.Filter{
			{
				Key:    "ship",
				Values: []interface{}{ID},
			},
		},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Unable to initialize ships (err: %v)", err))
		return consumption, ErrNotInitialized
	}
	if rows.Err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize ships (err: %v)", rows.Err))
		return consumption, ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var res string
	var fuel int

	override := false
	inconsistent := false

	sanity := make(map[string]int)

	for rows.Next() {
		err := rows.Scan(
			&res,
			&fuel,
		)

		if err != nil {
			sm.trace(logger.Error, fmt.Sprintf("Failed to initialize ship consumption from row (err: %v)", err))
			inconsistent = true

			continue
		}

		// Check for overrides.
		eConsForRes, ok := sanity[res]
		if ok {
			sm.trace(logger.Error, fmt.Sprintf("Prevented override of propulsion consumption for resource \"%s\" of \"%s\" from %d to %d", res, ID, fuel, eConsForRes))
			override = true

			continue
		}

		sanity[res] = fuel

		c := Consumption{
			ResourceAmount: ResourceAmount{
				Resource: res,
				Amount:   float32(fuel),
			},
		}
		consumption = append(consumption, c)
	}

	if override || inconsistent {
		return consumption, ErrInconsistentDB
	}

	return consumption, nil
}

// fetchEnginesForShip :
// Used in a similar way to `fetchConsumptionForShip` to
// fetch the engines associated to a ship.
//
// The `ID` defines the identifier of the ship for which
// the engines should be retrieved.
//
// The `proxy` allows to access to the DB.
//
// Returns the engines for this ship along with any error.
func (sm *ShipsModule) fetchEnginesForShip(ID string, proxy db.Proxy) ([]Engine, error) {
	engines := make([]Engine, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"propulsion",
			"speed",
			"min_level",
		},
		Table: "ships_propulsion",
		Filters: []db.Filter{
			{
				Key:    "ship",
				Values: []interface{}{ID},
			},
		},
		Ordering: "order by rank",
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Unable to initialize ships (err: %v)", err))
		return engines, ErrNotInitialized
	}
	if rows.Err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize ships (err: %v)", rows.Err))
		return engines, ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var propulsion string
	var speed, minLevel int

	override := false
	inconsistent := false

	sanity := make(map[string]bool)

	for rows.Next() {
		err := rows.Scan(
			&propulsion,
			&speed,
			&minLevel,
		)

		if err != nil {
			sm.trace(logger.Error, fmt.Sprintf("Failed to initialize ship engine from row (err: %v)", err))
			inconsistent = true

			continue
		}

		// Check for overrides.
		_, ok := sanity[propulsion]
		if ok {
			sm.trace(logger.Error, fmt.Sprintf("Prevented override of ship \"%s\" engine for propulsion \"%s\"", ID, propulsion))
			override = true

			continue
		}

		sanity[propulsion] = true

		// Attempt to fetch the propulsion for this engine.
		prop, ok := sm.propulsion[propulsion]
		if !ok {
			sm.trace(logger.Error, fmt.Sprintf("Invalid propulsion technology \"%s\" for \"%s\"", propulsion, ID))
			inconsistent = true

			continue
		}

		e := Engine{
			Propulsion: PropulsionDesc{
				Propulsion: propulsion,
				Increase:   prop,
			},
			MinLevel: minLevel,
			Speed:    speed,
		}
		engines = append(engines, e)
	}

	if override || inconsistent {
		return engines, ErrInconsistentDB
	}

	return engines, nil
}

// fetchDeploymentCostForShip :
// Used in a similar way to `fetchConsumptionForShip` to
// fetch the deployment cost associated to a ship.
//
// The `ID` defines the identifier of the ship for which
// the deployment costs should be retrieved.
//
// The `proxy` allows to access to the DB.
//
// Returns the deployment costs for this ship along with
// any error.
func (sm *ShipsModule) fetchDeploymentCostForShip(ID string, proxy db.Proxy) ([]Consumption, error) {
	depCosts := make([]Consumption, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"res",
			"cost",
		},
		Table: "ships_deployment_cost",
		Filters: []db.Filter{
			{
				Key:    "ship",
				Values: []interface{}{ID},
			},
		},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Unable to initialize ships (err: %v)", err))
		return depCosts, ErrNotInitialized
	}
	if rows.Err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize ships (err: %v)", rows.Err))
		return depCosts, ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var cost ResourceAmount

	override := false
	inconsistent := false

	sanity := make(map[string]bool)

	for rows.Next() {
		err := rows.Scan(
			&cost.Resource,
			&cost.Amount,
		)

		if err != nil {
			sm.trace(logger.Error, fmt.Sprintf("Failed to initialize ship engine from row (err: %v)", err))
			inconsistent = true

			continue
		}

		// Check for overrides.
		_, ok := sanity[cost.Resource]
		if ok {
			sm.trace(logger.Error, fmt.Sprintf("Prevented override of ship \"%s\" deployment cost for resource \"%s\"", ID, cost.Resource))
			override = true

			continue
		}

		sanity[cost.Resource] = true

		c := Consumption{
			ResourceAmount: cost,
		}
		depCosts = append(depCosts, c)
	}

	if override || inconsistent {
		return depCosts, ErrInconsistentDB
	}

	return depCosts, nil
}

// Ships :
// Used to retrieve the ships matching the input filters
// from the data model. Note that if the DB has not yet
// been polled to retrieve data, we will return an error.
// The process will consist in first fetching all the IDs
// of the ships matching the filters, and then build the
// rest of the data from the already fetched values.
//
// The `proxy` defines the DB to use to fetch the ships
// description.
//
// The `filters` represent the list of filters to apply
// to the data fecthing. This will select only part of
// all the available ships.
//
// Returns the list of ships matching the filters along
// with any error.
func (sm *ShipsModule) Ships(proxy db.Proxy, filters []db.Filter) ([]ShipDesc, error) {
	// Initialize the module if for some reasons it is still
	// not valid.
	if !sm.valid() {
		err := sm.Init(proxy, true)
		if err != nil {
			return []ShipDesc{}, err
		}
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table:   "ships",
		Filters: filters,
	}

	IDs, err := sm.fetchIDs(query, proxy)
	if err != nil {
		sm.trace(logger.Error, fmt.Sprintf("Unable to fetch ships (err: %v)", err))
		return []ShipDesc{}, err
	}

	// Now build the data from the fetched identifiers.
	descs := make([]ShipDesc, 0)
	for _, ID := range IDs {
		desc, err := sm.GetShipFromID(ID)

		if err != nil {
			sm.trace(logger.Error, fmt.Sprintf("Unable to fetch ship \"%s\" (err: %v)", ID, err))
			continue
		}

		descs = append(descs, desc)
	}

	return descs, nil
}

// GetShipFromID :
// Used to retrieve a single ship by its identifier. It is
// similar to calling the `Ships` method but is a bit faster
// as we don't request the DB at all.
//
// The `ID` defines the identifier of the ship to fetch.
//
// Returns the description of the ship corresponding to
// the input identifier along with any error.
func (sm *ShipsModule) GetShipFromID(ID string) (ShipDesc, error) {
	// Attempt to retrieve the ships from its identifier.
	upgradable, err := sm.getDependencyFromID(ID)

	if err != nil {
		return ShipDesc{}, ErrInvalidID
	}

	desc := ShipDesc{
		UpgradableDesc: upgradable,
	}

	cost, ok := sm.costs[ID]
	if !ok {
		return desc, ErrInvalidID
	}
	desc.Cost = cost

	rfShip, ok := sm.rfVSShips[ID]
	if ok {
		desc.RFVSShips = rfShip
	} else {
		desc.RFVSShips = make([]RapidFire, 0)
	}

	rfDefense, ok := sm.rfVSDefenses[ID]
	if ok {
		desc.RFVSDefenses = rfDefense
	} else {
		desc.RFVSDefenses = make([]RapidFire, 0)
	}

	props, ok := sm.characteristics[ID]
	if !ok {
		return desc, ErrInvalidID
	}
	desc.Cargo = props.cargo
	desc.Shield = props.shield
	desc.Weapon = props.weapon
	desc.Engines = props.engines
	desc.Consumption = props.consumption
	desc.Deployment = props.deployment

	return desc, nil
}
