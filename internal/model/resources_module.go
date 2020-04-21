package model

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// ResourcesModule :
// This structure is populated upon starting the server by
// fetching data from the DB to load in memory the values
// that hardly change during the life of the game. Values
// like that are typically resources, technologies, ships,
// defense, buildings, etc. All this data is set for the
// game and then used by players/accounts/fleets to build
// more complex information.
// We usually have to rely on the base elements to compute
// more complex stuff like the completion time when a unit
// needs to be built, or a technology researched. We also
// need to access this data to validate most of the actions
// that can be performed by users.
// Having all the data in the DB is usually not enough as
// otherwsie it forces to manipulate everything through
// scripts, which is not convenient. Moreover, we usually
// talk in terms of names rather than identifier. Rules
// to evolve a building for example involve the levle of
// specific building which are described by their name.
// As we don't know in advance the identifier of these
// buildings, it makes much more sense to have some sort
// of adapter that can handle conversion for us.
// This particular structure is responsible for loading
// the resources information from the DB, but other items
// in this package are designed in a similar way: they
// just handle different aspects of the game.
//
// The `prod` defines a map where the keys are resources
// identifiers (as loaded from the DB) while the values
// are base production level for each resource on a new
// planet. This value is expressed in units per hour.
//
// The `storage` defines a map where keys are resources
// identifiers while values are the base storage on a
// new planet for said resources.
//
// The `amount` defines a similar map where the values
// correspond to the initial amount of the resource that
// exists on the planet when it's created.
type ResourcesModule struct {
	associationTable
	baseModule

	prod    map[string]int
	storage map[string]int
	amount  map[string]int
}

// ResourceDesc :
// Defines the abstract representation of a resource which
// is bascially an identifier and the actual name of the
// resource plus some base properties of the resource.
//
// The `ID` defines the identifier of the resource in the
// table gathering all the in-game resources. This is used
// in most of the other operations referencing resources.
//
// The `Name` defines the human-readable name of the res
// as displayed to the user.
//
// The `BaseProd` defines the production without modifiers
// for this resource on each planet. It represents a way
// for the user to get resources without building anything
// else.
//
// The `BaseStorage` defines the base capacity to store
// the resource without any modifiers (usually hangars).
//
// The `BaseAmount` defines the base amount for this res
// that can be found on any new planet in the game.
type ResourceDesc struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	BaseProd    int    `json:"base_production"`
	BaseStorage int    `json:"base_storage"`
	BaseAmount  int    `json:"base_amount"`
}

// getModuleString :
// Used to retrieve a unique string characterizing this
// module to pass to the `baseModule`. This will allow
// to easily filter log messages coming from this system
// of the game.
//
// Returns the string representing this module.
func getModuleString() string {
	return "resources"
}

// NewResourcesModule :
// Used to create a new resources module initialized with
// no content (as no DB is provided yet). The module will
// stay invalid until the `init` method is called with a
// valid DB.
//
// The `log` defines the logging layer to forward to the
// base `baseModule` element.
func NewResourcesModule(log logger.Logger) *ResourcesModule {
	return &ResourcesModule{
		associationTable: associationTable{},
		baseModule:       newBaseModule(log, getModuleString()),
		prod:             nil,
		storage:          nil,
		amount:           nil,
	}
}

// valid :
// Refinement of the base `associationTable` valid method
// in order to perform some checks on the production and
// storage rules defined in this module.
//
// Returns `true` if the association table is valid and
// the internal resources as well.
func (rm *ResourcesModule) valid() bool {
	return rm.associationTable.valid() && len(rm.prod) > 0 && len(rm.storage) > 0 && len(rm.amount) > 0
}

