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

// // NewPlanet :
// // Used to perform the creation of the planet at
// // the specified coordinates. It will perform the
// // creation of the needed information such as the
// // planet's size and temperature based on input
// // coords.
// //
// // The `player` defines the identifier of the
// // player to which this planet will be assigned.
// //
// // The `coords` represent the desired position of
// // the planet to generate.
// //
// // The `homeworld` defines whether this planet
// // is the homeworld for a player (which indicates
// // that the name should be different from another
// // random planet).
// //
// // The `data` allow to generate a default amount
// // of resources on the planet.
// //
// // Returns the generated planet along with any error.
// func NewPlanet(player string, coords Coordinate, homeworld bool, data Instance) (*Planet, error) {
// 	// Create default properties.
// 	p := &Planet{
// 		ID:          uuid.New().String(),
// 		Player:      player,
// 		Coordinates: coords,
// 		Name:        getDefaultPlanetName(homeworld),
// 		Fields:      0,
// 		MinTemp:     0,
// 		MaxTemp:     0,
// 		Diameter:    0,
// 		Resources:   make(map[string]ResourceInfo, 0),
// 		Buildings:   make(map[string]BuildingInfo, 0),
// 		Ships:       make(map[string]ShipInfo, 0),
// 		Defenses:    make(map[string]DefenseInfo, 0),
// 	}

// 	// Generate diameter and fields count.
// 	err := p.generateData(data)

// 	return p, err
// }

// // fetchGeneralInfo :
// // Allows to fetch the general information of a planet
// // from the DB such as its diameter, name, coordinates
// // etc.
// //
// // The `data` defines the object to access the DB.
// //
// // Returns any error.
// func (p *Planet) fetchGeneralInfo(data Instance) error {
// 	// Create the query and execute it.
// 	query := db.QueryDesc{
// 		Props: []string{
// 			"player",
// 			"name",
// 			"min_temperature",
// 			"max_temperature",
// 			"fields",
// 			"galaxy",
// 			"solar_system",
// 			"position",
// 			"diameter",
// 		},
// 		Table: "planets",
// 		Filters: []db.Filter{
// 			{
// 				Key:    "id",
// 				Values: []interface{}{p.ID},
// 			},
// 		},
// 	}

// 	dbRes, err := data.Proxy.FetchFromDB(query)
// 	defer dbRes.Close()

// 	// Check for errors.
// 	if err != nil {
// 		return err
// 	}
// 	if dbRes.Err != nil {
// 		return dbRes.Err
// 	}

// 	// Populate the return value.
// 	var galaxy, system, position int

// 	for dbRes.Next() {
// 		err = dbRes.Scan(
// 			&p.Player,
// 			&p.Name,
// 			&p.MinTemp,
// 			&p.MaxTemp,
// 			&p.Fields,
// 			&galaxy,
// 			&system,
// 			&position,
// 			&p.Diameter,
// 		)

// 		if err != nil {
// 			return err
// 		}

// 		p.Coordinates = NewPlanetCoordinate(galaxy, system, position)
// 	}

// 	return nil
// }

// // fetchBuildingsUpgrades :
// // Similar to the `fetchGeneralInfo` method but used
// // to fetch the buildings upgrades for a planet.
// //
// // The `data` defines a way to access to the DB.
// //
// // Returns any error.
// func (p *Planet) fetchBuildingsUpgrades(data Instance) error {
// 	p.BuildingsUpgrade = make([]BuildingAction, 0)

// 	// Create the query and execute it.
// 	query := db.QueryDesc{
// 		Props: []string{
// 			"id",
// 		},
// 		Table: "construction_actions_buildings",
// 		Filters: []db.Filter{
// 			{
// 				Key:    "planet",
// 				Values: []interface{}{p.ID},
// 			},
// 		},
// 	}

// 	dbRes, err := data.Proxy.FetchFromDB(query)
// 	defer dbRes.Close()

// 	// Check for errors.
// 	if err != nil {
// 		return err
// 	}
// 	if dbRes.Err != nil {
// 		return dbRes.Err
// 	}

