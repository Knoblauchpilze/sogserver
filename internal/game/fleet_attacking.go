package game

import (
	"fmt"
	"math"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
)

// ErrFleetFightSimulationFailure : Indicates that an error has occurred
// while simulating a fleet fight.
var ErrFleetFightSimulationFailure = fmt.Errorf("Failure to simulate fleet fight")

// attack :
// Used to perform the attack of the input planet
// by this fleet. Note that only the units that
// are deployed on the planet at this moment will
// be included in the fight.
//
// The `p` represents the element to be attacked.
// It can either be a planet or a moon.
//
// The `data` allows to access to the DB data.
//
// Return any error along with the name of the
// script to execute to finalize the execution of
// the fleet.
func (f *Fleet) attack(p *Planet, data Instance) (string, error) {
	// Convert this fleet and the planet into valid
	// `defender` and `attacker` objects so that we
	// can use them to simulate the attack.
	a, err := f.toAttacker(data)
	if err != nil {
		return "", ErrFleetFightSimulationFailure
	}

	d, err := p.toDefender(data)
	if err != nil {
		return "", ErrFleetFightSimulationFailure
	}

	result, err := d.defend(&a)
	if err != nil {
		return "", ErrFleetFightSimulationFailure
	}

	// Handle the pillage of resources if the outcome
	// says so. Note that the outcome is expressed in
	// the defender's point of view.
	pillage := f.pillage(p, data, result.outcome)

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "fleet_fight_aftermath",
		Args: []interface{}{
			a.convertShips(f.ID),
			f.TargetCoords.Type,
			d.convertShips(),
			d.convertDefenses(),
			result.debris,
			pillage,
			result.outcome,
		},
	}

	err = data.Proxy.InsertToDB(query)

	return "fleet_return_to_base", err
}

// toDefender :
// Allows to transform the planet into a defender
// structure that can be used to simulate a fleet
// fight on the planet.
//
// The `data` allows to access to the DB data.
//
// Returns the defender object built from the
// planet along with any error.
func (p *Planet) toDefender(data Instance) (defender, error) {
	d := defender{
		indigenous:     make(shipsUnit, 0),
		reinforcements: make(shipsUnit, 0),
		defenses:       make([]defenseInFight, 0),
	}

	// Fetch the fighting technologies for the player
	// owning the planet so that we can update the
	// defenses and ships weapon/shield/hull values.
	shieldID, err := data.Technologies.GetIDFromName("shielding")
	if err != nil {
		return d, err
	}

	weaponID, err := data.Technologies.GetIDFromName("weapons")
	if err != nil {
		return d, err
	}

	armourID, err := data.Technologies.GetIDFromName("armour")
	if err != nil {
		return d, err
	}

	shieldLvl := p.technologies[shieldID]
	weaponLvl := p.technologies[weaponID]
	armourLvl := p.technologies[armourID]

	shieldIncrease := float64(1.0 + shieldLvl/100.0)
	weaponIncrease := float64(1.0 + weaponLvl/100.0)
	armourIncrease := float64(1.0 + armourLvl/100.0)

	metal, err := data.Resources.GetIDFromName("metal")
	if err != nil {
		return d, err
	}
	crystal, err := data.Resources.GetIDFromName("crystal")
	if err != nil {
		return d, err
	}

	// Convert ships.
	for _, s := range p.Ships {
		shield := int(math.Round(float64(s.Shield) * shieldIncrease))
		weapon := int(math.Round(float64(s.Weapon) * weaponIncrease))

		// Compute the base hull points for this ship: it is
		// derived from the cost in metal and crystal (the
		// cost in deuterium being accounted as `energy` to
		// provide for the construction of the ship).
		costs := s.Cost.ComputeCost(1)
		hp := 0

		for res, amount := range costs {
			if res == metal || res == crystal {
				hp += amount
			}
		}

		armour := int(math.Round(float64(hp) * armourIncrease))

		sif := shipInFight{
			// Empty fleet as this ship is not part of a fleet.
			Fleet:        "",
			Ship:         s.ID,
			Count:        s.Amount,
			Cargo:        s.Cargo,
			Shield:       shield,
			Weapon:       weapon,
			Hull:         armour,
			RFVSShips:    s.RFVSShips,
			RFVSDefenses: s.RFVSDefenses,
		}

		d.indigenous = append(d.indigenous, sif)
	}

	// Convert defenses.
	for _, def := range p.Defenses {
		shield := int(math.Round(float64(def.Shield) * shieldIncrease))
		weapon := int(math.Round(float64(def.Weapon) * weaponIncrease))

		// Compute the base hull points for this defense
		// system: it is derived from the cost in metal
		// and crystal (the cost in deuterium being used
		// as `energy` to provide for the construction of
		// the defense system).
		costs := def.Cost.ComputeCost(1)
		hp := 0

		for res, amount := range costs {
			if res == metal || res == crystal {
				hp += amount
			}
		}

		armour := int(math.Round(float64(hp) * armourIncrease))

		dif := defenseInFight{
			Planet:  p.ID,
			Defense: def.ID,
			Count:   def.Amount,
			Shield:  shield,
			Weapon:  weapon,
			Hull:    armour,
		}

		d.defenses = append(d.defenses, dif)
	}

	return d, nil
}

