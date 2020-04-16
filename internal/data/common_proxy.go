package data

import (
	"fmt"
	"oglike_server/internal/locker"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"
)

// fetchElementDependency :
// Used to fetch the dependencies for the input element. We don't
// need to know the element in itself, we will just use it as a
// filter when fetching elements from the table described in input.
//
// The `dbase` defines the DB that should be used to fetch data.
// We assume that this value is valid (i.e. not `nil`). A panic
// may be issued if this is not the case.
//
// The `element` represents an identifier that can be used to only
// get some matching in the `table` where dependencies should be
// fetched.
//
// The `filterName` defines the name of the column to which the
// `element` should be applied. This will basically be translated
// into something like `where filterName='element'` in the SQL
// query.
//
// The `table` describes the name of the table from which deps are
// to be fetched.
//
// Returns the list of dependencies as described in the DB for
// the input `element` along with any error (in which case the
// list of dependencies should be ignored).
func fetchElementDependency(dbase *db.DB, element string, filterName string, table string) ([]TechDependency, error) {
	// Check consistency.
	if element == "" {
		return []TechDependency{}, fmt.Errorf("Cannot fetch dependencies for invalid element")
	}

	// Build and execute the query.
	props := []string{
		"requirement",
		"level",
	}

	query := fmt.Sprintf("select %s from %s where %s='%s'", strings.Join(props, ", "), table, filterName, element)

	// Execute the query.
	rows, err := dbase.DBQuery(query)

	if err != nil {
		return []TechDependency{}, fmt.Errorf("Could not retrieve dependencies for \"%s\" (err: %v)", element, err)
	}

	// Populate the dependencies.
	var gError error

	deps := make([]TechDependency, 0)
	var dep TechDependency

	for rows.Next() {
		err = rows.Scan(
			&dep.ID,
			&dep.Level,
		)

		if err != nil {
			gError = fmt.Errorf("Could not retrieve info for dependency of \"%s\" (err: %v)", element, err)
		}

		deps = append(deps, dep)
	}

	return deps, gError
}

// fetchElementCost :
// Used to fetch the cost to build the input element. We don't
// really need to interpret the element, we will just fetch the
// table indicated by the input arguments and search for elems
// matching the `element` key.
//
// The `dbase` defines the DB that should be queried to retrieve
// data. We assume that this value is valid (i.e. not `nil`) and
// a panic may be issued if this is not the case.
//
// The `element` defines the filtering key that will be searched
// in the corresponding table of the DB.
//
// The `filterName` defines the name of the column to which the
// `element` should be applied. This will basically be translated
// into something like `where filterName='element'` in the SQL
// query.
//
// The `table` describes the name of the table from which the
// costs should be retrieved.
//
// Returns the list of costs registered for the element in the
// DB. In case the error value is not `nil` the list should be
// ignored.
func fetchElementCost(dbase *db.DB, element string, filterName string, table string) ([]ResourceAmount, error) {
	// Check consistency.
	if element == "" {
		return []ResourceAmount{}, fmt.Errorf("Cannot fetch costs for invalid element")
	}

	// Build and execute the query.
	props := []string{
		"res",
		"cost",
	}

	query := fmt.Sprintf("select %s from %s where %s='%s'", strings.Join(props, ", "), table, filterName, element)

	// Execute the query.
	rows, err := dbase.DBQuery(query)

	if err != nil {
		return []ResourceAmount{}, fmt.Errorf("Could not retrieve costs for \"%s\" (err: %v)", element, err)
	}

	// Populate the costs.
	var gError error

	costs := make([]ResourceAmount, 0)
	var cost ResourceAmount
	var amount float32

	for rows.Next() {
		err = rows.Scan(
			&cost.Resource,
			&amount,
		)

		if err != nil {
			gError = fmt.Errorf("Could not retrieve info for cost of \"%s\" (err: %v)", element, err)
		}

		cost.Amount = float32(amount)

		costs = append(costs, cost)
	}

	return costs, gError
}

