package game

import (
	"fmt"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
)

// ErrColonizationStatusUnknown : Indicates that we could not determine whether
// the colonization process could be executed.
var ErrColonizationStatusUnknown = fmt.Errorf("cannot determine colonization status")

// colonizationAuthorized :
// Used to verify whether the player is authorized to
// colonize at the coordinates indicated by the fleet.
// It includes verifying that the technology reached
// by the player allows one more planet and that the
// universe does not already contain a planet at the
// destination coordinates.
// In case the colonization is possible, the planet
// is directly created and returned to the user. In
// case the return value is `nil` it indicates that
// the colonization is not possible (except if the
// error returned is not `nil`).
//
// The `data` allows to fetch data from the DB.
//
// Returns the planet to colonize. If the return val
// is `nil` and the error as well it means that the
// colonization is not possible.
func (f *Fleet) colonizationAuthrozied(data Instance) (*Planet, error) {
	var planet *Planet

	// Fetch the player's data.
	p, err := NewPlayerFromDB(f.Player, data)
	if err != nil {
		return planet, ErrColonizationStatusUnknown
	}

	u, err := NewUniverseFromDB(f.Universe, data)
	if err != nil {
		return planet, ErrColonizationStatusUnknown
	}

	allowed, err := p.CanColonize(data)
	if err != nil {
		return planet, ErrColonizationStatusUnknown
	}

	if allowed {
		_, err = u.GetPlanetAt(f.TargetCoords, data)
		if err != nil && err != ErrPlanetNotFound {
			return planet, ErrColonizationStatusUnknown
		}

		allowed = (err == ErrPlanetNotFound)
	}

	if allowed {
		// In case the colonization can be performed we
		// will use the parent universe to create the
		// planet for this player at the specified coords.
		planet, err = NewPlanet(f.Player, f.TargetCoords, false, data)
		if err != nil {
			return planet, err
		}
	}

	return planet, nil
}

// canReturnFromColonization :
// Used to determine whether the fleet can return
// from the colonization operation. Colonizating
// a new location implies consuming one of the
// colony ship sent to colonize the planet. If the
// fleet is only composed of a single colony ship
// it means that no ships will actually come back
// to the source world.
// This method determines whether this is the case
// or if nothing will be left of the fleet.
//
// The `ships` module allows to access properties
// of the ships.
//
// Returns `true` if at least a ship from the fleet
// will return from the colonization operation and
// any error.
func (f *Fleet) canReturnFromColonization(ships *model.ShipsModule) (bool, error) {
	// Assume the fleet can return.
	onlyAColonyShip := true

	for id, s := range f.Ships {
		sd, err := ships.GetShipFromID(id)
		if err != nil {
			return false, err
		}

		if sd.Name != "colony ship" || s.Count > 1 {
			onlyAColonyShip = false
			break
		}
	}

	return !onlyAColonyShip, nil
}

// colonize :
// Used to perform the attempt at the colonization
// for a fleet. We will try to create a new planet
// for the player owning the fleet.
//
// The `data` allows to access to the DB.
//
// Return any error along with the name of the
// script to execute to finalize the execution of
// the fleet.
func (f *Fleet) colonize(data Instance) (string, error) {
	// The colonization operation is meant to create
	// a new planet for the player at the coordinates
	// indicated by the target of the fleet. There is
	// a possibility that the operation fails: this
	// can occur if another player already colonized
	// the planet before this fleet or if the level
	// of astrophysics of the player is not suited to
	// a new colony.
	script := "fleet_return_to_base"

	// If the fleet is not returning yet, process the
	// colonization operation.
	if !f.returning {
		// Fetch the player's data.
		p, err := f.colonizationAuthrozied(data)
		if err != nil {
			return "", ErrUnableToSimulateFleet
		}

		if p == nil {
			// In case the colonization cannot be performed
			// register a message for the player.
			query := db.InsertReq{
				Script: "fleet_colonization_failed",
				Args: []interface{}{
					f.ID,
				},
			}

			err = data.Proxy.InsertToDB(query)
			if err != nil {
				return "", err
			}
		} else {
			// Register this information in the DB.
			query := db.InsertReq{
				Script: "fleet_colonization_success",
				Args: []interface{}{
					f.ID,
					p,
					p.Resources,
				},
			}

			err = data.Proxy.InsertToDB(query)
			if err != nil {
				return "", err
			}

			// Determine whether we should apply a post-script
			// or not. In case the fleet is dismantled by the
			// colonization operation it is not needed.
			canReturn, err := f.canReturnFromColonization(data.Ships)
			if err != nil {
				return "", ErrUnableToSimulateFleet
			}

			if !canReturn {
				script = ""
			}
		}
	}

	return script, nil
}
