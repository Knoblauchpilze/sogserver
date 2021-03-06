package game

import (
	"fmt"
	"math"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"time"
)

// ErrFleetFightSimulationFailure : Indicates that an error has occurred
// while simulating a fleet fight.
var ErrFleetFightSimulationFailure = fmt.Errorf("failure to simulate fleet fight")

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
	// Make sure that the fleet is not returning to
	// its parent planet already.
	if f.returning {
		return "fleet_return_to_base", nil
	}

	// Convert this fleet and the planet into valid
	// `defender` and `attacker` objects so that we
	// can use them to simulate the attack.
	a, err := f.toAttacker(data)
	if err != nil {
		return "", ErrFleetFightSimulationFailure
	}

	d, err := p.toDefender(data, f.ArrivalTime)
	if err != nil {
		return "", ErrFleetFightSimulationFailure
	}

	result, err := d.defend(&a, f.ArrivalTime, data)
	if err != nil {
		return "", ErrFleetFightSimulationFailure
	}

	// Before handling the pillage we need to make
	// sure that any resources transported by the
	// fleet can still be transported. Indeed in
	// case a player choose to send resources when
	// attacking someone, there's a risk that the
	// destruction of some ships decrease the cargo
	// space available making it impossible to be
	// carrying the initial resources.
	enough := f.handleDumbMove(a)
	carried := make([]model.ResourceAmount, 0)
	pillaged := make([]model.ResourceAmount, 0)

	for _, res := range f.Cargo {
		carried = append(carried, res)
	}

	// We only need to handle pillaging in case the
	// dumb move handling function did not report
	// that the cargo capacity was not sufficient.
	if enough {
		// Handle the pillage of resources if the outcome
		// says so. Note that the outcome is expressed in
		// the defender's point of view.
		if result.outcome == Loss {
			pillaged, err = a.pillage(p, data)
			if err != nil {
				return "", ErrFleetFightSimulationFailure
			}

			// Append the pillaged resources to the cargo
			// carried by the fleet.
			pm := make(map[string]model.ResourceAmount)

			for _, r := range pillaged {
				e, ok := pm[r.Resource]

				if ok {
					e.Amount += r.Amount
				} else {
					e = r
				}

				pm[r.Resource] = e
			}

			for _, r := range carried {
				e, ok := pm[r.Resource]

				if ok {
					e.Amount += r.Amount
				} else {
					e = r
				}

				pm[r.Resource] = e
			}

			carried = make([]model.ResourceAmount, 0)
			for _, r := range pm {
				carried = append(carried, r)
			}
		}
	}

	// Post fight reports: we need one for each
	// participant and a special one for the owner
	// of the planet where the fight took place.
	err = d.generateReports(&a, result, pillaged, data.Proxy)
	if err != nil {
		return "", ErrFleetFightSimulationFailure
	}

	// Update the planet's data in the DB.
	query := db.InsertReq{
		Script: "planet_fight_aftermath",
		Args: []interface{}{
			p.ID,
			// We need to convert this arg to a string
			// because otherwise it gets marshalled as
			// `"planet"` which is different from the
			// expected `planet`: we can't rely on the
			// `Convertible` interface neither because
			// the process is to return a struct to be
			// marshalled so even if we were to return
			// a raw string it would still be packed
			// as `"planet"`.
			string(p.Coordinates.Type),
			d.convertShips(),
			d.convertDefenses(),
			pillaged,
			result.debris,
			result.moon,
			result.diameter,
			result.date,
		},
	}

	err = data.Proxy.InsertToDB(query)
	if err != nil {
		return "", ErrFleetFightSimulationFailure
	}

	// Update the reinforcements' data in the DB.
	// As we know that the reinforcements can't
	// be carrying resources, we will provide an
	// empty `pillage` slice.
	emptyRes := make([]model.ResourceAmount, 0)

	for _, fID := range d.fleets {
		// Execute the query to update the fleet.
		query = db.InsertReq{
			Script: "fleet_fight_aftermath",
			Args: []interface{}{
				fID,
				d.convertShipsForFleet(fID),
				emptyRes,
				result.outcome.String(),
			},
		}

		err = data.Proxy.InsertToDB(query)
		if err != nil {
			return "", ErrFleetFightSimulationFailure
		}
	}

	// Update the fleet's data in the DB.
	query = db.InsertReq{
		Script: "fleet_fight_aftermath",
		Args: []interface{}{
			f.ID,
			a.convertShipsForFleet(f.ID),
			carried,
			result.outcome.String(),
		},
	}

	err = data.Proxy.InsertToDB(query)

	return "", err
}