// 	// We now need to retrieve all the identifiers that matched
// 	// the input filters and then build the corresponding item
// 	// object for each one of them.
// 	var ID string
// 	IDs := make([]string, 0)

// 	for dbRes.Next() {
// 		err = dbRes.Scan(&ID)

// 		if err != nil {
// 			return err
// 		}

// 		IDs = append(IDs, ID)
// 	}

// 	for _, ID = range IDs {
// 		bu, err := NewBuildingActionFromDB(ID, data)

// 		if err != nil {
// 			return err
// 		}

// 		p.BuildingsUpgrade = append(p.BuildingsUpgrade, bu)
// 	}

// 	return nil
// }

// // fetchShipsUpgrades :
// // Similar to the `fetchGeneralInfo` method but used
// // to fetch the ships upgrades for a planet.
// //
// // The `data` defines a way to access to the DB.
// //
// // Returns any error.
// func (p *Planet) fetchShipsUpgrades(data Instance) error {
// 	p.ShipsConstruction = make([]ShipAction, 0)

// 	// Create the query and execute it.
// 	query := db.QueryDesc{
// 		Props: []string{
// 			"id",
// 		},
// 		Table: "construction_actions_ships",
// 		Filters: []db.Filter{
// 			{
// 				Key:    "planet",
// 				Values: []interface{}{p.ID},
// 			},
// 		},
// 		Ordering: "order by created_at",
// 	}

// 	dbRes, err := data.Proxy.FetchFromDB(query)
// 	defer dbRes.Close()

// 	// Check for errors.
// 	if err != nil {
// 		return err
// 	}
// 	if dbRes.Err != nil {
// 		return dbRes.Err
// 	}

// 	// We now need to retrieve all the identifiers that matched
// 	// the input filters and then build the corresponding item
// 	// object for each one of them.
// 	var ID string
// 	IDs := make([]string, 0)

// 	for dbRes.Next() {
// 		err = dbRes.Scan(&ID)

// 		if err != nil {
// 			return err
// 		}

// 		IDs = append(IDs, ID)
// 	}

// 	for _, ID = range IDs {
// 		su, err := NewShipActionFromDB(ID, data)

// 		if err != nil {
// 			return err
// 		}

// 		p.ShipsConstruction = append(p.ShipsConstruction, su)
// 	}

// 	return nil
// }

// // fetchDefensesUpgrades :
// // Similar to the `fetchGeneralInfo` method but used
// // to fetch the defenses upgrades for a planet.
// //
// // The `data` defines a way to access to the DB.
// //
// // Returns any error.
// func (p *Planet) fetchDefensesUpgrades(data Instance) error {
// 	p.DefensesConstruction = make([]DefenseAction, 0)

// 	// Create the query and execute it.
// 	query := db.QueryDesc{
// 		Props: []string{
// 			"id",
// 		},
// 		Table: "construction_actions_defenses",
// 		Filters: []db.Filter{
// 			{
// 				Key:    "planet",
// 				Values: []interface{}{p.ID},
// 			},
// 		},
// 		Ordering: "order by created_at",
// 	}

// 	dbRes, err := data.Proxy.FetchFromDB(query)
// 	defer dbRes.Close()

// 	// Check for errors.
// 	if err != nil {
// 		return err
// 	}
// 	if dbRes.Err != nil {
// 		return dbRes.Err
// 	}

// 	// We now need to retrieve all the identifiers that matched
// 	// the input filters and then build the corresponding item
// 	// object for each one of them.
// 	var ID string
// 	IDs := make([]string, 0)

// 	for dbRes.Next() {
// 		err = dbRes.Scan(&ID)

// 		if err != nil {
// 			return err
// 		}

// 		IDs = append(IDs, ID)
// 	}

// 	for _, ID = range IDs {
// 		du, err := NewDefenseActionFromDB(ID, data)

// 		if err != nil {
// 			return err
// 		}

// 		p.DefensesConstruction = append(p.DefensesConstruction, du)
// 	}

// 	return nil
// }

