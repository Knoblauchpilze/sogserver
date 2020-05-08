package game

import (
	"encoding/json"
	"fmt"
	"math"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"time"
)

// Component :
// Defines a single element participating to a fleet. This
// is usually defining a single wave of ship as it was set
// and joined the fleet. This component is composed of a
// certain number of ships and travels at a single speed.
// The most basic fleet has a single component (and must
// have at least one) but large fleets can benefit from
// being split into several parts as they might be able
// to better use their firepower.
//
// The `ID` represents the ID of the fleet component as
// defined in the DB. It allows to uniquely identify it.
//
// The `Player` defines the identifier of the player that
// launched this fleet component. It means that all the
// ships in this component belongs to the player.
//
// The `Source` defines the identifier of the starting
// planet or moon of this component. Unlike the target
// of the fleet which might not be a valid planet/moon
// yet a fleet must start from an existing location.
//
// The `SourceType` defines the kind of the source. It
// is either a planet or a moon and indicates to the
// underlying DB where the info for this component is
// to be updated.
//
// The `Speed` defines the travel speed of this fleet
// component. It is used to precisely determine how
// much this component impacts the final arrival time
// of the fleet and also for the consumption of fuel.
// The value is in the range `[0, 1]` representing a
// percentage of the maximum available speed.
//
// The `JoinedAt` defines the time at which this player
// has joined the main fleet and created this component.
//
// The `ReturnTime` defines the time at which the fleet
// component will return to the planet. This value does
// correspond to twice the duration of a single flight
// from the starting position to the destination.
//
// The `Ships` define the actual ships involved in this
// fleet component.
//
// The `Fleet` is used as an internal value allowing
// to determine to which fleet this component is linked.
//
// The `Consumption` defines the amount of resources
// that is needed for this component to move from its
// starting location to its destination.
//
// The `Cargo` holds the resources moved around by the
// fleet component. It accounts for all the cargo that
// is available on all ships.
//
// The `ArrivalTime` defines the expected arrival time
// of the component to its destination. It should be
// consistent with what's expected by the parent fleet
// and allows to slightly offset the arrival time if
// needed.
//
// The `Target` defines the destination of the fleet
// either through its own definition or by polling
// the parent fleet.
//
// The `Objective` defines the objective of this fleet
// component. It is used in case the fleet associated
// to the comp is not yet created to asses the purpose
// of the fleet.
//
// The `Name` defines a human readable name provided
// by the user to reference the fleet. It should be
// unique in a universe.
//
// The `flightTime` defines the flight time expressed
// in milliseconds. Note that it is somewhat redundant
// with the other time information (namely `JoinedAt`,
// `ArrivalTime` and `ReturnTime`) but actually it is
// meant to help the computations of these values.
type Component struct {
	ID          string                 `json:"id"`
	Player      string                 `json:"-"`
	Source      string                 `json:"source"`
	SourceType  Location               `json:"-"`
	Speed       float32                `json:"speed"`
	JoinedAt    time.Time              `json:"joined_at"`
	ReturnTime  time.Time              `json:"return_time"`
	Ships       ShipsInFleet           `json:"ships"`
	Fleet       string                 `json:"fleet"`
	Consumption []ConsumptionValue     `json:"-"`
	Cargo       []model.ResourceAmount `json:"cargo"`
	ArrivalTime time.Time              `json:"-"`
	Target      Coordinate             `json:"target"`
	Objective   string                 `json:"objective"`
	Name        string                 `json:"name"`
	flightTime  time.Duration
}

// Components :
// Convenience define to refer to a list of fleet
// components.
type Components []Component

// ConsumptionValue :
// Used as a convenience define to reference resource
// amount in a meaningful way.
type ConsumptionValue model.ResourceAmount

// ShipInFleet :
// Defines a single ship involved in a fleet component.
// This is basically the building blocks of fleet. This
// element defines a set of similar ships that joined a
// fleet from a single location (and thus belong to a
// single player).
//
// The `ID` defines the identifier of the ship that is
// involved in the fleet component.
//
// The `Count` defines how many ships of the specified
// type are involved.
type ShipInFleet struct {
	ID    string `json:"ship"`
	Count int    `json:"count"`
}

// ShipsInFleet :
// Convenience define to refer to a list of ships
// belonging to a fleet component. Allows to define
// some methods on this type to ease the consistency
// checks.
type ShipsInFleet []ShipInFleet

// ErrInvalidFleetComponent :
// Used to indicate that the fleet component's identifier
// provided in input is not valid.
var ErrInvalidFleetComponent = fmt.Errorf("Invalid fleet component with no identifier")

