package game

import (
	"fmt"
	"oglike_server/internal/model"
)

// shipInFight :
// Represent a ship in a fleet fight. Compared
// to the general descritpion we only keep the
// values of the shield, weapon and cargo as
// it is used during the fight. The multipliers
// due to the research level of the player that
// own the ship are included.
//
// The `Fleet` defines to which fleet the ship
// belongs.
//
// The `Ship` defines the ID of the ship as in
// the DB.
//
// The `Count` defines the number of ships to
// engage in the fight.
//
// The `Cargo` defines the cargo space provided
// by the ship.
//
// The `Shield` defines the consolidated shield
// value of the ship. This value includes the
// research level of the player owning the ship.
//
// The `Weapon` fills a similar purpose to the
// `Shield` but for the armament value.
//
// The `RFVSShips` defines the list of rapid
// fires this ship has against other ships.
//
// The `RFVSDefenses` defines the list of RF
// this ship has against defenses.
type shipInFight struct {
	Fleet        string
	Ship         string
	Count        int
	Cargo        int
	Shield       int
	Weapon       int
	RFVSShips    []model.RapidFire
	RFVSDefenses []model.RapidFire
}

// defenseInFight :
// Fills a similar role to `shipInFight` but
// for defenses. It contains the augmented
// values for the shield and weapon abilities
// of the defense along with the planet it
// is built on.
//
// The `Planet` defines the parent planet
// where the defense is built.
//
// The `Count` defines the number of defenses
// system engaged in the fight.
//
// The `Shield` defines the augmented value
// for the shield provided by this defense
// system.
//
// The `Weapon` defines the weapon value of
// the defense system.
type defenseInFight struct {
	Planet string
	Count  int
	Shield int
	Weapon int
}

// shipsUnit :
// An attacker is just a list of ships that
// are engaged in a fight. The order of the
// list does not matter (unlike in the other
// `attacker` struct as all ships will fire
// at the same time.
type shipsUnit []shipInFight

// attacker :
// An attacker can be seen as just an ordered
// list of ships unit. The order matters as
// the resolution of the attack will be made
// in order (meaning that the first elements
// of the slice will fire first). This can be
// quite important in case some ships have a
// large number of rapid fire so we want to
// put them first in the fight (so that they
// can reduce drastically the amount of some
// specific units).
type attacker struct {
	units []shipsUnit
}

// defender :
// A defender is an aggregate composed of a
// set of defense systems and some ships.
// Just like for the `attacker` case the
// order of ships in the `fleet`
type defender struct {
	units    shipsUnit
	defenses []defenseInFight
}

// FightOutcome :
// Defines a possible outcome for a fight. The
// outcome is always expressed in terms of the
// defender. It can be either a `Victory` which
// indicates that the attacking fleet has been
// destroyed, a `Draw` in case neither the
// attacking nor the defending fleets could be
// destroyed or a `Loss` in case the defending
// fleet has been eradicated..
type FightOutcome int

// Define the possible severity level for a log message.
const (
	Victory FightOutcome = iota
	Draw
	Loss
)

// fightResult :
// Describes the outcome of a fight between an
// attacking fleet and a defender. The outcome
// regroups both the debris field that might
// have been created along with the pillaged
// resources from the planet. The repartition
// between the incoming fleets is not processed
// at this step.
//
// The `debris` defines the resources that are
// dispersed in the debris field created by
// the fight. Might be empty in case no ships
// have been destroyed.
//
// The `pillage` defines the list of resources
// that have been pillaged by the attacking
// fleet. Might be empty in case of a draw or
// a defeat of the attacking fleet.
//
// The `outcome` defines a summary of the
// fight.
type fightResult struct {
	debris  []model.ResourceAmount
	pillage []model.ResourceAmount
	outcome FightOutcome
}

// defend :
// Used to perform the fight between the
// defender against an attack from the `a`
// attacker.
//
// The `a` defines the attacker that tries
// to eradicate the defender.
//
// The return value is `nil` in case the
// fight went well and contain the summary
// of the fight.
func (d *defender) defend(a *attacker) (fightResult, error) {
	// TODO: Implement this.
	return fightResult{}, fmt.Errorf("Not implemented")
}
