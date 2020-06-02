package game

import (
	"fmt"
	"math/rand"
	"oglike_server/internal/model"
	"time"
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
// The `Hull` defines the hit points for this
// ship given the technologies researched by
// the player owning this ship.
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
	Hull         int
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
// The `Defense` defines the identifier of
// the defense system so that we can know
// which lines to update in the DB at the
// end of the fight.
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
//
// The `Hull` defines the hit points for this
// defense system given the techs researched
// by the player owning this ship.
type defenseInFight struct {
	Planet  string
	Defense string
	Count   int
	Shield  int
	Weapon  int
	Hull    int
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
//
// The `units` define the list of individual
// group of ships composing this attacker.
//
// The `usedCargo` represents the amount of
// cargo used so far in the ships composing
// the attacker.
type attacker struct {
	units     []shipsUnit
	usedCargo float32
}

// defender :
// A defender is an aggregate composed of a
// set of defense systems and some ships.
// Unlike in the case of an attacker ships
// are not sorted in any particular order.
//
// The `seed` defines the seed to use for
// the RNG used by this defender. It will
// be used in any case a RN is needed so
// that we can replay the fight afterwards
// if needed.
//
// The `rng` defines a way to access to RN
// while simulating the fight. It can be
// used throughout the fight and is never
// reset in order to make sure that the
// fight is repeatable if needed.
//
// The `indigenous` defines the ships that
// are deployed on the moon/planet where
// the fight is taking plance and belong
// to the owner of the celestial body.
//
// The `reinforcements` define all ships
// that have been sent to defend the moon
// or planet where the fight is taking
// place but are not owned by the player
// of the planet/moon.
//
// The `defenses` define the list of the
// defense systems built on the planet or
// moon.
type defender struct {
	seed           int64
	rng            *rand.Rand
	indigenous     shipsUnit
	reinforcements shipsUnit
	defenses       []defenseInFight
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

// aftermathShip :
// Used as a convenience structure to be able
// to update the ship of a planet or a fleet
// in the aftermath of a fleet fight. It only
// defines the identifier of the ship and its
// final count so that we can update the info
// in the DB.
//
// The `ID` defines the identifier of the ship.
//
// The `Count` defines the number of ships
// that remain after the fight (should be at
// least `0`).
type aftermathShip struct {
	ID    string `json:"ship"`
	Count int    `json:"count"`
}

// aftermathDefense :
// Fills a similar purpose to `aftermathShip`
// but for the defense systems of a planet.
//
// The `ID` defines the identifier of the
// defense system.
//
// The `Coun/t` defines how many defenses of
// this type remains after the fight.
type aftermathDefense struct {
	ID    string `json:"defense"`
	Count int    `json:"count"`
}

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
// The `outcome` defines a summary of the
// fight.
type fightResult struct {
	debris  []model.ResourceAmount
	outcome FightOutcome
}

// maxCombatRounds : Indicates the maximum combat rounds
// that can occur.
var maxCombatRounds int = 6

// pillageRatio : Indicates the percentage of the resources
// of a planet that can be pillaged by an attacking fleet.
var pillageRatio float32 = 0.5

// defenseRebuilRatio : Indicates the chances for a damaged
// defense system to be rebuilt at the end of a fight.
var defenseRebuildRatio float32 = 0.7

// convertShips :
// Used to convert the ships registered for
// this attacker into a marshallable struct
// that can be used to modify the content of
// the DB.
//
// The `fleet` defines the identifier of the
// fleet to which the ships should belong in
// order to be considered for the marshalling.
//
// Returns the converted interface for ships.
func (a attacker) convertShips(fleet string) interface{} {
	ships := make([]aftermathShip, 0)

	// Note that we will traverse only the units
	// that are owned by the planet itself: the
	// fleets that might have come to defend will
	// be marshalled in a different step.
	for _, unit := range a.units {
		for _, s := range unit {
			// Only consider ships belonging to the
			// input fleet.
			if s.Fleet != fleet {
				continue
			}

			d := aftermathShip{
				ID:    s.Ship,
				Count: s.Count,
			}

			ships = append(ships, d)
		}
	}

	return ships
}

// pillage :
// Used to handle the pillage of the input
// planet by the attacker. We will use the
// remaining ships to compute the available
// cargo space and pillage as many resources
// as possible.
//
// The `p` defines the planet to pillage.
//
// The `data` defines a way to access the DB.
//
// Returns the resources pillaged.
func (a attacker) pillage(p *Planet, data Instance) ([]model.ResourceAmount, error) {
	pillage := make([]model.ResourceAmount, 0)

	// Use a dedicated handler to compute the
	// result of the pillage of the target of
	// the fleet.
	pp, err := newPillagingProps(a, data.Ships)
	if err != nil {
		return pillage, err
	}

	err = pp.pillage(p, pillageRatio, data)
	if err != nil {
		return pillage, err
	}

	return pp.collected, nil
}

// newDefender :
// Used to perform the creation of a new
// defender object that can be used in a
// fleet fight.
//
// Returns the created defender object.
func newDefender() defender {
	seed := time.Now().UTC().UnixNano()

	rngSource := rand.NewSource(seed)

	return defender{
		// Use the current time to seed the RNG used
		// for the fight.
		seed:           seed,
		rng:            rand.New(rngSource),
		indigenous:     make(shipsUnit, 0),
		reinforcements: make(shipsUnit, 0),
		defenses:       make([]defenseInFight, 0),
	}
}

// convertShips :
// Used to convert the ships registered for
// this defender into a marshallable struct
// that can be used to modify the content of
// the DB.
//
// Returns the converted interface for ships.
func (d *defender) convertShips() interface{} {
	ships := make([]aftermathShip, 0)

	// Note that we will traverse only the units
	// that are owned by the planet itself: the
	// fleets that might have come to defend will
	// be marshalled in a different step.
	for _, unit := range d.indigenous {
		d := aftermathShip{
			ID:    unit.Ship,
			Count: unit.Count,
		}

		ships = append(ships, d)
	}

	return ships
}

// convertDefenses :
// Used to convert the defenses registered for
// this defender into a marshallable structure
// that can be used to modify the content of
// the DB.
//
// Returns the converted interface for defenses.
func (d *defender) convertDefenses() interface{} {

	defs := make([]aftermathDefense, 0)

	for _, def := range d.defenses {
		d := aftermathDefense{
			ID:    def.Defense,
			Count: def.Count,
		}

		defs = append(defs, d)
	}

	return defs
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
	// The process of the combat is explained
	// quite thoroughly in the following link:
	// https://ogame.fandom.com/wiki/Combat
	//
	// Basically at each round all the units
	// will randomly choose a target and fire
	// a shot, lessening the shield value and
	// the hull plating if needed.
	// Destroyed units are removed at the end
	// of the round.
	fr := fightResult{
		debris:  make([]model.ResourceAmount, 0),
		outcome: Victory,
	}

	// Save the defenses so that we can try
	// to rebuild them at the end of the
	// fight.
	initDefs := make([]defenseInFight, len(d.defenses))
	copy(initDefs, d.defenses)

	for round := 0; round < maxCombatRounds; round++ {
		err := d.round(a)
		if err != nil {
			return fr, err
		}
	}

	// Rebuilt destroyed defense systems.
	d.reconstruct(initDefs, defenseRebuildRatio)

	return fr, nil
}

// round :
// Performs the simulation of a combat round
// between the input attacker and the base
// defender.
//
// The `a` is the attacker attacking us.
//
// Returns any error.
func (d *defender) round(a *attacker) error {

	type shipInFight struct {
		Fleet        string
		Ship         string
		Count        int
		Cargo        int
		Shield       int
		Weapon       int
		Hull         int
		RFVSShips    []model.RapidFire
		RFVSDefenses []model.RapidFire
	}

	type defenseInFight struct {
		Planet  string
		Defense string
		Count   int
		Shield  int
		Weapon  int
		Hull    int
	}

	// TODO: Implement this.
	return fmt.Errorf("Not implemented")
}

// reconstruct :
// Used to perform the reconstruction of the
// damaged defense systems on a planet after
// a fight. The input slice describes initial
// state of the systems before the fight.
//
// The `init` defines the initial state of
// the defense systems before the fight.
//
// The `rebuildRatio` defines the chances for
// a destroyed defense system to be rebuilt.
func (d *defender) reconstruct(init []defenseInFight, rebuildRatio float32) {
	// We first have to determine how many of each
	// defense system has been destroyed. This will
	// condition the number of trials that we have
	// to perform to handle the reconstruction of
	// the defense systems.
	destroyed := make([]int, len(init))

	// Note that we assume that the current state of
	// the `defender` still includes all the systems
	// just with a count of `0` if everything has
	// been destroyed.
	for id, def := range d.defenses {
		des := (init[id].Count - def.Count)
		destroyed[id] = des
	}

	// Reconstruct the defense systems based on the
	// amount that was destroyed.
	for id, gone := range destroyed {
		rebuilt := 0

		for i := 0; i < gone; i++ {
			if d.rng.Float32() < rebuildRatio {
				rebuilt++
			}
		}

		d.defenses[id].Count += rebuilt
	}
}