// ErrDuplicatedFleetComponent :
// Used to indicate that the fleet component's ID provided
// input is not unique in the DB.
var ErrDuplicatedFleetComponent = fmt.Errorf("Invalid not unique fleet component")

// ErrInsufficientCargo :
// Used to indicate that the fleet component is supposed
// to transport more resources than the available cargo
// space available.
var ErrInsufficientCargo = fmt.Errorf("Insufficient cargo space to hold resources")

// ErrInvalidCargo :
// Used to indicate that the amount of resources that should
// be carried by the fleet component is invalid (typically a
// negative value).
var ErrInvalidCargo = fmt.Errorf("Invalid cargo value for resource")

// ErrArrivalTimeMismatch :
// Used to indicate that the arrival time computed for a
// given fleet component is incompatible with the time its
// parent fleet should arrive.
var ErrArrivalTimeMismatch = fmt.Errorf("Fleet and component arrival times mismatch")

// ErrInvalidPropulsionSystem :
// Used to indicate that the propulsion system of a ship
// was not valid compared to the researched technologies
// on the planet.
var ErrInvalidPropulsionSystem = fmt.Errorf("Unknown propulsion system for ship")

// ErrInvalidObjective :
// Used to indicate that the objective associated to the
// fleet component is not valid.
var ErrInvalidObjective = fmt.Errorf("Invalid objective for component")

// ErrInvalidFleetForComponent :
// Used to indicate that the fleet provided in input
// for the validation of a component is not correct.
var ErrInvalidFleetForComponent = fmt.Errorf("Invalid fleet provided for component")

// ErrCargoNotMovable :
// Used to indicate that the resource specified as a
// cargo cannot be moved.
var ErrCargoNotMovable = fmt.Errorf("Resource cannot be moved by a fleet")

// valid :
// Used to verify that the ship assigned to a component
// are valid. It should contain a valid ship's ID and a
// non zero amount of ship.
//
// Returns `true` if the ship is valid.
func (sif ShipInFleet) valid() bool {
	return validUUID(sif.ID) && sif.Count > 0
}

// valid :
// Used to perform a chain validation on all the ships
// sets defined in the slice.
//
// Returns `true` if all individual components are
// valid.
func (sifs ShipsInFleet) valid() bool {
	for _, sif := range sifs {
		if !sif.valid() {
			return false
		}
	}

	return len(sifs) > 0
}

// valid :
// Used to perform a chain validation on all the elems
// for this slice.
//
// The `objective` defines the parent fleet's objective
// which should be matched by all components.
//
// The `target` defines the coordinates targeted by
// the parent fleet which should be reflected by all
// the individual components.
//
// Returns `true` if all individual components are
// valid.
func (fcs Components) valid(objective string, target Coordinate) bool {
	for _, comp := range fcs {
		if !comp.Valid() || comp.Objective != objective || comp.Target != target {
			return false
		}
	}

	return true
}

// newComponentFromDB :
// Used to fetch the content of the fleet component
// from the input DB and populate all internal fields
// from it. In case the DB cannot be fetched or some
// errors are encoutered, the return value can be used
// to get a description of the error.
//
// The `ID` defines the ID of the fleet component to
// get. It is fetched from the DB and should refer
// to an existing component.
//
// The `data` allows to actually perform the DB
// requests to fetch the component's data.
//
// Returns the component as fetched from the DB along
// with any errors.
func newComponentFromDB(ID string, data model.Instance) (Component, error) {
	// Create the fleet.
	c := Component{
		ID: ID,
	}

	// Consistency.
	if !validUUID(c.ID) {
		return c, ErrInvalidFleetComponent
	}

	// Fetch the fleet's content.
	err := c.fetchGeneralInfo(data)
	if err != nil {
		return c, err
	}

	err = c.fetchShips(data)
	if err != nil {
		return c, err
	}

	err = c.fetchCargo(data)
	if err != nil {
		return c, err
	}

	err = c.fetchFleetInfo(data)
	if err != nil {
		return c, err
	}

	return c, nil
}

// Valid :
// Used to determine whether the fleet component defined
// by this element is valid or not. We will check that
// the starting coordinate are valid and the each ship
// packet is also valid.
//
// Returns `true` if the component is valid.
func (fc *Component) Valid() bool {
	return validUUID(fc.ID) &&
		validUUID(fc.Player) &&
		validUUID(fc.Source) &&
		len(fc.SourceType) > 0 &&
		fc.Speed >= 0.0 && fc.Speed <= 1.0 &&
		fc.Ships.valid() &&
		validUUID(fc.Objective) &&
		// Allow for either a valid fleet identifier of
		// no identifier at all in case the fleet for
		// this component does not exist yet.
		(fc.Fleet == "" || validUUID(fc.Fleet))
}

