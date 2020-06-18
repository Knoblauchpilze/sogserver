package model

import (
	"encoding/json"
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// FleetObjectivesModule :
// This structure allows to manipulate the data related to
// the fleet objectives defined in the game. Such objectives
// are used to indicate which purpose fleets are serving.
// Each objective is retrieved from the DB and is hardly
// subject to modifications during the course of the game.
//
// The `hostile` defines whether fleet objectives are set
// to be hostile actions or not.
//
// The `directed` defines which of the objectives need a
// target planet.
//
// The `allowedShips` defines the ships that can be used
// for each fleet objective. Indeed some objectices are
// needing some specific capacities and at least a ship
// able to peform the action should be attached to the
// fleet so that it can perform its duty. The other ships
// can be seen as some sort of escort.
type FleetObjectivesModule struct {
	associationTable
	baseModule

	hostile  map[string]bool
	directed map[string]bool

	allowedShips map[string]map[string]bool
}

// Objective :
// Defines a fleet objective in the game. This type
// of element describes the purpose that a fleet is
// serving.
//
// The `ID` defines its identifier in the DB. It is
// used in most other table to actually reference a
// fleet objective.
//
// The `Name` defines the display name for the fleet
// objective. It already gives some summary of what
// is intended when a fleet has this purpose.
//
// The `Hostile` defines whether or not a fleet with
// this objective can be directed towards the planet
// of a single player.
//
// The `Directed` defines whether a planet should be
// associated to a fleet having this objective. If
// set to `false` it indicates that the fleet does
// not need to be directed to a planet.
//
// The `Allowed` defines the list of identifiers of
// ships that are allowed to perform this objective.
// It helps defining whether a fleet can have the
// defined objective given the ships that compose
// it.
type Objective struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Hostile  bool            `json:"hostile"`
	Directed bool            `json:"directed"`
	Allowed  map[string]bool `json:"-"`
}

// MarshalJSON :
// Implementation of the `Marshaler` interface to allow
// only specific information to be marshalled when the
// objective needs to be exported. This will mostly be
// used to convert the list of allowed ships from a map
// to a simple array.
//
// Returns the marshalled bytes for this objective along
// with any error.
func (o *Objective) MarshalJSON() ([]byte, error) {
	type outObjective struct {
		ID       string   `json:"id"`
		Name     string   `json:"name"`
		Hostile  bool     `json:"hostile"`
		Directed bool     `json:"directed"`
		Allowed  []string `json:"allowed_ships"`
	}

	// Copy the planet's data.
	oo := outObjective{
		ID:       o.ID,
		Name:     o.Name,
		Hostile:  o.Hostile,
		Directed: o.Directed,
		Allowed:  make([]string, 0),
	}

	// Make shallow copy of the allowed ships.
	for ship, allowed := range o.Allowed {
		// Append this ship if it is allowed to perform the mission.
		if allowed {
			oo.Allowed = append(oo.Allowed, ship)
		}
	}

	return json.Marshal(oo)
}

// CanBePerformedBy :
// Allows to determine whether the ship described by
// the input identifier can be used to perform this
// fleet objective.
// In case no rule can be found for this ship, we
// will conservatively assume that the ship cannot
// be used.
//
// The `ship` defines the identifier of the ship to
// be checked.
//
// Returns `true` if the ship can be used to serve
// this objective.
func (o *Objective) CanBePerformedBy(ship string) bool {
	// In case the `ship` does not exist in the table
	// for this objective we will get a default value
	// of `false`: this suits our conservative view
	// which will indicate that the ship is not usable
	// for this objective.
	return o.Allowed[ship]
}

// NewFleetObjectivesModule :
// Used to create a new fleet objectives module which is
// initialized with no content (as no DB is provided yet).
// The module will stay invalid until the `init` method
// is called with a valid DB.
//
// The `log` defines the logging layer to forward to the
// base `baseModule` element.
func NewFleetObjectivesModule(log logger.Logger) *FleetObjectivesModule {
	return &FleetObjectivesModule{
		associationTable: newAssociationTable(),
		baseModule:       newBaseModule(log, "fleets"),
		hostile:          nil,
		directed:         nil,
		allowedShips:     nil,
	}
}

// valid :
// Refinement of the base `associationTable` valid method
// in order to perform some checks on the hostile and the
// directionality of the objectives in this module.
//
// Returns `true` if the association table is valid and
// the internal resources as well.
func (fom *FleetObjectivesModule) valid() bool {
	return fom.associationTable.valid() && len(fom.hostile) > 0 && len(fom.directed) > 0 && len(fom.allowedShips) > 0
}

