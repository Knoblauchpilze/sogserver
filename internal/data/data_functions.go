package data

// valid :
// Used to determine whether the parameters defined for this
// fleet component are not obviously wrong. This method checks
// that the identifier provided for individual ships aren't
// empty or ill-formed and that the amount of each one is at
// least strictly positive.
//
// The `uni` represents the universe which should be attached
// to the fleet and will be used to verify that the starting
// position of the fleet component is consistent with possible
// coordinates in the universe.
//
// Returns `true` if the fleet component is valid.
func (fc *FleetComponent) valid(uni Universe) bool {
	// Check own identifier.
	if !validUUID(fc.ID) {
		return false
	}

	// Check the identifier of the player and parent fleet.
	if !validUUID(fc.FleetID) {
		return false
	}
	if !validUUID(fc.PlayerID) {
		return false
	}

	// Check the coordinates against the universe.
	if fc.Galaxy < 0 || fc.Galaxy >= uni.GalaxiesCount {
		return false
	}
	if fc.System < 0 || fc.System >= uni.GalaxySize {
		return false
	}
	if fc.Position < 0 || fc.Position >= uni.SolarSystemSize {
		return false
	}

	// Check the speed.
	if fc.Speed < 0 || fc.Speed > 1 {
		return false
	}

	// Now check individual ships.
	if len(fc.Ships) == 0 {
		return false
	}

	for _, ship := range fc.Ships {
		if !validUUID(ship.ShipID) || ship.Amount <= 0 {
			return false
		}
	}

	return true
}

// valid :
func (f *Fleet) valid(uni Universe) bool {
	// Check own identifier.
	if !validUUID(f.ID) {
		return false
	}

	// Check that the target is valid given the universe
	// into which the fleet is supposed to reside.
	if f.Galaxy < 0 || f.Galaxy >= uni.GalaxiesCount {
		return false
	}
	if f.System < 0 || f.System >= uni.GalaxySize {
		return false
	}
	if f.Position < 0 || f.Position >= uni.SolarSystemSize {
		return false
	}

	return true
}