// String :
// Implementation of the `Stringer` interface to make
// sure displaying this fleet component is easy.
//
// Returns the corresponding string.
func (fc Component) String() string {
	return fmt.Sprintf("[id: %s, player: %s, source: %s, type: %s]", fc.ID, fc.Player, fc.Source, fc.SourceType)
}

// consolidateConsumption :
// Used to perform the consolidation of the consumption
// required for this component to take off. It does not
// handle the fact that the parent planet actually has
// the needed fuel but only to compute it.
// The result of the computations will be saved in the
// component itself.
//
// The `data` allows to get information from the DB on
// the consumption of ships.
//
// The `p` defines the planet from where this component
// is starting the flight.
//
// Returns any error.
func (fc *Component) consolidateConsumption(data model.Instance, p *Planet) error {
	// Compute the distance between the starting position
	// and the destination of the flight.
	d := float64(p.Coordinates.distanceTo(fc.Target))

	// Now we can compute the total consumption by summing
	// the individual consumptions of ships.
	consumption := make(map[string]float64)

	for _, ship := range fc.Ships {
		sd, err := data.Ships.GetShipFromID(ship.ID)

		if err != nil {
			return err
		}

		for _, fuel := range sd.Consumption {
			// The values and formulas are extracted from here:
			// https://ogame.fandom.com/wiki/Talk:Fuel_Consumption
			// The flight time is expressed internally in millisecs.
			ftSec := float64(fc.flightTime) / float64(time.Second)

			sk := 35000.0 * math.Sqrt(d*10.0/float64(sd.Speed)) / (ftSec - 10.0)
			cons := float64(fuel.Amount*float32(ship.Count)) * d * math.Pow(1.0+sk/10.0, 2.0) / 35000.0

			ex := consumption[fuel.Resource]
			ex += cons
			consumption[fuel.Resource] = ex
		}
	}

	// Save the data in the component itself.
	fc.Consumption = make([]ConsumptionValue, 0)

	for res, fuel := range consumption {
		value := ConsumptionValue{
			Resource: res,
			Amount:   float32(fuel),
		}

		fc.Consumption = append(fc.Consumption, value)
	}

	return nil
}

// ConsolidateArrivalTime :
// Used to perform the update of the arrival time for
// this fleet component based on the technologies of
// the planet it starts from.
//
// The `data` allows to get information from the DB
// related to the propulsion used by each ships and
// their consumption.
//
// The `p` defines the planet from which this comp
// should start and will be used to update the values
// of the techs that should be used for computing a
// speed for each ship.
//
// Returns any error.
func (fc *Component) ConsolidateArrivalTime(data model.Instance, p *Planet) error {
	// Consistency.
	if fc.Source != p.ID || fc.SourceType != World {
		return ErrInvalidPlanet
	}

	// Update the time at which this component joined
	// the fleet.
	fc.JoinedAt = time.Now()

	// Compute the time of arrival for this component. It
	// is function of the percentage of the maximum speed
	// used by the ships and the slowest ship's speed.
	d := float64(p.Coordinates.distanceTo(fc.Target))

	// Compute the maximum speed of the fleet. This will
	// correspond to the speed of the slowest ship in the
	// component.
	maxSpeed := math.MaxInt32

	for _, ship := range fc.Ships {
		sd, err := data.Ships.GetShipFromID(ship.ID)

		if err != nil {
			return err
		}

		level, ok := p.technologies[sd.Propulsion.Propulsion]
		if !ok {
			return ErrInvalidPropulsionSystem
		}

		speed := sd.Propulsion.ComputeSpeed(sd.Speed, level)

		if speed < maxSpeed {
			maxSpeed = speed
		}
	}

	// Compute the duration of the flight given the distance.
	// Note that the speed percentage is interpreted as such:
	//  - 100% -> 10
	//  -  50% -> 5
	//  -  10% -> 1
	speedRatio := fc.Speed * 10.0
	flightTimeSec := 35000.0/float64(speedRatio)*math.Sqrt(float64(d)*10.0/float64(maxSpeed)) + 10.0

	// TODO: Hack to speed up fleets by a lot.
	flightTimeSec /= 200.0

	// Compute the flight time by converting this duration in
	// milliseconds: this will allow to keep more precision.
	fc.flightTime = time.Duration(1000.0*flightTimeSec) * time.Millisecond

	// The arrival time is just this duration in the future.
	// We will use the milliseconds in order to keep more
	// precision before rounding.
	fc.ArrivalTime = fc.JoinedAt.Add(fc.flightTime)

	// The return time is separated from the arrival time
	// by an additional full flight time.
	fc.ReturnTime = fc.ArrivalTime.Add(fc.flightTime)

	return nil
}

