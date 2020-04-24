package model

import (
	"fmt"
	"math"
	"math/rand"
	"oglike_server/pkg/db"

	"github.com/google/uuid"
)

// Planet :
// Define a planet which is an object within a certain
// universe and associated to a certain player. This
// element is the base where all fleets can originate
// and is usually the main source of production in any
// universe. It contains a set of resources, some ships
// that are deployed on it and a set of buildings.
// The user can also start some construction actions
// that can increase the number of ships present on the
// planet for example.
// A planet is a complex object which requires a lock
// to be correctly built so that no other process can
// temper with the data related to it.
//
// The `ID` defines the identifier of the planet within
// all the planets registered in og.
//
// The `Player` defines the identifier of the player
// who owns this planet. It is relative to an account
// and a universe.
//
// The `Coordinates` define the position of the planet
// in its parent universe. All needed coordinates are
// guaranteed to be valid within it.
//
// The `Name` of the planet as defined by the user.
//
// The `Fields` define the number of available fields in
// the planet. The number of used fields is computed from
// the infrastructure built on the planet but is derived
// from the actual level of each building.
//
// The `MinTemp` defines the minimum temperature of the
// planet in degrees.
//
// The `MaxTemp` defines the maximum temperature of the
// planet in degrees.
//
// The `Diameter` defines the diameter of the planet in
// kilometers.
//
// The `Resources` define the resources currently stored
// on the planet. This is basically the quantity available
// to produce some buildings, ships, etc.
//
// The `Buildings` defines the list of buildings currently
// built on the planet. Note that it does not provide info
// on buildings *being* built.
//
// The `Ships` defines the list of ships currently deployed
// on this planet. It does not include ships currently moving
// from or towards the planet.
//
// The `Defense` defines the list of defenses currently
// built on the planet. This does not include defenses
// *being* built.
type Planet struct {
	ID          string         `json:"id"`
	Player      string         `json:"player"`
	Coordinates Coordinate     `json:"coordinate"`
	Name        string         `json:"name"`
	Fields      int            `json:"fields"`
	MinTemp     int            `json:"min_temperature"`
	MaxTemp     int            `json:"max_temperature"`
	Diameter    int            `json:"diameter"`
	Resources   []ResourceInfo `json:"resources"`
	Buildings   []BuildingInfo `json:"buildings"`
	Ships       []ShipInfo     `json:"ships"`
	Defenses    []DefenseInfo  `json:"defenses"`
}

// ResourceInfo :
// Defines the information needed to qualify the amount
// of resource of a certain type existing on a planet.
// In addition to the base description of the resource,
// this object defines the current storage capacity of
// the planet along with the production given the set
// of buildings (and more specifically mines) built on
// the planet.
//
// The `Resource` defines the identifier of the resource
// that this association describes.
//
// The `Amount` defines how much of the resource is
// present on this planet.
//
// The `Storage` defines the maximum amount that can
// be stored on this planet. Note that the `Amount`
// can exceed the `Storage` in which case the prod is
// stopped.
//
// The `Production` defines the maximum theoretical
// production of this resource given the buildings
// existing on the planet. Note that this value is
// always set to the maximum possible and does not
// account for the fact that the production may be
// stopped (in case the storage is full for example).
// The production is given in units per hour.
type ResourceInfo struct {
	Resource   string `json:"resource"`
	Amount     int    `json:"amount"`
	Storage    int    `json:"storage"`
	Production int    `json:"production"`
}

// BuildingInfo :
// Defines the information about a building of a
// planet. It reuses most of the base description
// for a building with the addition of a level to
// indicate the current state reached on this
// planet.
//
// The `Level` defines the level reached by this
// building on a particular planet.
type BuildingInfo struct {
	BuildingDesc

	Level int `json:"level"`
}

// ShipInfo :
// Similar to the `BuildingInfo` but defines the
// amount of ships currently deployed on a planet.
// Note that this is a snapshot at a certain time
// and does not account for ships that may be
// built currently.
//
// The `Amount` defines how many of this ship are
// currently deployed on the planet.
type ShipInfo struct {
	ShipDesc

	Amount int `json:"amount"`
}

// DefenseInfo :
// Similar to the `ShipInfo` but defines the same
// properties for a defense system.
//
// The `Amount` defines how many of this defense
// system are currently deployed on the planet.
type DefenseInfo struct {
	DefenseDesc

	Amount int `json:"amount"`
}

