package game

// NewMoonFromDB :
// Used in a similar way to `NewPlanetFromDB` but to
// fetch the content of a moon. A moon is almost like
// a planet except its data is not registered in the
// same tables.
// If the DB does not contain any descritpion for the
// moon or something wrong happens an error is raised.
//
// The `ID` defines the identifier of the moon to get
// from the DB.
//
// The `data` allows to actually perform the DB
// requests to fetch the moon's data.
//
// Returns the moon as fetched from the DB along with
// any errors.
func NewMoonFromDB(ID string, data Instance) (Planet, error) {
	// Create the moon.
	m := Planet{
		ID: ID,
	}

	// Consistency.
	if !validUUID(m.ID) {
		return m, ErrInvalidElementID
	}

	// Fetch general info for this moon.
	err := m.fetchGeneralInfo(data)
	if err != nil {
		return m, err
	}

	// Fetch upgrade actions.
	err = m.fetchBuildingsUpgrades(data)
	if err != nil {
		return m, err
	}

	// No technologies research on moons.
	m.TechnologiesUpgrade = make([]TechnologyAction, 0)

	err = m.fetchShipsUpgrades(data)
	if err != nil {
		return m, err
	}

	err = m.fetchDefensesUpgrades(data)
	if err != nil {
		return m, err
	}

	// Fetch fleets.
	err = m.fetchIncomingFleets(data)
	if err != nil {
		return m, err
	}

	err = m.fetchSourceFleets(data)
	if err != nil {
		return m, err
	}

	// Fetch the moon's content.
	err = m.fetchResources(data)
	if err != nil {
		return m, err
	}

	err = m.fetchBuildings(data)
	if err != nil {
		return m, err
	}

	err = m.fetchTechnologies(data)
	if err != nil {
		return m, err
	}

	err = m.fetchShips(data)
	if err != nil {
		return m, err
	}

	err = m.fetchDefenses(data)
	if err != nil {
		return m, err
	}

	return m, err
}

// The following methods should be specialized:
// NewPlanet
// generateData
// fetchGeneralInfo
// fetchResources
// SaveToDB
// Convert
