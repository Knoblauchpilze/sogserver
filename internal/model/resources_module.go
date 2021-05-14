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
type ResourcesModule struct {
	associationTable
	baseModule

	// The `characteristics` defines the properties of the
	// resources such as their base production, storage or
	// whether or not it can be accumulated.
	characteristics map[string]resProps
}

// ResourceDesc :
// Defines the abstract representation of a resource which
// is bascially an identifier and the actual name of the
// resource plus some base properties of the resource.
type ResourceDesc struct {
	// The `ID` defines the identifier of the resource in the
	// table gathering all the in-game resources. This is used
	// in most of the other operations referencing resources.
	ID string `json:"id"`

	// The `Name` defines the human-readable name of the res
	// as displayed to the user.
	Name string `json:"name"`

	// The `BaseProd` defines the production without modifiers
	// for this resource on each planet. It represents a way
	// for the user to get resources without building anything
	// else.
	BaseProd int `json:"base_production"`

	// The `BaseStorage` defines the base capacity to store
	// the resource without any modifiers (usually hangars).
	BaseStorage int `json:"base_storage"`

	// The `BaseAmount` defines the base amount for this res
	// that can be found on any new planet in the game.
	BaseAmount int `json:"base_amount"`

	// The `Movable` defines whether this resources can be
	// plundered by an attacking fleet or transported by an
	// allied fleet to another planet.
	Movable bool `json:"movable"`

	// The `Storable` defines whether the resources can be
	// stored on a planet or not.
	Storable bool `json:"storable"`

	// The `Dispersable` defines whether this resource can
	// be scattered in a debris field after a fight.
	Dispersable bool `json:"dispersable"`

	// The `Scalable` defines whether or not the values
	// that use this resource will be affected by the
	// coefficient of universe (like economy speed).
	Scalable bool `json:"scalable"`
}

// resProps :
// Used internally to save properties of the resources.
// It describes most of the properties of the resource
// as defined in the `ResourceDesc` without the ID and
// name.
type resProps struct {
	// The `prod` defines the base production for this res.
	prod int

	// The `storage` defines the available storage for the
	// resource by default.
	storage int

	// The `amount` defines the base amount of this res in
	// any new planet.
	amount int

	// The `movable` defines whether this resource can be
	// moved from a planet to another.
	movable bool

	// The `storable` defines whether or not this res is
	// allowed to be stored. Resources that do not have
	// this property will not be accumulated by the game
	// on a planet. This can represent resources where
	// the production is actually used at each time step
	// like energy for example.
	storable bool

	// The `dispersable` prop defines whether this res
	// can be dispersed in a debris field after a fight
	// for example.
	dispersable bool

	// The `scalable` defines whether or not the values
	// that use this resource will be affected by the
	// coefficient of universe (like economy speed).
	scalable bool
}

// ResourceAmount :
// Holds a convenience structure allowing to talk of
// a resource and the associated amount.
//
// The `Resource` defines the ID of the resource that
// is consumed.
//
// The `Amount` defines how much of the resource is
// to be provided.
type ResourceAmount struct {
	Resource string  `json:"resource"`
	Amount   float32 `json:"amount"`
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
		associationTable: newAssociationTable(),
		baseModule:       newBaseModule(log, "resources"),
		characteristics:  nil,
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
	return rm.associationTable.valid() && len(rm.characteristics) > 0
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
	rm.characteristics = make(map[string]resProps)

	// Perform the DB query through a dedicated DB proxy.
	query := db.QueryDesc{
		Props: []string{
			"id",
			"name",
			"base_production",
			"base_storage",
			"base_amount",
			"movable",
			"storable",
			"dispersable",
			"economy_scalable",
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
		rm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize resources module (err: %v)", rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var ID, name string
	var props resProps

	override := false
	inconsistent := false

	for rows.Next() {
		err = rows.Scan(
			&ID,
			&name,
			&props.prod,
			&props.storage,
			&props.amount,
			&props.movable,
			&props.storable,
			&props.dispersable,
			&props.scalable,
		)

		if err != nil {
			rm.trace(logger.Error, fmt.Sprintf("Failed to initialize resource from row (err: %v)", err))
			inconsistent = true

			continue
		}

		// Check whether a resource with this identifier exists.
		if rm.existsID(ID) {
			rm.trace(logger.Error, fmt.Sprintf("Prevented override of resource \"%s\"", ID))
			override = true

			continue
		}

		// Register this resource in the association table.
		err = rm.registerAssociation(ID, name)
		if err != nil {
			rm.trace(logger.Error, fmt.Sprintf("Cannot register resource \"%s\" (id: \"%s\") (err: %v)", name, ID, err))
			inconsistent = true

			continue
		}

		rm.characteristics[ID] = props
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
	if !rm.existsID(id) {
		rm.trace(logger.Error, fmt.Sprintf("Cannot retrieve desc for resource \"%s\"", id))
		return ResourceDesc{}, ErrNotFound
	}

	// We assume at this point that the identifier (and
	// thus the name) both exists so we discard errors.
	name, _ := rm.getNameFromID(id)

	res := ResourceDesc{
		ID:   id,
		Name: name,
	}

	props, ok := rm.characteristics[id]
	if !ok {
		return res, ErrInvalidID
	}
	res.BaseProd = props.prod
	res.BaseStorage = props.storage
	res.BaseAmount = props.amount
	res.Movable = props.movable
	res.Storable = props.storable
	res.Dispersable = props.dispersable
	res.Scalable = props.scalable

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
	id, err := rm.GetIDFromName(name)
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

		props, ok := rm.characteristics[ID]
		if !ok {
			rm.trace(logger.Error, fmt.Sprintf("Unable to fetch characteristics for resource \"%s\"", ID))
			continue
		} else {
			desc.BaseProd = props.prod
			desc.BaseStorage = props.storage
			desc.BaseAmount = props.amount
			desc.Movable = props.movable
			desc.Storable = props.storable
			desc.Dispersable = props.dispersable
			desc.Scalable = props.scalable
		}

		descs = append(descs, desc)
	}

	return descs, nil
}

// Description :
// Used to build a string describing the input set
// of resources with a syntax similar to the below
// example:
// `X res_name_1, Y res_name_2 and Z res_name_3`.
//
// The `resources` defines the list of resources
// for which a description string should be built.
//
// Returns the output string along with any error.
func (rm *ResourcesModule) Description(resources []ResourceAmount) (string, error) {
	out := ""

	if len(out) == 0 {
		return "no resources", nil
	}

	count := len(resources)

	for id, res := range resources {
		r, err := rm.GetResourceFromID(res.Resource)
		if err != nil {
			return "no resources", err
		}

		if out != "" {
			if id == count-1 {
				out += " and "
			} else {
				out += ", "
			}
		}

		out += fmt.Sprintf("%d %s", int(res.Amount), r.Name)
	}

	return out, nil
}
