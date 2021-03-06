package game

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"time"
)

// Fleet :
// Defines a fleet in the OG context. A fleet is composed of
// several ships and several fleets can be grouped into a so
// called ACS operation (for Alliance Combat System).
// A fleet is characterized by a source location which has
// to be a valid planet/moon and a target which can either
// be another planet or moon or an unassigned location for
// some objectives.
type Fleet struct {
	// The `ID` represents a way to uniquely identify the fleet.
	ID string `json:"id"`

	// The `Universe` defines the ID of the universe this fleet
	// belongs to. This allows to check the coordinates of the
	// source and target to be sure that they're valid.
	Universe string `json:"universe"`

	// The `Objective` defines the identifier of the objective
	// this fleet is serving. It is checked to make sure it is
	// consistent with the rest of the data (typically if no
	// target is provided, the objective should allow it).
	Objective string `json:"objective"`

	// The `Player` defines the identifier of the player to
	// which this fleet is related to. It should also be the
	// owner of the source of the fleet.
	Player string `json:"player"`

	// The `Source` defines the identifier of the fleet's
	// source. It should correspond to a valid planet or
	// moon within the specified universe.
	Source string `json:"source"`

	// The `SourceType` defines the type of element that is
	// represented by the source. It is used to distinguish
	// between a planet and a moon as starting coordinates.
	SourceType Location `json:"source_type"`

	// The `TargetCoords` defines the coordinates of the
	// target destination of the fleet. It can either be
	// the only indication of the purpose of the fleet in
	// case the objective allows a fleet not directed to
	// a target or the coordinates of the planet/moon to
	// which this fleet is directed to.
	TargetCoords Coordinate `json:"target_coordinates"`

	// The `Target` defines the identifier of the planet
	// or moon to whicht his fleet is directed. Note that
	// this value may be empty if the objective allows it.
	Target string `json:"target"`

	// The `ACS` defines the identifier of the ACS into
	// which this fleet is included. This value might be
	// empty in case the fleet is not related to any ACS.
	ACS string `json:"acs"`

	// The `Speed` defines the speed percentage that is
	// used by the fleet to travel at. It will be used
	// when computing the flight time and the consumption.
	// This value should be in the range `[0; 1]`.
	Speed float32 `json:"speed"`

	// The `CreatedAt` defines the time at which the
	// fleet was created. It is the launch time.
	CreatedAt time.Time `json:"created_at"`

	// The `ArrivalTime` represents the time at which
	// the fleet should arrive at its destination. This
	// value is computed in the server and any data
	// provided when registering the fleet is overriden.
	ArrivalTime time.Time `json:"arrival_time"`

	// The `DeploymentTime` defines the duration in seconds
	// of the deployment phase of this fleet. This might be
	// `0` in case the fleet does not deploy but for example
	// in the case of expeditions or ACS defend, the fleet
	// will be deployed for a certain amount of time at its
	// destination. This duration is taken into account when
	// the `ReturnTime` is computed.
	DeploymentTime int `json:"deployment_time"`

	// The `ReturnTime` defines the time at which the
	// fleet will be back to its starting location in
	// case the fleet proceeds to its destination. This
	// value may be updated if the user calls back the
	// fleet beforehand, or ignored altogether in case
	// the fleet is no longer able to perform its duty
	// (like a completely destroyed fleet during a fight
	// or a deployment mission).
	ReturnTime time.Time `json:"return_time"`

	// The `Ships` defines the ships that are part of
	// the fleet along with the amount of each one of
	// them included in the fleet.
	Ships ShipsInFleet `json:"ships"`

	// The `Consumption` defines a slice containing all
	// the fuel needed for this fleet. It contains the
	// list of resources that need to be existing on the
	// launch body to be able to create the fleet. This
	// value is computed internally.
	Consumption []model.ResourceAmount `json:"-"`

	// The `Cargo` defines the requested resources to
	// be transported by the fleet. It will be checked
	// against the available resources on the source
	// body.
	Cargo map[string]model.ResourceAmount `json:"cargo"`

	// The `flightTime` represents the duration of the
	// flight from the source destination to the target
	// in seconds. This should be the interval between
	// the `CreatedAt` and `ArrivalTime` but also from
	// `ArrivalTime` to `ReturnTime`.
	flightTime time.Duration

	// The `returning` allows to determine whether the
	// fleet is returning to its source or not.
	returning bool

	// The `deployed` allows to determine whether the
	// fleet is currently deployed to another planet.
	deployed bool
}

// ShipInFleet :
// Defines a single ship involved in a fleet. This
// is the building blocks of fleets: it defines the
// ID of the ship and the number of ships of this
// type that are included in a fleet.
// All the ships belong to a single player and are
// launched from a single planet.
type ShipInFleet struct {
	// The `ID` defines the identifier of the ship that
	// is involved in the fleet.
	ID string `json:"ship"`

	// The `Count` defines how many ships of this type
	// are involved.
	Count int `json:"count"`
}