// ErrInvalidPlanet :
// Used to indicate an ill-formed planet with no
// associated identifier.
var ErrInvalidPlanet = fmt.Errorf("Invalid planet with no identifier")

// getDefaultPlanetName :
// Used to retrieve a default name for a planet. The
// generated name will be different based on whether
// the planet is a homeworld or no.
//
// The `isHomeworld` is `true` if we should generate
// a name for the homeworld and `false` otherwise.
//
// Returns a string corresponding to the name of the
// planet.
func getDefaultPlanetName(isHomeWorld bool) string {
	if isHomeWorld {
		return "homeworld"
	}

	return "planet"
}

// getPlanetTemperatureAmplitude :
// Used to retrieve the default planet temperature's
// amplitude. Basically the interval between the min
// and max temperature will always be equal to this
// value.
//
// Returns the default temperature amplitude for the
// planets.
func getPlanetTemperatureAmplitude() int {
	return 50
}

// NewPlanetFromDB :
// Used to fetch the content of the planet from the
// input DB and populate all internal fields from it.
// In case the DB cannot be fetched or some errors
// are encoutered, the return value will include a
// description of the error.
//
// The `ID` defines the identifier of the planet to
// create. It should be fetched from the DB and is
// assumed to refer to an existing planet.
//
// The `data` allows to actually perform the DB
// requests to fetch the planet's data.
//
// Returns the planet as fetched from the DB along
// with any errors.
func NewPlanetFromDB(ID string, data Instance) (Planet, error) {
	// Create the planet.
	p := Planet{
		ID: ID,
	}

	// TODO: Update planet's upgrade action. It used to be done
	// this way in the planets' prody:
	// query := fmt.Sprintf("SELECT update_resources_for_planet('%s')", planetID)
	// query = fmt.Sprintf("SELECT update_building_upgrade_action('%s')", planetID)
	// query = fmt.Sprintf("SELECT update_ship_upgrade_action('%s')", planetID)
	// query = fmt.Sprintf("SELECT update_defense_upgrade_action('%s')", planetID)

	// Fetch the planet's content.
	err := p.fetchResources(data)
	if err != nil {
		return p, err
	}

	err = p.fetchBuildings(data)
	if err != nil {
		return p, err
	}

	err = p.fetchShips(data)
	if err != nil {
		return p, err
	}

	err = p.fetchDefenses(data)
	if err != nil {
		return p, err
	}

	return p, nil
}

// NewPlanet :
// Used to perform the creation of the planet at
// the specified coordinates. It will perform the
// creation of the needed information such as the
// planet's size and temperature based on input
// coords.
//
// The `player` defines the identifier of the
// player to which this planet will be assigned.
//
// The `coords` represent the desired position of
// the planet to generate.
//
// Returns the generated planet.
func NewPlanet(player string, coords Coordinate) *Planet {
	// Create default properties.
	p := &Planet{
		ID:          uuid.New().String(),
		Player:      player,
		Coordinates: coords,
		Name:        getDefaultPlanetName(coords.isNull()),
		Fields:      0,
		MinTemp:     0,
		MaxTemp:     0,
		Diameter:    0,
		Resources:   make([]ResourceInfo, 0),
		Buildings:   make([]BuildingInfo, 0),
		Ships:       make([]ShipInfo, 0),
		Defenses:    make([]DefenseInfo, 0),
	}

	// Generate diameter and fields count.
	p.generateData()

	return p
}

// AverageTemperature :
// Returns the average temperature for this planet.
func (p *Planet) AverageTemperature() float32 {
	return float32(p.MinTemp+p.MaxTemp) / 2.0
}

// RemainingFields :
// Returns the number of remaining fields on the planet
// given the current buildings on it. Note that it does
// not include the potential upgrade actions.
func (p *Planet) RemainingFields() int {
	// Accumulate the total used fields.
	used := 0

	for _, b := range p.Buildings {
		used += b.Level
	}

	return p.Fields - used
}