// performWithLock :
// Used to exectue the specified query on the provided DB by
// making sure that the lock on the specified resource is
// acquired and released when needed.
//
// The `resource` represents an identifier of the resoure to
// access with the query: this method makes sure that a lock
// on this resource is created and handled adequately.
//
// The `dbase` represents the DB into which the query should
// be performed. Should not be `nil` otheriwse a panic will
// be issued.
//
// The `query` represents the operation to perform on the DB
// which should be protected with a lock. It should consist
// in a valid SQL query.
//
// The `cl` allows to acquire a locker on the resource to
// make sure that a single routine is executing the update
// on the input resource at once.
//
// Returns any error occurring during the process.
func performWithLock(resource string, dbase *db.DB, query string, cl *locker.ConcurrentLocker) error {
	// Prevent invalid resource identifier.
	if resource == "" {
		return fmt.Errorf("Cannot update resources for invalid empty id")
	}

	// Acquire a lock on this resource.
	resLock := cl.Acquire(resource)
	defer cl.Release(resLock)

	// Perform the update: we will wrap the function inside
	// a dedicated handler to make sure that we don't lock
	// the resource more than necessary.
	var err error
	var errRelease error
	var errExec error

	func() {
		resLock.Lock()
		defer func() {
			if rawErr := recover(); rawErr != nil {
				err = fmt.Errorf("Error occured while executing query (err: %v)", rawErr)
			}
			errRelease = resLock.Release()
		}()

		// Perform the update.
		_, errExec = dbase.DBExecute(query)
	}()

	// Return any error.
	if errExec != nil {
		return fmt.Errorf("Could not perform operation on resource \"%s\" (err: %v)", resource, errExec)
	}
	if errRelease != nil {
		return fmt.Errorf("Could not release locker protecting resource \"%s\" (err: %v)", resource, err)
	}
	if err != nil {
		return fmt.Errorf("Could not execute operation on resource \"%s\" (err: %v)", resource, err)
	}

	return nil
}

// initResourcesFromDB :
// Used to query information from the DB and fetch data
// related to the resources. It is used to retrieve part
// of the data model to memory so as to speed up various
// processes.
// In case the DB cannot be contacted the process will
// fail.
//
// The `dbase` represents the DB from which info about
// resources should be fetched.
//
// The `log` allows to notify errors and info to the
// user in case of any failure.
//
// Returns a map representing all the data related to
// a resource along with any error.
func initResourcesFromDB(dbase *db.DB, log logger.Logger) (map[string]ResourceDesc, error) {
	resources := make(map[string]ResourceDesc)

	if dbase == nil {
		return resources, fmt.Errorf("Could not initialize resources from DB, no DB provided")
	}

	// Prepare the query to execute on the DB.
	props := []string{
		"id",
		"name",
		"base_production",
		"base_storage",
		"base_amount",
	}
	table := "resources"

	query := fmt.Sprintf("select %s from %s", strings.Join(props, ", "), table)

	rows, err := dbase.DBQuery(query)
	if err != nil {
		return resources, fmt.Errorf("Could not initialize resources from DB (err: %v)", err)
	}

	// Traverse the rows and store each resource in
	// the output map.
	var res ResourceDesc

	for rows.Next() {
		err = rows.Scan(
			&res.ID,
			&res.Name,
			&res.BaseProd,
			&res.BaseStorage,
			&res.BaseAmount,
		)

		if err != nil {
			log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for resource (err: %v)", err))
			continue
		}

		if existing, ok := resources[res.Name]; ok {
			log.Trace(logger.Warning, fmt.Sprintf("Overriding resource \"%s\" with id \"%s\" (existing \"%s\")", res.Name, res.ID, existing.ID))
		}

		resources[res.Name] = res
	}

	return resources, nil
}

