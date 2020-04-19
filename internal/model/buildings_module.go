package model

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// BuildingsModule :
// Refines the concept of `progressCostsModule` for the
// particular case of buildings. A building is an item
// that can be built on a planet and which allows some
// effects on various aspects of the game.
// Typical buildings are mines (which allow to produce
// a certain amount of resource), storages (which can
// be used to increase the amount of resources that can
// be stored on a planet), etc.
// Each level of a building occupies a field on its
// parent planet. The higher the level the more effects
// it has.
//
// The `production` allows to store all the production
// rules for buildings handled by the module.
//
// The `storage` fills a similar purpose but for the
// storage effects a building has.
type BuildingsModule struct {
	progressCostsModule

	production map[string][]ProductionRule
	storage    map[string][]StorageRule
}

// BuildingDesc :
// Defines the abstract representation of a building
// with its name and unique identifier. It provides
// some info about the effects that this building has
// on production, storage, etc.
//
// The `ID` defines the unique id for this building.
//
// The `Name` defines a human readable name for the
// building.
//
// The `BuildingDeps` defines a list of identifiers
// which represent the buildings (and their associated
// level) which need to be available for this building
// to be built. It is some sort of representation of
// the tech-tree.
//
// The `TechnologiesDeps` fills a similar purpose but
// register dependencies on technologies.
//
// The `Cost` allows to compute the cost of this item
// at any level.
//
// The `Production` defines the list of resources that
// see their production (either positive or negative)
// modified with each upgrade of this building.
//
// The `Storage` defines how upgrading this building
// impacts the storage capacities of the planet.
type BuildingDesc struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	BuildingsDeps    []Dependency     `json:"buildings_dependencies"`
	TechnologiesDeps []Dependency     `json:"technologies_dependencies"`
	Cost             ProgressCost     `json:"cost"`
	Production       []ProductionRule `json:"production"`
	Storage          []StorageRule    `json:"storage"`
}

// ProductionRule :
// Used to define the rule to produce some quantity
// of a resource for a building. Some of the in-game
// buildings are able to generate resources meaning
// that they provide a certain amount or a resource
// at each step of time (usually using the second as
// time unit).
// The higher the level of the building, the more of
// the resource will be produced.
//
// The `Resource` defines an identifier of the res
// that is produced by the element. It should be
// linked to an actual existing resource in the
// `resources` table.
//
// The `InitProd` defines the base production at the
// `0`-th level for this resource. This is the base
// from where the production gains are scaled.
//
// The `ProgressionRule` defines the base associated
// to the exponential growth in production for the
// resource. The larger this value the quicker the
// production will rise with each level.
//
// The `TemperatureCoeff` defines a coefficient to
// apply to the production which depends on the temp
// of the planet where the production rule is applied.
// If this value is positive it means that the hotter
// the planet is, the more production for this res is
// to be expected, while the effect is reversed if the
// coefficient is negative.
// A coefficient of `0` means that the temperature of
// the planet does not have any impact on the resource
// prod.
//
// The `TemperatureOffset` is used in conjunction with
// the `TemperatureCoeff` and allows to provide some
// boost in the computation.
// Typically the temperature dependent part of the
// production of a resource looks something like this:
// `TemperatureCoeff * T + TemperatureOffset`.
type ProductionRule struct {
	Resource          string
	InitProd          int
	ProgressionRule   float32
	TemperatureCoeff  float32
	TemperatureOffset float32
}

// newProductionRule :
// Creates a new production rule which does not produce
// anything for the specified resource.
//
// The `res` defines the identifier of the resource to
// bind to this production rule.
//
// Returns the created production rule.
func newProductionRule(res string) ProductionRule {
	return ProductionRule{
		Resource:          res,
		InitProd:          0,
		ProgressionRule:   1.0,
		TemperatureCoeff:  0.0,
		TemperatureOffset: 1.0,
	}
}

// StorageRule :
// Used to define the prgoression rule for a storage.
// It defines the way storage scale with the level as
// it increases.
//
// The `Resource` defines the id of the resource that
// the storage hold.
//
// The `InitStorage` defines some base storage that
// is used in the formula to compute the capacity at
// each level.
//
// The `Multiplier` defines another parameter for the
// formula to compute the capacity.
//
// The `Progress` defines the exponential constant in
// used to make the storage capacity progres.
type StorageRule struct {
	Resource    string
	InitStorage int
	Multiplier  float32
	Progress    float32
}

