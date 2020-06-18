package model

import (
	"fmt"
	"math"
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
// The `allowedOnPlanet` defines whether the buildings
// are allowed on a planet.
//
// The `allowedOnMoon` is similar to the above param
// but defines whether buildings can be built on moons.
//
// The `production` allows to store all the production
// rules for buildings handled by the module.
//
// The `storage` fills a similar purpose but for the
// storage effects a building has.
//
// The `fields` defines the rules to increase the fields
// available on a planet for each level of a building.
type BuildingsModule struct {
	progressCostsModule

	allowedOnPlanet map[string]bool
	allowedOnMoon   map[string]bool
	production      map[string][]ProductionRule
	storage         map[string][]StorageRule
	fields          map[string]FieldsRule
}

// BuildingDesc :
// Defines the abstract representation of a building
// with its name and unique identifier. It provides
// some info about the effects that this building has
// on production, storage, etc.
//
// The `AllowedOnPlanet` defines whether the item
// can be built on a planet.
//
// The `AllowedOnMoon` is similar to the above param
// but for the construction status on moons.
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
//
// The `Fields` defines the rules that are assigned to
// an increase in the available field for each level
// of this building built.
type BuildingDesc struct {
	UpgradableDesc

	AllowedOnPlanet bool             `json:"allowed_on_planet"`
	AllowedOnMoon   bool             `json:"allowed_on_moon"`
	Cost            ProgressCost     `json:"cost"`
	Production      []ProductionRule `json:"production,omitempty"`
	Storage         []StorageRule    `json:"storage,omitempty"`
	Fields          FieldsRule       `json:"fields,omitempty"`
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

// ComputeProduction :
// Used to perform the computation of the resources that
// are produced by the level `level` of the element that
// is described by the input production rule.
// The level is clamped to be in the range `[0; +inf[` if
// this is not already the case.
//
// The `level` for which the production should be computed.
// It is clamped to be positive.
//
// The `temperature` defines the average temperature of
// the planet where the production is evaluated. It is
// used to determine the temperature dependent part of the
// resource production.
//
// Returns the amount of resource that are produced by the
// selected rule with the specified level and temperature
// values.
func (pr ProductionRule) ComputeProduction(level int, temperature float32) float32 {
	// Clamp the input level.
	fLevel := math.Max(0.0, float64(level))
	fInitProd := float64(pr.InitProd)

	// Compute both parts of the production (temperature
	// dependent and independent).
	tempDep := float64(pr.TemperatureOffset + temperature*pr.TemperatureCoeff)
	tempIndep := fInitProd * fLevel * math.Pow(float64(pr.ProgressionRule), fLevel)

	prod := tempDep * tempIndep

	return float32(prod)
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

// newStorageRule :
// Creates a new storage rule which does not allow to
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

// ComputeStorage :
// Used to perform the computation of the amount of res
// that can be held at the specified level.
//
// The `level` for which the storage capacity should be
// computed.
//
// Returns the amount of resource that can be held for
// the specified level by this storage.
func (sr StorageRule) ComputeStorage(level int) int {
	factor := float64(sr.Multiplier) * math.Exp(float64(sr.Progress)*float64(level))
	return sr.InitStorage * int(math.Floor(factor))
}

// FieldsRule :
// Defines the rule to update the fields of a planet
// or a moon based on the level of a building. It is
// used to change the available number of fields.
//
// The `Multiplier` defines the slope of the linear
// function defining the increase in fields.
//
// The `Constant` defines the ordinate of the fields
// increase rule.
type FieldsRule struct {
	Multiplier float32
	Constant   int
}

// newFieldsRule :
// Creates a new fields rule which does not lead to
// any increase in fields count based on the level
// of a building.
//
// Returns the created fields rule.
func newFieldsRule() FieldsRule {
	return FieldsRule{
		Multiplier: 0.0,
		Constant:   0,
	}
}

// ComputeFields :
// Used to perform the computation of the additional
// fields provided by the level of the building in
// input.
//
// The `level` for which the increase in fields needs
// to be computed.
//
// Returns the number of fields added by building the
// input level of the building.
func (fr FieldsRule) ComputeFields(level int) int {
	prevFields := float64(0.0)
	if level > 0 {
		prevFields = float64(fr.Multiplier)*float64(level-1) + float64(fr.Constant)
	}

	fFields := float64(fr.Multiplier)*float64(level) + float64(fr.Constant)

	return int(math.Floor(fFields - prevFields))
}

// NewBuildingsModule :
// Creates a new module allowing to handle buildings
// defined in the game. It applies the abstract set
// of functions for upgradable and progress costs
// model to the specific case of buildings, while in
// addition providing information about the storage
// and production effects of each building.
//
// The `log` defines the logging layer to forward to the
// base `progressCostsModule` element.
func NewBuildingsModule(log logger.Logger) *BuildingsModule {
	return &BuildingsModule{
		progressCostsModule: *newProgressCostsModule(log, Building, "buildings"),
		allowedOnPlanet:     nil,
		allowedOnMoon:       nil,
		production:          nil,
		storage:             nil,
		fields:              nil,
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
	return bm.progressCostsModule.valid() &&
		len(bm.allowedOnPlanet) > 0 &&
		len(bm.allowedOnMoon) > 0 &&
		len(bm.production) > 0 &&
		len(bm.storage) > 0 &&
		len(bm.fields) > 0
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
func (bm *BuildingsModule) Init(proxy db.Proxy, force bool) error {
	// Prevent reload if not needed.
	if bm.valid() && !force {
		return nil
	}

	// Initialize internal values.
	bm.allowedOnPlanet = make(map[string]bool)
	bm.allowedOnMoon = make(map[string]bool)
	bm.production = make(map[string][]ProductionRule)
	bm.storage = make(map[string][]StorageRule)
	bm.fields = make(map[string]FieldsRule)

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
	err = bm.progressCostsModule.Init(proxy, force)
	if err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Failed to initialize base module (err: %v)", err))
		return err
	}

	// Initialize allowed on planets and moons status.
	err = bm.initAllowedStatus(proxy)
	if err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Unable to initialize allowed status (err: %v)", err))
		return err
	}

	// Perform the initialization of the production rules.
	err = bm.initProduction(proxy)
	if err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Unable to initialize production rules (err: %v)", err))
		return err
	}

	// Update the storage rules.
	err = bm.initStorage(proxy)
	if err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Unable to initialize storage rules (err: %v)", err))
		return err
	}

	// Update the fields increase rules.
	err = bm.initFields(proxy)
	if err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Unable to initialize fields rules (err: %v)", err))
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
			inconsistent = true

			continue
		}

		// Check whether a building with this identifier exists.
		if bm.existsID(ID) {
			bm.trace(logger.Error, fmt.Sprintf("Prevented override of building \"%s\"", ID))
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

// initAllowedStatus :
// Used to fetch for each building whether it can be built on
// planets and moons. It is used to make sure that buildings
// are built only on relevant celestial bodies.
//
// The `proxy` defines a convenient way to access to the DB.
//
// Returns any error.
func (bm *BuildingsModule) initAllowedStatus(proxy db.Proxy) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
			"buildable_on_planet",
			"buildable_on_moon",
		},
		Table:   "buildings",
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Unable to initialize allowed construction rules (err: %v)", err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize allowed construction rules (err: %v)", rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var ID string
	var allowedOnP, allowedOnM bool

	override := false
	inconsistent := false

	sanity := make(map[string]bool)

	for rows.Next() {
		err := rows.Scan(
			&ID,
			&allowedOnP,
			&allowedOnM,
		)

		if err != nil {
			bm.trace(logger.Error, fmt.Sprintf("Failed to initialize allowed construction rules from row (err: %v)", err))
			continue
		}

		// Check whether a building with this identifier exists.
		if !bm.existsID(ID) {
			bm.trace(logger.Error, fmt.Sprintf("Cannot register allowed construction rule for \"%s\" not defined in DB", ID))
			inconsistent = true

			continue
		}

		// Check for overrides.
		_, ok := sanity[ID]
		if ok {
			bm.trace(logger.Error, fmt.Sprintf("Prevented override of allowed construction rule for building \"%s\"", ID))
			override = true

			continue
		}

		sanity[ID] = true

		// Register this value.
		bm.allowedOnPlanet[ID] = allowedOnP
		bm.allowedOnMoon[ID] = allowedOnM
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
			bm.trace(logger.Error, fmt.Sprintf("Cannot register storage rule for \"%s\" not defined in DB", ID))
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

// initFields :
// Similar to the above functions but handles the loading of
// the fields increase rules associated to buildings. Rules
// are checked to make sure that they reference existing
// buildings.
//
// The `proxy` defines a convenient way to access to the DB.
//
// Returns any error.
func (bm *BuildingsModule) initFields(proxy db.Proxy) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"element",
			"multiplier",
			"constant",
		},
		Table:   "buildings_fields_progress",
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Unable to initialize fields rules (err: %v)", err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize fields rules (err: %v)", rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var ID string
	var rule FieldsRule

	override := false
	inconsistent := false

	sanity := make(map[string]bool)

	for rows.Next() {
		err := rows.Scan(
			&ID,
			&rule.Multiplier,
			&rule.Constant,
		)

		if err != nil {
			bm.trace(logger.Error, fmt.Sprintf("Failed to initialize fields rules from row (err: %v)", err))
			continue
		}

		// Check whether a building with this identifier exists.
		if !bm.existsID(ID) {
			bm.trace(logger.Error, fmt.Sprintf("Cannot register fields rule for \"%s\" not defined in DB", ID))
			inconsistent = true

			continue
		}

		// Check for overrides.
		_, ok := sanity[ID]
		if ok {
			bm.trace(logger.Error, fmt.Sprintf("Prevented override of fields rule for building \"%s\"", ID))
			override = true

			continue
		}

		sanity[ID] = true

		// Register this value.
		bm.fields[ID] = rule
	}

	if override || inconsistent {
		return ErrInconsistentDB
	}

	return nil
}

