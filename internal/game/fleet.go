package game

import (
	"database/sql"
	"fmt"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"time"
)

// Fleet :
// Defines a fleet in the OG context. A fleet is composed of
// several ships, that can be grouped in distinct components.
// Several players can join a fleet meaning that the starting
// point of the fleet cannot be expressed as a single coord.
// However a fleet always has a single destination which is
// reached by all the components at the same time.
//
// The `ID` represents a way to uniquely identify the fleet.
//
// The `Name` defines the name that the user provided when
// the fleet was created. It might be empty in case no name
// was provided.
//
// The `Universe` defines the identifier of the universe this
// fleet belongs to. Indeed a fleet is linked to some coords
// which are linked to a universe. It also is used to make
// sure that only players of this universe can participate in
// the fleet.
//
// The `Objective` is a string defining the action intended
// for this fleet. It is a way to determine which purpose
// the fleet serves. This string represents an identifier
// in the objectives description table.
//
// The `Target` defines the destination coordinates of the
// fleet. Note that depending on the objective of the fleet
// it might not always refer to an existing planet.
//
// The `Body` attribute defines the potential identifier
// of the planet or moon this fleet is targetting. This
// value is meant to be used in case a celestial body is
// already existing at the location indicated by `target`
// coordinates. It is left empty in case nothing exists
// there (typically in the case of a colonization mission).
//
// The `ArrivalTime` describes the time at which the fleet
// is meant to reach its destination without taking into
// account the potential delays.
//
// The `Comps` defines the list of components defining the
// fleet. Each component correspond to some ships starting
// from a single location and travelling as a single unit.
type Fleet struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Universe    string     `json:"universe"`
	Objective   string     `json:"objective"`
	Target      Coordinate `json:"target_coordinates"`
	Body        string     `json:"target,omitempty"`
	ArrivalTime time.Time  `json:"arrival_time"`
	Comps       Components `json:"components"`
}

// ErrInvalidFleet :
// Used to indicate that the fleet provided in input is
// not valid.
var ErrInvalidFleet = fmt.Errorf("Invalid fleet with no identifier")

// ErrDuplicatedFleet :
// Used to indicate that the fleet's identifier provided
// input is not unique in the DB.
var ErrDuplicatedFleet = fmt.Errorf("Invalid not unique fleet")

// ErrNoShipToPerformObjective :
// Indicates that none of the ships taking part in the
// fleet are able to perform the fleet's objective.
var ErrNoShipToPerformObjective = fmt.Errorf("No ships can perform the fleet's objective")

// ErrInvalidTargetForObjective :
// Used to indicate that the objective is not valid
// compared to the target of a fleet.
var ErrInvalidTargetForObjective = fmt.Errorf("Target cannot be used for fleet's objective")

// purpose :
// Convenience define to refer to the purpose of a fleet
// which mimics the objectives of a fleet.
type purpose string

// Possible values of a fleet's purpose.
const (
	deployment   purpose = "deployment"
	transport    purpose = "transport"
	colonization purpose = "colonization"
	expedition   purpose = "expedition"
	acsDefend    purpose = "ACS defend"
	acsAttack    purpose = "ACS attack"
	harvesting   purpose = "harvesting"
	attacking    purpose = "attacking"
	espionage    purpose = "espionage"
	destroy      purpose = "destroy"
)

// NewFleetFromDB :
// Used to fetch the content of the fleet from
// the input DB and populate all internal fields
// from it. In case the DB cannot be fetched or
// some errors are encoutered, the return value
// will include a description of the error.
//
// The `ID` defines the ID of the fleet to get.
// It is fetched from the DB and should refer
// to an existing fleet.
//
// The `data` allows to actually perform the DB
// requests to fetch the fleet's data.
//
// Returns the fleet as fetched from the DB along
// with any errors.
func NewFleetFromDB(ID string, data model.Instance) (Fleet, error) {
	// Create the fleet.
	f := Fleet{
		ID: ID,
	}

	// Consistency.
	if !validUUID(f.ID) {
		return f, ErrInvalidFleet
	}

	// Fetch the fleet's content.
	err := f.fetchGeneralInfo(data)
	if err != nil {
		return f, err
	}

	err = f.fetchComponents(data)
	if err != nil {
		return f, err
	}

	return f, nil
}

