package model

// Instance :
// Defines an instance of the data model which contains
// modules allowing to handle various aspects of it. It
// is usually created once to regroup all the data of
// the game in a single easy-to-use object.
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
type Instance struct {
	Buildings    *BuildingsModule
	Technologies *TechnologiesModule
	Ships        *ShipsModule
	Defenses     *DefensesModule
	Resources    *ResourcesModule
}