// newProductionRule :
// Creates a new stroage rule which does not allow to
// store anything for the specified resource.
//
// The `res` defines the identifier of the resource to
// bind to this storage rule.
//
// Returns the created storage rule.
func newStorageRule(res string) StorageRule {
	return StorageRule{
		Resource:    res,
		InitStorage: 0,
		Multiplier:  0.0,
		Progress:    0.0,
	}
}

// newBuildingsModule :
// Creates a new module allowing to handle buildings
// defined in the game. It applies the abstract set
// of functions for upgradable and progress costs
// model to the specific case of buildings, while in
// addition providing information about the storage
// and production effects of each building.
//
// The `log` defines the logging layer to forward to the
// base `progressCostsModule` element.
func newBuildingsModule(log logger.Logger) *BuildingsModule {
	return &BuildingsModule{
		progressCostsModule: *newProgressCostsModule(log, Building, "buildings"),
		production:          nil,
		storage:             nil,
	}
}

// valid :
// Refinement of the base `progressCostsModule` method
// in order to perform some checks on the effects that
// are loaded in this module.
//
// Returns `true` if the progress costs module is valid
// and the internal resources as well.
func (bm *BuildingsModule) valid() bool {
	return bm.progressCostsModule.valid() && len(bm.production) > 0 && len(bm.storage) > 0
}

// Init :
// Provide some more implementation to retrieve data from
// the DB by fetching the production and storage rules of
// each building. It will also parse the table defining
// the buildings in order to allow the rest of the data
// to be attached to the right elements.
//
// The `dbase` represents the main data source to use
// to initialize the buildings data.
//
// The `force` allows to erase any existing information
// and reload everything from the DB.
//
// Returns any error.
func (bm *BuildingsModule) Init(dbase *db.DB, force bool) error {
	if dbase == nil {
		bm.trace(logger.Error, fmt.Sprintf("Unable to initialize module from nil DB"))
		return db.ErrInvalidDB
	}

	// Prevent reload if not needed.
	if bm.valid() && !force {
		return nil
	}

	// Initialize internal values.
	bm.production = make(map[string][]ProductionRule)
	bm.storage = make(map[string][]StorageRule)

	proxy := db.NewProxy(dbase)

	// Load the names and base information for each building.
	// This operation is performed first so that the rest of
	// the data can be checked against the actual list of
	// buildings that are defined in the game.
	err := bm.initNames(proxy)
	if err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Could not initialize module (err: %v)", err))
		return err
	}

	// Perform the initialization of the progression rules,
	// and various data from the base handlers.
	err = bm.progressCostsModule.Init(dbase, force)
	if err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Failed to initialize base progression module (err: %v)", err))
		return err
	}

	// Perform the initialization of the production rules.
	err = bm.initProduction(proxy)
	if err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Unable to initialize production rules (err: %v)", err))
		return err
	}

	// And finally update the storage rules.
	err = bm.initStorage(proxy)
	if err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Unable to initialize storage rules (err: %v)", err))
		return err
	}

	return nil
}

// initNames :
// Used to perform the fetching of the definition of buildings
// from the input proxy. It will only get some basic info about
// the buildings such as their names and identifiers in order
// to populate the minimum information to fact-check the rest
// of the data (like the production rules, etc.).
//
// The `proxy` defines a convenient way to access to the DB.
//
// Returns any error.
func (bm *BuildingsModule) initNames(proxy db.Proxy) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
			"name",
		},
		Table:   "buildings",
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Unable to initialize buildings (err: %v)", err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize buildings (err: %v)", rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var ID, name string

	override := false
	inconsistent := false

	for rows.Next() {
		err := rows.Scan(
			&ID,
			&name,
		)

		if err != nil {
			bm.trace(logger.Error, fmt.Sprintf("Failed to initialize building from row (err: %v)", err))
			continue
		}

		// Check whether a building with this identifier exists.
		if bm.existsID(ID) {
			bm.trace(logger.Error, fmt.Sprintf("Overriding building \"%s\"", ID))
			override = true

			continue
		}

		// Register this building in the association table.
		err = bm.registerAssociation(ID, name)
		if err != nil {
			bm.trace(logger.Error, fmt.Sprintf("Cannot register building \"%s\" (id: \"%s\") (err: %v)", name, ID, err))
			inconsistent = true
		}
	}

	if override || inconsistent {
		return ErrInconsistentDB
	}

	return nil
}

