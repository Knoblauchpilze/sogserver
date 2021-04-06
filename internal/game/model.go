package game

import (
	"fmt"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"time"
)

// Instance :
// Defines an instance of the data model which contains
// modules allowing to handle various aspects of it. It
// is usually created once to regroup all the data of
// the game in a single easy-to-use object.
//
// The `Proxy` defines a way to access to the DB in
// case the information present in this element does
// not cover all the needs.
//
// The `Buildings` defines the object to use to access
// to the buildings information for the game.
//
// The `Technologies` defines a similar object but to
// access to technologies.
//
// The `Ships` defines the possible ships in the game.
//
// The `Defense` defines the defense system that can
// be built on a planet.
//
// The `Resources` defines the module to access to all
// available resources in the game.
//
// The `Objectives` defines the module to access to all
// the fleet objectives defined in the game.
//
// The `Messages` defines the module to access to all
// the messages defined in the game.
//
// The `log` defines a logger object to use to notify
// information or errors to the user.
//
// The `waiter` allows to lock this instance which will
// prevent any unauthorized use of the DB.
type Instance struct {
	Proxy        db.Proxy
	Countries    *model.CountriesModule
	Buildings    *model.BuildingsModule
	Technologies *model.TechnologiesModule
	Ships        *model.ShipsModule
	Defenses     *model.DefensesModule
	Resources    *model.ResourcesModule
	Objectives   *model.FleetObjectivesModule
	Messages     *model.MessagesModule

	log    logger.Logger
	waiter *locker
}

// actionKind :
// Describes the possible kind for an action to perform.
// It is linked to the upgrade action and the fleets to
// process in the game.
type actionKind string

// Define the possible kind of actions.
const (
	planetBuilding actionKind = "building_upgrade"
	moonBuilding   actionKind = "building_upgrade_moon"
	technology     actionKind = "technology_upgrade"
	planetShip     actionKind = "ship_upgrade"
	moonShip       actionKind = "ship_upgrade_moon"
	planetDefense  actionKind = "defense_upgrade"
	moonDefense    actionKind = "defense_upgrade_moon"
	fleet          actionKind = "fleet"
	acsFleet       actionKind = "acs_fleet"
)

// locker :
// Defines a common locker that can be used to protect
// from concurrent accesses in a single-user fashion.
type locker struct {
	waiter chan struct{}
}

// newLocker :
// Performs the creation of a new locker with a status
// set to unlocked.
//
// Returns the created locker.
func newLocker() *locker {
	l := locker{
		make(chan struct{}, 1),
	}

	l.waiter <- struct{}{}

	return &l
}

// lock :
// Used to perform the lock of the resource managed
// by this element.
func (l *locker) lock() {
	<-l.waiter
}

// unlock :
// Used to release the resource managed by this lock.
func (l *locker) unlock() {
	l.waiter <- struct{}{}
}

// NewInstance :
// Used to create a default instance of a data model
// with a valid waiter object. Nothing else is set
// to a meaningful value.
//
// The `proxy` represents the DB object to use for
// this instance.
//
// The `log` defines a way to notify information
// and errors if needed.
//
// Returns the created instance.
func NewInstance(proxy db.Proxy, log logger.Logger) Instance {
	i := Instance{
		Proxy: proxy,

		log:    log,
		waiter: newLocker(),
	}

	return i
}

// trace :
// Wrapper around the internal logger method to be
// able to provide always the same `module` string.
//
// The `level` defines the severity of the message
// to log.
//
// The `msg` defines the content of the log.
func (i Instance) trace(level logger.Severity, msg string) {
	i.log.Trace(level, "lock", msg)
}

// Lock :
// Used to acquire the lock on this object. This
// method will block until the lock is acquired.
// It will also perform an update of the actions
// that are outstanding when the lock is acquired.
func (i Instance) Lock() {
	i.trace(logger.Verbose, fmt.Sprintf("Acquiring lock on DB"))
	i.waiter.lock()
	i.trace(logger.Verbose, fmt.Sprintf("Acquired lock on DB"))

	// Schedule the execution of the outstanding actions
	// now that the lock is acquired.
	err := i.scheduleActions()
	if err != nil {
		i.trace(logger.Error, fmt.Sprintf("Unable to execute outstanding actions (err: %v)", err))
	}
}

