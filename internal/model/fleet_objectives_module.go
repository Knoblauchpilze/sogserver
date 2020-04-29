package model

import (
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
type FleetObjectivesModule struct {
	associationTable
	baseModule
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
type Objective struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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
	}
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
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
			"name",
		},
		Table:   "fleet_objectives",
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
		)

		if err != nil {
			fom.trace(logger.Error, fmt.Sprintf("Failed to initialize objective from row (err: %v)", err))
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
		}
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
		ID:   id,
		Name: name,
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
	id, err := fom.getIDFromName(name)
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
// The `proxy` defines the DB to use to fetch the res
// description.
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
		Table:   "fleet_objectives",
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

		descs = append(descs, desc)
	}

	return descs, nil
}