// toDefender :
// Allows to transform the planet into a defender
// structure that can be used to simulate a fleet
// fight on the planet.
//
// The `data` allows to access to the DB data.
//
// The `moment` defines the time at which the
// conversion should be performed. It helps to
// determine which fleet can be included in the
// reinforcements for the planet.
//
// Returns the defender object built from the
// planet along with any error.
func (p *Planet) toDefender(data Instance, moment time.Time) (defender, error) {
	// Fetch the universe of this planet.
	var uni string
	var err error

	if p.Moon {
		uni, err = UniverseOfMoon(p.ID, data)
	} else {
		uni, err = UniverseOfPlanet(p.ID, data)
	}

	if err != nil {
		return defender{}, err
	}

	d, err := newDefender(moment.UnixNano(), uni, data)
	if err != nil {
		return d, err
	}

	// Assign the main defender: it is the player
	// owning this celestial body.
	d.mainDef = p.Player
	d.location = p.ID
	d.moon = p.Moon

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
		// Only account for ships where at least one item
		// is present on the defender.
		if s.Amount <= 0 {
			continue
		}

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
		// Only account for ships where at least one item
		// is present on the defender.
		if def.Amount <= 0 {
			continue
		}

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

	// Fetch reinforcements. We need to fetch each fleet
	// incoming to the planet and see if any has already
	// arrived with a ACS defend objective.
	acsDefend, err := data.Objectives.GetIDFromName("ACS defend")
	if err != nil {
		return d, err
	}

	playersRegistered := make(map[string]bool)

	for _, fID := range p.IncomingFleets {
		// Fetch the fleet.
		f, err := NewFleetFromDB(fID, data)

		if err != nil {
			return d, err
		}

		// Check whether this fleet is coming for an ACS
		// defend operation.
		if f.Objective != acsDefend {
			continue
		}

		// Check whether the fleet is already deployed
		// near this planet at the time of the fight.
		endOfDeployment := f.ArrivalTime.Add(time.Duration(f.DeploymentTime) * time.Second)

		if f.ArrivalTime.After(moment) || endOfDeployment.Before(moment) {
			continue
		}

		// This fleet is part of the reinforcement forces
		// for this fight. Note that we *don't* register
		// the main defender as a participant.
		if f.Player != d.mainDef {
			_, ok := playersRegistered[f.Player]
			if !ok {
				d.participants = append(d.participants, f.Player)
				playersRegistered[f.Player] = true
			}
		}

		d.fleets = append(d.fleets, f.ID)

		// We need to include the ships of this fleet
		// into the defender state.
		for _, s := range f.Ships {
			sd, err := data.Ships.GetShipFromID(s.ID)
			if err != nil {
				return d, err
			}

			player, err := NewPlayerFromDB(f.Player, data)
			if err != nil {
				return d, err
			}

			shieldTech := player.Technologies[shieldID]
			weaponTech := player.Technologies[weaponID]
			armourTech := player.Technologies[armourID]

			shieldIncrease := float64(1.0 + shieldTech.Level/100.0)
			weaponIncrease := float64(1.0 + weaponTech.Level/100.0)
			armourIncrease := float64(1.0 + armourTech.Level/100.0)

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
				Fleet:        f.ID,
				Ship:         s.ID,
				Count:        s.Count,
				Cargo:        sd.Cargo,
				Shield:       shield,
				Weapon:       weapon,
				Hull:         armour,
				RFVSShips:    sd.RFVSShips,
				RFVSDefenses: sd.RFVSDefenses,
			}

			d.reinforcements = append(d.reinforcements, sif)
		}
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
	a := attacker{
		participants: []string{f.Player},
		fleets:       []string{f.ID},
		usedCargo:    f.usedCargoSpace(),
		log:          data.log,
	}

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
		// Only account for ships where at least one item
		// is present on the attacker: this should always
		// be the case but just to be on the safe side.
		if s.Count <= 0 {
			continue
		}

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
			Fleet:        f.ID,
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

// handleDumbMove :
// Despite its quite provocative name this method
// is used to perform the verification that the
// ships remaining in the attacker fleet after
// the fight is still enough to actually carry
// the resources initially carried by the fleet.
// This should usually not happen as it's quite
// unwise to send loaded ships to combat.
//
// The `a` attacker is what remains of the fleet.
//
// Returns a boolean indicating whether the fleet
// has enough cargo space for the inital res.
func (f *Fleet) handleDumbMove(a attacker) bool {
	// Compute the remaining cargo space for this
	// fleet.
	remainingCargo := 0

	for _, units := range a.units {
		for _, unit := range units {
			remainingCargo += unit.Count * unit.Cargo
		}
	}

	// In case the cargo space is still sufficient
	// it's okay.
	if float32(remainingCargo) >= a.usedCargo || len(f.Cargo) == 0 {
		return true
	}

	// There is not enough space left: we will try
	// to reduce in the same proportion the amount
	// of each resource carried to solve this.
	toShave := a.usedCargo - float32(remainingCargo)
	toShavePerRes := toShave / float32(len(f.Cargo))

	for rID, res := range f.Cargo {
		res.Amount -= toShavePerRes
		f.Cargo[rID] = res
	}

	return false
}