// Init :
// Implementation of the `DBModule` interface to allow
// fetching information from the input DB and load to
// local memory.
//
// The `proxy` represents the main data source to use
// to initialize the fleet objectives data.
//
// The `force` allows to erase any existing information
// and reload everything from the DB.
//
// Returns any error.
func (fom *FleetObjectivesModule) Init(proxy db.Proxy, force bool) error {
	// Prevent reload if not needed.
	if fom.valid() && !force {
		return nil
	}

	// Initialize internal values.
	fom.hostile = make(map[string]bool)
	fom.directed = make(map[string]bool)
	fom.allowedShips = make(map[string]map[string]bool)

	// Initialize general information for objectives.
	err := fom.fetchObjectives(proxy)
	if err != nil {
		fom.trace(logger.Error, fmt.Sprintf("Could not initialize module (err: %v)", err))
		return err
	}

	// Initialize allowed ships for each objective.
	err = fom.fetchAllowedShips(proxy)
	if err != nil {
		fom.trace(logger.Error, fmt.Sprintf("Could not initialize module (err: %v)", err))
		return err
	}

	return nil
}

// fetchObjectives :
// Used internally to fetch the fleet objectives from
// the DB. The input proxy will be used to access the
// information and populate internal tables.
//
// The `proxy` defines a convenient way to access to
// the DB.
//
// Returns any error.
func (fom *FleetObjectivesModule) fetchObjectives(proxy db.Proxy) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
			"name",
			"hostile",
			"directed",
		},
		Table:   "fleets_objectives",
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		fom.trace(logger.Error, fmt.Sprintf("Unable to initialize objectives (err: %v)", err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		fom.trace(logger.Error, fmt.Sprintf("Invalid query to initialize objectives (err: %v)", rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var obj Objective

	override := false
	inconsistent := false

	for rows.Next() {
		err := rows.Scan(
			&obj.ID,
			&obj.Name,
			&obj.Hostile,
			&obj.Directed,
		)

		if err != nil {
			fom.trace(logger.Error, fmt.Sprintf("Failed to initialize objective from row (err: %v)", err))
			inconsistent = true

			continue
		}

		// Check whether a objective with this identifier exists.
		if fom.existsID(obj.ID) {
			fom.trace(logger.Error, fmt.Sprintf("Prevented override of objective \"%s\"", obj.ID))
			override = true

			continue
		}

		// Register this objective in the association table.
		err = fom.registerAssociation(obj.ID, obj.Name)
		if err != nil {
			fom.trace(logger.Error, fmt.Sprintf("Cannot register objective \"%s\" (id: \"%s\") (err: %v)", obj.Name, obj.ID, err))
			inconsistent = true

			continue
		}

		// Register the directed and hostile status.
		eh, hok := fom.hostile[obj.ID]
		ed, dok := fom.directed[obj.ID]

		if hok {
			fom.trace(logger.Error, fmt.Sprintf("Overriding hostile flag for \"%s\" (%t to %t)", obj.ID, eh, obj.Hostile))
			override = true
		}
		if dok {
			fom.trace(logger.Error, fmt.Sprintf("Overriding directed flag for \"%s\" (%t to %t)", obj.ID, ed, obj.Directed))
			override = true
		}

		fom.hostile[obj.ID] = obj.Hostile
		fom.directed[obj.ID] = obj.Directed
	}

	if override || inconsistent {
		return ErrInconsistentDB
	}

	return nil
}

// fetchAllowedShips :
// Used internally to fetch the list of ships that
// can be used to perform a specific fleet objective.
// This will allow to make sure that at least a ship
// in a fleet can be used to perform the mission.
//
// The `proxy` defines a convenient way to access to
// the DB.
//
// Returns any error.
func (fom *FleetObjectivesModule) fetchAllowedShips(proxy db.Proxy) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"ship",
			"objective",
			"usable",
		},
		Table:   "ships_usage",
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		fom.trace(logger.Error, fmt.Sprintf("Unable to initialize allowed ships (err: %v)", err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		fom.trace(logger.Error, fmt.Sprintf("Invalid query to initialize allowed ships (err: %v)", rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var ship, obj string
	var usable bool

	override := false
	inconsistent := false

	for rows.Next() {
		err := rows.Scan(
			&ship,
			&obj,
			&usable,
		)

		if err != nil {
			fom.trace(logger.Error, fmt.Sprintf("Failed to initialize objective from row (err: %v)", err))
			inconsistent = true

			continue
		}

		// Check whether a objective with this identifier exists.
		if !fom.existsID(obj) {
			fom.trace(logger.Error, fmt.Sprintf("Cannot interpret allowed ship with invalid objective \"%s\"", obj))
			inconsistent = true

			continue
		}

		allowedFor, ok := fom.allowedShips[obj]
		if !ok {
			allowedFor = make(map[string]bool)
		}

		e, ok := allowedFor[ship]
		if ok {
			fom.trace(logger.Error, fmt.Sprintf("Overriding allowed status ship \"%s\" for objective \"%s\" (%t to %t)", ship, obj, e, usable))
			override = true
		}

		allowedFor[ship] = usable
		fom.allowedShips[obj] = allowedFor
	}

	if override || inconsistent {
		return ErrInconsistentDB
	}

	return nil
}

// GetObjectiveFromID :
// Used to retrieve information on the fleet objective
// that corresponds to the input identifier. If no such
// objective exists an error is returned.
//
// The `id` defines the identifier of the fleet objective
// to fetch.
//
// Returns the description for this objective along with
// any errors.
func (fom *FleetObjectivesModule) GetObjectiveFromID(id string) (Objective, error) {
	// Find this element in the association table.
	if !fom.existsID(id) {
		fom.trace(logger.Error, fmt.Sprintf("Cannot retrieve desc for objective \"%s\"", id))
		return Objective{}, ErrNotFound
	}

	// We assume at this point that the identifier (and
	// thus the name) both exists so we discard errors.
	name, _ := fom.getNameFromID(id)

	// If the key does not exist the zero value will be
	// assigned to the left operands which is okay (and
	// even desired).
	res := Objective{
		ID:       id,
		Name:     name,
		Hostile:  fom.hostile[id],
		Directed: fom.directed[id],
		Allowed:  fom.allowedShips[id],
	}

	return res, nil
}

// GetObjectiveFromName :
// Calls internally the `GetObjectiveFromID` in order
// to forward the call to the above method. Failures
// happen in similar cases.
//
// The `name` defines the name of the objective for
// which a description should be provided.
//
// Returns the description for this objective along
// with any errors.
func (fom *FleetObjectivesModule) GetObjectiveFromName(name string) (Objective, error) {
	// Find this element in the association table.
	id, err := fom.GetIDFromName(name)
	if err != nil {
		fom.trace(logger.Error, fmt.Sprintf("Cannot retrieve desc for objective \"%s\" (err: %v)", name, err))
		return Objective{}, ErrNotFound
	}

	return fom.GetObjectiveFromID(id)
}

// Objectives :
// Used to retrieve the objectives matching the input
// filters from the data model. Note that if the DB
// has not yet been polled to retrieve data, we will
// return an error.
//
// The `proxy` defines the DB to use to fetch fleets
// objectives description.
//
// The `filters` represent the list of filters to apply
// to the data fecthing. This will select only part of
// all the available objectives.
//
// Returns the list of objectives matching the filters
// along with any error.
func (fom *FleetObjectivesModule) Objectives(proxy db.Proxy, filters []db.Filter) ([]Objective, error) {
	// Initialize the module if for some reasons it is still
	// not valid.
	if !fom.valid() {
		err := fom.Init(proxy, true)
		if err != nil {
			return []Objective{}, err
		}
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table:   "fleets_objectives",
		Filters: filters,
	}

	IDs, err := fom.fetchIDs(query, proxy)
	if err != nil {
		fom.trace(logger.Error, fmt.Sprintf("Unable to fetch objectives (err: %v)", err))
		return []Objective{}, err
	}

	// Now build the data from the fetched identifiers.
	descs := make([]Objective, 0)
	for _, ID := range IDs {
		name, err := fom.getNameFromID(ID)
		if err != nil {
			fom.trace(logger.Error, fmt.Sprintf("Unable to fetch objective \"%s\" (err: %v)", ID, err))
			continue
		}

		desc := Objective{
			ID:   ID,
			Name: name,
		}

		h, ok := fom.hostile[ID]
		if !ok {
			fom.trace(logger.Error, fmt.Sprintf("Unable to fetch hostile status for objective \"%s\"", ID))
			continue
		} else {
			desc.Hostile = h
		}

		d, ok := fom.directed[ID]
		if !ok {
			fom.trace(logger.Error, fmt.Sprintf("Unable to fetch directed status for objective \"%s\"", ID))
			continue
		} else {
			desc.Directed = d
		}

		allowed, ok := fom.allowedShips[ID]
		if ok {
			desc.Allowed = allowed
		} else {
			desc.Allowed = make(map[string]bool)
		}

		descs = append(descs, desc)
	}

	return descs, nil
}