// toAttacker :
// Allows to transform the fleet into an attacker
// structure that can be used to simulate a fleet
// fight on a planet.
//
// The `data` allows to access to the DB data.
//
// Returns the attacker object built from the
// planet along with any error.
func (f *Fleet) toAttacker(data Instance) (attacker, error) {
	a := attacker{}

	// A fleet only has a single batch of ships.
	// In order to have several attacking units
	// one should use the `ACS` mechanism.
	a.units = make([]shipsUnit, 1)
	a.units[0] = make([]shipInFight, 0)

	// Fetch the fighting technologies for the player
	// owning the planet so that we can update the
	// defenses and ships weapon/shield/hull values.
	shieldID, err := data.Technologies.GetIDFromName("shielding")
	if err != nil {
		return a, err
	}

	weaponID, err := data.Technologies.GetIDFromName("weapons")
	if err != nil {
		return a, err
	}

	armourID, err := data.Technologies.GetIDFromName("armour")
	if err != nil {
		return a, err
	}

	// The technologies need to be fetched from the
	// player owning this fleet.
	p, err := NewPlayerFromDB(f.Player, data)
	if err != nil {
		return a, err
	}

	shield := p.Technologies[shieldID]
	weapon := p.Technologies[weaponID]
	armour := p.Technologies[armourID]

	shieldIncrease := float64(1.0 + shield.Level/100.0)
	weaponIncrease := float64(1.0 + weapon.Level/100.0)
	armourIncrease := float64(1.0 + armour.Level/100.0)

	metal, err := data.Resources.GetIDFromName("metal")
	if err != nil {
		return a, err
	}
	crystal, err := data.Resources.GetIDFromName("crystal")
	if err != nil {
		return a, err
	}

	// Convert ships.
	for _, s := range f.Ships {
		sd, err := data.Ships.GetShipFromID(s.ID)
		if err != nil {
			return a, err
		}

		shield := int(math.Round(float64(sd.Shield) * shieldIncrease))
		weapon := int(math.Round(float64(sd.Weapon) * weaponIncrease))

		// Compute the base hull points for this ship: it is
		// derived from the cost in metal and crystal (the
		// cost in deuterium being accounted as `energy` to
		// provide for the construction of the ship).
		costs := sd.Cost.ComputeCost(1)
		hp := 0

		for res, amount := range costs {
			if res == metal || res == crystal {
				hp += amount
			}
		}

		armour := int(math.Round(float64(hp) * armourIncrease))

		sif := shipInFight{
			// Empty fleet as this ship is not part of a fleet.
			Fleet:        "",
			Ship:         s.ID,
			Count:        s.Count,
			Cargo:        sd.Cargo,
			Shield:       shield,
			Weapon:       weapon,
			Hull:         armour,
			RFVSShips:    sd.RFVSShips,
			RFVSDefenses: sd.RFVSDefenses,
		}

		a.units[0] = append(a.units[0], sif)
	}

	return a, nil
}

// pillage :
// Used to handle the pillage of the input
// planet by this fleet. We assume that the
// ships available to perform the pillage
// are up-to-date in this fleet and that the
// resources on the planet are up-to-date as
// well.
//
// The `p` defines the planet to pillage.
//
// The `data` defines a way to access the DB.
//
// The `result` defines the result of the
// fight of the fleet on the input planet.
// Note that this is expressed from the
// defender's point of view.
//
// Returns the resources pillaged.
func (f *Fleet) pillage(p *Planet, data Instance, result FightOutcome) []model.ResourceAmount {
	pillage := make([]model.ResourceAmount, 0)

	// If the outcome indicates that the fleet
	// could not pass the planet's defenses we
	// can't pillage anything.
	if result != Loss {
		return pillage
	}

	// TODO: Implement this.
	return pillage
}