// initProgressCostsFromDB :
// Used to query information from the DB and fetch info
// for the construction costs of an in-game element. As
// the structure for most of the elements is similar we
// can mutualize part of the code to fetch the info.
// This method assumes that the element for which the
// costs should be retrieved follow a progression rule
// where each level costs more than the previous one.
// We thus need to fetch both the initial cost and the
// rule to compute the cost at the next level. The info
// is assumed to be registered in two different tables
// that should be provided as input.
// In case the DB cannot be contacted an error is sent
// back but a valid map is still returned (it is empty
// though).
//
// The `dbase` represents the DB from which info about
// elements costs should be fetched.
//
// The `log` allows to notify errors and info to the
// user in case of any failure.
//
// The `propName` defines the name of the property that
// describes the identifier of the element in the DB.
// It will be used as the main key in the output map.
//
// The `tableForProgress` defines the name of the DB
// table defining the costs for the progression rules
// of each element.
//
// The `tableForCosts` define the name of the DB table
// defining the initial costs for each element. Along
// with the `tableForProgress` it defines all the info
// needed to compute the cost of a given level of the
// element.
//
// Returns a map representing the associated cost for
// each element along with the progression rule to use
// to compute the next level costs.
func initProgressCostsFromDB(dbase *db.DB, log logger.Logger, propName string, tableForProgress string, tableForCosts string) (map[string]ConstructionCost, error) {
	costs := make(map[string]ConstructionCost)

	if dbase == nil {
		return costs, fmt.Errorf("Could not fetch construction costs from DB, no DB provided")
	}

	// First retrieve the elements progression rule.
	props := []string{
		propName,
		"progress",
	}

	query := fmt.Sprintf("select %s from %s", strings.Join(props, ", "), tableForProgress)

	rows, err := dbase.DBQuery(query)
	if err != nil {
		return costs, fmt.Errorf("Could not initialize element construction costs from DB (err: %v)", err)
	}

	var ID string
	var progress float32

	for rows.Next() {
		err = rows.Scan(
			&ID,
			&progress,
		)

		if err != nil {
			log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for element cost (err: %v)", err))
			continue
		}

		existing, ok := costs[ID]
		if ok {
			log.Trace(logger.Error, fmt.Sprintf("Overriding progression rule for element \"%s\" (existing was %f, new is %f)", ID, existing.ProgressionRule, progress))
		}

		costs[ID] = ConstructionCost{
			make(map[string]int),
			progress,
		}
	}

	// Now populate the costs for each element.
	props = []string{
		propName,
		"res",
		"cost",
	}

	query = fmt.Sprintf("select %s from %s", strings.Join(props, ", "), tableForCosts)

	rows, err = dbase.DBQuery(query)
	if err != nil {
		return costs, fmt.Errorf("Could not initialize element constructions costs from DB (err: %v)", err)
	}

	var res string
	var cost int

	for rows.Next() {
		err = rows.Scan(
			&ID,
			&res,
			&cost,
		)

		if err != nil {
			log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for element construction cost (err: %v)", err))
			continue
		}

		elem, ok := costs[ID]

		if !ok {
			log.Trace(logger.Error, fmt.Sprintf("Cannot define cost of %d of resource \"%s\" for element \"%s\" not found in progression rules", cost, res, ID))
			continue
		}

		existing, ok := elem.InitCosts[res]
		if ok {
			log.Trace(logger.Error, fmt.Sprintf("Overriding cost for resource \"%s\" in element \"%s\" (existing was %d, new is %d)", res, ID, existing, cost))
		}

		elem.InitCosts[res] = cost

		costs[ID] = elem
	}

	return costs, nil
}