// ShipsInFleet :
// Convenience define to refer to a list of ships
// belonging to a fleet component. Allows to define
// some methods on this type to ease the consistency
// checks.
type ShipsInFleet map[string]ShipInFleet

// valid :
// Determines whether the ship is valid. By valid we
// only mean obvious syntax errors.
//
// Returns any error or `nil` if the ship seems valid.
func (sif ShipInFleet) valid() error {
	if !validUUID(sif.ID) {
		return ErrInvalidElementID
	}
	if sif.Count <= 0 {
		return ErrInvalidShipCount
	}

	return nil
}

// valid :
// Used to perform a chain validation on all the ships
// sets defined in the slice.
//
// Returns `nil` if all individual components are
// valid.
func (sifs ShipsInFleet) valid() error {
	for _, sif := range sifs {
		if err := sif.valid(); err != nil {
			return err
		}
	}

	if len(sifs) == 0 {
		return ErrNoShipsInFleet
	}

	return nil
}

// convert :
// Used to convert the map to a simple slice
// of ships in order to use in the DB scripts
// for example.
//
// Returns the equivalent slice of ships.
func (sifs ShipsInFleet) convert() []ShipInFleet {
	ships := make([]ShipInFleet, 0)

	for _, s := range sifs {
		ships = append(ships, s)
	}

	return ships
}

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

// ErrInvalidShipCount : Indicates that an invalid number of ships is requested.
var ErrInvalidShipCount = fmt.Errorf("invalid number of ships requested for fleet")

// ErrNoShipsInFleet : Indicates that no ships are associated to a fleet.
var ErrNoShipsInFleet = fmt.Errorf("no ships associated to fleet")

// ErrInvalidUniverseForFleet : Indicates that no valid universe is provided for a fleet.
var ErrInvalidUniverseForFleet = fmt.Errorf("no valid universe for fleet")

// ErrInvalidObjectiveForFleet : Indicates that the objective provided for a fleet is not valid.
var ErrInvalidObjectiveForFleet = fmt.Errorf("no valid objective for a fleet")

// ErrInvalidPlayerForFleet : Indicates that the player provided for a fleet is not valid.
var ErrInvalidPlayerForFleet = fmt.Errorf("no valid player for fleet")

// ErrInvalidSourceForFleet : Indicates that the source of a fleet is not valid.
var ErrInvalidSourceForFleet = fmt.Errorf("source for fleet is not valid")

// ErrInvalidSourceTypeForFleet : Indicates that the source type for a fleet is not valid.
var ErrInvalidSourceTypeForFleet = fmt.Errorf("source type for fleet is not valid")

// ErrInvalidTargetForFleet : Indicates that the target of a fleet is not valid.
var ErrInvalidTargetForFleet = fmt.Errorf("target for fleet is not valid")

// ErrInvalidTargetTypeForFleet : Indicates that the target type for a fleet is not valid.
var ErrInvalidTargetTypeForFleet = fmt.Errorf("target type for fleet is not valid")

// ErrInvalidSpeedForFleet : Indicates that the speed provided for a fleet is not valid.
var ErrInvalidSpeedForFleet = fmt.Errorf("speed for fleet is not valid")

// ErrInvalidCargoForFleet : Indicates that a cargo resource is invalid for a fleet.
var ErrInvalidCargoForFleet = fmt.Errorf("invalid cargo value for fleet")

// ErrFleetDirectedTowardsSource : Indicates that the source is identical to the target of a fleet.
var ErrFleetDirectedTowardsSource = fmt.Errorf("target is identical to source for fleet")

// ErrFleetInMultipleACS : Indicates that the fleet is linked to multiple ACS.
var ErrFleetInMultipleACS = fmt.Errorf("fleet is linked to multiple ACS")

// ErrNonExistingObjective : Indicates that the objective does not exist for this fleet.
var ErrNonExistingObjective = fmt.Errorf("inexisting fleet objective")

// ErrNoShipToPerformObjective : Indicates that no ship can be used to perform the fleet's objective.
var ErrNoShipToPerformObjective = fmt.Errorf("no ships can perform the fleet's objective")

// ErrInvalidTargetForObjective : Indicates that the target is not consistent with the fleet's objective.
var ErrInvalidTargetForObjective = fmt.Errorf("target cannot be used for fleet's objective")

// ErrInvalidCargoForObjective : Indicates that the cargo specified for the fleet is not
// consistent with the fleet's objective.
var ErrInvalidCargoForObjective = fmt.Errorf("cargo is not consistent with fleet's objective")

// ErrCargoNotMovable : Indicates that one of the resource defined in the cargo is not movable.
var ErrCargoNotMovable = fmt.Errorf("resource cannot be moved by a fleet")