// NewEmptyFleet :
// Used to perform the creation of a fleet with
// minimalistic properties from the specified ID.
// Nothing else is set apart from the ID of the
// fleet.
//
// The `ID` defines the identifier of the fleet.
//
// The `data` allows to register the locker for
// this fleet.
//
// Returns the created fleet along with any errors.
func NewEmptyFleet(ID string, data model.Instance) (Fleet, error) {
	// Consistency.
	if !validUUID(ID) {
		return Fleet{}, ErrInvalidFleet
	}

	// Create the fleet.
	f := Fleet{
		ID: ID,
	}

	return f, nil
}

// Valid :
// Defines whether this fleet is valid given the bounds
// for the coordinates that are admissible in the uni
// this fleet evolves in.
//
// Returns `true` if the fleet is valid.
func (f *Fleet) Valid(uni Universe) bool {
	return validUUID(f.ID) &&
		validUUID(f.Universe) &&
		validUUID(f.Objective) &&
		f.Target.valid(uni.GalaxiesCount, uni.GalaxySize, uni.SolarSystemSize) &&
		f.Comps.valid(f.Objective, f.Target)
}

// String :
// Implementation of the `Stringer` interface to make
// sure displaying this fleet is easy.
//
// Returns the corresponding string.
func (f Fleet) String() string {
	return fmt.Sprintf("[id: %s, universe: %s, target: %s]", f.ID, f.Universe, f.Target)
}

// fetchGeneralInfo :
// Used internally when building a fleet from the
// DB to retrieve general information such as the
// objective and target of the fleet.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (f *Fleet) fetchGeneralInfo(data model.Instance) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"name",
			"uni",
			"objective",
			"target_galaxy",
			"target_solar_system",
			"target_position",
			"target",
			"target_type",
			"arrival_time",
		},
		Table: "fleets",
		Filters: []db.Filter{
			{
				Key:    "id",
				Values: []string{f.ID},
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

	// Scan the fleet's data.
	var g, s, p int
	var loc Location

	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return ErrInvalidFleet
	}

	// Note that we have to query the `planet` in a nullable
	// string in order to account for cases where the string
	// is not filled (typically for undirected objectives).
	var pl sql.NullString

	err = dbRes.Scan(
		&f.Name,
		&f.Universe,
		&f.Objective,
		&g,
		&s,
		&p,
		&pl,
		&loc,
		&f.ArrivalTime,
	)

	var errC error
	f.Target, errC = newCoordinate(g, s, p, loc)
	if errC != nil {
		return errC
	}

	if pl.Valid {
		f.Body = pl.String
	}

	// Make sure that it's the only fleet.
	if dbRes.Next() {
		return ErrDuplicatedFleet
	}

	return err
}

