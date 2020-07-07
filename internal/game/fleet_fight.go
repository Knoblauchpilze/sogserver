package game

import (
	"fmt"
	"math"
	"math/rand"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
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
// The `participants` defines the list of
// players registered in this attack.
//
// The `fleets` represents the list of IDs
// of fleets that are part of the fight as
// reinforcement on the defender.
//
// The `units` define the list of individual
// group of ships composing this attacker.
//
// The `usedCargo` represents the amount of
// cargo used so far in the ships composing
// the attacker.
//
// The `log` allows to notify information
// during the simulation of the fight.
type attacker struct {
	participants []string
	fleets       []string
	units        []shipsUnit
	usedCargo    float32
	log          logger.Logger
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
// The `mainDef` defines the identifier
// of the main defender in the fight. it
// corresponds to the player owning the
// planet/moon where the fight is taking
// place.
//
// The `location` defines the identifier
// of the location where the fight is
// taking place: can be either an ID of
// a planet or a moon.
//
// The `moon` is `true` if the `location`
// refers to a moon.
//
// The `indigenous` defines the ships that
// are deployed on the moon/planet where
// the fight is taking plance and belong
// to the owner of the celestial body.
//
// The `participants` defines the list
// of all players participating to the
// defense of the planet/moon. Should
// correspond to the players owning a
// fleet in the `fleets` array.
//
// The `fleets` represents the list of IDs
// of fleets that are part of the fight as
// reinforcement on the defender.
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
//
// The `shipsToRuins` defines the ratio
// of the ships' construction costs that
// end up in the debris field in case a
// ship explodes.
//
// The `defensesToRuins` plays a similar
// role for defenses.
//
// The `debris` allow to build the debris
// field generated by the fight along the
// way of simulating it.
//
// The `log` allows to notify information
// during the simulation of the fight.
//
// The `rountCound` defines the current
// fight round for the defender.
type defender struct {
	seed            int64
	rng             *rand.Rand
	mainDef         string
	location        string
	moon            bool
	indigenous      shipsUnit
	participants    []string
	fleets          []string
	reinforcements  shipsUnit
	defenses        []defenseInFight
	shipsToRuins    float32
	defensesToRuins float32
	debris          map[string]model.ResourceAmount
	log             logger.Logger
	roundCount      int
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

// Define the possible fight outcomes.
const (
	Victory FightOutcome = iota
	Draw
	Loss
)

// String :
// Implementation of the stringer interface in
// order to provide human-readable messages.
//
// Returns the string corresponding to the `fo`
// outcome.
func (fo FightOutcome) String() string {
	switch fo {
	case Victory:
		return "victory"
	case Draw:
		return "draw"
	case Loss:
		return "loss"
	}

	return "\"unknown\""
}

// aftermathShip :
// Allow to count ships belonging to each
// fleet after a fight: it helps building
// the report for each participant to a
// fleet fight.
//
// The `Fleet` defines the fleet to which
// the ship belongs.
//
// The `ID` defines the identifier of the
// ship.
//
// The `Count` defines how many of this
// ships remain after the fight.
type aftermathShip struct {
	Fleet string `json:"fleet"`
	ID    string `json:"ship"`
	Count int    `json:"count"`
}

// aftermathDefense :
// Fills a similar purpose to `ShipInFleet`
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
// The `date` defines the moment in time at
// which the fight took place.
//
// The `debris` defines the resources that are
// dispersed in the debris field created by
// the fight. Might be empty in case no ships
// have been destroyed.
//
// The `outcome` defines a summary of the
// fight.
//
// The `moon` defines whether or not a moon was
// created during the fight.
//
// The `diameter` defines the diameter of the
// moon that has been created by this fight.
//
// The `rebuilt` defines how many defense
// systems have been rebuilt after the fight.
type fightResult struct {
	date     time.Time
	debris   []model.ResourceAmount
	outcome  FightOutcome
	moon     bool
	diameter int
	rebuilt  int
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

// minMoonDiameter : Defines the minimum diameter for a moon.
var minMoonDiameter float32 = 3464.0

// maxMoonDiameter : Defines the maximum diameter for a moon.
var maxMoonDiameter float32 = 8944.0

// convertShipsForFleet :
// Used to convert the ships registered for
// this attacker into a marshallable struct
// that can be used to modify the content of
// the DB.
//
// The `fleet` defines the identifier of the
// fleet to which the ships should belong in
// order to be considered for the marshalling.
// In case it is empty all ships will be
//
// Returns the converted interface for ships.
func (a attacker) convertShipsForFleet(fleet string) interface{} {
	ships := make([]ShipInFleet, 0)

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

			// Use the case where all ships have been
			// destroyed to update the corresponding
			// entry in the DB for this fleet.
			sif := ShipInFleet{
				ID:    s.Ship,
				Count: s.Count,
			}

			ships = append(ships, sif)
		}
	}

	return ships
}

// convertShips :
// Used to convert the ships registered for
// this attacker into a marshallable struct
// that can be used to modify the content of
// the DB. Unlike the above method no fleet
// is considered in particular.
//
// The `fleet` defines the identifier of the
// fleet to which the ships should belong in
// order to be considered for the marshalling.
// In case it is empty all ships will be
//
// Returns the converted interface for ships.
func (a attacker) convertShips() []aftermathShip {
	ships := make([]aftermathShip, 0)

	// Note that we will traverse only the units
	// that are owned by the planet itself: the
	// fleets that might have come to defend will
	// be marshalled in a different step.
	for _, unit := range a.units {
		for _, s := range unit {
			sif := aftermathShip{
				Fleet: s.Fleet,
				ID:    s.Ship,
				Count: s.Count,
			}

			ships = append(ships, sif)
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
// The `seed` defines the base for the
// RNG to use for the fight.
//
// The `uni` defines the identifier of
// the universe associated to this def.
// It helps fetching information about
// the percentage of units going into
// debris.
//
// The `data` allows to access to the DB.
//
// Returns the created defender object
// along with any error.
func newDefender(seed int64, uni string, data Instance) (defender, error) {
	rngSource := rand.NewSource(seed)

	d := defender{
		// Use the current time to seed the RNG used
		// for the fight.
		seed:           seed,
		rng:            rand.New(rngSource),
		indigenous:     make(shipsUnit, 0),
		participants:   make([]string, 0),
		fleets:         make([]string, 0),
		reinforcements: make(shipsUnit, 0),
		defenses:       make([]defenseInFight, 0),
		debris:         make(map[string]model.ResourceAmount),
		log:            data.log,
		roundCount:     0,
	}

	// Fetch the multipliers of the parent
	// universe.
	mul, err := NewMultipliersFromDB(uni, data)
	if err != nil {
		return d, err
	}

	d.shipsToRuins = mul.ShipsToRuins
	d.defensesToRuins = mul.DefensesToRuins

	return d, nil
}

// convertShips :
// Used to convert the ships registered for
// this defender into a marshallable struct
// that can be used to modify the content of
// the DB.
//
// Returns the converted interface for ships.
func (d *defender) convertShips() []ShipInFleet {
	ships := make([]ShipInFleet, 0)

	// Note that we will traverse only the units
	// that are owned by the planet itself: the
	// fleets that might have come to defend will
	// be marshalled in a different step.
	for _, unit := range d.indigenous {
		// We want to consider all the ships even the
		// ones that don't have any remaining element
		// because it will be stored in the DB and be
		// used to reduce the total if all ships have
		// been destroyed.
		def := ShipInFleet{
			ID:    unit.Ship,
			Count: unit.Count,
		}

		ships = append(ships, def)
	}

	return ships
}

// convertShipsForFleet :
// Used to convert the ships registered for
// the input fleet as reinforcements of the
// defender. The content is returned into a
// structure that can be easily marshalled.
//
// The `fleet` defines the identifier of the
// fleet for which reinforcement ships sent
// to defend the planet should be fetched.
//
// Returns the converted interface for ships.
func (d *defender) convertShipsForFleet(fleet string) interface{} {
	ships := make([]ShipInFleet, 0)

	// Note that we will traverse only the units
	// that are owned by the planet itself: the
	// fleets that might have come to defend will
	// be marshalled in a different step.
	for _, s := range d.reinforcements {
		// Only consider ships belonging to the
		// input fleet.
		if s.Fleet != fleet {
			continue
		}

		ship := ShipInFleet{
			ID:    s.Ship,
			Count: s.Count,
		}

		ships = append(ships, ship)
	}

	return ships
}

// convertReinforcementShips :
// Used to convert the defender ships sent
// to protect the defender into a single
// array registering the ships and their
// parent fleet.
//
// Returns the converted interface for ships.
func (d defender) convertReinforcementShips() []aftermathShip {
	ships := make([]aftermathShip, 0)

	// Note that we will traverse only the units
	// that are owned by the planet itself: the
	// fleets that might have come to defend will
	// be marshalled in a different step.
	for _, s := range d.reinforcements {
		sif := aftermathShip{
			Fleet: s.Fleet,
			ID:    s.Ship,
			Count: s.Count,
		}

		ships = append(ships, sif)
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
		// Similarly to the `convertShips` method
		// we will use the case where a defense
		// does not have any remaining system on
		// the planet to update the count in the
		// DB so we will consider this case.
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
		defenses:          make([]unit, 0),
		defensesIDs:       make([]int, len(d.defenses)),
		indigenous:        make([]unit, 0),
		indigenousIDs:     make([]int, len(d.indigenous)),
		reinforcements:    make([]unit, 0),
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

		curDef := make([]unit, d.Count)

		for i := 0; i < d.Count; i++ {
			u := unit{
				id:      d.Defense,
				shield:  float32(d.Shield),
				weapon:  float32(d.Weapon),
				hull:    float32(d.Hull),
				explode: hullDamageToExplode * float32(d.Hull),
				rf:      -1, // No RFs for defenses.
			}

			curDef[i] = u
		}

		du.defenses = append(du.defenses, curDef...)
	}

	// Convert indigenous ships.
	for id, shp := range d.indigenous {
		du.indigenousIDs[id] = len(du.indigenous)

		curShp := make([]unit, shp.Count)

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

			curShp[i] = u
		}

		du.indigenous = append(du.indigenous, curShp...)
	}

	// Convert reinforcement ships.
	for id, shp := range d.reinforcements {
		du.reinforcementsIDs[id] = len(du.reinforcements)

		curShp := make([]unit, shp.Count)

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

			curShp[i] = u
		}
		du.reinforcements = append(du.reinforcements, curShp...)
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
// The `moment` defines when the fight is
// actually taking place. It is used as a
// way to ensure that various computations
// always use the same timestamp throughout
// the fight.
//
// The `data` defines a way to access to
// the ships and defense systems information
// from the DB.
//
// The return value is `nil` in case the
// fight went well and contain the summary
// of the fight.
func (d *defender) defend(a *attacker, moment time.Time, data Instance) (fightResult, error) {
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
		date:     moment,
		debris:   make([]model.ResourceAmount, 0),
		outcome:  Victory,
		moon:     false,
		diameter: 0,
		rebuilt:  0,
	}

	// Save the defenses so that we can try
	// to rebuild them at the end of the
	// fight.
	initDefs := make([]defenseInFight, len(d.defenses))
	copy(initDefs, d.defenses)

	round := 0
	defDestroyed := false
	attDestroyed := false
	var err error
	for ; round < maxCombatRounds && !defDestroyed && !attDestroyed; round++ {
		defDestroyed, attDestroyed, err = d.round(a, data)

		if err != nil {
			return fr, err
		}
	}

	// Update the result of the fight based on
	// whether the defender was destroyed and
	// the number of rounds.
	// Indeed the possible outcomes are:
	//  1. def destroyed, att destroyed
	//  2. def destroyed, att not destroyed
	//  3. def not destroyed, att destroyed
	//  4. def not destroyed, att not destroyed
	//
	// Case 1. is a draw, case 2. is a loss for
	// the defender, case 3. is a win for the
	// defender and case 4. is a draw.
	if defDestroyed && !attDestroyed {
		fr.outcome = Loss
	} else if !defDestroyed && attDestroyed {
		fr.outcome = Victory
	} else {
		// Case 1 and 4.
		fr.outcome = Draw
	}

	d.log.Trace(logger.Verbose, "fight", fmt.Sprintf("Fight at %v took %d round(s)", time.Unix(0, d.seed), round))
	d.log.Trace(logger.Verbose, "fight", fmt.Sprintf("Result of fight at %v is \"%s\"", time.Unix(0, d.seed), fr.outcome))

	// Assign the debris field computed during
	// the simulation of the fight. Also take
	// into account the creation of a moon.
	tot := float32(0.0)

	for _, res := range d.debris {
		fr.debris = append(fr.debris, res)

		tot += res.Amount
	}

	moonChance := (tot / 100000.0) / 100.0
	moonChance = float32(math.Min(float64(moonChance), 0.2))
	getLucky := d.rng.Float32()

	if getLucky > moonChance {
		d.log.Trace(logger.Verbose, "fight", fmt.Sprintf("Failed to generate moon (out of %f)", moonChance))
	} else {
		fr.moon = true

		// The diameter is computed based on the maximum
		// size of the moon and the minimum size with a
		// linear progression between both and based on
		// the actual moon chance.
		interval := maxMoonDiameter - minMoonDiameter

		fDiameter := math.Round(float64(minMoonDiameter + interval*getLucky/moonChance))
		fDiameter = math.Min(math.Max(float64(minMoonDiameter), fDiameter), float64(maxMoonDiameter))

		fr.diameter = int(fDiameter)

		d.log.Trace(logger.Verbose, "fight", fmt.Sprintf("Created moon with diameter %d (%f chances out of %f)", fr.diameter, getLucky, moonChance))
	}

	// Rebuilt destroyed defense systems.
	fr.rebuilt = d.reconstruct(initDefs, defenseRebuildRatio, data)

	return fr, nil
}

// round :
// Performs the simulation of a combat round
// between the input attacker and the base
// defender.
//
// The `a` is the attacker attacking us.
//
// The `data` defines a way to access to
// the ships and defense systems information
// from the DB.
//
// Returns any error along with a boolean
// indicating whether the defender or the
// attacker has been destroyed during this
// round.
func (d *defender) round(a *attacker, data Instance) (bool, bool, error) {
	d.roundCount++

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

	// In case the defender does not have any
	// unit to defend, the attacker is winning.
	// Similarly, if the attacker does not have
	// any attacking units left, the defender
	// is winning.
	if defCount == 0 || attCount == 0 {
		return defCount == 0, attCount == 0, nil
	}

	// First the attacker.
	for _, u := range au.ships {
		reshoot := true

		// While the unit can shoot we will
		// continue shooting.
		for reshoot {

			id := d.rng.Int31n(defCount)

			var target *unit
			if id < defRange {
				target = &du.defenses[id]
			} else if id < indigenousRange {
				target = &du.indigenous[id-defRange]
			} else {
				target = &du.reinforcements[id-indigenousRange]
			}

			alreadyDead := (target.hull <= 0.0)

			u.fire(target, d.rng.Float32())
			reshoot = au.rf(u, target.id, d.rng.Float32())

			if target.hull <= 0.0 && !alreadyDead {
				if id < defRange {
					dd, err := data.Defenses.GetDefenseFromID(target.id)
					if err == nil {
						d.log.Trace(logger.Verbose, "fight", fmt.Sprintf("Destroyed 1 \"%s\"", dd.Name))
					}
				} else {
					sd, err := data.Ships.GetShipFromID(target.id)
					if err == nil {
						d.log.Trace(logger.Verbose, "fight", fmt.Sprintf("Destroyed 1 \"%s\"", sd.Name))
					}
				}
			}
		}
	}

	// Then defender.
	for _, def := range du.defenses {
		reshoot := true

		for reshoot {
			id := d.rng.Int31n(attCount)
			target := &au.ships[id]

			def.fire(target, d.rng.Float32())
			reshoot = du.rf(def, target.id, d.rng.Float32())

			if target.hull <= 0.0 {
				sd, err := data.Ships.GetShipFromID(target.id)
				if err == nil {
					d.log.Trace(logger.Verbose, "fight", fmt.Sprintf("Destroyed 1 \"%s\"", sd.Name))
				}
			}
		}
	}

	for _, i := range du.indigenous {
		reshoot := true

		for reshoot {
			id := d.rng.Int31n(attCount)
			target := &au.ships[id]

			i.fire(target, d.rng.Float32())
			reshoot = du.rf(i, target.id, d.rng.Float32())

			if target.hull <= 0.0 {
				sd, err := data.Ships.GetShipFromID(target.id)
				if err == nil {
					d.log.Trace(logger.Verbose, "fight", fmt.Sprintf("Destroyed 1 \"%s\"", sd.Name))
				}
			}
		}
	}

	for _, r := range du.reinforcements {
		reshoot := true

		for reshoot {
			id := d.rng.Int31n(attCount)
			target := &au.ships[id]

			r.fire(target, d.rng.Float32())
			reshoot = du.rf(r, target.id, d.rng.Float32())

			if target.hull <= 0.0 {
				sd, err := data.Ships.GetShipFromID(target.id)
				if err == nil {
					d.log.Trace(logger.Verbose, "fight", fmt.Sprintf("Destroyed 1 \"%s\"", sd.Name))
				}
			}
		}
	}

	// For each unit that has been destroyed
	// we need to increase the debris field.
	err := d.addShipsDebris(au.ships, data)
	if err != nil {
		return false, false, err
	}

	for _, def := range du.defenses {
		if def.hull <= 0.0 {
			ds, err := data.Defenses.GetDefenseFromID(def.id)

			if err != nil {
				return false, false, err
			}

			for res, amount := range ds.Cost.InitCosts {
				ex, ok := d.debris[res]

				// Make sure that the resources can be
				// dispersed in a debris field.
				r, err := data.Resources.GetResourceFromID(res)
				if err != nil {
					return false, false, err
				}

				if !r.Dispersable {
					continue
				}

				deb := d.defensesToRuins * float32(amount)

				if ok {
					ex.Amount += deb
				} else {
					ex = model.ResourceAmount{
						Resource: res,
						Amount:   deb,
					}
				}

				d.debris[res] = ex
			}
		}
	}

	err = d.addShipsDebris(du.indigenous, data)
	if err != nil {
		return false, false, err
	}

	err = d.addShipsDebris(du.reinforcements, data)
	if err != nil {
		return false, false, err
	}

	msg := fmt.Sprintf("Debris after round %d are: ", d.roundCount)
	for _, res := range d.debris {
		r, err := data.Resources.GetResourceFromID(res.Resource)
		if err == nil {
			msg = fmt.Sprintf("%s %f %s", msg, res.Amount, r.Name)
		}
	}
	d.log.Trace(logger.Verbose, "fight", msg)

	// Convert back the units and save back
	// to the defender and attacker.
	defDestroyed, err := du.update(d)
	if err != nil {
		return defDestroyed, false, err
	}

	attDestroyed, err := au.update(a)
	return defDestroyed, attDestroyed, err
}

// addDebris :
// Used to scan the input slice and add the
// destroyed units to the debris field of
// this defender based on the input ratio.
//
// The `units` are the units to scan for a
// collection of destroyed units.
//
// The `data` allows to access properties
// of ships and resources.
//
// Returns any error and the total amount
// of debris generated.
func (d *defender) addShipsDebris(units []unit, data Instance) error {
	// Traverse the input list of units.
	for _, u := range units {
		// If the unit is destroyed add it to
		// the debris field.
		if u.hull <= 0.0 {
			desc, err := data.Ships.GetShipFromID(u.id)

			if err != nil {
				return err
			}

			for res, amount := range desc.Cost.InitCosts {
				ex, ok := d.debris[res]

				// Make sure that the resources can be
				// dispersed in a debris field.
				r, err := data.Resources.GetResourceFromID(res)
				if err != nil {
					return err
				}

				if !r.Dispersable {
					continue
				}

				deb := d.shipsToRuins * float32(amount)

				if ok {
					ex.Amount += deb
				} else {
					ex = model.ResourceAmount{
						Resource: res,
						Amount:   deb,
					}
				}

				d.debris[res] = ex
			}
		}
	}

	return nil
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
//
// The `data` allows to notify information
// about the rebuilt defense systems to the
// user.
//
// Returns the amount of defense systmes that
// have been rebuilt.
func (d *defender) reconstruct(init []defenseInFight, rebuildRatio float32, data Instance) int {
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

	reconstructed := 0

	// Reconstruct the defense systems based on the
	// amount that was destroyed.
	for id, gone := range destroyed {
		if gone <= 0 {
			continue
		}

		rebuilt := 0

		for i := 0; i < gone; i++ {
			if d.rng.Float32() < rebuildRatio {
				rebuilt++
			}
		}

		dd, err := data.Defenses.GetDefenseFromID(d.defenses[id].Defense)
		if err == nil {
			d.log.Trace(logger.Verbose, "fight", fmt.Sprintf("Rebuilt %d \"%s\" (%d destroyed)", rebuilt, dd.Name, gone))
		}

		d.defenses[id].Count += rebuilt

		reconstructed += rebuilt
	}

	return reconstructed
}

// generateReports :
// Used to perform the generation of the reports after
// a fight for all participants. It will analyze the
// result of the fight so that a comprehensive report
// is generated.
//
// The `a` defines the attacker that was involved in
// the attack. It contains the remains of the attacking
// ships.
//
// The `d` defines the remains of the defender fleet
// for this fight.
//
// The `fr` defines the final result of the fight.
//
// The `pillage` defines the result of the pillage
// performed by the attacker. Might be empty.
//
// The `proxy` allows to perform queries on the DB.
//
// Returns any error.
func (d *defender) generateReports(a *attacker, fr fightResult, pillage []model.ResourceAmount, proxy db.Proxy) error {
	// We need to generate a report for the attacker and
	// one for each defender. Each report is divided into
	// several parts:
	//  1. the header
	//  2. the participants
	//  3. the status
	//  4. the footer
	//  5. the final report
	// Failure to generate any part of any of the reports
	// will be reported. For convenience we will use the
	// dedicated DB script which will make sure that no
	// report can be generated incompletely.
	// Note that the structure of the report was taken
	// from this link:
	// https://lng.xooit.com/t1488-Mettre-en-page-un-RC-avec-ogame-winner.htm
	// Which provide useful information in the case of a
	// single attacker and defender. We extrapolated for
	// the case of an ACS operation or when some defenders
	// are also participating to the fight.

	// Convert the remaining ships of both the attackers
	// and the defenders.
	remains := a.convertShips()
	defenderRemains := d.convertReinforcementShips()

	remains = append(remains, defenderRemains...)

	// Create the involved fleets and players.
	type reportPlayer struct {
		Player string `json:"player"`
	}
	type reportFleet struct {
		Fleet string `json:"fleet"`
	}

	players := make([]reportPlayer, 0)
	for _, p := range a.participants {
		rp := reportPlayer{
			Player: p,
		}

		players = append(players, rp)
	}

	for _, p := range d.participants {
		rp := reportPlayer{
			Player: p,
		}

		players = append(players, rp)
	}

	attackingFleets := make([]reportFleet, 0)
	for _, f := range a.fleets {
		rf := reportFleet{
			Fleet: f,
		}

		attackingFleets = append(attackingFleets, rf)
	}

	defendingFleets := make([]reportFleet, 0)
	for _, f := range d.fleets {
		rf := reportFleet{
			Fleet: f,
		}

		defendingFleets = append(defendingFleets, rf)
	}

	kind := "planet"
	if d.moon {
		kind = "moon"
	}

	// TODO: We should somehow split the players
	// into attackers and defenders.

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "fight_report",
		Args: []interface{}{
			players,
			attackingFleets,
			defendingFleets,
			d.mainDef,
			d.location,
			kind,
			fr.date,
			fmt.Sprintf("%s", fr.outcome),
			remains,
			d.convertShips(),
			d.convertDefenses(),
			pillage,
			fr.debris,
			fr.rebuilt,
		},
	}

	return proxy.InsertToDB(query)
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

// update :
// Used to perform the update of the attacker
// struct from the attacker unit block. Note
// that we assume that the attacker units obj
// is actually related to the input attacker.
//
// The `a` defines the attacker unit to update.
//
// Returns any error and a boolean indicating
// whether all the units for the attacker have
// been destroyed.
func (au attackerUnits) update(a *attacker) (bool, error) {
	// Consistency.
	if len(a.units) != len(au.unitsIDs) {
		return false, ErrInvalidAttackerStruct
	}

	for id, u := range a.units {
		if len(u) != len(au.unitsIDs[id]) {
			return false, ErrInvalidAttackerStruct
		}
	}

	destroyed := true

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

			if destroyed && remaining > 0 {
				destroyed = false
			}

			u.Count = remaining
			un[shpID] = u
		}

		a.units[id] = un
	}

	return destroyed, nil
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
// Returns any error and a boolean indicating
// whether all the units for the defender have
// been destroyed.
func (du defenderUnits) update(d *defender) (bool, error) {
	// Consistency.
	if len(d.defenses) != len(du.defensesIDs) {
		return false, ErrInvalidDefenderStruct
	}
	if len(d.indigenous) != len(du.indigenousIDs) {
		return false, ErrInvalidDefenderStruct
	}
	if len(d.reinforcements) != len(du.reinforcementsIDs) {
		return false, ErrInvalidDefenderStruct
	}

	destroyed := true

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

		if destroyed && remaining > 0 {
			destroyed = false
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

		if destroyed && remaining > 0 {
			destroyed = false
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

		if destroyed && remaining > 0 {
			destroyed = false
		}

		d.reinforcements[id].Count = remaining
	}

	return destroyed, nil
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