// ErrInsufficientCargoForFleet : Indicates that the fleet has insufficient cargo space.
var ErrInsufficientCargoForFleet = fmt.Errorf("insufficient cargo space to hold resources in fleet")

// ErrFuelNotMovable : Indicates that one of the resource defined in the fuel is not movable.
var ErrFuelNotMovable = fmt.Errorf("resource cannot be used as fuel by a fleet")

// ErrInsufficientTankForFleet : Indicates that the fleet has insufficient fuel tank space.
var ErrInsufficientTankForFleet = fmt.Errorf("insufficient fuel tank space to hold resources in fleet")

// ErrInvalidPropulsionSystem : Indicates that the propulsion system of a ship is not compatible with
// the researched technologies of the starting location of a fleet.
var ErrInvalidPropulsionSystem = fmt.Errorf("unknown propulsion system for ship for a fleet")

// ErrFleetObjectiveNotCorrect : Indicates that the fleet objective associated to the fleet is not
// correct and cannot be performed by the server.
var ErrFleetObjectiveNotCorrect = fmt.Errorf("unknown fleet objective in simulation phase")

// ErrUnableToSimulateFleet : Indicates that an error occurred while simulating the fleet.
var ErrUnableToSimulateFleet = fmt.Errorf("error occurred while simulating fleet")

// ErrMultipliersError : Indicates that an error occurred while fetching the multipliers for the
// parent universe of the fleet.
var ErrMultipliersError = fmt.Errorf("error occurred while fetching universe's multipliers")

// Valid :
// Determines whether the fleet is valid. By valid we only
// mean obvious syntax errors.
//
// Returns any error or `nil` if the fleet seems valid.
func (f *Fleet) Valid(uni Universe) error {
	if !validUUID(f.ID) {
		return ErrInvalidElementID
	}
	if !validUUID(f.Universe) {
		return ErrInvalidUniverseForFleet
	}
	if !validUUID(f.Objective) {
		return ErrInvalidObjectiveForFleet
	}
	if !validUUID(f.Player) {
		return ErrInvalidPlayerForFleet
	}
	if !validUUID(f.Source) {
		return ErrInvalidSourceForFleet
	}
	if !existsLocation(f.SourceType) {
		return ErrInvalidSourceTypeForFleet
	}
	if !f.TargetCoords.valid(uni.GalaxiesCount, uni.GalaxySize, uni.SolarSystemSize) {
		return ErrInvalidCoordinates
	}
	if f.Target != "" && !validUUID(f.Target) {
		return ErrInvalidTargetForFleet
	}
	if !existsLocation(f.TargetCoords.Type) {
		return ErrInvalidTargetTypeForFleet
	}
	if f.Speed <= 0.0 || f.Speed > 1.0 {
		return ErrInvalidSpeedForFleet
	}
	if err := f.Ships.valid(); err != nil {
		return err
	}
	for _, c := range f.Cargo {
		if !validUUID(c.Resource) || c.Amount <= 0.0 {
			return ErrInvalidCargoForFleet
		}
	}

	if f.Target == f.Source {
		return ErrFleetDirectedTowardsSource
	}

	return nil
}

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
func NewFleetFromDB(ID string, data Instance) (Fleet, error) {
	// Create the fleet.
	f := Fleet{
		ID: ID,
	}

	// Consistency.
	if !validUUID(f.ID) {
		return f, ErrInvalidElementID
	}

	// Fetch the fleet's content.
	err := f.fetchGeneralInfo(data)
	if err != nil {
		return f, err
	}

	err = f.fetchACSInfo(data)
	if err != nil {
		return f, err
	}

	err = f.fetchShips(data)
	if err != nil {
		return f, err
	}

	err = f.fetchCargo(data)
	if err != nil {
		return f, err
	}

	return f, nil
}

// cargoSpace :
// Returns the available cargo space for this
// fleet.
//
// The `data` allows to access to the DB data.
//
// Returns the amount of cargo for this fleet
// along with any error.
func (f *Fleet) cargoSpace(data Instance) (int, error) {
	cargo := 0

	for _, ship := range f.Ships {
		sd, err := data.Ships.GetShipFromID(ship.ID)

		if err != nil {
			return cargo, err
		}

		cargo += (ship.Count * sd.Cargo)
	}

	return cargo, nil
}