// Validate :
// Used to make sure that the component can be created
// from the input planet. The check of the number of
// ships required by the component against the actual
// ships count on the planet will be verified along
// with the fuel consumption and resources loading.
//
// The `data` allows to access to the DB if needed.
//
// The `source` defines the planet attached to this
// fleet: it represents the mandatory location from
// where the component is launched.
//
// The `target` defines the potential planet to which
// this component is directed. It might be `nil` in
// case the objective allows it (typically a colonize
// operation).
//
// The `f` defines the parent fleet for this component.
//
// Returns any error.
func (fc *Component) Validate(data model.Instance, source *Planet, target *Planet, f *Fleet) error {
	// Consistency.
	if fc.Source != source.ID || fc.SourceType != World {
		return ErrInvalidPlanet
	}
	if fc.Fleet != f.ID {
		return ErrInvalidFleetForComponent
	}

	// Update consumption.
	err := fc.consolidateConsumption(data, source)
	if err != nil {
		return err
	}

	// Make sure that the cargo defined for this fleet
	// component can be stored in the ships.
	totCargo := 0

	for _, ship := range fc.Ships {
		sd, err := data.Ships.GetShipFromID(ship.ID)

		if err != nil {
			return err
		}

		totCargo += (ship.Count * sd.Cargo)
	}

	var totNeeded float32
	for _, res := range fc.Cargo {
		if res.Amount < 0.0 {
			return ErrInvalidCargo
		}

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
		return ErrInsufficientCargo
	}

	// Check that the arrival time for this component
	// is consistent with what's expected by the fleet.
	// TODO: Relax the constraint to allow fleet offset.
	if fc.ArrivalTime != f.ArrivalTime {
		return ErrArrivalTimeMismatch
	}

	// Make sure that the objective of this component
	// is consistent with what's defined in the DB.
	obj, err := data.Objectives.GetObjectiveFromID(fc.Objective)
	if err != nil {
		return ErrInvalidObjective
	}

	// Verify that the objective is consistent with
	// the origin and destination of the component:
	// we don't want player to attack themselves for
	// example.
	if obj.Directed && target == nil {
		return ErrInvalidObjective
	}
	if obj.Hostile && target == nil {
		return ErrInvalidObjective
	}
	if obj.Hostile && source.Player == target.Player {
		return ErrInvalidObjective
	}

	// Validate the amount of fuel available on the
	// planet compared to the amount required and
	// that there are enough resources to be taken
	// from the planet.
	// TODO: Hack to allow creation of fleets without checks.
	// return source.validateComponent(fc.Consumption, fc.Cargo, fc.Ships, data)
	return nil
}

// fetchGeneralInfo :
// Used to populate the internal data of the fleet
// component with info from the DB. It will assume
// the identifier for this fleet component is at
// least provided.
//
// The `data` allows to access to the DB.
//
// Returns any error.
func (fc *Component) fetchGeneralInfo(data model.Instance) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"fleet",
			"player",
			"source",
			"source_type",
			"speed",
			"joined_at",
			"return_time",
		},
		Table: "fleet_elements",
		Filters: []db.Filter{
			{
				Key:    "id",
				Values: []string{fc.ID},
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

	// Scan the fleet component's data.
	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return ErrInvalidFleetComponent
	}

	err = dbRes.Scan(
		&fc.Fleet,
		&fc.Player,
		&fc.Source,
		&fc.SourceType,
		&fc.Speed,
		&fc.JoinedAt,
		&fc.ReturnTime,
	)

	// Make sure that it's the only fleet component.
	if dbRes.Next() {
		return ErrDuplicatedFleetComponent
	}

	return err
}