// Buildings :
// Used to retrieve the buildings matching the input filters
// from the data model. Note that if the DB has not yet been
// polled to retrieve data, we will return an error.
// The process will consist in first fetching the identifiers
// of the buildings matching the filters, and then build the
// rest of the data from the already fetched values.
//
// The `proxy` defines the DB to use to fetch the buildings
// description.
//
// The `filters` represent the list of filters to apply to
// the data fecthing. This will select only part of all the
// available buildings.
//
// Returns the list of buildings matching the filters along
// with any error.
func (bm *BuildingsModule) Buildings(proxy db.Proxy, filters []db.Filter) ([]BuildingDesc, error) {
	// We will first perform a query on the DB to get all the
	// identifiers that matche the input criteria and then use
	// the returned values to build the buildings description.
	// We will also try to initialize this module if needed.
	if !bm.valid() {
		err := bm.Init(proxy, true)
		if err != nil {
			return []BuildingDesc{}, err
		}
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table:   "buildings",
		Filters: filters,
	}

	IDs, err := bm.fetchIDs(query, proxy)
	if err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Unable to fetch buildings (err: %v)", err))
		return []BuildingDesc{}, err
	}

	// Now build the data from the fetched identifiers.
	descs := make([]BuildingDesc, 0)
	for _, ID := range IDs {
		desc, err := bm.GetBuildingFromID(ID)

		if err != nil {
			bm.trace(logger.Error, fmt.Sprintf("Unable to fetch building \"%s\" (err: %v)", ID, err))
			continue
		}

		descs = append(descs, desc)
	}

	return descs, nil
}