// // fetchIncomingFleets :
// // Used to fetch the incoming fleets from the DB for
// // this planet. This include only the fleets having
// // their target set for this planet. Hostiles along
// // with friendly fleets will be fetched.
// //
// // The `data` defines the object to access the DB
// // if needed.
// //
// // Returns any error.
// func (p *Planet) fetchIncomingFleets(data Instance) error {
// 	p.IncomingFleets = make([]string, 0)

// 	// Create the query and execute it.
// 	query := db.QueryDesc{
// 		Props: []string{
// 			"id",
// 		},
// 		Table: "fleets",
// 		Filters: []db.Filter{
// 			{
// 				Key:    "target",
// 				Values: []interface{}{p.ID},
// 			},
// 			{
// 				Key:    "target_type",
// 				Values: []interface{}{"planet"},
// 			},
// 		},
// 		Ordering: "order by arrival_time desc",
// 	}

// 	dbRes, err := data.Proxy.FetchFromDB(query)
// 	defer dbRes.Close()

// 	// Check for errors.
// 	if err != nil {
// 		return err
// 	}
// 	if dbRes.Err != nil {
// 		return dbRes.Err
// 	}

// 	var ID string

// 	for dbRes.Next() {
// 		err = dbRes.Scan(&ID)

// 		if err != nil {
// 			return err
// 		}

// 		p.IncomingFleets = append(p.IncomingFleets, ID)
// 	}

// 	return nil
// }

// // fetchSourceFleets :
// // Similar to the `fetchIncomingFleets` but retrieves
// // the fleets that starts from this planet.
// //
// // The `data` defines the object to access the DB if
// // needed.
// //
// // Returns any error.
// func (p *Planet) fetchSourceFleets(data Instance) error {
// 	p.SourceFleets = make([]string, 0)

// 	// Create the query and execute it.
// 	query := db.QueryDesc{
// 		Props: []string{
// 			"id",
// 		},
// 		Table: "fleets",
// 		Filters: []db.Filter{
// 			{
// 				Key:    "source",
// 				Values: []interface{}{p.ID},
// 			},
// 			{
// 				Key:    "source_type",
// 				Values: []interface{}{"planet"},
// 			},
// 		},
// 	}

// 	dbRes, err := data.Proxy.FetchFromDB(query)
// 	defer dbRes.Close()

// 	// Check for errors.
// 	if err != nil {
// 		return err
// 	}
// 	if dbRes.Err != nil {
// 		return dbRes.Err
// 	}

// 	var ID string

// 	for dbRes.Next() {
// 		err = dbRes.Scan(&ID)

// 		if err != nil {
// 			return err
// 		}

// 		p.SourceFleets = append(p.SourceFleets, ID)
// 	}

// 	return nil
// }

// // fetchResources :
// // Similar to the `fetchGeneralInfo` but handles the
// // retrieval of the planet's resources data.
// //
// // The `data` defines the object to access the DB.
// //
// // Returns any error.
// func (p *Planet) fetchResources(data Instance) error {
// 	p.Resources = make(map[string]ResourceInfo, 0)

// 	// The server guarantees that any action that takes
// 	// or bring resources to the planet will perform an
// 	// update of the resources for this planet. But in
// 	// case nothing happens on the planet, we have to
// 	// make sure that the resources are still updated.
// 	// This is done here. Note that it might not be
// 	// super useful as we don't really know if this
// 	// planet has just been updated due to an action
// 	// or fleet.
// 	err := data.updateResourcesForPlanet(p.ID)
// 	if err != nil {
// 		return err
// 	}

// 	// Create the query and execute it.
// 	query := db.QueryDesc{
// 		Props: []string{
// 			"res",
// 			"amount",
// 			"production",
// 			"storage_capacity",
// 		},
// 		Table: "planets_resources",
// 		Filters: []db.Filter{
// 			{
// 				Key:    "planet",
// 				Values: []interface{}{p.ID},
// 			},
// 		},
// 	}

// 	dbRes, err := data.Proxy.FetchFromDB(query)
// 	defer dbRes.Close()

// 	// Check for errors.
// 	if err != nil {
// 		return err
// 	}
// 	if dbRes.Err != nil {
// 		return dbRes.Err
// 	}