// initFixedCostsFromDB :
// Used in a similar way to the `initProgressCostsFromDB`
// but assumes that the elements to fetch follow a fixed
// costs rule where a single price is fixed for any unit
// of the element.
// The information is assumed to be registered in a single
// table in the DB which should be provided in input.
//
// The `dbase` is the DB from which information should
// be fetched.
//
// The `log` defines a way to inform the user of errors
// and notifications.
//
// The `propName` define the name of the field that is
// used to identify the element in the table. This is
// used as a key in the output map.
//
// The `table` defines the table inside the DB that is
// to be queried to fetch the costs.
//
// Returns a map where keys are the `propName` values
// extracted from the table along with the associated
// resources. Any error is also returned (in which case
// the map should be ignored).
func initFixedCostsFromDB(dbase *db.DB, log logger.Logger, propName string, table string) (map[string]FixedCost, error) {
	costs := make(map[string]FixedCost)

	if dbase == nil {
		return costs, fmt.Errorf("Could not fetch fixed costs from DB, no DB provided")
	}

	// Retrieve the elements from the provided table progression rule.
	props := []string{
		propName,
		"res",
		"cost",
	}

	query := fmt.Sprintf("select %s from %s", strings.Join(props, ", "), table)

	rows, err := dbase.DBQuery(query)
	if err != nil {
		return costs, fmt.Errorf("Could not initialize element fixed costs from DB (err: %v)", err)
	}

	var ID string
	var res string
	var cost int

	for rows.Next() {
		err = rows.Scan(
			&ID,
			&res,
			&cost,
		)

		if err != nil {
			log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for element cost (err: %v)", err))
			continue
		}

		// Check whether this element exists in the table.
		elem, ok := costs[ID]
		if !ok {
			elem = FixedCost{
				make(map[string]int),
			}
		}

		// Check whether the resource is already defined for
		// this element.
		if existing, ok := elem.InitCosts[res]; ok {
			log.Trace(logger.Error, fmt.Sprintf("Overriding fixed cost for res \"%s\" in element \"%s\" (existing was %d, new is %d)", res, ID, existing, cost))
		}

		elem.InitCosts[res] = cost

		costs[ID] = elem
	}

	return costs, nil
}

// initBuildingsProductionRulesFromDB :
// Similar to `initBuildingCostsFromDB` but used to
// query information about the production gains for
// each building from the DB.
// In case the DB cannot be contacted an error is sent
// back but a valid map is still returned (it is empty
// though).
//
// The `dbase` represents the DB from which info about
// buildings production should be fetched.
//
// The `log` allows to notify errors and info to the
// user in case of any failure.
//
// Returns a map representing the associated prod for
// each building. As a building can produce several
// resources, the key (corresponding to the building's
// identifier) can have several associated values (i.e.
// production rules).
func initBuildingsProductionRulesFromDB(dbase *db.DB, log logger.Logger) (map[string][]ProductionRule, error) {
	prodRules := make(map[string][]ProductionRule)

	if dbase == nil {
		return prodRules, fmt.Errorf("Could not buildings production rules from DB, no DB provided")
	}

	// First retrieve the buildings progression rule.
	props := []string{
		"building",
		"res",
		"base",
		"progress",
		"temperature_coeff",
		"temperature_offset",
	}
	table := "buildings_gains_progress"

	query := fmt.Sprintf("select %s from %s", strings.Join(props, ", "), table)

	rows, err := dbase.DBQuery(query)
	if err != nil {
		return prodRules, fmt.Errorf("Could not initialize buildings production rules from DB (err: %v)", err)
	}

	var buildingID string
	var res string
	var base int
	var progress float32
	var tempCoeff float32
	var tempOffset float32

	// This map allows to make sure that we don't override for
	// a single building the rule associated to a particular
	// resource.
	resForBuildings := make(map[string]map[string]bool)

	for rows.Next() {
		err = rows.Scan(
			&buildingID,
			&res,
			&base,
			&progress,
			&tempCoeff,
			&tempOffset,
		)

		if err != nil {
			log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for building production rule (err: %v)", err))
			continue
		}

		// Make sure that this resource has not already been defined
		// for this building.
		resForBuilding, ok := resForBuildings[buildingID]
		if !ok {
			// Obvisouly not defined yet as even the building does not
			// have any associated production rules.
			resForBuilding = make(map[string]bool)
		} else {
			_, ok = resForBuilding[res]
			if ok {
				log.Trace(logger.Error, fmt.Sprintf("Overriding production rule for resource \"%s\" for building \"%s\"", res, buildingID))
			}
		}

		// In any case we will register this resource as defined for
		// the current building.
		resForBuilding[res] = true
		resForBuildings[buildingID] = resForBuilding

		// If the building already exists, we will append a new prod
		// rule for the corresponding resource. If the building does
		// not exist yet we will create it.
		buildingRules, ok := prodRules[buildingID]
		if !ok {
			buildingRules = make([]ProductionRule, 0)
		}

		rule := ProductionRule{
			Resource:          res,
			InitProd:          base,
			ProgressionRule:   progress,
			TemperatureCoeff:  tempCoeff,
			TemperatureOffset: tempOffset,
		}

		buildingRules = append(buildingRules, rule)

		// Save back the building to the map.
		prodRules[buildingID] = buildingRules
	}

	return prodRules, nil
}

