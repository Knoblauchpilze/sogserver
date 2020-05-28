package game

import (
	"math"
	"math/rand"
	"oglike_server/pkg/db"
)

// destroy :
// Used to attempt to destroy the moon in argument
// by this fleet. This may fail and result in the
// total destruction of the fleet.
//
// The `m` represents the moon to destroy.
//
// The `data` defines a way to access to the DB.
//
// Return any error along with the name of the
// script to execute to finalize the execution of
// the fleet.
func (f *Fleet) destroy(m *Planet, data Instance) (string, error) {
	// The destruction operation is first performing
	// a regular attack of the fleet and defense on
	// the planet and then performing the destruction
	// attempt.
	// The formula governing the chances of the fleet
	// or the moon to get destroyed are taken from:
	// https://ogame.fandom.com/wiki/Moon#Formulae

	// Consistency.
	if m.Coordinates.Type != Moon {
		return "", ErrUnableToSimulateFleet
	}

	// First attack the moon.
	script, err := f.attack(m, data)
	if err != nil {
		return script, err
	}

	// Check whether there are some remaining ships
	// to perform the destruction.
	dsInfo, err := data.Ships.GetIDFromName("deathstar")
	if err != nil {
		return "", ErrUnableToSimulateFleet
	}

	ds, ok := f.Ships[dsInfo]

	// Perform the destruction in case there are
	// still some deathstars available.
	if ok {
		// Compute the chance to destroy the moon
		// and the fleet.
		moonDestroyChance := (100.0 - math.Sqrt(float64(m.Diameter))) * math.Sqrt(float64(ds.Count))
		dsDestroyChance := 0.5 * math.Sqrt(float64(m.Diameter))

		source := rand.NewSource(int64(m.Coordinates.generateSeed()))
		rng := rand.New(source)
		moonDestroyedProb := rng.Float64()
		dsDestroyedProb := rng.Float64()

		moonDestroyed := (moonDestroyedProb < moonDestroyChance)
		dsDestroyed := (dsDestroyedProb < dsDestroyChance)

		// Notify the result of the destruction both
		// for the fleet and the moon.
		query := db.InsertReq{
			Script: "fleet_destroy",
			Args: []interface{}{
				f.ID,
				moonDestroyed,
				dsDestroyed,
			},
		}

		err = data.Proxy.InsertToDB(query)
		if err != nil {
			return "", err
		}
	}

	return "fleet_return_to_base", nil
}