// 	// Populate the return value.
// 	var res ResourceInfo

// 	sanity := make(map[string]bool)

// 	for dbRes.Next() {
// 		err = dbRes.Scan(
// 			&res.Resource,
// 			&res.Amount,
// 			&res.Production,
// 			&res.Storage,
// 		)

// 		if err != nil {
// 			return err
// 		}

// 		_, ok := sanity[res.Resource]
// 		if ok {
// 			return model.ErrInconsistentDB
// 		}
// 		sanity[res.Resource] = true

// 		p.Resources[res.Resource] = res
// 	}

// 	return nil
// }

// // fetchBuildings :
// // Similar to the `fetchGeneralInfo` but handles the
// // retrieval of the planet's buildings data.
// //
// // The `data` defines the object to access the DB.
// //
// // Returns any error.
// func (p *Planet) fetchBuildings(data Instance) error {
// 	p.Buildings = make(map[string]BuildingInfo, 0)

// 	// Create the query and execute it.
// 	query := db.QueryDesc{
// 		Props: []string{
// 			"building",
// 			"level",
// 		},
// 		Table: "planets_buildings",
// 		Filters: []db.Filter{
// 			{
// 				Key:    "planet",
// 				Values: []interface{}{p.ID},
// 			},
// 		},
// 	}

// 	dbRes, err := data.Proxy.FetchFromDB(query)
// 	defer dbRes.Close()

// 	// Check for errors.
// 	if err != nil {
// 		return err
// 	}
// 	if dbRes.Err != nil {
// 		return dbRes.Err
// 	}

// 	// Populate the return value.
// 	var ID string
// 	var b BuildingInfo

// 	sanity := make(map[string]bool)

// 	for dbRes.Next() {
// 		err = dbRes.Scan(
// 			&ID,
// 			&b.Level,
// 		)

// 		if err != nil {
// 			return err
// 		}

// 		_, ok := sanity[ID]
// 		if ok {
// 			return model.ErrInconsistentDB
// 		}
// 		sanity[ID] = true

// 		desc, err := data.Buildings.GetBuildingFromID(ID)
// 		if err != nil {
// 			return err
// 		}

// 		b.BuildingDesc = desc

// 		p.Buildings[ID] = b
// 	}

// 	return nil
// }

// // fetchShips :
// // Similar to the `fetchGeneralInfo` but handles the
// // retrieval of the planet's ships data.
// //
// // The `data` defines the object to access the DB.
// //
// // Returns any error.
// func (p *Planet) fetchShips(data Instance) error {
// 	p.Ships = make(map[string]ShipInfo, 0)

// 	// Create the query and execute it.
// 	query := db.QueryDesc{
// 		Props: []string{
// 			"ship",
// 			"count",
// 		},
// 		Table: "planets_ships",
// 		Filters: []db.Filter{
// 			{
// 				Key:    "planet",
// 				Values: []interface{}{p.ID},
// 			},
// 		},
// 	}

// 	dbRes, err := data.Proxy.FetchFromDB(query)
// 	defer dbRes.Close()

// 	// Check for errors.
// 	if err != nil {
// 		return err
// 	}
// 	if dbRes.Err != nil {
// 		return dbRes.Err
// 	}

// 	// Populate the return value.
// 	var ID string
// 	var s ShipInfo

// 	sanity := make(map[string]bool)

// 	for dbRes.Next() {
// 		err = dbRes.Scan(
// 			&ID,
// 			&s.Amount,
// 		)

// 		if err != nil {
// 			return err
// 		}

// 		_, ok := sanity[ID]
// 		if ok {
// 			return model.ErrInconsistentDB
// 		}
// 		sanity[ID] = true

// 		desc, err := data.Ships.GetShipFromID(ID)
// 		if err != nil {
// 			return err
// 		}

// 		s.ShipDesc = desc

// 		p.Ships[ID] = s
// 	}

// 	return nil
// }

