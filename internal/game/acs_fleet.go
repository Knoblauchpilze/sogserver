package game

import (
	"fmt"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"

	"github.com/google/uuid"
)

// ACSFleet :
// Describe an ACS fleet with all its components.
// Compared to a regular fleet such a fleet can
// be joined by several players who can all add
// some ships from different planet to the fleet.
// All the ships will arrive at the same time to
// the destination and be considered as a single
// larger fleet.
//
// The `ID` defines the identifier for this ACS
// fleet.
//
// The `Universe` defines the identifier of the
// universe to which this fleet belongs.
//
// The `Objective` represents the consolidated
// objective of all the components of the fleet.
// For now only the `ACSAttack` case is handled.
//
// The `Target` defines the identifier of the
// element targeted by this fleet: can either
// be a planet or a moon on the specified uni.
//
// The `target_type` defines the type of the
// target for this ACS fleet. It helps finding
// out where the `target` should be fetched.
//
// The `Fleets` define the identifiers of the
// individual fleets that joined this ACS.
type ACSFleet struct {
	ID         string   `json:"id"`
	Universe   string   `json:"universe"`
	Objective  string   `json:"objective"`
	Target     string   `json:"target"`
	TargetType Location `json:"source_type"`
	Fleets     []string `json:"components"`
}

// Valid :
// Determines whether the fleet is valid. By valid we
// only mean obvious syntax errors.
//
// Returns any error or `nil` if the fleet seems valid.
func (f *ACSFleet) Valid() error {
	if !validUUID(f.ID) {
		return ErrInvalidElementID
	}
	if !validUUID(f.Universe) {
		return ErrInvalidUniverseForFleet
	}
	if !validUUID(f.Objective) {
		return ErrInvalidObjectiveForFleet
	}
	if !validUUID(f.Target) {
		return ErrInvalidTargetForFleet
	}
	if !existsLocation(f.TargetType) {
		return ErrInvalidTargetTypeForFleet
	}

	return nil
}

// NewACSFleetFromDB :
// Used to retrieve the information related to the
// ACS fleet described by the input `ID`. In case
// no such fleet can be found an error is raised.
//
// The `ID` defines the identifier of the ACS fleet
// to fetch from the DB.
//
// The `data` provides a way to access to the DB.
//
// Returns the built ACS fleet and any error.
func NewACSFleetFromDB(ID string, data Instance) (ACSFleet, error) {
	// Create the fleet.
	f := ACSFleet{
		ID: ID,
	}

	// Consistency.
	if !validUUID(f.ID) {
		return f, ErrInvalidElementID
	}

	// Fetch the ACS fleet's content.
	err := f.fetchGeneralInfo(data)
	if err != nil {
		return f, err
	}

	err = f.fetchFleets(data)
	if err != nil {
		return f, err
	}

	return f, nil
}

// NewACSFleet :
// Perform the creation of a new ACS fleet from the
// input fleet. We assume that the input fleet will
// be the first component for the ACS so most of the
// fields will be equalized from the input data.
//
// The `fleet` defines the first (and for now unique)
// component of the ACS operation.
//
// Return the created ACS operation.
func NewACSFleet(fleet *Fleet) ACSFleet {
	return ACSFleet{
		ID:         uuid.New().String(),
		Universe:   fleet.Universe,
		Target:     fleet.Target,
		TargetType: fleet.TargetCoords.Type,
	}
}

// fetchGeneralInfo :
// Used internally when building an ACS fleet from
// the DB to retrieve general information such as
// the objective and target of the fleet.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (f *ACSFleet) fetchGeneralInfo(data Instance) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"universe",
			"objective",
			"target",
			"target_type",
		},
		Table: "fleets_acs",
		Filters: []db.Filter{
			{
				Key:    "id",
				Values: []interface{}{f.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}
	if dbRes.Err != nil {
		return dbRes.Err
	}

	// Scan the ACS fleet's data.
	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return ErrElementNotFound
	}

	err = dbRes.Scan(
		&f.Universe,
		&f.Objective,
		&f.Target,
		&f.TargetType,
	)

	// Make sure that it's the only ACS fleet.
	if dbRes.Next() {
		return ErrDuplicatedElement
	}

	return err
}

// fetchFleets :
// Similar to `fetchGeneralInfo` but allows to
// fetch the individual fleet components that
// have joined the ACS.
//
// The `data` allows to access to the DB.
//
// Returns any error.
func (f *ACSFleet) fetchFleets(data Instance) error {
	f.Fleets = make([]string, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"fleet",
		},
		Table: "fleets_acs_components",
		Filters: []db.Filter{
			{
				Key:    "acs",
				Values: []interface{}{f.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}
	if dbRes.Err != nil {
		return dbRes.Err
	}

	// Populate the return value.
	var fleet string

	for dbRes.Next() {
		err = dbRes.Scan(
			&fleet,
		)

		if err != nil {
			return err
		}

		f.Fleets = append(f.Fleets, fleet)
	}

	return nil
}

// SaveToDB :
// Used to save the content of the fleet provided
// in argument as a component of this ACS fleet.
// It is very similar to saving a fleet with some
// different script that handled the additional
// operations to perform.
//
// The `proxy` allows to access to the DB.
//
// Returns any error.
func (f *ACSFleet) SaveToDB(fleet *Fleet, proxy db.Proxy) error {
	// Convert the cargo to a marshallable slice.
	resources := make([]model.ResourceAmount, 0)
	for _, res := range fleet.Cargo {
		resources = append(resources, res)
	}

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_acs_fleet",
		Args: []interface{}{
			fleet,
			fleet.Ships,
			resources,
			fleet.Consumption,
		},
	}

	err := proxy.InsertToDB(query)

	// Analyze the error in order to provide some
	// comprehensive message.
	// TODO: Probably refine even more the errors ?
	dbe, ok := err.(db.Error)
	if !ok {
		return err
	}

	dee, ok := dbe.Err.(db.DuplicatedElementError)
	if ok {
		switch dee.Constraint {
		case "fleets_pkey":
			return ErrDuplicatedElement
		}

		return dee
	}

	fkve, ok := dbe.Err.(db.ForeignKeyViolationError)
	if ok {
		switch fkve.ForeignKey {
		case "uni":
			return ErrNonExistingUniverse
		case "objective":
			return ErrNonExistingObjective
		case "player":
			return ErrNonExistingPlayer
		}

		return fkve
	}

	return dbe
}

// ValidateFleet :
// Used to perform the validation of the ACS fleet
// and verify that it is valid. This method is used
// to make sure that the arrival time of a new comp
// is valid compared to the existing data.
// No information is persisted to the DB yet, only
// verified against existing elements.
//
// The `fleet` represents the component to add to
// the ACS fleet.
//
// The `data` allows to access to the DB.
//
// Returns any error.
func (f *ACSFleet) ValidateFleet(fleet *Fleet, data Instance) error {
	// TODO: Implement validation of ACS fleet.
	return fmt.Errorf("Not implemented")
}

// simulate :
// Used to perform the execution of this ACS
// fleet on its target.
//
// The `data` allows to access to the DB.
//
// Returns any error.
func (f *ACSFleet) simulate(data Instance) error {
	// TODO: Implement the simulation of the ACS fleet.
	return fmt.Errorf("Not implemented")
}
