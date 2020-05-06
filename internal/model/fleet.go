package model

import (
	"database/sql"
	"fmt"
	"oglike_server/internal/locker"
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
// The `Planet` attribute defines the potential identifier
// of the planet this fleet is targetting. This value is
// meant to be used in case a planet already exist at the
// location indicated by the `target` coordinates. It is
// left empty in case nothing exists there (typically in
// the case of a colonization mission).
//
// The `ArrivalTime` describes the time at which the fleet
// is meant to reach its destination without taking into
// account the potential delays.
//
// The `Comps` defines the list of components defining the
// fleet. Each component correspond to some ships starting
// from a single location and travelling as a single unit.
//
// The `mode` defines whether the locker on the planet's
// resources should be kept as long as this object exist
// or only during the acquisition of the resources.
//
// The `locker` defines the object to use to prevent a
// concurrent process to access to the resources of the
// planet.
type Fleet struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Universe    string     `json:"universe"`
	Objective   string     `json:"objective"`
	Target      Coordinate `json:"target"`
	Planet      string     `json:"planet,omitempty"`
	ArrivalTime time.Time  `json:"arrival_time"`
	Comps       Components `json:"components"`
	mode        accessMode
	locker      *locker.Lock
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

// newFleetFromDB :
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
func newFleetFromDB(ID string, data Instance, mode accessMode) (Fleet, error) {
	// Create the fleet.
	f := Fleet{
		ID:   ID,
		mode: mode,
	}

	// Consistency.
	if f.ID == "" {
		return f, ErrInvalidFleet
	}

	// Acquire the lock on the planet from the DB.
	var err error
	f.locker, err = data.Locker.Acquire(f.ID)
	if err != nil {
		return f, err
	}
	f.locker.Lock()

	defer func() {
		// Release the locker if needed.
		if f.mode == ReadOnly {
			err = f.locker.Unlock()
		}
	}()

	// Fetch the fleet's content.
	err = f.fetchGeneralInfo(data)
	if err != nil {
		return f, err
	}

	err = f.fetchComponents(data)
	if err != nil {
		return f, err
	}

	return f, nil
}

// NewReadOnlyFleet :
// Used to wrap the call arounf `newFleetFromDB`
// to create a fleet with a read only mode. This
// will indicate that the data will not be used
// to modify anything on this fleet and thus the
// resource can be locked for minimum overhead.
//
// The `ID` defines the identifier of the fleet
// to fetch from the DB.
//
// The `data` defines a way to access to the DB.
//
// Returns the fleet fetched from the DB along
// with any errors.
func NewReadOnlyFleet(ID string, data Instance) (Fleet, error) {
	return newFleetFromDB(ID, data, ReadOnly)
}

// NewReadWriteFleet :
// Similar to the `NewReadOnlyFleet` but tells
// that the fleet will be used to modify it and
// thus the locker is kept after exiting this
// function.
// The user should call the `Close` method when
// done with the modifications so as to allow
// other processes to access the fleet.
//
// The `ID` defines the identifier of the fleet
// to fetch from the DB.
//
// The `data` defines a way to access to the DB.
//
// Returns the fleet fetched from the DB along
// with any errors.
func NewReadWriteFleet(ID string, data Instance) (Fleet, error) {
	return newFleetFromDB(ID, data, ReadWrite)
}

// NewEmptyReadWriteFleet :
// Used to perform the creation of a minimalistic
// fleet from the specified identifier. Nothing
// else is set apart from the locker which is
// acquired in read write mode.
//
// The `ID` defines the identifier of the fleet.
//
// The `data` allows to register the locker for
// this fleet.
//
// Returns the created fleet along with any errors.
func NewEmptyReadWriteFleet(ID string, data Instance) (Fleet, error) {
	// Consistency.
	if !validUUID(ID) {
		return Fleet{}, ErrInvalidFleet
	}

	// Create the fleet.
	f := Fleet{
		ID:   ID,
		mode: ReadWrite,
	}

	// Try to acquire the lock for this fleet.
	var err error
	f.locker, err = data.Locker.Acquire(f.ID)
	if err != nil {
		return f, err
	}
	f.locker.Lock()

	return f, nil
}

// Close :
// Implementation of the `Closer` interface allowing
// to release the lock this fleet may still detain
// on the DB resources.
func (f *Fleet) Close() error {
	// Only release the locker in case the access mode
	// indicates so.
	var err error

	if f.mode == ReadWrite && f.locker != nil {
		err = f.locker.Unlock()
	}

	return err
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
func (f *Fleet) fetchGeneralInfo(data Instance) error {
	// Consistency.
	if f.ID == "" {
		return ErrInvalidFleet
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"name",
			"uni",
			"objective",
			"target_galaxy",
			"target_solar_system",
			"target_position",
			"planet",
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
		&f.ArrivalTime,
	)

	f.Target = NewCoordinate(g, s, p)
	if pl.Valid {
		f.Planet = pl.String
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
func (f *Fleet) fetchComponents(data Instance) error {
	// Consistency.
	if f.ID == "" {
		return ErrInvalidFleet
	}

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
		Planet      string    `json:"planet,omitempty"`
		ArrivalTime time.Time `json:"arrival_time"`
	}{
		ID:          f.ID,
		Name:        f.Name,
		Universe:    f.Universe,
		Objective:   f.Objective,
		Galaxy:      f.Target.Galaxy,
		System:      f.Target.System,
		Position:    f.Target.Position,
		Planet:      f.Planet,
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
func (f *Fleet) Validate(data Instance) error {
	// Consistency.
	if f.ID == "" {
		return ErrInvalidFleet
	}

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
			if obj.canBePerformedBy(ship.ID) {
				return nil
			}
		}
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
func (f *Fleet) simulate(p *Planet, data Instance) error {
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
		return fmt.Errorf("Not implemented")
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
func (f *Fleet) persistToDB(data Instance) error {
	// TODO: Handle this.
	return fmt.Errorf("Not implemented")
}