// initBuildingsStorageRulesFromDB :
// Used to query information from the DB and fetch
// the progression for the various storage defined
// in the DB. This will allow to compute for each
// level the capacity for the storage associated
// to each resource.
// In case the DB cannot be contacted an error is
// returned but a valid map is still returned (it
// is empty though).
//
// The `dbase` represents the DB from which info
// about storage capacities should be fetched.
//
// The `log` allows to notify errors and info to
// the user in case of any failure.
//
// Returns a map representing for each storage
// its associated properties along with any
// error.
func initBuildingsStorageRulesFromDB(dbase *db.DB, log logger.Logger) (map[string][]StorageRule, error) {
	storageRules := make(map[string][]StorageRule)

	if dbase == nil {
		return storageRules, fmt.Errorf("Could not initialize storage capacities from DB, no DB provided")
	}

	// Prepare the query to execute on the DB.
	props := []string{
		"building",
		"res",
		"base",
		"multiplier",
		"progress",
	}

	table := "buildings_storage_progress"

	query := fmt.Sprintf("select %s from %s", strings.Join(props, ", "), table)

	rows, err := dbase.DBQuery(query)
	if err != nil {
		return storageRules, fmt.Errorf("Could not initialize storage capacities from DB (err: %v)", err)
	}

	// Traverse the rows and store each storage rule in the output map.
	var buildingID string
	var rule StorageRule

	// This map allows to make sure that we don't override for
	// a single building the storage rule associated to a single
	// resource.
	storageForBuildings := make(map[string]map[string]bool)

	for rows.Next() {
		err = rows.Scan(
			&buildingID,
			&rule.Resource,
			&rule.InitStorage,
			&rule.Multiplier,
			&rule.Progress,
		)

		if err != nil {
			log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for storage capacity (err: %v)", err))
			continue
		}

		// Make sure that this resource has not already been defined
		// for this building.
		storageForBuilding, ok := storageForBuildings[buildingID]
		if !ok {
			// Obvisouly not defined yet as even the building does not
			// have any associated storage rules.
			storageForBuilding = make(map[string]bool)
		} else {
			_, ok = storageForBuilding[rule.Resource]
			if ok {
				log.Trace(logger.Error, fmt.Sprintf("Overriding storage rule for resource \"%s\" for building \"%s\"", rule.Resource, buildingID))
			}
		}

		// In any case we will register this resource as defined for
		// the current building.
		storageForBuilding[rule.Resource] = true
		storageForBuildings[buildingID] = storageForBuilding

		// If the building already exists, we will append a new
		// storage rule for the corresponding resource. If the
		// building does not exist yet we will create it.
		buildingRules, ok := storageRules[buildingID]
		if !ok {
			buildingRules = make([]StorageRule, 0)
		}

		buildingRules = append(buildingRules, rule)

		// Save back the building to the map.
		storageRules[buildingID] = buildingRules
	}

	return storageRules, nil
}
