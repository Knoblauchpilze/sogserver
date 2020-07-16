package game

import (
	"fmt"
	"math"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
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

// maxFleetDelay : Indicates the maximum delay in percentage of the remaining flight time that
// a component of an ACS fleet is allowed to cause.
var maxFleetDelay float32 = 1.3

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

	if deltaT > maxFleetDelay {
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

	mul, err := NewMultipliersFromDB(fleet.Universe, data)
	if err != nil {
		return ErrMultipliersError
	}

	// On the other hand if the fleet is actually
	// faster than the actual arrival time we need
	// to update the consumption and flight time
	// to match the current arrival time as closely
	// as possible.
	fleet.Speed *= deltaT

	err = fleet.ConsolidateArrivalTime(data, source, mul.Fleet)
	if err != nil {
		return err
	}

	// We shouldn't need to revalidate the data as
	// we will reduce the speed of the fleet and
	// thus burn less fuel in all likelihood.
	err = fleet.consolidateConsumption(data, source, mul)
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
	save := fleet.ArrivalTime

	d := -fleet.ArrivalTime.Sub(acs.arrivalTime)
	fleet.ArrivalTime = acs.arrivalTime
	fleet.CreatedAt = fleet.CreatedAt.Add(d)

	data.log.Trace(logger.Verbose, "fleet", fmt.Sprintf("Changed arrival time for \"%s\" from %v to %v", fleet.ID, save, fleet.ArrivalTime))

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
	fleets := make([]*Fleet, 0)
	cargo := float32(0.0)

	for _, f := range acs.Fleets {
		fleet, err := NewFleetFromDB(f, data)
		if err != nil {
			return ErrFleetFightSimulationFailure
		}

		cargo += fleet.usedCargoSpace()

		fleets = append(fleets, &fleet)
	}

	// Create the attacker structure from the fleets.
	// We know that the fleets are ordered by their
	// desired joining time so we can just traverse
	// the slice from the beginning to the end.
	a := attacker{
		participants: make([]string, 0),
		fleets:       make([]string, 0),
		units:        make([]shipsUnit, 0),
		usedCargo:    cargo,
	}

	attackers := make(map[string]bool)

	for _, f := range fleets {
		att, err := f.toAttacker(data)
		if err != nil {
			return ErrFleetFightSimulationFailure
		}

		a.units = append(a.units, att.units...)

		// Register new participants and fleets.
		for _, p := range att.participants {
			_, ok := attackers[p]
			if !ok {
				a.participants = append(a.participants, p)
				attackers[p] = true
			}
		}

		a.fleets = append(a.fleets, f.ID)
	}

	// Create the defender from the planet.
	d, err := p.toDefender(data, acs.arrivalTime)
	if err != nil {
		return ErrFleetFightSimulationFailure
	}

	result, err := d.defend(&a, acs.arrivalTime, data)
	if err != nil {
		return ErrFleetFightSimulationFailure
	}

	// Just like regular fleets we need to
	// handle the resources already carried
	// by the fleets composing the ACS.
	for _, f := range fleets {
		f.handleDumbMove(a)
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

	// The pillage will be split equally between the
	// fleets based on their available cargo space.
	repartition := make(map[string][]model.ResourceAmount)

	if len(pillage) > 0 {
		repartition, err = acs.allocatePillage(pillage, fleets, data)
		if err != nil {
			return ErrFleetFightSimulationFailure
		}
	}

	// Update the fleets so that we can batch save
	// them back to the DB.
	err = acs.updateFleetsAfterFight(a, fleets, data)
	if err != nil {
		return ErrFleetFightSimulationFailure
	}

	// Post fight reports.
	err = d.generateReports(&a, result, pillage, data.Proxy)
	if err != nil {
		return ErrFleetFightSimulationFailure
	}

	// Update the planet's data in the DB.
	query := db.InsertReq{
		Script: "planet_fight_aftermath",
		Args: []interface{}{
			p.ID,
			string(p.Coordinates.Type),
			d.convertShips(),
			d.convertDefenses(),
			pillage,
			result.debris,
			result.moon,
			result.diameter,
			result.date,
		},
	}

	err = data.Proxy.InsertToDB(query)
	if err != nil {
		return ErrFleetFightSimulationFailure
	}

	// Update each fleet's data in the DB.
	for _, f := range fleets {
		// Execute the query to update the fleet with
		// its associated data.
		pillageForFleet := repartition[f.ID]
		if pillageForFleet == nil {
			pillageForFleet = make([]model.ResourceAmount, 0)
		}

		query = db.InsertReq{
			Script: "fleet_fight_aftermath",
			Args: []interface{}{
				f.ID,
				a.convertShipsForFleet(f.ID),
				pillageForFleet,
				fmt.Sprintf("%s", result.outcome),
			},
		}

		err = data.Proxy.InsertToDB(query)
		if err != nil {
			return ErrFleetFightSimulationFailure
		}
	}

	// Update the ACS operation in the DB.
	query = db.InsertReq{
		Script: "acs_fleet_fight_aftermath",
		Args: []interface{}{
			acs.ID,
		},
	}

	err = data.Proxy.InsertToDB(query)
	if err != nil {
		return ErrFleetFightSimulationFailure
	}

	return nil
}

// allocatePillage :
// Used to perform the allocation of the resources
// plundered by an ACS operation between all the
// fleets that participated to it. The goal is to
// split it equally based on the cargo available
// for each fleet.
//
// The `pillage` represents the amount of resources
// plundered from the planet.
//
// The `fleets` represents the fleets objects to
// associate to each element of the ACS attack.
//
// The `data` allows to access to the DB if needed.
//
// Returns a map where each key corresponds to a
// key of a fleet participating in this ACS op and
// the value corresponds to the pillaged resources
// carried by the fleet. Also any error is returned.
func (acs *ACSFleet) allocatePillage(pillage []model.ResourceAmount, fleets []*Fleet, data Instance) (map[string][]model.ResourceAmount, error) {
	// Each fleet will receive an amount of pillage
	// based on the cargo capacity of each of them.
	repartition := make(map[string][]model.ResourceAmount)
	toAllocate := make(map[string]map[string]float32)

	totalCargo := float32(0.0)
	totalUsed := float32(0.0)
	cargoForFleets := make(map[string]float32)

	for _, f := range fleets {
		cargo, err := f.cargoSpace(data)
		if err != nil {
			return repartition, err
		}

		used := f.usedCargoSpace()
		available := float32(cargo) - used

		cargoForFleets[f.ID] = available
		totalUsed += available
		totalCargo += float32(cargo)

		// Build the resources carried by the fleet.
		resForFleet := make(map[string]float32)
		for _, res := range f.Cargo {
			resForFleet[res.Resource] = res.Amount
		}

		toAllocate[f.ID] = resForFleet
	}

	// We need to dispatch the resources that were
	// pillaged among the fleets. We will start by
	// assigning a fair amount of resources to the
	// fleet and then adjust to allocate the rest
	// in case the capacity is filled.
	leftToCollect := float32(0.0)
	for _, res := range pillage {
		leftToCollect += res.Amount
	}

	resourcesTypesToCarry := len(pillage)
	fleetsToAllocate := len(fleets)
	available := totalCargo - totalUsed

	for available > 0.1 && leftToCollect > 0.1 {
		// The amount of resources collected in this
		// pass is a fair share of all the resources
		// types available.
		toCollect := available / float32(fleetsToAllocate*resourcesTypesToCarry)

		// Collect each resource: the actual amount
		// collected might be smaller in case there's
		// not enough resources in the field.
		for id := range pillage {
			res := &pillage[id]

			// Allocate the share to each fleet.
			for _, f := range fleets {
				resForFleet := toAllocate[f.ID]
				carried := resForFleet[res.Resource]

				// Cannot collect more that what's available.
				amount := math.Min(float64(res.Amount), float64(toCollect))
				space := cargoForFleets[f.ID]

				if space <= 0.0 {
					continue
				}

				collectedAmount := float32(math.Min(amount, float64(space)))

				available -= collectedAmount
				res.Amount -= collectedAmount
				leftToCollect -= collectedAmount

				// If one resource is depleted, decrease the
				// available resource type: this will allow
				// to speed up the repartition of the other
				// resource types.
				if res.Amount <= 0.0 {
					resourcesTypesToCarry--
				}
				if space <= collectedAmount {
					fleetsToAllocate--
				}

				// Collect the resources.
				carried += collectedAmount
				resForFleet[res.Resource] = carried
				toAllocate[f.ID] = resForFleet
				cargoForFleets[f.ID] -= space
			}
		}
	}

	// Convert the map of resources collected to a
	// slice.
	for fID, resources := range toAllocate {
		share := make([]model.ResourceAmount, 0)

		for res, amount := range resources {
			r := model.ResourceAmount{
				Resource: res,
				Amount:   amount,
			}

			share = append(share, r)
		}

		repartition[fID] = share
	}

	return repartition, nil
}

// updateFleetsAfterFight :
// Used to perform the update of the fleets as
// a result of the fight process. The attacker
// provided as input corresponds to the state
// of the fleet after the fight and the `fleets`
// represent the initial list of ships and state
// of the fleets in this ACS component. We will
// basically override the ships information as
// it is described in the `a` object.
//
// The `a` object represents the state of the
// ACS operation after the fight.
//
// The `fleets` define the initial state of the
// fleets as they were before the fight.
//
// Allows to access the logger.
//
// Returns any error.
func (acs *ACSFleet) updateFleetsAfterFight(a attacker, fleets []*Fleet, data Instance) error {
	// For each fleet we will fetch the remaining
	// ships from the attacker.
	for _, f := range fleets {
		// Log previous state of fleet.
		for _, s := range f.Ships {
			data.log.Trace(logger.Verbose, "acs", fmt.Sprintf("Fleet \"%s\" contained %d \"%s\"", f.ID, s.Count, s.ID))
		}

		f.Ships = make(map[string]ShipInFleet)

		for _, units := range a.units {
			for _, unit := range units {
				if unit.Fleet != f.ID {
					continue
				}

				ships, ok := f.Ships[unit.Ship]
				if !ok {
					ships = ShipInFleet{
						ID:    unit.Ship,
						Count: unit.Count,
					}
				} else {
					ships.Count += unit.Count
				}

				f.Ships[unit.Ship] = ships
			}
		}

		// Log current state of fleet.
		for _, s := range f.Ships {
			data.log.Trace(logger.Verbose, "acs", fmt.Sprintf("Fleet \"%s\" now contains %d \"%s\"", f.ID, s.Count, s.ID))
		}
	}

	return nil
}