// fetchShips :
// Used to populate the internal data of the fleet
// component with info from the DB. It will assume
// the identifier for this fleet component is at
// least provided.
//
// The `data` allows to access to the DB.
//
// Returns any error.
func (fc *Component) fetchShips(data model.Instance) error {
	fc.Ships = make([]ShipInFleet, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"ship",
			"count",
		},
		Table: "fleet_ships",
		Filters: []db.Filter{
			{
				Key:    "fleet_element",
				Values: []string{fc.ID},
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
	var sif ShipInFleet

	for dbRes.Next() {
		err = dbRes.Scan(
			&sif.ID,
			&sif.Count,
		)

		if err != nil {
			return err
		}

		fc.Ships = append(fc.Ships, sif)
	}

	return nil
}

// fetchCargo :
// Serves a similar purpose to `fetchShips` but is
// dedicated to fetching the cargo for this comp.
//
// The `data` allows to access to the DB.
//
// Returns any error.
func (fc *Component) fetchCargo(data model.Instance) error {
	fc.Cargo = make([]model.ResourceAmount, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"resource",
			"amount",
		},
		Table: "fleet_resources",
		Filters: []db.Filter{
			{
				Key:    "fleet_element",
				Values: []string{fc.ID},
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
	var ra model.ResourceAmount

	for dbRes.Next() {
		err = dbRes.Scan(
			&ra.Resource,
			&ra.Amount,
		)

		if err != nil {
			return err
		}

		fc.Cargo = append(fc.Cargo, ra)
	}

	return nil
}

// fetchFleetInfo :
// Allows to fetch the general info that is stored
// in the parent fleet of a component such as the
// objective of the fleet, its name, etc.
//
// The `data` allows to access to the DB.
//
// Returns any error.
func (fc *Component) fetchFleetInfo(data model.Instance) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"name",
			"objective",
			"target_galaxy",
			"target_solar_system",
			"target_position",
			"target_type",
			"arrival_time",
		},
		Table: "fleets",
		Filters: []db.Filter{
			{
				Key:    "id",
				Values: []string{fc.Fleet},
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

	// Scan the fleet's information.
	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return ErrInvalidFleetComponent
	}

	var g, s, p int
	var loc Location

	err = dbRes.Scan(
		&fc.Name,
		&fc.Objective,
		&g,
		&s,
		&p,
		&loc,
		&fc.ArrivalTime,
	)

	var errC error
	fc.Target, errC = newCoordinate(g, s, p, loc)
	if errC != nil {
		return errC
	}

	fc.flightTime = fc.ArrivalTime.Sub(fc.JoinedAt)

	// Make sure that it's the only fleet.
	if dbRes.Next() {
		return ErrDuplicatedFleet
	}

	return nil
}

// Convert :
// Implementation of the `db.Convertible` interface
// from the DB package in order to only include fields
// that need to be marshalled in the fleet component's
// creation.
//
// Returns the converted version of the planet which
// only includes relevant fields.
func (fc *Component) Convert() interface{} {
	return struct {
		ID         string    `json:"id"`
		Fleet      string    `json:"fleet"`
		Player     string    `json:"player"`
		Source     string    `json:"source"`
		SourceType Location  `json:"source_type"`
		Speed      float32   `json:"speed"`
		JoinedAt   time.Time `json:"joined_at"`
		ReturnTime time.Time `json:"return_time"`
	}{
		ID:         fc.ID,
		Fleet:      fc.Fleet,
		Player:     fc.Player,
		Source:     fc.Source,
		SourceType: fc.SourceType,
		Speed:      fc.Speed,
		JoinedAt:   fc.JoinedAt,
		ReturnTime: fc.ReturnTime,
	}
}

// MarshalJSON :
// Implementation of the `Marshaler` interface to allow
// only specific information to be marshalled when the
// component needs to be exported. It fills a similar
// role to the `Convert` method but only to provide a
// clean interface to the outside world where only
// relevant info is provided.
//
// Returns the marshalled bytes for this fleet component
// along with any error.
func (fc *Component) MarshalJSON() ([]byte, error) {
	type lightComponent struct {
		ID         string                 `json:"id"`
		Source     string                 `json:"source"`
		SourceType Location               `json:"source_type"`
		Speed      float32                `json:"speed"`
		JoinedAt   time.Time              `json:"joined_at"`
		ReturnTime time.Time              `json:"return_time"`
		Ships      ShipsInFleet           `json:"ships"`
		Cargo      []model.ResourceAmount `json:"cargo"`
	}

	// Copy the planet's data.
	lc := lightComponent{
		ID:         fc.ID,
		Source:     fc.Source,
		SourceType: fc.SourceType,
		Speed:      fc.Speed,
		JoinedAt:   fc.JoinedAt,
		ReturnTime: fc.ReturnTime,
		Ships:      fc.Ships,
		Cargo:      fc.Cargo,
	}

	return json.Marshal(lc)
}