// initProduction :
// Used to fetch the production rules associated to buildings
// from the DB. It will check each production rule to make
// sure that it is associated to an existing building.
//
// The `proxy` defines a convenient way to access to the DB.
//
// Returns any error.
func (bm *BuildingsModule) initProduction(proxy db.Proxy) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"element",
			"res",
			"base",
			"progress",
			"temperature_coeff",
			"temperature_offset",
		},
		Table:   "buildings_gains_progress",
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Unable to initialize production rules (err: %v)", err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize production rules (err: %v)", rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var ID string
	var rule ProductionRule

	override := false
	inconsistent := false

	sanity := make(map[string]map[string]bool)

	for rows.Next() {
		err := rows.Scan(
			&ID,
			&rule.Resource,
			&rule.InitProd,
			&rule.ProgressionRule,
			&rule.TemperatureCoeff,
			&rule.TemperatureOffset,
		)

		if err != nil {
			bm.trace(logger.Error, fmt.Sprintf("Failed to initialize production rules from row (err: %v)", err))
			continue
		}

		// Check whether a building with this identifier exists.
		if !bm.existsID(ID) {
			bm.trace(logger.Error, fmt.Sprintf("Cannot register production rule for \"%s\" not defined in DB", ID))
			inconsistent = true

			continue
		}

		// Check for overrides.
		eRules, ok := sanity[ID]
		if !ok {
			eRules = make(map[string]bool)
			eRules[rule.Resource] = true
		} else {
			_, ok := eRules[rule.Resource]

			if ok {
				bm.trace(logger.Error, fmt.Sprintf("Prevented override of production rule for resource \"%s\" for \"%s\"", rule.Resource, ID))
				override = true

				continue
			}

			eRules[rule.Resource] = true
		}

		sanity[ID] = eRules

		// Register this value.
		prodRules, ok := bm.production[ID]

		if !ok {
			prodRules = make([]ProductionRule, 0)
		}

		prodRules = append(prodRules, rule)
		bm.production[ID] = prodRules
	}

	if override || inconsistent {
		return ErrInconsistentDB
	}

	return nil
}

// initStorage :
// Similar to `initProduction` but handles the loading of the
// storage rules associated to buildings. Any rule is checked
// to make sure that it references an existing building.
//
// The `proxy` defines a convenient way to access to the DB.
//
// Returns any error.
func (bm *BuildingsModule) initStorage(proxy db.Proxy) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"element",
			"res",
			"base",
			"multiplier",
			"progress",
		},
		Table:   "buildings_storage_progress",
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Unable to initialize storage rules (err: %v)", err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize storage rules (err: %v)", rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var ID string
	var rule StorageRule

	override := false
	inconsistent := false

	sanity := make(map[string]map[string]bool)

	for rows.Next() {
		err := rows.Scan(
			&ID,
			&rule.Resource,
			&rule.InitStorage,
			&rule.Multiplier,
			&rule.Progress,
		)

		if err != nil {
			bm.trace(logger.Error, fmt.Sprintf("Failed to initialize storage rules from row (err: %v)", err))
			continue
		}

		// Check whether a building with this identifier exists.
		if !bm.existsID(ID) {
			bm.trace(logger.Error, fmt.Sprintf("Cannot register stroage rule for \"%s\" not defined in DB", ID))
			inconsistent = true

			continue
		}

		// Check for overrides.
		eRules, ok := sanity[ID]
		if !ok {
			eRules = make(map[string]bool)
			eRules[rule.Resource] = true
		} else {
			_, ok := eRules[rule.Resource]

			if ok {
				bm.trace(logger.Error, fmt.Sprintf("Prevented override of storage rule for resource \"%s\" for \"%s\"", rule.Resource, ID))
				override = true

				continue
			}

			eRules[rule.Resource] = true
		}

		sanity[ID] = eRules

		// Register this value.
		storageRules, ok := bm.storage[ID]

		if !ok {
			storageRules = make([]StorageRule, 0)
		}

		storageRules = append(storageRules, rule)
		bm.storage[ID] = storageRules
	}

	if override || inconsistent {
		return ErrInconsistentDB
	}

	return nil
}