// Unlock :
// Used to release the lock previously acquired
// on this object.
func (i Instance) Unlock() {
	i.trace(logger.Verbose, fmt.Sprintf("Releasing lock on DB"))
	i.waiter.unlock()
	i.trace(logger.Verbose, fmt.Sprintf("Released lock on DB"))
}

// scheduleActions :
// Used to perform the execution of all the pending
// actions in the actions queue. It will fetch all
// actions that have completed but not yet been
// executed and process them in order.
//
// Returns any error.
func (i Instance) scheduleActions() error {
	// First, we need to fetch the identifiers of the
	// actions that have completed but have not yet
	// been executed.
	now := time.Now()

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"action",
			"type",
			"completion_time",
		},
		Table: "actions_queue",
		Filters: []db.Filter{
			{
				Key:      "completion_time",
				Values:   []interface{}{now},
				Operator: db.LessThan,
			},
		},
		Ordering: "order by completion_time",
	}

	dbRes, err := i.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}
	if dbRes.Err != nil {
		return dbRes.Err
	}

	// Analyze actions to perform.
	var action string
	var kind actionKind
	var completion time.Time

	for dbRes.Next() {
		err = dbRes.Scan(
			&action,
			&kind,
			&completion,
		)

		if err != nil {
			return err
		}

		switch kind {
		case planetBuilding:
			err = i.performBuildingAction(action, "planet")
		case moonBuilding:
			err = i.performBuildingAction(action, "moon")
		case technology:
			err = i.performTechnologyAction(action)
		case planetShip:
			err = i.performShipAction(action, "planet")
		case moonShip:
			err = i.performShipAction(action, "moon")
		case planetDefense:
			err = i.performDefenseAction(action, "planet")
		case moonDefense:
			err = i.performDefenseAction(action, "moon")
		case fleet:
			err = i.performFleetAction(action)
		case acsFleet:
			err = i.performACSFleetAction(action)
		default:
			i.trace(logger.Error, fmt.Sprintf("Unknown action \"%s\" with kind \"%s\" not processed", action, kind))
		}

		if err != nil {
			i.trace(logger.Error, fmt.Sprintf("Failed to perform action \"%s\" (err: %v)", action, err))
		}
	}

	return nil
}

// updateResourcesForPlanet :
// Used to perform the update of the resources for the
// input planet in the DB.
//
// The `planet` defines the identifier of the planet.
//
// Returns any error.
func (i Instance) updateResourcesForPlanet(planet string) error {
	i.trace(logger.Verbose, fmt.Sprintf("Updating resources for planet %s", planet))

	update := db.InsertReq{
		Script: "update_resources_for_planet",
		Args: []interface{}{
			planet,
		},
		SkipReturn: true,
	}

	err := i.Proxy.InsertToDB(update)

	return err
}

// performBuildingAction :
// Used to perform the execution of the action related
// to the input identifier. It should correspond to a
// building action otherwise the update will fail.
//
// The `action` defines the identifier of the action
// to perform.
//
// The `location` defines where the action is taking
// place, i.e. either a planet or a moon.
//
// Returns any error.
func (i Instance) performBuildingAction(action string, location string) error {
	i.trace(logger.Verbose, fmt.Sprintf("Executing action %s (type: \"building\", location: \"%s\")", action, location))

	update := db.InsertReq{
		Script: "update_building_upgrade_action",
		Args: []interface{}{
			action,
			location,
		},
		SkipReturn: true,
	}

	err := i.Proxy.InsertToDB(update)

	return err
}

// performTechnologyAction :
// Similar to the `performBuildingAction` but for
// the case of technology actions. Note that this
// is the only case where there's no need to give
// a location indication as a technology is always
// researched from a planet.
//
// The `action` defines the ID of the action.
//
// Returns any error.
func (i Instance) performTechnologyAction(action string) error {
	i.trace(logger.Verbose, fmt.Sprintf("Executing action %s (type: \"technology\", location: \"planet\")", action))

	update := db.InsertReq{
		Script: "update_technology_upgrade_action",
		Args: []interface{}{
			action,
		},
		SkipReturn: true,
	}

	err := i.Proxy.InsertToDB(update)

	return err
}

