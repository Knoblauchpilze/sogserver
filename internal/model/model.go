package model

import (
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
// The `Objectives` defines the module to access to all
// the fleet objectives defined in the game.
//
// The `Locker` defines an object to use to protect
// resources of the DB from concurrent accesses. It
// is used to guarantee that a single process is able
// for example to update the information of a planet
// at any time. It helps preventing data races when
// performing actions on shared elements of the game.
type Instance struct {
	Proxy        db.Proxy
	Buildings    *BuildingsModule
	Technologies *TechnologiesModule
	Ships        *ShipsModule
	Defenses     *DefensesModule
	Resources    *ResourcesModule
	Objectives   *FleetObjectivesModule
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