// usedCargoSpace :
// Returns the used cargo space by the resources
// carried by the fleet. Note that this is the
// raw value of the cargo without any control to
// determine whether the resources can actually
// be moved or carried by the fleet.
//
// The `data` allows to access to the DB data.
//
// Returns the amount of used  cargo for this fleet
// along with any error.
func (f *Fleet) usedCargoSpace() float32 {
	cargo := float32(0.0)

	for _, c := range f.Cargo {
		cargo += c.Amount
	}

	return cargo
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
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"universe",
			"objective",
			"player",
			"source",
			"source_type",
			"target_galaxy",
			"target_solar_system",
			"target_position",
			"target",
			"target_type",
			"speed",
			"created_at",
			"arrival_time",
			"deployment_time",
			"return_time",
			"is_returning",
			"is_deployed",
		},
		Table: "fleets",
		Filters: []db.Filter{
			{
				Key:    "id",
				Values: []interface{}{f.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)

	// Check for errors.
	if err != nil {
		return err
	}
	defer dbRes.Close()

	if dbRes.Err != nil {
		return dbRes.Err
	}

	// Scan the fleet's data.
	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return ErrElementNotFound
	}

	// Note that we have to query the `target` in a nullable
	// string in order to account for cases where the string
	// is not filled (typically for undirected objectives).
	var ta sql.NullString

	var g, s, p int
	var loc Location

	err = dbRes.Scan(
		&f.Universe,
		&f.Objective,
		&f.Player,
		&f.Source,
		&f.SourceType,
		&g,
		&s,
		&p,
		&ta,
		&loc,
		&f.Speed,
		&f.CreatedAt,
		&f.ArrivalTime,
		&f.DeploymentTime,
		&f.ReturnTime,
		&f.returning,
		&f.deployed,
	)

	var errC error
	f.TargetCoords, errC = newCoordinate(g, s, p, loc)
	if errC != nil {
		return errC
	}

	if ta.Valid {
		f.Target = ta.String
	}

	f.flightTime = f.ArrivalTime.Sub(f.CreatedAt)

	// Make sure that it's the only fleet.
	if dbRes.Next() {
		return ErrDuplicatedElement
	}

	return err
}

// fetchACSInfo :
// Similar to `fetchGeneralInfo` but allows to
// fetch the ACS component for this fleet.
//
// The `data` defines the object to access the
// DB.
//
// Returns any error.
func (f *Fleet) fetchACSInfo(data Instance) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"fac.acs",
		},
		Table: "fleets f left join fleets_acs_components fac on f.id = fac.fleet",
		Filters: []db.Filter{
			{
				Key:    "f.id",
				Values: []interface{}{f.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)

	// Check for errors.
	if err != nil {
		return err
	}
	defer dbRes.Close()

	if dbRes.Err != nil {
		return dbRes.Err
	}

	// Scan the fleet's data.
	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return ErrElementNotFound
	}

	// The element might be empty in case the fleet is
	// not part of any ACS.
	var acs sql.NullString

	err = dbRes.Scan(
		&acs,
	)

	if acs.Valid {
		f.ACS = acs.String
	}

	// Make sure that it's the only ACS for this fleet.
	if dbRes.Next() {
		return ErrFleetInMultipleACS
	}

	return err
}