// Init :
// Implementation of the `DBModule` interface to allow
// fetching information from the input DB and load to
// local memory.
//
// The `proxy` represents the main data source to use
// to initialize the resources data.
//
// The `force` allows to erase any existing information
// and reload everything from the DB.
//
// Returns any error.
func (rm *ResourcesModule) Init(proxy db.Proxy, force bool) error {
	// Prevent reload if not needed.
	if rm.valid() && !force {
		return nil
	}

	// Initialize internal values.
	rm.prod = make(map[string]int)
	rm.storage = make(map[string]int)
	rm.amount = make(map[string]int)

	// Perform the DB query through a dedicated DB proxy.
	query := db.QueryDesc{
		Props: []string{
			"id",
			"name",
			"base_production",
			"base_storage",
			"base_amount",
		},
		Table:   "resources",
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	// Check for errors.
	if err != nil {
		rm.trace(logger.Error, fmt.Sprintf("Unable to initialize resources module (err: %v)", err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		rm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize resources module (err: %v)", err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var ID, name string
	var prod, storage, amount int

	override := false
	inconsistent := false

	for rows.Next() {
		err = rows.Scan(
			&ID,
			&name,
			&prod,
			&storage,
			&amount,
		)

		if err != nil {
			rm.trace(logger.Error, fmt.Sprintf("Failed to initialize resource from row (err: %v)", err))
			continue
		}

		// Check for overrides.
		ep, pok := rm.prod[ID]
		es, sok := rm.storage[ID]
		ea, aok := rm.amount[ID]

		if pok {
			rm.trace(logger.Error, fmt.Sprintf("Overriding base production for \"%s\" (%d to %d)", ID, ep, prod))
			override = true
		}
		if sok {
			rm.trace(logger.Error, fmt.Sprintf("Overriding base storage for \"%s\" (%d to %d)", ID, es, storage))
			override = true
		}
		if aok {
			rm.trace(logger.Error, fmt.Sprintf("Overriding base amount for \"%s\" (%d to %d)", ID, ea, amount))
			override = true
		}

		rm.prod[ID] = prod
		rm.storage[ID] = storage
		rm.amount[ID] = amount

		err := rm.registerAssociation(ID, name)
		if err != nil {
			rm.trace(logger.Error, fmt.Sprintf("Cannot register resource \"%s\" (id: \"%s\") (err: %v)", name, ID, err))
			inconsistent = true
		}
	}

	if override || inconsistent {
		return ErrInconsistentDB
	}

	return nil
}

// GetResourceFromID :
// Used to retrieve information on the resource that
// corresponds to the input identifier. If no resource
// with this ID exists an error is returned.
//
// The `id` defines the identifier of the resource to
// fetch.
//
// Returns the description for this resource along with
// any errors.
func (rm *ResourcesModule) GetResourceFromID(id string) (ResourceDesc, error) {
	// Find this element in the association table.
	if rm.existsID(id) {
		rm.trace(logger.Error, fmt.Sprintf("Cannot retrieve desc for resource \"%s\"", id))
		return ResourceDesc{}, ErrNotFound
	}

	// We assume at this point that the identifier (and
	// thus the name) both exists so we discard errors.
	name, _ := rm.getNameFromID(id)

	// If the key does not exist the zero value will be
	// assigned to the left operands which is okay (and
	// even desired).
	res := ResourceDesc{
		ID:          id,
		Name:        name,
		BaseProd:    rm.prod[id],
		BaseStorage: rm.storage[id],
		BaseAmount:  rm.amount[id],
	}

	return res, nil
}

// GetResourceFromName :
// Calls internally the `GetResourceFromID` in order
// to forward the call to the above method. Failures
// happen in similar cases.
//
// The `name` defines the name of the resource for
// which a description should be provided.
//
// Returns the description for this resource along
// with any errors.
func (rm *ResourcesModule) GetResourceFromName(name string) (ResourceDesc, error) {
	// Find this element in the association table.
	id, err := rm.getIDFromName(name)
	if err != nil {
		rm.trace(logger.Error, fmt.Sprintf("Cannot retrieve desc for resource \"%s\" (err: %v)", name, err))
		return ResourceDesc{}, ErrNotFound
	}

	return rm.GetResourceFromID(id)
}

// Resources :
// Used to retrieve the resources matching the input
// filters from the data model. Note that if the DB
// has not yet been polled to retrieve data, we will
// return an error.
// The process will consist in first fetching all the
// IDs of the resources matching the filters, and then
// build the rest of the data from the already fetched
// values.
//
// The `proxy` defines the DB to use to fetch the res
// description.
//
// The `filters` represent the list of filters to apply
// to the data fecthing. This will select only part of
// all the available resources.
//
// Returns the list of resources matching the filters
// along with any error.
func (rm *ResourcesModule) Resources(proxy db.Proxy, filters []db.Filter) ([]ResourceDesc, error) {
	// Initialize the module if for some reasons it is still
	// not valid.
	if !rm.valid() {
		err := rm.Init(proxy, true)
		if err != nil {
			return []ResourceDesc{}, err
		}
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table:   "resources",
		Filters: filters,
	}

	IDs, err := rm.fetchIDs(query, proxy)
	if err != nil {
		rm.trace(logger.Error, fmt.Sprintf("Unable to fetch resources (err: %v)", err))
		return []ResourceDesc{}, err
	}

	// Now build the data from the fetched identifiers.
	descs := make([]ResourceDesc, 0)
	for _, ID := range IDs {
		name, err := rm.getNameFromID(ID)
		if err != nil {
			rm.trace(logger.Error, fmt.Sprintf("Unable to fetch resource \"%s\" (err: %v)", ID, err))
			continue
		}

		desc := ResourceDesc{
			ID:   ID,
			Name: name,
		}

		prod, ok := rm.prod[ID]
		if !ok {
			rm.trace(logger.Error, fmt.Sprintf("Unable to fetch base production for resource \"%s\"", ID))
			continue
		} else {
			desc.BaseProd = prod
		}

		storage, ok := rm.storage[ID]
		if !ok {
			rm.trace(logger.Error, fmt.Sprintf("Unable to fetch base storage for resource \"%s\"", ID))
			continue
		} else {
			desc.BaseStorage = storage
		}

		amount, ok := rm.amount[ID]
		if !ok {
			rm.trace(logger.Error, fmt.Sprintf("Unable to fetch base amount for resource \"%s\"", ID))
			continue
		} else {
			desc.BaseAmount = amount
		}

		descs = append(descs, desc)
	}

	return descs, nil
}
