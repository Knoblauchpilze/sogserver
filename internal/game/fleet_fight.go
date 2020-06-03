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

// unit :
// Describes a fighting unit in a fight. It is
// used to describe either a ship or a defense
// in a simple way so that we can use it during
// a fight round.
//
// The `id` defines the type of this unit as
// defined by the data model. It can either
// represent an ID for a defense or a ship
// and it does not matter at the unit level.
//
// The `shield` defines the shield value that
// is remaining for the unit.
//
// The `weapon` value defines the weapon value
// for the unit.
//
// The `hull` defines the remaining hull point
// of this unit.
//
// The `explode` represents the value of the
// hull points below which the unit becomes
// unstable and has chances to explode.
//
// The `rf` defines the index of the RFs for
// this unit in the array attached to the def
// or attacker structure.
type unit struct {
	id      string
	shield  float32
	weapon  float32
	hull    float32
	explode float32
	rf      int
}

// Describes the rapid fire that a unit can
// have against ships and defenses.
// This structure defines a single map which
// define the RFs against ships and defenses
// indifferently.
type rf map[string]int

// attackerUnits :
// Convenience structure to handle the attacker
// units. Each attribute is a reflection of the
// attacker's base attributes.
type attackerUnits struct {
	rfs      []rf
	ships    []unit
	unitsIDs [][]int
}

// defenderUnits :
// Convenience structure to handle the defender
// units. Each attribute is a reflection of the
// defender's base attributes.
type defenderUnits struct {
	rfs               []rf
	defenses          []unit
	defensesIDs       []int
	indigenous        []unit
	indigenousIDs     []int
	reinforcements    []unit
	reinforcementsIDs []int
}

// ErrInvalidDefenderStruct : Indicates that the structure provided
// to update from defender units is not correct.
var ErrInvalidDefenderStruct = fmt.Errorf("Invalid defender structure to update from units")

// ErrInvalidAttackerStruct : Indicates that the structure provided
// to update from attacker units is not correct.
var ErrInvalidAttackerStruct = fmt.Errorf("Invalid attacker structure to update from units")

// maxCombatRounds : Indicates the maximum combat rounds
// that can occur.
var maxCombatRounds int = 6

// pillageRatio : Indicates the percentage of the resources
// of a planet that can be pillaged by an attacking fleet.
var pillageRatio float32 = 0.5

// defenseRebuilRatio : Indicates the chances for a damaged
// defense system to be rebuilt at the end of a fight.
var defenseRebuildRatio float32 = 0.7

// hullDamageToExplode : Indicates the maximum probability
// at which a ship or a defense system starts to become
// unstable and has a risk to explode.
var hullDamageToExplode float32 = 0.7

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