// fetchComponents :
// Used to fetch data related to a fleet: this is
// meant to represent all the individual components
// of the fleest meaning the various waves of ships
// that have been created by the players that joined
// the fleet.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (f *Fleet) fetchComponents(data model.Instance) error {
	f.Comps = make([]Component, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table: "fleet_elements",
		Filters: []db.Filter{
			{
				Key:    "fleet",
				Values: []string{f.ID},
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

	// Extract each component for this fleet.
	var ID string
	IDs := make([]string, 0)

	for dbRes.Next() {
		err = dbRes.Scan(&ID)

		if err != nil {
			return err
		}

		IDs = append(IDs, ID)
	}

	f.Comps = make([]Component, 0)

	for _, ID = range IDs {
		comp, err := newComponentFromDB(ID, data)

		if err != nil {
			return err
		}

		f.Comps = append(f.Comps, comp)
	}

	return nil
}

// Convert :
// Implementation of the `db.Convertible` interface
// from the DB package in order to only include fields
// that need to be marshalled in the fleet's creation.
//
// Returns the converted version of the planet which
// only includes relevant fields.
func (f *Fleet) Convert() interface{} {
	return struct {
		ID          string    `json:"id"`
		Name        string    `json:"name"`
		Universe    string    `json:"uni"`
		Objective   string    `json:"objective"`
		Galaxy      int       `json:"target_galaxy"`
		System      int       `json:"target_solar_system"`
		Position    int       `json:"target_position"`
		Target      string    `json:"target,omitempty"`
		TargetType  Location  `json:"target_type"`
		ArrivalTime time.Time `json:"arrival_time"`
	}{
		ID:          f.ID,
		Name:        f.Name,
		Universe:    f.Universe,
		Objective:   f.Objective,
		Galaxy:      f.Target.Galaxy,
		System:      f.Target.System,
		Position:    f.Target.Position,
		Target:      f.Body,
		TargetType:  f.Target.Type,
		ArrivalTime: f.ArrivalTime,
	}
}

// Validate :
// Used to make sure that the data contained in this
// fleet is valid. It will mostly consist in checking
// whether at least one ship contained in the comps
// that are associated to it are allowed to perform
// the mission's objective.
//
// The `data` allows to access data from the DB.
//
// Returns an error in case the fleet is not valid
// and `nil` otherwise (indicating that no obvious
// errors were detected).
func (f *Fleet) Validate(data model.Instance) error {
	// Retrieve this fleet's objective's description.
	obj, err := data.Objectives.GetObjectiveFromID(f.Objective)
	if err != nil {
		return err
	}

	// We will return as early as possible because
	// for now there's nothing else to check. Make
	// sure to change the loop if it ever changes.
	for _, comp := range f.Comps {
		for _, ship := range comp.Ships {
			if obj.CanBePerformedBy(ship.ID) {
				return nil
			}
		}
	}

	// Make sure that the location of the target
	// is consistent with the objective.
	if purpose(obj.Name) == harvesting && f.Target.Type != Debris {
		return ErrInvalidTargetForObjective
	}
	if purpose(obj.Name) != harvesting && f.Target.Type == Debris {
		return ErrInvalidTargetForObjective
	}
	if purpose(obj.Name) == destroy && f.Target.Type != Moon {
		return ErrInvalidTargetForObjective
	}

	return ErrNoShipToPerformObjective
}

// simulate :
// Used to perform the simulation of this fleet on
// the input planet. This will simulate the fight
// or any effect that this fleet might have on the
// planet.
//
// The `p` represents the planet this fleet is
// directed to. Providing an invalid planet will
// make the simulation fail.
//
// The `data` allows to access the data from the
// DB if needed.
//
// Returns any error.
func (f *Fleet) simulate(p *Planet, data model.Instance) error {
	// We need to check the objective of this fleet
	// and perform the adequate processing.
	obj, err := data.Objectives.GetObjectiveFromID(f.Objective)
	if err != nil {
		return err
	}

	var script string

	// TODO: Handle missing cases.
	switch obj.Name {
	case "deployment":
		script = "fleet_deployment"
	case "transport":
		script = "fleet_transport"
	case "colonization":
		return fmt.Errorf("Not implemented")
	case "expedition":
		return fmt.Errorf("Not implemented")
	case "ACSdefend'":
		return fmt.Errorf("Not implemented")
	case "ACSattack'":
		return fmt.Errorf("Not implemented")
	case "harvesting":
		script = "fleet_harvesting"
	case "attacking":
		return fmt.Errorf("Not implemented")
	case "espionage":
		return fmt.Errorf("Not implemented")
	case "destroy":
		return fmt.Errorf("Not implemented")
	}

	// Execute the script allowing to perform
	// the objective of the fleet.
	query := db.InsertReq{
		Script: script,
		Args: []interface{}{
			f.ID,
		},
	}

	err = data.Proxy.InsertToDB(query)

	return err
}

// persistToDB :
// Used to persist the content of this fleet to
// the DB. Most of the process will be done by
// calling the dedicated script with the valid
// data.
//
// The `data` allows to access to the DB to be
// able to save the fleet's data.
//
// Returns any error.
func (f *Fleet) persistToDB(data model.Instance) error {
	// TODO: Handle this.
	return fmt.Errorf("Not implemented")
}
