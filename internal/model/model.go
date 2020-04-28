package model

import (
	"oglike_server/internal/locker"
	"oglike_server/pkg/db"
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
// The `Locker` defines an object to use to protect
// resources of the DB from concurrent accesses. It
// is used to guarantee that a single process is able
// for example to update the information of a planet
// at any time. It helps preventing data races when
// performing actions on shared elements of a uni.
type Instance struct {
	Proxy        db.Proxy
	Buildings    *BuildingsModule
	Technologies *TechnologiesModule
	Ships        *ShipsModule
	Defenses     *DefensesModule
	Resources    *ResourcesModule
	Locker       *locker.ConcurrentLocker
}

// accessMode :
// Describes the possible ways to access to the
// resources of a planet. This allows to determine
// when to release the locker on the planet's data.
type accessMode int

// Define the possible severity level for a log message.
const (
	ReadOnly accessMode = iota
	ReadWrite
)
