package game

import (
	"fmt"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"time"

	"github.com/google/uuid"
)

// ErrPlanetIsNotAMoon : Indicates that an attempt to save a planet
// as a moon was detected.
var ErrPlanetIsNotAMoon = fmt.Errorf("Cannot save planet as moon")

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
	err := m.fetchMoonInfo(data)
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
	err = m.fetchMoonResources(data)
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

// NewMoon :
// Used to perform the creation of the moon at
// the specified coordinates. As a moon always
// has to be associated to a planet we figure
// it made sense to build the moon from its
// parent planet.
//
// The `p` defines the parent planet of this
// moon.
//
// The `diameter` defines the size of the moon
// to create.
//
// Returns the generated moon.
func NewMoon(p *Planet, diameter int) *Planet {
	// Create default properties.
	m := &Planet{
		ID:          uuid.New().String(),
		Player:      p.Player,
		Coordinates: p.Coordinates,
		Name:        "moon",
		// Only a single field is generated at the
		// creation of the moon. The rest of the
		// fields will be added whenever a lunar
		// base is built.
		Fields:    1,
		MinTemp:   p.MinTemp,
		MaxTemp:   p.MaxTemp,
		Diameter:  diameter,
		Resources: make(map[string]ResourceInfo, 0),
		Buildings: make(map[string]BuildingInfo, 0),
		Ships:     make(map[string]ShipInfo, 0),
		Defenses:  make(map[string]DefenseInfo, 0),

		BuildingsUpgrade:     make([]BuildingAction, 0),
		TechnologiesUpgrade:  make([]TechnologyAction, 0),
		ShipsConstruction:    make([]ShipAction, 0),
		DefensesConstruction: make([]DefenseAction, 0),

		SourceFleets:   make([]string, 0),
		IncomingFleets: make([]string, 0),

		technologies: p.technologies,

		moon:   true,
		planet: p.ID,
	}

	m.Coordinates.Type = Moon

	return m
}

// fetchMoonInfo :
// Allows to fetch the general information of a moon
// from the DB such as its diameter, name, coordinates
// etc.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (p *Planet) fetchMoonInfo(data Instance) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"m.planet",
			"p.player",
			"m.name",
			"p.min_temperature",
			"p.max_temperature",
			"m.fields",
			"p.galaxy",
			"p.solar_system",
			"p.position",
			"m.diameter",
		},
		Table: "moons m inner join planets p on m.planet=p.id",
		Filters: []db.Filter{
			{
				Key:    "m.id",
				Values: []interface{}{p.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}
	if dbRes.Err != nil {
		return dbRes.Err
	}

	// Populate the return value.
	var galaxy, system, position int

	for dbRes.Next() {
		err = dbRes.Scan(
			&p.planet,
			&p.Player,
			&p.Name,
			&p.MinTemp,
			&p.MaxTemp,
			&p.Fields,
			&galaxy,
			&system,
			&position,
			&p.Diameter,
		)

		if err != nil {
			return err
		}

		p.Coordinates = NewMoonCoordinate(galaxy, system, position)
	}

	return nil
}

// fetchMoonResources :
// Similar to the `fetchGeneralInfo` but handles the
// retrieval of the moon's resources data.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (p *Planet) fetchMoonResources(data Instance) error {
	p.Resources = make(map[string]ResourceInfo, 0)

	// Do not update the resources of the moon: as
	// there is no production of resources on a moon
	// we know that any update is brought by a fleet.
	// And thus as all fleets actions have already
	// been processed we know that the resources are
	// up-to-date.

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"res",
			"amount",
		},
		Table: "moons_resources",
		Filters: []db.Filter{
			{
				Key:    "moon",
				Values: []interface{}{p.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}
	if dbRes.Err != nil {
		return dbRes.Err
	}

	// Populate the return value.
	var res ResourceInfo

	sanity := make(map[string]bool)

	for dbRes.Next() {
		err = dbRes.Scan(
			&res.Resource,
			&res.Amount,
		)

		if err != nil {
			return err
		}

		// Update the storage from the
		// base values and the prod as
		// we know that nothing ever
		// happens on the moon.
		info, err := data.Resources.GetResourceFromID(res.Resource)
		if err != nil {
			return err
		}

		res.Storage = float32(info.BaseStorage)
		res.Production = 0

		_, ok := sanity[res.Resource]
		if ok {
			return model.ErrInconsistentDB
		}
		sanity[res.Resource] = true

		p.Resources[res.Resource] = res
	}

	return nil
}

// SaveMoonToDB :
// Used to save the content of this moon to
// the DB. In case an error is raised during
// the operation a comprehensive error is
// returned.
//
// The `proxy` allows to access to the DB.
//
// Returns any error.
func (p *Planet) SaveMoonToDB(proxy db.Proxy) error {
	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_moon",
		Args: []interface{}{
			p,
			p.Resources,
			time.Now(),
		},
	}

	err := proxy.InsertToDB(query)

	// Analyze the error in order to provide some
	// comprehensive message.
	dbe, ok := err.(db.Error)
	if !ok {
		return err
	}

	dee, ok := dbe.Err.(db.DuplicatedElementError)
	if ok {
		switch dee.Constraint {
		case "moons_planet_key":
			return ErrDuplicatedElement
		}

		return dee
	}

	fkve, ok := dbe.Err.(db.ForeignKeyViolationError)
	if ok {
		switch fkve.ForeignKey {
		case "planet":
			return ErrNonExistingPlanet
		}

		return fkve
	}

	return dbe
}