// generateData :
// Used to generate the size associated to a planet. The size
// is a general notion including both its actual diameter and
// also the temperature on the surface of the planet. Both
// values depend on the actual position of the planet in the
// parent solar system.
func (p *Planet) generateData() {
	// Create a random source to be used for the generation of
	// the planet's properties. We will use a procedural algo
	// which will be based on the position of the planet in its
	// parent universe.
	source := rand.NewSource(int64(p.Coordinates.generateSeed()))
	rng := rand.New(source)

	// The table of the dimensions of the planet are inspired
	// from this link:
	// https://ogame.fandom.com/wiki/Colonizing_in_Redesigned_Universes
	var min int
	var max int
	var stdDev int

	switch p.Coordinates.Position {
	case 0:
		// Range [96; 172], average 134.
		min = 96
		max = 172
		stdDev = max - min
	case 1:
		// Range [104; 176], average 140.
		min = 104
		max = 176
		stdDev = max - min
	case 2:
		// Range [112; 182], average 147.
		min = 112
		max = 182
		stdDev = max - min
	case 3:
		// Range [118; 208], average 163.
		min = 118
		max = 208
		stdDev = max - min
	case 4:
		// Range [133; 232], average 182.
		min = 133
		max = 232
		stdDev = max - min
	case 5:
		// Range [152; 248], average 200.
		min = 152
		max = 248
		stdDev = max - min
	case 6:
		// Range [156; 262], average 204.
		min = 156
		max = 262
		stdDev = max - min
	case 7:
		// Range [150; 246], average 198.
		min = 150
		max = 246
		stdDev = max - min
	case 8:
		// Range [142; 232], average 187.
		min = 142
		max = 232
		stdDev = max - min
	case 9:
		// Range [136; 210], average 173.
		min = 136
		max = 210
		stdDev = max - min
	case 10:
		// Range [125; 186], average 156.
		min = 125
		max = 186
		stdDev = max - min
	case 11:
		// Range [114; 172], average 143.
		min = 114
		max = 172
		stdDev = max - min
	case 12:
		// Range [100; 168], average 134.
		min = 100
		max = 168
		stdDev = max - min
	case 13:
		// Range [90; 164], average 127.
		min = 96
		max = 164
		stdDev = max - min
	case 14:
		fallthrough
	default:
		// Assume default case if the `15th` position
		// Range [90; 164], average 134.
		min = 90
		max = 164
		stdDev = max - min
	}

	mean := (max + min) / 2
	p.Fields = mean + int(math.Round(rng.NormFloat64()*float64(stdDev)))

	// The diameter is derived from the fields count with a random part.
	p.Diameter = 100*p.Fields + int(math.Round(float64(100.0*rand.Float32())))

	// The temperatures are described in the following link:
	// https://ogame.fandom.com/wiki/Temperature
	switch p.Coordinates.Position {
	case 0:
		// Range [220; 260], average 240.
		min = 220
		max = 260
		stdDev = max - min
	case 1:
		// Range [170; 210], average 190.
		min = 170
		max = 210
		stdDev = max - min
	case 2:
		// Range [120; 160], average 140.
		min = 120
		max = 160
		stdDev = max - min
	case 3:
		// Range [70; 110], average 90.
		min = 70
		max = 110
		stdDev = max - min
	case 4:
		// Range [60; 100], average 80.
		min = 60
		max = 100
		stdDev = max - min
	case 5:
		// Range [50; 90], average 70.
		min = 50
		max = 90
		stdDev = max - min
	case 6:
		// Range [40; 80], average 60.
		min = 40
		max = 80
		stdDev = max - min
	case 7:
		// Range [30; 70], average 50.
		min = 30
		max = 70
		stdDev = max - min
	case 8:
		// Range [20; 60], average 40.
		min = 20
		max = 60
		stdDev = max - min
	case 9:
		// Range [10; 50], average 30.
		min = 10
		max = 50
		stdDev = max - min
	case 10:
		// Range [0; 40], average 20.
		min = 0
		max = 40
		stdDev = max - min
	case 11:
		// Range [-10; 30], average 10.
		min = -10
		max = 30
		stdDev = max - min
	case 12:
		// Range [-50; -10], average -30.
		min = -50
		max = -10
		stdDev = max - min
	case 13:
		// Range [-90; -50], average -70.
		min = -90
		max = -50
		stdDev = max - min
	case 14:
		fallthrough
	default:
		// Assume default case if the `15th` position
		// Range [-130; -90], average -110.
		min = -130
		max = -90
		stdDev = max - min
	}

	mean = (max + min) / 2
	p.MaxTemp = mean + int(math.Round(rng.NormFloat64()*float64(stdDev)))
	p.MinTemp = p.MaxTemp - getPlanetTemperatureAmplitude()
}