// fetchShips :
// Similar to `fetchGeneralInfo` but allows to
// fetch the ships associated to the fleet.
//
// The `data` allows to access to the DB.
//
// Returns any error.
func (f *Fleet) fetchShips(data Instance) error {
	f.Ships = make(map[string]ShipInFleet)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"ship",
			"count",
		},
		Table: "fleets_ships",
		Filters: []db.Filter{
			{
				Key:    "fleet",
				Values: []interface{}{f.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)

	// Check for errors.
	if err != nil {
		return err
	}
	defer dbRes.Close()

	if dbRes.Err != nil {
		return dbRes.Err
	}

	// Populate the return value.
	var sif ShipInFleet

	for dbRes.Next() {
		err = dbRes.Scan(
			&sif.ID,
			&sif.Count,
		)

		if err != nil {
			return err
		}

		s, ok := f.Ships[sif.ID]
		if ok {
			s.Count += sif.Count
		} else {
			s = sif
		}

		f.Ships[sif.ID] = s
	}

	return nil
}

// fetchCargo :
// Similar to `fetchGeneralInfo` but allows to
// fetch the cargo associated to the fleet.
//
// The `data` allows to access to the DB.
//
// Returns any error.
func (f *Fleet) fetchCargo(data Instance) error {
	f.Cargo = make(map[string]model.ResourceAmount)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"resource",
			"amount",
		},
		Table: "fleets_resources",
		Filters: []db.Filter{
			{
				Key:    "fleet",
				Values: []interface{}{f.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)

	// Check for errors.
	if err != nil {
		return err
	}
	defer dbRes.Close()

	if dbRes.Err != nil {
		return dbRes.Err
	}

	// Populate the return value.
	var ra model.ResourceAmount

	for dbRes.Next() {
		err = dbRes.Scan(
			&ra.Resource,
			&ra.Amount,
		)

		if err != nil {
			return err
		}

		// We consider that we will add any resource
		// if it already exist.
		res, ok := f.Cargo[ra.Resource]

		if !ok {
			res = ra
		} else {
			res.Amount += ra.Amount
		}

		f.Cargo[ra.Resource] = res
	}

	return nil
}

// SaveToDB :
// Used to save the content of this fleet to
// the DB. In case an error is raised during
// the operation a comprehensive error is
// returned.
//
// The `proxy` allows to access to the DB.
//
// Returns any error.
func (f *Fleet) SaveToDB(proxy db.Proxy) error {
	// Convert the cargo to a marshallable slice.
	resources := make([]model.ResourceAmount, 0)
	for _, res := range f.Cargo {
		resources = append(resources, res)
	}

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_fleet",
		Args: []interface{}{
			f,
			f.Ships.convert(),
			resources,
			f.Consumption,
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

// MarshalJSON :
// Implementation of the `json.Marshaler` interface
// in order to convert some of the internal props to
// an output format easier to manipulate.
//
// Returns the marshalled bytes for this planet along
// with any error.
func (f *Fleet) MarshalJSON() ([]byte, error) {
	type outFleet struct {
		ID             string                 `json:"id"`
		Universe       string                 `json:"universe"`
		Objective      string                 `json:"objective"`
		Player         string                 `json:"player"`
		Source         string                 `json:"source"`
		SourceType     Location               `json:"source_type"`
		TargetCoords   Coordinate             `json:"target_coordinates"`
		Target         string                 `json:"target"`
		ACS            string                 `json:"acs"`
		Speed          float32                `json:"speed"`
		CreatedAt      time.Time              `json:"created_at"`
		ArrivalTime    time.Time              `json:"arrival_time"`
		DeploymentTime int                    `json:"deployment_time"`
		ReturnTime     time.Time              `json:"return_time"`
		Ships          []ShipInFleet          `json:"ships"`
		Cargo          []model.ResourceAmount `json:"cargo,omitempty"`
	}

	// Copy the fleet's data.
	of := outFleet{
		ID:             f.ID,
		Universe:       f.Universe,
		Objective:      f.Objective,
		Player:         f.Player,
		Source:         f.Source,
		SourceType:     f.SourceType,
		TargetCoords:   f.TargetCoords,
		Target:         f.Target,
		ACS:            f.ACS,
		Speed:          f.Speed,
		CreatedAt:      f.CreatedAt,
		ArrivalTime:    f.ArrivalTime,
		DeploymentTime: f.DeploymentTime,
		ReturnTime:     f.ReturnTime,
		Ships:          f.Ships.convert(),
	}

	for _, r := range f.Cargo {
		of.Cargo = append(of.Cargo, r)
	}

	return json.Marshal(of)
}

// UnmarshalJSON :
// Implementation of the `json.Unmarshaler` interface
// in order to allow the passing of the cargo values
// for a fleet through a slice rather than a map. It
// will then be converted for internal use within
// this function.
//
// The `raw` represents the data to unmarshal.
//
// Returns any error.
func (f *Fleet) UnmarshalJSON(raw []byte) error {
	// Define an input structure that will be used to
	// populate the fleet's data.
	type inFleet struct {
		ID             string                 `json:"id"`
		Universe       string                 `json:"universe"`
		Objective      string                 `json:"objective"`
		Player         string                 `json:"player"`
		Source         string                 `json:"source"`
		SourceType     Location               `json:"source_type"`
		TargetCoords   Coordinate             `json:"target_coordinates"`
		Target         string                 `json:"target"`
		ACS            string                 `json:"acs"`
		Speed          float32                `json:"speed"`
		CreatedAt      time.Time              `json:"created_at"`
		ArrivalTime    time.Time              `json:"arrival_time"`
		DeploymentTime int                    `json:"deployment_time"`
		ReturnTime     time.Time              `json:"return_time"`
		Ships          []ShipInFleet          `json:"ships"`
		Cargo          []model.ResourceAmount `json:"cargo"`
	}

	var in inFleet

	err := json.Unmarshal(raw, &in)
	if err != nil {
		return err
	}

	// Equalize field.
	f.ID = in.ID
	f.Universe = in.Universe
	f.Objective = in.Objective
	f.Player = in.Player
	f.Source = in.Source
	f.SourceType = in.SourceType
	f.TargetCoords = in.TargetCoords
	f.Target = in.Target
	f.ACS = in.ACS
	f.Speed = in.Speed
	f.CreatedAt = in.CreatedAt
	f.ArrivalTime = in.ArrivalTime
	f.DeploymentTime = in.DeploymentTime
	f.ReturnTime = in.ReturnTime

	f.Ships = make(map[string]ShipInFleet)
	for _, s := range in.Ships {
		ship, ok := f.Ships[s.ID]

		if !ok {
			ship = s
		} else {
			ship.Count += s.Count
		}

		f.Ships[s.ID] = ship
	}

	f.Cargo = make(map[string]model.ResourceAmount)
	for _, c := range in.Cargo {
		res, ok := f.Cargo[c.Resource]

		if !ok {
			res = c
		} else {
			res.Amount += c.Amount
		}

		f.Cargo[c.Resource] = res
	}

	return nil
}

// Convert :
// Implementation of the `db.Convertible` interface
// from the DB package in order to only include fields
// that need to be marshalled in the fleet's creation.
//
// Returns the converted version of the fleet which
// only includes relevant fields.
func (f *Fleet) Convert() interface{} {
	return struct {
		ID             string    `json:"id"`
		Universe       string    `json:"universe"`
		Objective      string    `json:"objective"`
		Player         string    `json:"player"`
		Source         string    `json:"source"`
		SourceType     Location  `json:"source_type"`
		Galaxy         int       `json:"target_galaxy"`
		System         int       `json:"target_solar_system"`
		Position       int       `json:"target_position"`
		Target         string    `json:"target,omitempty"`
		TargetType     Location  `json:"target_type"`
		Speed          float32   `json:"speed"`
		ArrivalTime    time.Time `json:"arrival_time"`
		DeploymentTime int       `json:"deployment_time"`
		ReturnTime     time.Time `json:"return_time"`
		IsReturning    bool      `json:"is_returning"`
		IsDeployed     bool      `json:"is_deployed"`
	}{
		ID:             f.ID,
		Universe:       f.Universe,
		Objective:      f.Objective,
		Player:         f.Player,
		Source:         f.Source,
		SourceType:     f.SourceType,
		Galaxy:         f.TargetCoords.Galaxy,
		System:         f.TargetCoords.System,
		Position:       f.TargetCoords.Position,
		Target:         f.Target,
		TargetType:     f.TargetCoords.Type,
		Speed:          f.Speed,
		ArrivalTime:    f.ArrivalTime,
		DeploymentTime: f.DeploymentTime,
		ReturnTime:     f.ReturnTime,
		IsReturning:    f.returning,
		IsDeployed:     f.deployed,
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
// The `source` defines the planet attached to this
// fleet: it represents the mandatory location from
// where the fleet is launched.
//
// The `target` defines the potential planet to which
// this fleet is directed. It might be `nil` in case
// the objective allows it (typically in case of a
// colonization operation).
//
// Returns an error in case the fleet is not valid
// and `nil` otherwise (indicating that no obvious
// errors were detected).
func (f *Fleet) Validate(data Instance, source *Planet, target *Planet) error {
	// Fetch the multipliers to apply for this fleet.
	mul, err := NewMultipliersFromDB(f.Universe, data)
	if err != nil {
		return ErrMultipliersError
	}

	// Update consumption.
	err = f.consolidateConsumption(data, source, mul)
	if err != nil {
		return err
	}

	// Retrieve this fleet's objective's description
	// and check that at least a ship is able to be
	// used to perform the objective.
	obj, err := data.Objectives.GetObjectiveFromID(f.Objective)
	if err != nil {
		return err
	}

	objDoable := false
	for id := range f.Ships {
		objDoable = obj.CanBePerformedBy(id)

		if objDoable {
			break
		}
	}

	if !objDoable {
		return ErrNoShipToPerformObjective
	}

	// Update deployment time based on the objective
	// of the fleet (as all objectives do not allow a
	// deployment phase).
	if purpose(obj.Name) != "ACS defend" && purpose(obj.Name) != "expedition" {
		f.DeploymentTime = 0
	}

	// Make sure that the location of the target
	// is consistent with the objective.
	if purpose(obj.Name) == harvesting && f.TargetCoords.Type != Debris {
		return ErrInvalidTargetForObjective
	}
	if purpose(obj.Name) != harvesting && f.TargetCoords.Type == Debris {
		return ErrInvalidTargetForObjective
	}
	if purpose(obj.Name) == destroy && f.TargetCoords.Type != Moon {
		return ErrInvalidTargetForObjective
	}
	if obj.Directed && target == nil {
		return ErrInvalidTargetForObjective
	}
	if !obj.Directed && target != nil {
		return ErrInvalidTargetForObjective
	}
	if obj.Hostile && target == nil {
		return ErrInvalidTargetForObjective
	}
	if obj.Hostile && source.Player == target.Player {
		return ErrInvalidTargetForObjective
	}
	// Prevent ACS defend operation where the fleet
	// carries some resources.
	if purpose(obj.Name) == acsDefend && f.usedCargoSpace() > 0.0 {
		return ErrInvalidCargoForObjective
	}
	// We also prevent attacking fleet to be carrying
	// resources in case they join an ACS operation.
	if f.ACS != "" && f.usedCargoSpace() > 0.0 {
		return ErrInvalidCargoForObjective
	}

	// In the case of a harvesting mission we need
	// to make sure that the debris fields actually
	// exist.
	if purpose(obj.Name) == harvesting {
		_, err = NewDebrisFieldFromCoords(f.TargetCoords, f.Universe, data)

		if err != nil && err != ErrElementNotFound {
			return err
		}

		if err == ErrElementNotFound {
			return ErrInvalidTargetForObjective
		}
	}

	// In the case of a colonization operation,
	// prevent the colonization of the exact
	// location of the source planet.
	if purpose(obj.Name) == colonization {
		if f.TargetCoords == source.Coordinates {
			return ErrInvalidTargetForObjective
		}
	}

	// Make sure that the cargo defined for this fleet
	// component can be stored in the ships.
	totCargo, err := f.cargoSpace(data)
	if err != nil {
		return err
	}

	if f.Cargo == nil {
		f.Cargo = make(map[string]model.ResourceAmount)
	}

	var totNeeded float32
	for _, res := range f.Cargo {
		rDesc, err := data.Resources.GetResourceFromID(res.Resource)
		if err != nil {
			return err
		}
		if !rDesc.Movable {
			return ErrCargoNotMovable
		}

		totNeeded += res.Amount
	}

	if totNeeded > float32(totCargo) {
		return ErrInsufficientCargoForFleet
	}

	// Make sure that the amount of fuel needed for
	// this fleet can be stored in the tanks. We use
	// a separate tank to store the fuel.
	totNeeded = 0
	for _, res := range f.Consumption {
		rDesc, err := data.Resources.GetResourceFromID(res.Resource)
		if err != nil {
			return err
		}
		if !rDesc.Movable {
			return ErrFuelNotMovable
		}

		totNeeded += res.Amount
	}

	if totNeeded > float32(totCargo) {
		return ErrInsufficientTankForFleet
	}

	// Validate the amount of fuel available on the
	// planet compared to the amount required and
	// that there are enough resources to be taken
	// from the planet.
	return source.validateFleet(f.Consumption, f.Cargo, f.Ships, data)
}

// consolidateConsumption :
// Used to perform the consolidation of the consumption
// required for this fleet to take off. It does not
// handle the fact that the parent planet actually has
// the needed fuel but only to compute it.
// The result of the computations will be saved in the
// fleet itself.
//
// The `data` allows to get information from the DB on
// the consumption of ships.
//
// The `p` defines the planet from where this fleet is
// starting the flight.
//
// The `mul` defines a multipliers to apply to various
// quantities based on the parent universe (typically
// in the case of the flight time or consumption ratio).
//
// Returns any error.
func (f *Fleet) consolidateConsumption(data Instance, p *Planet, mul Multipliers) error {
	// Compute the distance between the starting position
	// and the destination of the flight.
	d := float64(p.Coordinates.distanceTo(f.TargetCoords))

	// Compute the raw flight time in seconds. Internally
	// it is expressed in milliseconds and also already
	// updated with the parent universe's fleets speed.
	rawFlightTimeSec := float64(f.flightTime) / (float64(time.Second) * float64(mul.Fleet))

	// Now we can compute the total consumption by summing
	// the individual consumptions of ships.
	consumption := make(map[string]float64)

	for _, ship := range f.Ships {
		sd, err := data.Ships.GetShipFromID(ship.ID)
		if err != nil {
			return err
		}

		// Retrieve the best suited engine for the ship.
		engine := sd.SelectEngine(p.technologies)
		speed := engine.ComputeSpeed(p.technologies)

		for _, fuel := range sd.Consumption {
			// The values and formulas are extracted from here:
			// https://ogame.fandom.com/wiki/Talk:Fuel_Consumption
			sk := 35000.0 * math.Sqrt(d*10.0/float64(speed)) / (rawFlightTimeSec - 10.0)
			cons := float64(fuel.Amount*float32(ship.Count)) * d * math.Pow(1.0+sk/10.0, 2.0) / 35000.0

			ex := consumption[fuel.Resource]
			ex += cons
			consumption[fuel.Resource] = ex
		}
	}

	// Handle the deployment time of the fleet.
	if f.DeploymentTime > 0 {
		// For each ship, add the consumption knowing that the
		// deployment time is expressed in seconds.
		dTimeH := float64(f.DeploymentTime) / 3600.0

		for _, ship := range f.Ships {
			sd, err := data.Ships.GetShipFromID(ship.ID)
			if err != nil {
				return err
			}

			for _, fuel := range sd.Deployment {
				ex := consumption[fuel.Resource]
				ex += float64(fuel.Amount) * dTimeH * float64(ship.Count)
				consumption[fuel.Resource] = ex
			}
		}
	}

	// Save the data in the fleet itself. We will also
	// use the consumption ratio to scale the amount of
	// fuel needed.
	f.Consumption = make([]model.ResourceAmount, 0)

	for res, fuel := range consumption {
		value := model.ResourceAmount{
			Resource: res,
			Amount:   float32(fuel) * mul.Consumption,
		}

		f.Consumption = append(f.Consumption, value)
	}

	return nil
}

// ConsolidateArrivalTime :
// Used to perform the update of the arrival time for
// this fleet based on the technologies of the planet
// it starts from.
// We will assume that the input `p` corresponds to
// the source location of the fleet.
//
// The `data` allows to get information from the DB
// related to the propulsion used by each ships and
// their consumption.
//
// The `p` defines the planet from which this fleet
// should start and will be used to update the values
// of the techs that should be used for computing a
// speed for each ship.
//
// The `ratio` defines a multiplier to apply to the
// computed flight time to compensate for the fact
// that some universes try to make things faster
// through a raw multiplier.
//
// Returns any error.
func (f *Fleet) ConsolidateArrivalTime(data Instance, p *Planet, ratio float32) error {
	// Update the time at which this component joined
	// the fleet.
	f.CreatedAt = time.Now()

	// Compute the time of arrival for this component. It
	// is function of the percentage of the maximum speed
	// used by the ships and the slowest ship's speed.
	d := float64(p.Coordinates.distanceTo(f.TargetCoords))

	// Compute the maximum speed of the fleet. This will
	// correspond to the speed of the slowest ship in the
	// component.
	maxSpeed := math.MaxInt32

	for _, ship := range f.Ships {
		sd, err := data.Ships.GetShipFromID(ship.ID)

		if err != nil {
			return err
		}

		engine := sd.SelectEngine(p.technologies)
		speed := engine.ComputeSpeed(p.technologies)

		if speed < maxSpeed {
			maxSpeed = speed
		}
	}

	// Compute the duration of the flight given the distance.
	// Note that the speed percentage is interpreted as such:
	//  - 100% -> 10
	//  -  50% -> 5
	//  -  10% -> 1
	speedRatio := f.Speed * 10.0
	flightTimeSec := 35000.0/float64(speedRatio)*math.Sqrt(float64(d)*10.0/float64(maxSpeed)) + 10.0

	// Apply the multiplier for the speed of the fleet based
	// on the parent universe's value.
	flightTimeSec *= float64(ratio)

	// Compute the flight time by converting this duration in
	// milliseconds: this will allow to keep more precision.
	f.flightTime = time.Duration(1000.0*flightTimeSec) * time.Millisecond

	// The arrival time is just this duration in the future.
	// We will use the milliseconds in order to keep more
	// precision before rounding.
	f.ArrivalTime = f.CreatedAt.Add(f.flightTime)

	// The return time is separated from the arrival time
	// by an additional full flight time and the duration
	// of the deployment. Note that we will consolidate the
	// deployment time at his step: indeed only some fleet
	// objectives allow for a deployment.
	obj, err := data.Objectives.GetObjectiveFromID(f.Objective)
	if err != nil {
		return ErrNonExistingObjective
	}

	purp := purpose(obj.Name)
	if purp != expedition && purp != acsDefend {
		f.DeploymentTime = 0
	}

	f.ReturnTime = f.ArrivalTime.Add(f.flightTime).Add(time.Duration(f.DeploymentTime) * time.Second)

	// The fleet is not yet retruning from its mission
	// nor already deployed.
	f.returning = false
	f.deployed = false

	return nil
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

	// In order to process the fleet we need to know
	// whether some server side processing should be
	// done and the script to use to update the data
	// for the fleet if any.
	// We will perform the simulation of the fleet
	// first and then use the specified script to
	// finalize the operation. In case the fleet is
	// not needing a simulation phase, we directly
	// run the script.
	var script string

	switch obj.Name {
	case "deployment":
		script = "fleet_deployment"
	case "transport":
		script = "fleet_transport"
	case "harvesting":
		script, err = f.harvest(data)
	case "colonization":
		script, err = f.colonize(data)
	case "attacking":
		script, err = f.attack(p, data)
	case "espionage":
		script, err = f.spy(p, data)
	case "destroy":
		script, err = f.destroy(p, data)
	case "expedition":
		return fmt.Errorf("not implemented")
	case "ACS defend":
		script = "fleet_acs_defend"
	case "ACS attack":
		// The `ACS` are not actually valid objectives
		// for a fleet. It is converted when creating
		// the fleet and the corresponding ACS element.
		// We will handle the case of an unknown fleet
		// objective in a similar fashion.
		fallthrough
	default:
		return ErrFleetObjectiveNotCorrect
	}

	// Make sure that no errors occured while the
	// fleet was being simulated.
	if err != nil {
		return err
	}

	// Execute the script allowing to perform the
	// objective of the fleet if any was provided.
	if script == "" {
		return nil
	}

	query := db.InsertReq{
		Script: script,
		Args: []interface{}{
			f.ID,
		},
	}

	err = data.Proxy.InsertToDB(query)

	return err
}
