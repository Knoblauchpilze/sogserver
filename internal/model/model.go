package model

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
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

// Lock :
// Used to acquire the lock on this object. This
// method will block until the lock is acquired.
func (i Instance) Lock() {
	i.log.Trace(logger.Verbose, "lock", fmt.Sprintf("Acquiring lock on DB"))
	i.waiter.lock()
	i.log.Trace(logger.Verbose, "lock", fmt.Sprintf("Acquired lock on DB"))
}

// Unlock :
// Used to release the lock previously acquired
// on this object.
func (i Instance) Unlock() {
	i.log.Trace(logger.Verbose, "lock", fmt.Sprintf("Releasing lock on DB"))
	i.waiter.unlock()
	i.log.Trace(logger.Verbose, "lock", fmt.Sprintf("Released lock on DB"))
}

// UpdateResourcesForPlanet :
// Used to perform the update of the resources for the
// input planet in the DB.
//
// The `planet` defines the identifier of the planet.
//
// Returns any error.
func (i Instance) UpdateResourcesForPlanet(planet string) error {
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

// UpdateBuildingsForPlanet :
// Used to perform the update of the buildings for
// the input planet in the DB.
//
// The `planet` defines the identifier of the planet.
//
// Returns any error.
func (i Instance) UpdateBuildingsForPlanet(planet string) error {
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

// UpdateTechnologiesForPlayer :
// Used to perform the update of the technologies
// for the input player in the DB.
//
// The `player` defines the ID of the player.
//
// Returns any error.
func (i Instance) UpdateTechnologiesForPlayer(player string) error {
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

// UpdateShipsForPlanet :
// Used to perform the update of the ships for the
// input planet in the DB.
//
// The `planet` defines the ID of the planet.
//
// Returns any error.
func (i Instance) UpdateShipsForPlanet(planet string) error {
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

// UpdateDefensesForPlanet :
// Used to perform the update of the defenses for
// the input planet in the DB.
//
// The `planet` defines the ID of the planet.
//
// Returns any error.
func (i Instance) UpdateDefensesForPlanet(planet string) error {
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