// GetBuildingFromID :
// Used to retrieve a single building by its identifier. It
// is similar to calling the `Buildings` method but is quite
// faster as we don't request the DB at all.
//
// The `ID` defines the identifier of the building to fetch.
//
// Returns the description of the building corresponding to
// the input identifier along with any error.
func (bm *BuildingsModule) GetBuildingFromID(ID string) (BuildingDesc, error) {
	// Attempt to retrieve the building from its identifier.
	upgradable, err := bm.getDependencyFromID(ID)

	if err != nil {
		return BuildingDesc{}, ErrInvalidID
	}

	desc := BuildingDesc{
		UpgradableDesc: upgradable,
	}

	cost, ok := bm.costs[ID]
	if !ok {
		return desc, ErrInvalidID
	}
	desc.Cost = cost

	desc.AllowedOnPlanet = bm.allowedOnPlanet[ID]
	desc.AllowedOnMoon = bm.allowedOnMoon[ID]

	prod, ok := bm.production[ID]
	if ok {
		desc.Production = prod
	} else {
		desc.Production = make([]ProductionRule, 0)
	}

	storage, ok := bm.storage[ID]
	if ok {
		desc.Storage = storage
	} else {
		desc.Storage = make([]StorageRule, 0)
	}

	fields, ok := bm.fields[ID]
	if ok {
		desc.Fields = fields
	} else {
		desc.Fields = newFieldsRule()
	}

	return desc, nil
}
