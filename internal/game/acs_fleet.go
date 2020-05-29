package game

import (
	"fmt"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"time"

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
//
// The `arrivalTime` is computed from all the
// fleets already assigned to the ACS and is
// the estimated arrival time of all of them
// at the destination.
type ACSFleet struct {
	ID          string   `json:"id"`
	Universe    string   `json:"universe"`
	Objective   string   `json:"objective"`
	Target      string   `json:"target"`
	TargetType  Location `json:"source_type"`
	Fleets      []string `json:"components"`
	arrivalTime time.Time
}

// ErrACSOperationMismatch : Indicates that the fleet is not added to the correct ACS operation.
var ErrACSOperationMismatch = fmt.Errorf("Mismatch in fleet's ACS compared to actual ACS operation")

// ErrACSUniverseMismacth : Indicates that the universe of the ACS and the fleet mismatch.
var ErrACSUniverseMismacth = fmt.Errorf("Mismatch in fleet universe compared to ACS operation")

// ErrACSObjectiveMismacth : Indicates that the objective of the ACS and the fleet mismatch.
var ErrACSObjectiveMismacth = fmt.Errorf("Mismatch in fleet objective compared to ACS operation")

// ErrACSTargetMismacth : Indicates that the target of the ACS and the fleet mismatch.
var ErrACSTargetMismacth = fmt.Errorf("Mismatch in fleet target compared to ACS operation")

// ErrACSTargetTypeMismacth : Indicates that the target type of the ACS and the fleet mismatch.
var ErrACSTargetTypeMismacth = fmt.Errorf("Mismatch in fleet target type compared to ACS operation")

// ErrACSFleetDelayedTooMuch : Indicates that the fleet would delay the ACS by too much time.
var ErrACSFleetDelayedTooMuch = fmt.Errorf("Fleet would delay ACS operation too much")

// Valid :
// Determines whether the fleet is valid. By valid we
// only mean obvious syntax errors.
//
// Returns any error or `nil` if the fleet seems valid.
func (acs *ACSFleet) Valid() error {
	if !validUUID(acs.ID) {
		return ErrInvalidElementID
	}
	if !validUUID(acs.Universe) {
		return ErrInvalidUniverseForFleet
	}
	if !validUUID(acs.Objective) {
		return ErrInvalidObjectiveForFleet
	}
	if !validUUID(acs.Target) {
		return ErrInvalidTargetForFleet
	}
	if !existsLocation(acs.TargetType) {
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
	acs := ACSFleet{
		ID: ID,
	}

	// Consistency.
	if !validUUID(acs.ID) {
		return acs, ErrInvalidElementID
	}

	// Fetch the ACS fleet's content.
	err := acs.fetchGeneralInfo(data)
	if err != nil {
		return acs, err
	}

	err = acs.fetchFleets(data)
	if err != nil {
		return acs, err
	}

	err = acs.fetchArrivalTime(data)
	if err != nil {
		return acs, err
	}

	return acs, nil
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
	acs := ACSFleet{
		ID:         uuid.New().String(),
		Universe:   fleet.Universe,
		Objective:  fleet.Objective,
		Target:     fleet.Target,
		TargetType: fleet.TargetCoords.Type,
		Fleets:     make([]string, 0),
	}

	// Make sure that the input fleet is linked to this
	// ACS now.
	fleet.ACS = acs.ID

	return acs
}

// fetchGeneralInfo :
// Used internally when building an ACS fleet from
// the DB to retrieve general information such as
// the objective and target of the fleet.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (acs *ACSFleet) fetchGeneralInfo(data Instance) error {
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
				Values: []interface{}{acs.ID},
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
		&acs.Universe,
		&acs.Objective,
		&acs.Target,
		&acs.TargetType,
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
func (acs *ACSFleet) fetchFleets(data Instance) error {
	acs.Fleets = make([]string, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"fleet",
		},
		Table: "fleets_acs_components",
		Filters: []db.Filter{
			{
				Key:    "acs",
				Values: []interface{}{acs.ID},
			},
		},
		Ordering: "order by joined_at",
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

		acs.Fleets = append(acs.Fleets, fleet)
	}

	return nil
}

// fetchArrivalTime :
// Similar to `fetchGeneralInfo` but allows to
// fetch the information relative to the arrival
// time of the ACS.
//
// The `data` allows to access to the DB.
//
// Return any errors.
func (acs *ACSFleet) fetchArrivalTime(data Instance) error {
	// In order to consolidate the arrival time from
	// the registered components. We assume that the
	// info in the DB is consistent and we can just
	// fetch it from the first fleet (as all others
	// should be the same).
	query := db.QueryDesc{
		Props: []string{
			"f.arrival_time",
		},
		Table: "fleets f inner join fleets_acs_components fac on f.id = fac.fleet",
		Filters: []db.Filter{
			{
				Key:    "fac.acs",
				Values: []interface{}{acs.ID},
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

	// Fetch the arrival time: we should have at
	// least a component registered in the ACS.
	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return ErrElementNotFound
	}

	err = dbRes.Scan(
		&acs.arrivalTime,
	)

	return err
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
func (acs *ACSFleet) SaveToDB(fleet *Fleet, proxy db.Proxy) error {
	// Convert the cargo to a marshallable slice.
	resources := make([]model.ResourceAmount, 0)
	for _, res := range fleet.Cargo {
		resources = append(resources, res)
	}

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_acs_fleet",
		Args: []interface{}{
			acs.ID,
			fleet,
			fleet.Ships.convert(),
			resources,
			fleet.Consumption,
		},
	}

	err := proxy.InsertToDB(query)

	// Analyze the error in order to provide some
	// comprehensive message.
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
		case "universe":
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
// The `source` defines the source planet for the
// fleet: it is used in case the flight time for
// the new fleet should be updated.
//
// The `data` allows to access to the DB.
//
// Returns any error.
func (acs *ACSFleet) ValidateFleet(fleet *Fleet, source *Planet, data Instance) error {
	// Make sure that the common properties for the
	// fleet are consistent.
	if fleet.ACS != acs.ID {
		return ErrACSOperationMismatch
	}
	if fleet.Universe != acs.Universe {
		return ErrACSUniverseMismacth
	}
	if fleet.Objective != acs.Objective {
		return ErrACSObjectiveMismacth
	}
	if fleet.Target != acs.Target {
		return ErrACSTargetMismacth
	}
	if fleet.TargetCoords.Type != acs.TargetType {
		return ErrACSTargetTypeMismacth
	}

	// In case there's no elements yet in the ACS
	// the component is now declared valid.
	if len(acs.Fleets) == 0 {
		return nil
	}

	// Compare the arrival time of the fleet with
	// the currently estimated arrival time: when
	// we add this component the time difference
	// should not delay the arrival time by more
	// than 30%.
	now := time.Now()

	timeToArrival := acs.arrivalTime.Sub(now)
	newTimeToArrival := fleet.ArrivalTime.Sub(now)

	deltaT := float32(newTimeToArrival) / float32(timeToArrival)

	if deltaT > 1.3 {
		return ErrACSFleetDelayedTooMuch
	}

	// The fleet does not delay the fleet too much.
	// We still have too cases to handle: either
	// the new fleet *does* delay the fleet: in this
	// case nothing is left to do, the script used
	// to perform the insertion of the fleet will
	// handle the modification of the arrival time
	// of the other fleets adequately.
	if deltaT >= 1.0 {
		return nil
	}

	// On the other hand if the fleet is actually
	// faster than the actual arrival time we need
	// to update the consumption and flight time
	// to match the current arrival time as closely
	// as possible.
	fleet.Speed *= deltaT

	err := fleet.ConsolidateArrivalTime(data, source)
	if err != nil {
		return err
	}

	// We shouldn't need to revalidate the data as
	// we will reduce the speed of the fleet and
	// thus burn less fuel in all likelihood.
	err = fleet.consolidateConsumption(data, source)
	if err != nil {
		return err
	}

	// Due to some numerical inaccuracies (maybe in
	// the way we handle the flight time) we can be
	// modifying slightly the actual arrival time.
	// To prevent that we will force afterwards the
	// arrival time to be precisely what it was. It
	// is not a big issue as the error we noted was
	// quite small (but possibly in the second-ish
	// region).
	d := -fleet.ArrivalTime.Sub(acs.arrivalTime)
	fleet.ArrivalTime = acs.arrivalTime
	fleet.CreatedAt = fleet.CreatedAt.Add(d)

	return nil
}

// simulate :
// Used to perform the execution of this ACS
// fleet on its target.
//
// The `p` describes the element that will be
// attacked. It can either be a planet or a
// moon.
//
// The `data` allows to access to the DB.
//
// Returns any error.
func (acs *ACSFleet) simulate(p *Planet, data Instance) error {
	// We first need to fetch all the fleets that
	// belong to this ACS.
	fleets := make([]Fleet, 0)
	cargo := float32(0.0)

	for _, f := range acs.Fleets {
		fleet, err := NewFleetFromDB(f, data)
		if err != nil {
			return ErrFleetFightSimulationFailure
		}

		cargo += fleet.usedCargoSpace()

		fleets = append(fleets, fleet)
	}

	// Create the attacker structure from the fleets.
	// We know that the fleets are ordered by their
	// desired joining time so we can just traverse
	// the slice from the beginning to the end.
	a := attacker{
		units:     make([]shipsUnit, 0),
		usedCargo: cargo,
	}

	for _, f := range fleets {
		att, err := f.toAttacker(data)
		if err != nil {
			return ErrFleetFightSimulationFailure
		}

		a.units = append(a.units, att.units...)
	}

	// Create the defender from the planet.
	d, err := p.toDefender(data)
	if err != nil {
		return ErrFleetFightSimulationFailure
	}

	result, err := d.defend(&a)
	if err != nil {
		return ErrFleetFightSimulationFailure
	}

	// Handle the pillage of resources if the outcome
	// says so. Note that the outcome is expressed in
	// the defender's point of view.
	pillage := make([]model.ResourceAmount, 0)

	if result.outcome == Loss {
		pillage, err = a.pillage(p, data)
		if err != nil {
			return ErrFleetFightSimulationFailure
		}
	}

	// TODO: Split pillage equally and handle the save
	// script (maybe use the `fleet_fight_aftermath` to
	// save each fleet).
	fmt.Println(fmt.Sprintf("Pillage is %v", pillage))

	// Create the query and execute it.
	return fmt.Errorf("Not implemented")
}