// // fetchShips :
// // Similar to the `fetchGeneralInfo` but handles the
// // retrieval of the planet's defenses data.
// //
// // The `data` defines the object to access the DB.
// //
// // Returns any error.
// func (p *Planet) fetchDefenses(data Instance) error {
// 	p.Defenses = make(map[string]DefenseInfo, 0)

// 	// Create the query and execute it.
// 	query := db.QueryDesc{
// 		Props: []string{
// 			"defense",
// 			"count",
// 		},
// 		Table: "planets_defenses",
// 		Filters: []db.Filter{
// 			{
// 				Key:    "planet",
// 				Values: []interface{}{p.ID},
// 			},
// 		},
// 	}

// 	dbRes, err := data.Proxy.FetchFromDB(query)
// 	defer dbRes.Close()

// 	// Check for errors.
// 	if err != nil {
// 		return err
// 	}
// 	if dbRes.Err != nil {
// 		return dbRes.Err
// 	}

// 	// Populate the return value.
// 	var ID string
// 	var d DefenseInfo

// 	sanity := make(map[string]bool)

// 	for dbRes.Next() {
// 		err = dbRes.Scan(
// 			&ID,
// 			&d.Amount,
// 		)

// 		if err != nil {
// 			return err
// 		}

// 		_, ok := sanity[ID]
// 		if ok {
// 			return model.ErrInconsistentDB
// 		}
// 		sanity[ID] = true

// 		desc, err := data.Defenses.GetDefenseFromID(ID)
// 		if err != nil {
// 			return err
// 		}

// 		d.DefenseDesc = desc

// 		p.Defenses[ID] = d
// 	}

// 	return nil
// }

// // SaveToDB :
// // Used to save the content of this planet to
// // the DB. In case an error is raised during
// // the operation a comprehensive error is
// // returned.
// //
// // The `proxy` allows to access to the DB.
// //
// // Returns any error.
// func (p *Planet) SaveToDB(proxy db.Proxy) error {
// 	// Create the query and execute it.
// 	query := db.InsertReq{
// 		Script: "create_planet",
// 		Args: []interface{}{
// 			p,
// 			p.Resources,
// 			time.Now(),
// 		},
// 	}

// 	err := proxy.InsertToDB(query)

// 	// Analyze the error in order to provide some
// 	// comprehensive message.
// 	dbe, ok := err.(db.Error)
// 	if !ok {
// 		return err
// 	}

// 	dee, ok := dbe.Err.(db.DuplicatedElementError)
// 	if ok {
// 		switch dee.Constraint {
// 		case "planets_pkey":
// 			return ErrDuplicatedElement
// 		}

// 		return dee
// 	}

// 	fkve, ok := dbe.Err.(db.ForeignKeyViolationError)
// 	if ok {
// 		switch fkve.ForeignKey {
// 		case "player":
// 			return ErrNonExistingPlayer
// 		}

// 		return fkve
// 	}

// 	return dbe
// }

// // Convert :
// // Implementation of the `db.Convertible` interface
// // from the DB package in order to only include fields
// // that need to be marshalled in the planet's creation.
// //
// // Returns the converted version of the planet which
// // only includes relevant fields.
// func (p *Planet) Convert() interface{} {
// 	return struct {
// 		ID       string `json:"id"`
// 		Player   string `json:"player"`
// 		Name     string `json:"name"`
// 		MinTemp  int    `json:"min_temperature"`
// 		MaxTemp  int    `json:"max_temperature"`
// 		Fields   int    `json:"fields"`
// 		Galaxy   int    `json:"galaxy"`
// 		System   int    `json:"solar_system"`
// 		Position int    `json:"position"`
// 		Diameter int    `json:"diameter"`
// 	}{
// 		ID:       p.ID,
// 		Player:   p.Player,
// 		Name:     p.Name,
// 		MinTemp:  p.MinTemp,
// 		MaxTemp:  p.MaxTemp,
// 		Fields:   p.Fields,
// 		Galaxy:   p.Coordinates.Galaxy,
// 		System:   p.Coordinates.System,
// 		Position: p.Coordinates.Position,
// 		Diameter: p.Diameter,
// 	}
// }
