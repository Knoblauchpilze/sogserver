package model

import (
	"fmt"
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
// The `waiter` allows to lock this instance which will
// prevent any unauthorized use of the DB.
type Instance struct {
	Proxy        db.Proxy
	Buildings    *BuildingsModule
	Technologies *TechnologiesModule
	Ships        *ShipsModule
	Defenses     *DefensesModule
	Resources    *ResourcesModule
	Objectives   *FleetObjectivesModule

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
	building   actionKind = "building_upgrade"
	technology actionKind = "technology_upgrade"
	ship       actionKind = "ship_upgrade"
	defense    actionKind = "defense_upgrade"
	fleet      actionKind = "fleet"
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
		case building:
		case technology:
		case ship:
		case defense:
		case fleet:
		default:
			i.trace(logger.Error, fmt.Sprintf("Unknown action \"%s\" with kind \"%s\" not processed", action, kind))
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
	// Perform the update of the building upgrade actions.
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

// updateBuildingsForPlanet :
// Used to perform the update of the buildings for
// the input planet in the DB.
//
// The `planet` defines the identifier of the planet.
//
// Returns any error.
func (i Instance) updateBuildingsForPlanet(planet string) error {
	update := db.InsertReq{
		Script: "update_building_upgrade_action",
		Args: []interface{}{
			planet,
			"planet",
		},
		SkipReturn: true,
	}

	err := i.Proxy.InsertToDB(update)

	return err
}

// updateTechnologiesForPlayer :
// Used to perform the update of the technologies
// for the input player in the DB.
//
// The `player` defines the ID of the player.
//
// Returns any error.
func (i Instance) updateTechnologiesForPlayer(player string) error {
	update := db.InsertReq{
		Script: "update_technology_upgrade_action",
		Args: []interface{}{
			player,
		},
		SkipReturn: true,
	}

	err := i.Proxy.InsertToDB(update)

	return err
}

// updateShipsForPlanet :
// Used to perform the update of the ships for the
// input planet in the DB.
//
// The `planet` defines the ID of the planet.
//
// Returns any error.
func (i Instance) updateShipsForPlanet(planet string) error {
	update := db.InsertReq{
		Script: "update_ship_upgrade_action",
		Args: []interface{}{
			planet,
			"planet",
		},
		SkipReturn: true,
	}

	err := i.Proxy.InsertToDB(update)

	return err
}

// updateDefensesForPlanet :
// Used to perform the update of the defenses for
// the input planet in the DB.
//
// The `planet` defines the ID of the planet.
//
// Returns any error.
func (i Instance) updateDefensesForPlanet(planet string) error {
	update := db.InsertReq{
		Script: "update_defense_upgrade_action",
		Args: []interface{}{
			planet,
			"planet",
		},
		SkipReturn: true,
	}

	err := i.Proxy.InsertToDB(update)

	return err
}