// fetchResources :
// Used internally when building a planet from the
// DB to update the resources existing on the planet.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (p *Planet) fetchResources(data Instance) error {
	// Consistency.
	if p.ID == "" {
		return ErrInvalidPlanet
	}

	p.Resources = make([]ResourceInfo, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"res",
			"amount",
			"production",
			"storage_capacity",
		},
		Table: "planets_resources",
		Filters: []db.Filter{
			{
				Key:    "planet",
				Values: []string{p.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}

	// Populate the return value.
	var res ResourceInfo

	for dbRes.Next() {
		err = dbRes.Scan(
			&res.Resource,
			&res.Amount,
			&res.Production,
			&res.Storage,
		)

		if err != nil {
			return err
		}

		p.Resources = append(p.Resources, res)
	}

	return nil
}

// fetchBuildings :
// Similar to the `fetchResources` but handles the
// retrieval of the planet's buildings data.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (p *Planet) fetchBuildings(data Instance) error {
	// Consistency.
	if p.ID == "" {
		return ErrInvalidPlanet
	}

	p.Buildings = make([]BuildingInfo, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"building",
			"level",
		},
		Table: "planets_buildings",
		Filters: []db.Filter{
			{
				Key:    "planet",
				Values: []string{p.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}

	// Populate the return value.
	var ID string
	var b BuildingInfo

	for dbRes.Next() {
		err = dbRes.Scan(
			&ID,
			&b.Level,
		)

		if err != nil {
			return err
		}

		desc, err := data.Buildings.getBuildingFromID(ID)
		if err != nil {
			return err
		}

		b.BuildingDesc = desc

		// TODO: Update cost, production and storage.

		p.Buildings = append(p.Buildings, b)
	}

	return nil
}

// fetchShips :
// Similar to the `fetchResources` but handles the
// retrieval of the planet's ships data.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (p *Planet) fetchShips(data Instance) error {
	// Consistency.
	if p.ID == "" {
		return ErrInvalidPlanet
	}

	p.Ships = make([]ShipInfo, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"ship",
			"count",
		},
		Table: "planets_ships",
		Filters: []db.Filter{
			{
				Key:    "planet",
				Values: []string{p.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}

	// Populate the return value.
	var ID string
	var s ShipInfo

	for dbRes.Next() {
		err = dbRes.Scan(
			&ID,
			&s.Amount,
		)

		if err != nil {
			return err
		}

		desc, err := data.Ships.getShipFromID(ID)
		if err != nil {
			return err
		}

		s.ShipDesc = desc

		p.Ships = append(p.Ships, s)
	}

	return nil
}

// fetchShips :
// Similar to the `fetchResources` but handles the
// retrieval of the planet's defenses data.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (p *Planet) fetchDefenses(data Instance) error {
	// Consistency.
	if p.ID == "" {
		return ErrInvalidPlanet
	}

	p.Defenses = make([]DefenseInfo, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"defense",
			"count",
		},
		Table: "planets_defenses",
		Filters: []db.Filter{
			{
				Key:    "planet",
				Values: []string{p.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}

	// Populate the return value.
	var ID string
	var d DefenseInfo

	for dbRes.Next() {
		err = dbRes.Scan(
			&ID,
			&d.Amount,
		)

		if err != nil {
			return err
		}

		desc, err := data.Defenses.getDefenseFromID(ID)
		if err != nil {
			return err
		}

		d.DefenseDesc = desc

		p.Defenses = append(p.Defenses, d)
	}

	return nil
}

// GetBuilding :
// Retrieves the building from the input identifier.
//
// The `ID` defines the identifier of the building
// to fetch from the planet.
//
// Returns the building description corresponding
// to the input identifier along with any error.
func (p *Planet) GetBuilding(ID string) (BuildingInfo, error) {
	// Traverse the list of buildings attached to the
	// planet and search for the input ID.
	for _, b := range p.Buildings {
		if b.ID == ID {
			return b, nil
		}
	}

	return BuildingInfo{}, ErrInvalidID
}