// convertToUnits :
// Used to perform the conversion of the data
// of this attacker into a `attackerUnits` to
// be used in a fight.
//
// Returns the created structure.
func (a *attacker) convertToUnits() attackerUnits {
	au := attackerUnits{
		rfs:      make([]rf, 0),
		ships:    make([]unit, 0),
		unitsIDs: make([][]int, len(a.units)),
	}

	count := 0
	rfs := make(map[string]int)

	for id, un := range a.units {
		au.unitsIDs[id] = make([]int, len(un))

		for _, u := range un {
			count += u.Count

			// Register the rapid fire for this ship
			// if it does not exist yet.
			_, ok := rfs[u.Ship]
			if !ok {
				rfs[u.Ship] = len(au.rfs)

				var rfObj rf
				rfObj = make(map[string]int)

				for _, r := range u.RFVSShips {
					rfObj[r.Receiver] = r.RF
				}
				for _, r := range u.RFVSDefenses {
					rfObj[r.Receiver] = r.RF
				}

				au.rfs = append(au.rfs, rfObj)
			}
		}
	}

	au.ships = make([]unit, count)
	processed := 0

	for id, un := range a.units {
		for shpID, u := range un {
			au.unitsIDs[id][shpID] = processed

			for id := 0; id < u.Count; id++ {
				rfID, ok := rfs[u.Ship]
				if !ok {
					rfID = -1
				}

				au.ships[processed+id] = unit{
					id:      u.Ship,
					shield:  float32(u.Shield),
					weapon:  float32(u.Weapon),
					hull:    float32(u.Hull),
					explode: hullDamageToExplode * float32(u.Hull),
					rf:      rfID,
				}
			}

			processed += u.Count
		}
	}

	return au
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

// convertToUnits :
// Used to perform the conversion of the data
// of this defender into a `defenderUnits` to
// be used in a fight.
//
// Returns the created structure.
func (d *defender) convertToUnits() defenderUnits {
	du := defenderUnits{
		rfs:               make([]rf, 0),
		defenses:          make([]unit, len(d.defenses)),
		defensesIDs:       make([]int, len(d.defenses)),
		indigenous:        make([]unit, len(d.indigenous)),
		indigenousIDs:     make([]int, len(d.indigenous)),
		reinforcements:    make([]unit, len(d.reinforcements)),
		reinforcementsIDs: make([]int, len(d.reinforcements)),
	}

	// Build the RFs table.
	rfs := make(map[string]int)
	for _, shp := range d.indigenous {
		_, ok := rfs[shp.Ship]
		if !ok {
			rfs[shp.Ship] = len(du.rfs)

			var rfObj rf
			rfObj = make(map[string]int)

			for _, r := range shp.RFVSShips {
				rfObj[r.Receiver] = r.RF
			}
			for _, r := range shp.RFVSDefenses {
				rfObj[r.Receiver] = r.RF
			}

			du.rfs = append(du.rfs, rfObj)
		}
	}

	for _, shp := range d.reinforcements {
		_, ok := rfs[shp.Ship]
		if !ok {
			rfs[shp.Ship] = len(du.rfs)

			var rfObj rf
			rfObj = make(map[string]int)

			for _, r := range shp.RFVSShips {
				rfObj[r.Receiver] = r.RF
			}
			for _, r := range shp.RFVSDefenses {
				rfObj[r.Receiver] = r.RF
			}

			du.rfs = append(du.rfs, rfObj)
		}
	}

	// Convert defenses.
	for id, d := range d.defenses {
		du.defensesIDs[id] = len(du.defenses)

		for i := 0; i < d.Count; i++ {
			u := unit{
				id:      d.Defense,
				shield:  float32(d.Shield),
				weapon:  float32(d.Weapon),
				hull:    float32(d.Hull),
				explode: hullDamageToExplode * float32(d.Hull),
				rf:      -1, // No RFs for defenses.
			}

			du.defenses[id] = u
		}
	}

	// Convert indigenous ships.
	for id, shp := range d.indigenous {
		du.indigenousIDs[id] = len(du.indigenous)

		for i := 0; i < shp.Count; i++ {
			rfID, ok := rfs[shp.Ship]
			if !ok {
				rfID = -1
			}

			u := unit{
				id:      shp.Ship,
				shield:  float32(shp.Shield),
				weapon:  float32(shp.Weapon),
				hull:    float32(shp.Hull),
				explode: hullDamageToExplode * float32(shp.Hull),
				rf:      rfID,
			}

			du.indigenous[id] = u
		}
	}

	// Convert reinforcement ships.
	for id, shp := range d.reinforcements {
		du.reinforcementsIDs[id] = len(du.reinforcements)

		for i := 0; i < shp.Count; i++ {
			rfID, ok := rfs[shp.Ship]
			if !ok {
				rfID = -1
			}

			u := unit{
				id:      shp.Ship,
				shield:  float32(shp.Shield),
				weapon:  float32(shp.Weapon),
				hull:    float32(shp.Hull),
				explode: hullDamageToExplode * float32(shp.Hull),
				rf:      rfID,
			}

			du.reinforcements[id] = u
		}
	}

	return du
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
	// Create the equivalent structures for the
	// attacker and the defender.
	du := d.convertToUnits()
	au := a.convertToUnits()

	// Perform the simulation of the round.
	// We want all the units of the attacker
	// to fire on the defender and then all
	// the defender units to fire on the att.
	// No units are removed until the end of
	// the round we just update the chances
	// of explosion of each unit.
	defRange := int32(len(du.defenses))
	indigenousRange := defRange + int32(len(du.indigenous))
	reinforcementsRange := indigenousRange + int32(len(du.reinforcements))

	defCount := reinforcementsRange
	attCount := int32(len(au.ships))

	reshoot := true

	// First the attacker.
	for _, u := range au.ships {
		// While the unit can shoot we will
		// continue shooting.
		for reshoot {

			id := d.rng.Int31n(defCount)

			var target *unit
			if id < defRange {
				target = &du.defenses[id]
			} else if id < indigenousRange {
				target = &du.indigenous[id]
			} else {
				target = &du.reinforcements[id]
			}

			u.fire(target, d.rng.Float32())
			reshoot = au.rf(u, target.id, d.rng.Float32())
		}
	}

	// Then defender.
	for _, def := range du.defenses {
		for reshoot {
			id := d.rng.Int31n(attCount)
			target := &au.ships[id]

			def.fire(target, d.rng.Float32())
			reshoot = du.rf(def, target.id, d.rng.Float32())
		}
	}

	for _, i := range du.indigenous {
		for reshoot {
			id := d.rng.Int31n(attCount)
			target := &au.ships[id]

			i.fire(target, d.rng.Float32())
			reshoot = du.rf(i, target.id, d.rng.Float32())
		}
	}

	for _, r := range du.reinforcements {
		for reshoot {
			id := d.rng.Int31n(attCount)
			target := &au.ships[id]

			r.fire(target, d.rng.Float32())
			reshoot = du.rf(r, target.id, d.rng.Float32())
		}
	}

	// Convert back the units and save back
	// to the defender and attacker.
	err := du.update(d)
	if err != nil {
		return err
	}

	return au.update(a)
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

// update :
// Used to perform the update of the attacker
// struct from the attacker unit block. Note
// that we assume that the attacker units obj
// is actually related to the input attacker.
//
// The `a` defines the attacker unit to update.
//
// Returns any error.
func (au attackerUnits) update(a *attacker) error {
	// Consistency.
	if len(a.units) != len(au.unitsIDs) {
		return ErrInvalidAttackerStruct
	}

	for id, u := range a.units {
		if len(u) != len(au.unitsIDs[id]) {
			return ErrInvalidAttackerStruct
		}
	}

	for id, un := range a.units {
		for shpID, u := range un {
			start := au.unitsIDs[id][shpID]
			end := au.unitsIDs[id][shpID] + u.Count

			remaining := 0

			for i := start; i < end; i++ {
				if au.ships[i].hull > 0.0 {
					remaining++
				}
			}

			u.Count = remaining

			un[shpID] = u
		}

		a.units[id] = un
	}

	return nil
}

// rf :
// Used to perform the rf simulation for the
// input unit against the provided target.
//
// The `u` defines the unit for which the rf
// simulation should be performed.
//
// The `target` defines the identifier of
// the ship/defense system to target.
//
// The `rng` defines a random number that
// can be used during the simulation.
//
// Returns `true` if the unit can reshoot.
func (au attackerUnits) rf(u unit, target string, rng float32) bool {
	// If the unit does not have a rf value
	// it never reshoots.
	if u.rf < 0 {
		return false
	}

	rfData := au.rfs[u.rf]

	prob, ok := rfData[target]
	if !ok {
		// No rapid fire against this target.
		return false
	}

	trial := (float32(prob) - 1.0) / float32(prob)

	// If the `trial` is passed the unit will reshoot.
	return rng < trial
}

// update :
// Used to perform the update of the defender
// struct from the defender unit block. Note
// that we assume that the defender units obj
// is actually related to the input defender.
//
// The `d` defines the defender unit to update.
//
// Returns any error.
func (du defenderUnits) update(d *defender) error {
	// Consistency.
	if len(d.defenses) != len(du.defensesIDs) {
		return ErrInvalidDefenderStruct
	}
	if len(d.indigenous) != len(du.indigenousIDs) {
		return ErrInvalidDefenderStruct
	}
	if len(d.reinforcements) != len(du.reinforcementsIDs) {
		return ErrInvalidDefenderStruct
	}

	// Update defenses.
	for id, def := range d.defenses {
		start := du.defensesIDs[id]
		end := du.defensesIDs[id] + def.Count

		remaining := 0

		for i := start; i < end; i++ {
			if du.defenses[i].hull > 0.0 {
				remaining++
			}
		}

		d.defenses[id].Count = remaining
	}

	// Update indigenous ships.
	for id, shp := range d.indigenous {
		start := du.indigenousIDs[id]
		end := du.indigenousIDs[id] + shp.Count

		remaining := 0

		for i := start; i < end; i++ {
			if du.indigenous[i].hull > 0.0 {
				remaining++
			}
		}

		d.indigenous[id].Count = remaining
	}

	// Update reinforcements
	for id, shp := range d.reinforcements {
		start := du.reinforcementsIDs[id]
		end := du.reinforcementsIDs[id] + shp.Count

		remaining := 0

		for i := start; i < end; i++ {
			if du.reinforcements[i].hull > 0.0 {
				remaining++
			}
		}

		d.reinforcements[id].Count = remaining
	}

	return nil
}

// rf :
// Used to perform the rf simulation for the
// input unit against the provided target.
//
// The `u` defines the unit for which the rf
// simulation should be performed.
//
// The `target` defines the identifier of
// the ship/defense system to target.
//
// The `rng` defines a random number that
// can be used during the simulation.
//
// Returns `true` if the unit can reshoot.
func (du defenderUnits) rf(u unit, target string, rng float32) bool {
	// If the unit does not have a rf value
	// it never reshoots.
	if u.rf < 0 {
		return false
	}

	rfData := du.rfs[u.rf]

	prob, ok := rfData[target]
	if !ok {
		// No rapid fire against this target.
		return false
	}

	trial := (float32(prob) - 1.0) / float32(prob)

	// If the `trial` is passed the unit will reshoot.
	return rng < trial
}

// fire :
// Used to perform the shooting of the input
// unit by the receiver unit.
//
// The `target` defines the target to shoot
// at..
//
// The `rnd` defines a random number that
// can be used during the shooting process.
func (u *unit) fire(target *unit, rnd float32) {
	// Deflect the shot if the weaponry
	// of the shooting unit is less than
	// 1% of the shield of the defense.
	if u.weapon < 0.01*target.shield {
		return
	}

	// Make the shield absorbs the shot.
	toAbsorb := target.shield - u.weapon
	target.shield = toAbsorb

	// In case the shot has been absorbed
	// we can move on.
	if target.shield >= 0.0 {
		return
	}

	// The shot penetrates the hull. We
	// need to reset the shield to `0`
	// and decrease the hull points.
	target.shield = 0.0
	target.hull += toAbsorb

	// In case the ship has now negative
	// hull points, mark it as destroyed.
	if target.hull < 0.0 {
		target.hull = 0.0

		return
	}

	// Otherwise, check whether the
	// ship explodes if the hull is
	// damaged enough.
	if target.hull < target.explode {
		// explode = iHull * hullDamageToExplode
		// <=>
		// iHull = explode / hullDamageToExplode
		// <=>
		// 1 - hull / iHull
		// <=>
		// 1 - hull / (explode / hullDamageToExplode)
		prob := 1.0 - hullDamageToExplode*target.hull/target.explode

		if rnd < prob {
			// Boom.
			target.hull = 0.0
		}
	}
}