// performShipAction :
// Similar to the `performBuildingAction` but for
// the case of ship actions.
//
// The `action` defines the ID of the action.
//
// The `location` defines where the action is set
// to occur, i.e. either a planet or a moon.
//
// Returns any error.
func (i Instance) performShipAction(action string, location string) error {
	i.trace(logger.Verbose, fmt.Sprintf("Executing action %s (type: \"ship\", location: \"%s\")", action, location))

	update := db.InsertReq{
		Script: "update_ship_upgrade_action",
		Args: []interface{}{
			action,
			location,
		},
		SkipReturn: true,
	}

	err := i.Proxy.InsertToDB(update)

	return err
}

// performDefenseAction :
// Similar to the `performBuildingAction` but for
// the case of defense actions.
//
// The `action` defines the ID of the action.
//
// The `location` defines where the action is set
// to occur, i.e. either a planet or a moon.
//
// Returns any error.
func (i Instance) performDefenseAction(action string, location string) error {
	i.trace(logger.Verbose, fmt.Sprintf("Executing action %s (type: \"defense\", location: \"%s\")", action, location))

	update := db.InsertReq{
		Script: "update_defense_upgrade_action",
		Args: []interface{}{
			action,
			location,
		},
		SkipReturn: true,
	}

	err := i.Proxy.InsertToDB(update)

	return err
}

// performFleetAction :
// Used to perform the simulation of the effects
// of the fleet described by the input ID. It is
// fetched from the actions queue and when this
// method is reached we know that everything is
// up-to-date until this point.
//
// The `ID` defines the identifier of the fleet
// to simulate.
//
// Returns any error.
func (i Instance) performFleetAction(ID string) error {
	i.trace(logger.Verbose, fmt.Sprintf("Executing fleet %s", ID))

	// Retrieve the fleet corresponding to the ID
	// in argument.
	f, err := NewFleetFromDB(ID, i)
	if err != nil {
		return err
	}

	// Retrieve the target planet of the fleet to
	// be able to simulate it. In case the fleet
	// does not have a target, use a `nil` value.
	var p *Planet

	if f.Target != "" {
		valid := true
		var rp Planet

		switch f.TargetCoords.Type {
		case World:
			rp, err = NewPlanetFromDB(f.Target, i)
		case Moon:
			rp, err = NewMoonFromDB(f.Target, i)
		default:
			// Probably debris, do nothing.
			valid = false
		}

		if err != nil {
			return err
		}

		if valid {
			p = &rp
		}
	}

	return f.simulate(p, i)
}

// performACSFleetAction :
// Used to perform the simulation of the ACS
// fleet described by the input action. It is
// similar to the process performed in the
// `performFleetAction` but for the ACS case.
//
// The `ID` defines the identifier of the fleet
// to simulate.
//
// Returns any error.
func (i Instance) performACSFleetAction(ID string) error {
	i.trace(logger.Verbose, fmt.Sprintf("Executing ACS fleet %s", ID))

	// Retrieve the ACS fleet corresponding to
	// the ID in argument.
	acs, err := NewACSFleetFromDB(ID, i)
	if err != nil {
		return err
	}

	// Retrieve the target planet of the fleet to
	// be able to simulate it. In case of a `ACS`
	// fleet we *know* it will exist.
	var p *Planet

	if acs.TargetType == World || acs.TargetType == Moon {
		valid := true
		var rp Planet

		switch acs.TargetType {
		case World:
			rp, err = NewPlanetFromDB(acs.Target, i)
		case Moon:
			rp, err = NewMoonFromDB(acs.Target, i)
		default:
			// Probably debris, do nothing.
			valid = false
		}

		if err != nil {
			return err
		}

		if valid {
			p = &rp
		}
	}

	return acs.simulate(p, i)
}
