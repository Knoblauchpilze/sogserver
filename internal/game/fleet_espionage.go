package game

import (
	"fmt"
	"math"
	"math/rand"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// spyInfo : Convenience define to refer to the
// information level visible in a report based
// on the number of probes sent.
type spyInfo int

// Define the possible information visible in a
// spy report.
const (
	Materials spyInfo = iota
	Fleets
	Defenses
	Buildings
	Researches
)

// String :
// Implementation of the string interface for
// the `spyInfo` type.
//
// Returns the created string.
func (si spyInfo) String() string {
	switch si {
	case Materials:
		return "\"materials\""
	case Fleets:
		return "\"fleets\""
	case Defenses:
		return "\"defenses\""
	case Buildings:
		return "\"buildings\""
	case Researches:
		return "\"researches\""
	}

	return "\"unknown\""
}

// spyModule :
// Used as a convenience object to handle the
// possible otucomes of an espionage operation.
//
// The `spyEspionageTechLevel` defines the tech
// level of espionage reached by the spy.
//
// The `targetEspionageTechLevel` defines the
// tech of espionage reached by the spied.
//
// The `probes` defines the number of probes
// sent by the spy.
//
// The `ships` defines the number of ships on
// the spied planet.
type spyModule struct {
	spyEspionageTechLevel    int
	targetEspionageTechLevel int
	probes                   int
	ships                    int
}

// newSpyModule :
// Used to create a new spy module for the input
// player and target.
//
// The `fleet` defines the espionage fleet.
//
// The `target` defines the celestial body that
// is spied.
//
// The `data` allows to access information from
// the data model.
//
// Returns the spy module along with any errors.
func newSpyModule(fleet *Fleet, target *Planet, data Instance) (spyModule, error) {
	sm := spyModule{
		spyEspionageTechLevel:    0,
		targetEspionageTechLevel: 0,
		probes:                   0,
		ships:                    0,
	}

	// Fetch the espionage tech level for both the
	// spyer and the target player.
	techID, err := data.Technologies.GetIDFromName("espionage")
	if err != nil {
		return sm, err
	}

	p, err := NewPlayerFromDB(fleet.Player, data)
	if err != nil {
		return sm, err
	}

	sm.spyEspionageTechLevel = p.Technologies[techID].Level
	sm.targetEspionageTechLevel = target.technologies[techID]

	// Fetch the number of probes sent in this
	// espionage mission.
	probeID, err := data.Ships.GetIDFromName("espionage probe")
	if err != nil {
		return sm, err
	}

	for _, s := range fleet.Ships {
		if s.ID == probeID {
			sm.probes += s.Count
		}
	}

	// Gather the number of ships of the player
	// deployed on the planet. We consider that
	// defending fleets do not participate in
	// the chances of counter-espionage.
	for _, s := range target.Ships {
		sm.ships += s.Amount
	}

	return sm, nil
}

// infoLevel :
// Used to compute and retrieve the level of
// info that will be provided in the spying
// report based on the properties of this
// spy module.
//
// Returns the level of information that will
// be provided.
func (sm spyModule) infoLevel() spyInfo {
	sEspDif := sm.spyEspionageTechLevel - sm.targetEspionageTechLevel
	espDif := int(math.Abs(float64(sEspDif)))

	det := sm.probes + sEspDif*espDif

	// This information is directly extracted
	// from here:
	// https://ogame.fandom.com/wiki/Espionage
	switch {
	case det == 2:
		return Fleets
	case det < 5:
		return Defenses
	case det < 7:
		return Buildings
	case det >= 7:
		return Researches
	default:
		// Only resources visible.
		return Materials
	}
}

// counterEspionageProbability :
// Used to perform the computation of the
// probability of counter-espionage of the
// input fleet.
//
// Returns the counter-espionage probability
// in the range `[0; 1]` where `0` indicates
// no chances of counter-espionage.
func (sm spyModule) counterEspionageProbability() float32 {
	// The formula seems to be unknown as some
	// tests with similar conditions provide
	// various counter-espionage percentages.
	// The formula we use comes from this link:
	// https://ogame.fandom.com/wiki/Counterespionage
	fProbes := float64(sm.probes)
	fShips := float64(sm.ships)
	fE := float64(sm.targetEspionageTechLevel)
	fO := float64(sm.spyEspionageTechLevel)

	prob := math.Pow(2.0, fE-fO) * fShips * fProbes * 0.25 / 100.0
	prob = math.Min(1.0, math.Max(0.0, prob))

	return float32(prob)
}

// ErrFleetEspionageSimulationFailure : Indicates that an error has occurred
// while simulating an espionage operation.
var ErrFleetEspionageSimulationFailure = fmt.Errorf("failure to simulate fleet espionage")

// spy :
// Used to perform a spying operation on the planet
// by the fleet. This can lead to a fight in case
// the fleet is spotted.
//
// The `p` represents the planet to spy.
//
// The `data` allows to access information from the
// data model.
//
// Return any error along with the name of the
// script to execute to finalize the execution of
// the fleet.
func (f *Fleet) spy(p *Planet, data Instance) (string, error) {
	// Spying a planet is a way to gather intelligence
	// on a planet. Depending on the level of the tech
	// espionage of both the player sending the fleet
	// and the player owning the planet some info will
	// or will not be available.
	// The number of probes also changes that.
	// Finally the number of ships deployed on the
	// planet and the number of probes determine the
	// chances of counter-espionage which can lead to
	// a fight between the ships deployed at the dest
	// of the fleet and the fleet itself.

	// The process is only relevant is the fleet is
	// not already returning to its base.
	if !f.returning {
		sm, err := newSpyModule(f, p, data)
		if err != nil {
			return "", ErrFleetEspionageSimulationFailure
		}

		// Determine the information level provided by the
		// spying operation.
		il := sm.infoLevel()

		data.log.Trace(logger.Verbose, "spy", fmt.Sprintf("Spy level for fleet \"%s\" is %s", f.ID, il))

		// Compute the counter-espionage probability.
		ce := sm.counterEspionageProbability()
		source := rand.NewSource(f.ArrivalTime.UnixNano())
		rng := rand.New(source)

		// Notify an espionage report in the DB.
		query := db.InsertReq{
			Script: "espionage_report",
			Args: []interface{}{
				f.ID,
				int(math.Round(float64(ce) * 100.0)),
				il,
			},
		}

		err = data.Proxy.InsertToDB(query)
		if err != nil {
			return "", ErrFleetFightSimulationFailure
		}

		if rng.Float32() <= ce {
			// The fleet is detected, proceed to a fight
			// between the attacker fleet and the target
			// defense.
			return f.attack(p, data)
		}
	}

	return "fleet_return_to_base", nil
}
