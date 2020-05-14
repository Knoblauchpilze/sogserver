package game

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"oglike_server/internal/model"
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
// In addition to the lock on the planet, we assume in
// this object that the locker on the player owning the
// planet is already acquired. This allows to fetch the
// properties linked to it in a safe way without having
// to worry whether we should acquire the lock and meet
// potential dead lock situations as we would have a
// situation like:
//  - the player locks itself and then planets.
//  - the planet locks itself and then the player.
// In case this assumption is not met we might run in
// trouble and use incorrect data (typically if another
// planet finished a technology while a fleet fight is
// processed on another one, we might use outdated
// info).
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
// The `Fields` define the number of available fields
// in the planet. The number of used fields is computed
// from the infrastructure built on the planet but is
// derived from the actual level of each building.
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
// The `Defenses` defines the list of defenses currently
// built on the planet. This does not include defenses
// *being* built.
//
// The `BuildingsUpgrade` defines the list of upgrade action
// registered on this planet for buildings. This array might
// be empty in case no building is being upgraded.
//
// The `TechnologiesUpgrade` defines the list of upgrade
// action currently registered for this player.
//
// The `ShipsConstruction` defines the list of outstanding
// ships construction actions.
//
// The `DefensesConstruction` defines a similar list for
// defense systems on this planet.
//
// The `SourceFleets` defines the identifiers of all the
// fleets in which ships from this planet are taking an
// active part. It does not define anything beyond the
// ID of the fleets. The data can be fetched through the
// standard fleets route.
//
// The `IncomingFleets` defines a similar list of IDs
// but for fleets that will land on this planet.
//
// The `technologies` defines the technologies that have
// already been researched by the player owning this tech.
// It helps in various cases to be able to fetch player's
// info about technology.
type Planet struct {
	ID                   string                  `json:"id"`
	Player               string                  `json:"player"`
	Coordinates          Coordinate              `json:"coordinate"`
	Name                 string                  `json:"name"`
	Fields               int                     `json:"fields"`
	MinTemp              int                     `json:"min_temperature"`
	MaxTemp              int                     `json:"max_temperature"`
	Diameter             int                     `json:"diameter"`
	Resources            Resources               `json:"-"`
	Buildings            map[string]BuildingInfo `json:"-"`
	Ships                map[string]ShipInfo     `json:"-"`
	Defenses             map[string]DefenseInfo  `json:"-"`
	BuildingsUpgrade     []BuildingAction        `json:"-"`
	TechnologiesUpgrade  []TechnologyAction      `json:"-"`
	ShipsConstruction    []ShipAction            `json:"-"`
	DefensesConstruction []DefenseAction         `json:"-"`
	SourceFleets         []string                `json:"-"`
	IncomingFleets       []string                `json:"-"`
	technologies         map[string]int
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
	Resource   string  `json:"resource"`
	Amount     float32 `json:"amount"`
	Storage    float32 `json:"storage"`
	Production float32 `json:"production"`
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
	model.BuildingDesc

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
	model.ShipDesc

	Amount int `json:"amount"`
}

// DefenseInfo :
// Similar to the `ShipInfo` but defines the same
// properties for a defense system.
//
// The `Amount` defines how many of this defense
// system are currently deployed on the planet.
type DefenseInfo struct {
	model.DefenseDesc

	Amount int `json:"amount"`
}

// ErrNotEnoughResources : Indicates that there are not enough resources available.
var ErrNotEnoughResources = fmt.Errorf("Not enough resources available")

// ErrTechDepsNotMet : Indicates that the tech dependencies are not met.
var ErrTechDepsNotMet = fmt.Errorf("Action dependencies not met")

// ErrNoCost : Indicates that the action does not define any cost.
var ErrNoCost = fmt.Errorf("No cost provided for action")

// ErrNotEnoughFuel : Indicates that there's not enough fuel for the fleet.
var ErrNotEnoughFuel = fmt.Errorf("Not enough fuel for fleet")

// ErrNotEnoughShips : Indicates that there's not enough ships for the fleet.
var ErrNotEnoughShips = fmt.Errorf("Not enough ships for fleet")

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

// Resources :
// Convenience define to refer to a slice of resource
// info. It is used to make that that the array does
// get converted before getting sent to the DB.
type Resources map[string]ResourceInfo

// Convert :
// Implementation of the `db.Convertible` interface
// to allow the import of `ResourceInfo` struct to
// the DB. It mostly consists in changing the name
// of the `Resource` field.
//
// Returns the marshallable slice.
func (r Resources) Convert() interface{} {
	type resForDB struct {
		Resource   string  `json:"res"`
		Amount     float32 `json:"amount"`
		Storage    float32 `json:"storage_capacity"`
		Production float32 `json:"production"`
	}

	out := make([]resForDB, 0)

	for _, res := range r {
		out = append(
			out,
			resForDB{
				Resource:   res.Resource,
				Amount:     res.Amount,
				Storage:    res.Storage,
				Production: res.Production,
			},
		)
	}

	return out
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

	// Consistency.
	if !validUUID(p.ID) {
		return p, ErrInvalidElementID
	}

	// Fetch general info for this planet. It will
	// allow to fetch the player's identifier and
	// be able to update the technologies actions.
	err := p.fetchGeneralInfo(data)
	if err != nil {
		return p, err
	}

	// Fetch upgrade actions for this planet.
	err = p.fetchBuildingsUpgrades(data)
	if err != nil {
		return p, err
	}

	err = p.fetchTechnologiesUpgrades(data)
	if err != nil {
		return p, err
	}

	err = p.fetchShipsUpgrades(data)
	if err != nil {
		return p, err
	}

	err = p.fetchDefensesUpgrades(data)
	if err != nil {
		return p, err
	}

	// Fetch fleets.
	err = p.fetchIncomingFleets(data)
	if err != nil {
		return p, err
	}

	err = p.fetchSourceFleets(data)
	if err != nil {
		return p, err
	}

	// Fetch the planet's content.
	err = p.fetchResources(data)
	if err != nil {
		return p, err
	}

	err = p.fetchBuildings(data)
	if err != nil {
		return p, err
	}

	err = p.fetchTechnologies(data)
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

	return p, err
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
// The `homeworld` defines whether this planet
// is the homeworld for a player (which indicates
// that the name should be different from another
// random planet).
//
// The `data` allow to generate a default amount
// of resources on the planet.
//
// Returns the generated planet along with any error.
func NewPlanet(player string, coords Coordinate, homeworld bool, data Instance) (*Planet, error) {
	// Create default properties.
	p := &Planet{
		ID:          uuid.New().String(),
		Player:      player,
		Coordinates: coords,
		Name:        getDefaultPlanetName(homeworld),
		Fields:      0,
		MinTemp:     0,
		MaxTemp:     0,
		Diameter:    0,
		Resources:   make(map[string]ResourceInfo, 0),
		Buildings:   make(map[string]BuildingInfo, 0),
		Ships:       make(map[string]ShipInfo, 0),
		Defenses:    make(map[string]DefenseInfo, 0),
	}

	// Generate diameter and fields count.
	err := p.generateData(data)

	return p, err
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
//
// The `data ` will help generating the resources that are
// present on the planet upon starting it. It corresponds to
// a starting budget for the player to start building some
// stuff on the planet.
//
// Returns any error.
func (p *Planet) generateData(data Instance) error {
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

	switch p.Coordinates.Position {
	case 0:
		// Range [96; 172], average 134.
		min = 96
		max = 172
	case 1:
		// Range [104; 176], average 140.
		min = 104
		max = 176
	case 2:
		// Range [112; 182], average 147.
		min = 112
		max = 182
	case 3:
		// Range [118; 208], average 163.
		min = 118
		max = 208
	case 4:
		// Range [133; 232], average 182.
		min = 133
		max = 232
	case 5:
		// Range [152; 248], average 200.
		min = 152
		max = 248
	case 6:
		// Range [156; 262], average 204.
		min = 156
		max = 262
	case 7:
		// Range [150; 246], average 198.
		min = 150
		max = 246
	case 8:
		// Range [142; 232], average 187.
		min = 142
		max = 232
	case 9:
		// Range [136; 210], average 173.
		min = 136
		max = 210
	case 10:
		// Range [125; 186], average 156.
		min = 125
		max = 186
	case 11:
		// Range [114; 172], average 143.
		min = 114
		max = 172
	case 12:
		// Range [100; 168], average 134.
		min = 100
		max = 168
	case 13:
		// Range [90; 164], average 127.
		min = 96
		max = 164
	case 14:
		fallthrough
	default:
		// Assume default case if the `15th` position
		// Range [90; 164], average 134.
		min = 90
		max = 164
	}

	// Now that we have a valid range we should attempt to pick
	// random values in it. We would like to use a truncated
	// normal distribution for this matter. But it is not readily
	// available in Go so we will try to be clever (and probably
	// be a bit wrong from a mathematical point of view). We can
	// use the `rng.NormFloat64` function which generates normally
	// distributed values: the problem is that these values are
	// in the range `]-inf; +inf[` as expected from a NDF. However
	// we know that the range `]-3 * sigma; +3 * sigma[` will be
	// containing `99.37%` of the values. See here for more info:
	// https://en.wikipedia.org/wiki/68%E2%80%9395%E2%80%9399.7_rule)
	// So if we want *almost* all values to lie in the range
	// `[min; max]` we can use a standard deviation equal to
	// `(max - min) / 6` and we should be good to go.
	// A bit of clamping to sharpen the edges and it's close enough
	// for our purposes.
	stdDev := (max - min) / 6
	mean := (max + min) / 2

	fFields := float64(mean) + rng.NormFloat64()*float64(stdDev)
	clFFields := math.Max(float64(min), math.Min(float64(max), fFields))
	p.Fields = int(math.Round(clFFields))

	// The diameter is derived from the fields count with a random part.
	p.Diameter = 100*p.Fields + int(math.Round(float64(100.0*rand.Float32())))

	// The temperatures are described in the following link:
	// https://ogame.fandom.com/wiki/Temperature
	switch p.Coordinates.Position {
	case 0:
		// Range [220; 260], average 240.
		min = 220
		max = 260
	case 1:
		// Range [170; 210], average 190.
		min = 170
		max = 210
	case 2:
		// Range [120; 160], average 140.
		min = 120
		max = 160
	case 3:
		// Range [70; 110], average 90.
		min = 70
		max = 110
	case 4:
		// Range [60; 100], average 80.
		min = 60
		max = 100
	case 5:
		// Range [50; 90], average 70.
		min = 50
		max = 90
	case 6:
		// Range [40; 80], average 60.
		min = 40
		max = 80
	case 7:
		// Range [30; 70], average 50.
		min = 30
		max = 70
	case 8:
		// Range [20; 60], average 40.
		min = 20
		max = 60
	case 9:
		// Range [10; 50], average 30.
		min = 10
		max = 50
	case 10:
		// Range [0; 40], average 20.
		min = 0
		max = 40
	case 11:
		// Range [-10; 30], average 10.
		min = -10
		max = 30
	case 12:
		// Range [-50; -10], average -30.
		min = -50
		max = -10
	case 13:
		// Range [-90; -50], average -70.
		min = -90
		max = -50
	case 14:
		fallthrough
	default:
		// Assume default case if the `15th` position
		// Range [-130; -90], average -110.
		min = -130
		max = -90
	}

	// We will follow a similar process to the one described
	// for fields generation. Note that the ranges are used
	// to determine the value of the maximum temperature and
	// that the minimum one is just computed given a default
	// temperature amplitude.
	stdDev = (max - min) / 6
	mean = (max + min) / 2

	fMaxTemp := float64(mean) + rng.NormFloat64()*float64(stdDev)
	clFMaxTemp := math.Max(float64(min), math.Min(float64(max), fMaxTemp))
	p.MaxTemp = int(math.Round(clFMaxTemp))
	p.MinTemp = p.MaxTemp - 50

	// Generate default resources for this planet.
	p.Resources = make(map[string]ResourceInfo, 0)

	allRes, err := data.Resources.Resources(data.Proxy, []db.Filter{})
	if err != nil {
		return err
	}

	for _, res := range allRes {
		desc := ResourceInfo{
			Resource:   res.ID,
			Amount:     float32(res.BaseAmount),
			Storage:    float32(res.BaseStorage),
			Production: float32(res.BaseProd),
		}

		p.Resources[res.ID] = desc
	}

	return nil
}

// fetchGeneralInfo :
// Allows to fetch the general information of a planet
// from the DB such as its diameter, name, coordinates
// etc.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (p *Planet) fetchGeneralInfo(data Instance) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"player",
			"name",
			"min_temperature",
			"max_temperature",
			"fields",
			"galaxy",
			"solar_system",
			"position",
			"diameter",
		},
		Table: "planets",
		Filters: []db.Filter{
			{
				Key:    "id",
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

		p.Coordinates = NewPlanetCoordinate(galaxy, system, position)
	}

	return nil
}

// fetchBuildingsUpgrades :
// Similar to the `fetchGeneralInfo` method but used
// to fetch the buildings upgrades for a planet.
//
// The `data` defines a way to access to the DB.
//
// Returns any error.
func (p *Planet) fetchBuildingsUpgrades(data Instance) error {
	p.BuildingsUpgrade = make([]BuildingAction, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table: "construction_actions_buildings",
		Filters: []db.Filter{
			{
				Key:    "planet",
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

	// We now need to retrieve all the identifiers that matched
	// the input filters and then build the corresponding item
	// object for each one of them.
	var ID string
	IDs := make([]string, 0)

	for dbRes.Next() {
		err = dbRes.Scan(&ID)

		if err != nil {
			return err
		}

		IDs = append(IDs, ID)
	}

	for _, ID = range IDs {
		bu, err := NewBuildingActionFromDB(ID, data)

		if err != nil {
			return err
		}

		p.BuildingsUpgrade = append(p.BuildingsUpgrade, bu)
	}

	return nil
}

// fetchTechnologiesUpgrades :
// Similar to the `fetchGeneralInfo` method but used
// to fetch the technologies upgrades for a planet.
//
// The `data` defines a way to access to the DB.
//
// Returns any error.
func (p *Planet) fetchTechnologiesUpgrades(data Instance) error {
	p.TechnologiesUpgrade = make([]TechnologyAction, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table: "construction_actions_technologies",
		Filters: []db.Filter{
			{
				Key:    "player",
				Values: []interface{}{p.Player},
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

	// We now need to retrieve all the identifiers that matched
	// the input filters and then build the corresponding item
	// object for each one of them.
	var ID string
	IDs := make([]string, 0)

	for dbRes.Next() {
		err = dbRes.Scan(&ID)

		if err != nil {
			return err
		}

		IDs = append(IDs, ID)
	}

	for _, ID = range IDs {
		tu, err := NewTechnologyActionFromDB(ID, data)

		if err != nil {
			return err
		}

		p.TechnologiesUpgrade = append(p.TechnologiesUpgrade, tu)
	}

	return nil
}

// fetchShipsUpgrades :
// Similar to the `fetchGeneralInfo` method but used
// to fetch the ships upgrades for a planet.
//
// The `data` defines a way to access to the DB.
//
// Returns any error.
func (p *Planet) fetchShipsUpgrades(data Instance) error {
	p.ShipsConstruction = make([]ShipAction, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table: "construction_actions_ships",
		Filters: []db.Filter{
			{
				Key:    "planet",
				Values: []interface{}{p.ID},
			},
		},
		Ordering: "order by created_at",
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

	// We now need to retrieve all the identifiers that matched
	// the input filters and then build the corresponding item
	// object for each one of them.
	var ID string
	IDs := make([]string, 0)

	for dbRes.Next() {
		err = dbRes.Scan(&ID)

		if err != nil {
			return err
		}

		IDs = append(IDs, ID)
	}

	for _, ID = range IDs {
		su, err := NewShipActionFromDB(ID, data)

		if err != nil {
			return err
		}

		p.ShipsConstruction = append(p.ShipsConstruction, su)
	}

	return nil
}

// fetchDefensesUpgrades :
// Similar to the `fetchGeneralInfo` method but used
// to fetch the defenses upgrades for a planet.
//
// The `data` defines a way to access to the DB.
//
// Returns any error.
func (p *Planet) fetchDefensesUpgrades(data Instance) error {
	p.DefensesConstruction = make([]DefenseAction, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table: "construction_actions_defenses",
		Filters: []db.Filter{
			{
				Key:    "planet",
				Values: []interface{}{p.ID},
			},
		},
		Ordering: "order by created_at",
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

	// We now need to retrieve all the identifiers that matched
	// the input filters and then build the corresponding item
	// object for each one of them.
	var ID string
	IDs := make([]string, 0)

	for dbRes.Next() {
		err = dbRes.Scan(&ID)

		if err != nil {
			return err
		}

		IDs = append(IDs, ID)
	}

	for _, ID = range IDs {
		du, err := NewDefenseActionFromDB(ID, data)

		if err != nil {
			return err
		}

		p.DefensesConstruction = append(p.DefensesConstruction, du)
	}

	return nil
}

// fetchIncomingFleets :
// Used to fetch the incoming fleets from the DB for
// this planet. This include only the fleets having
// their target set for this planet. Hostiles along
// with friendly fleets will be fetched.
//
// The `data` defines the object to access the DB
// if needed.
//
// Returns any error.
func (p *Planet) fetchIncomingFleets(data Instance) error {
	p.IncomingFleets = make([]string, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table: "fleets",
		Filters: []db.Filter{
			{
				Key:    "target",
				Values: []interface{}{p.ID},
			},
			{
				Key:    "target_type",
				Values: []interface{}{"planet"},
			},
		},
		Ordering: "order by arrival_time desc",
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

	var ID string

	for dbRes.Next() {
		err = dbRes.Scan(&ID)

		if err != nil {
			return err
		}

		p.IncomingFleets = append(p.IncomingFleets, ID)
	}

	return nil
}

// fetchSourceFleets :
// Similar to the `fetchIncomingFleets` but retrieves
// the fleets that starts from this planet.
//
// The `data` defines the object to access the DB if
// needed.
//
// Returns any error.
func (p *Planet) fetchSourceFleets(data Instance) error {
	p.SourceFleets = make([]string, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table: "fleets",
		Filters: []db.Filter{
			{
				Key:    "source",
				Values: []interface{}{p.ID},
			},
			{
				Key:    "source_type",
				Values: []interface{}{"planet"},
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

	var ID string

	for dbRes.Next() {
		err = dbRes.Scan(&ID)

		if err != nil {
			return err
		}

		p.SourceFleets = append(p.SourceFleets, ID)
	}

	return nil
}

// fetchResources :
// Similar to the `fetchGeneralInfo` but handles the
// retrieval of the planet's resources data.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (p *Planet) fetchResources(data Instance) error {
	p.Resources = make(map[string]ResourceInfo, 0)

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
			&res.Production,
			&res.Storage,
		)

		if err != nil {
			return err
		}

		_, ok := sanity[res.Resource]
		if ok {
			return model.ErrInconsistentDB
		}
		sanity[res.Resource] = true

		p.Resources[res.Resource] = res
	}

	return nil
}

// fetchBuildings :
// Similar to the `fetchGeneralInfo` but handles the
// retrieval of the planet's buildings data.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (p *Planet) fetchBuildings(data Instance) error {
	p.Buildings = make(map[string]BuildingInfo, 0)

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
	var ID string
	var b BuildingInfo

	sanity := make(map[string]bool)

	for dbRes.Next() {
		err = dbRes.Scan(
			&ID,
			&b.Level,
		)

		if err != nil {
			return err
		}

		_, ok := sanity[ID]
		if ok {
			return model.ErrInconsistentDB
		}
		sanity[ID] = true

		desc, err := data.Buildings.GetBuildingFromID(ID)
		if err != nil {
			return err
		}

		b.BuildingDesc = desc

		p.Buildings[ID] = b
	}

	return nil
}

// fetchTechnologies :
// Similar to the `fetchGeneralInfo` but handles the
// retrieval of the technologies researched by the
// player owning the planet.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (p *Planet) fetchTechnologies(data Instance) error {
	p.technologies = make(map[string]int)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"technology",
			"level",
		},
		Table: "players_technologies",
		Filters: []db.Filter{
			{
				Key:    "player",
				Values: []interface{}{p.Player},
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
	var tech string
	var level int

	sanity := make(map[string]int)

	for dbRes.Next() {
		err = dbRes.Scan(
			&tech,
			&level,
		)

		if err != nil {
			return err
		}

		_, ok := sanity[tech]
		if ok {
			return model.ErrInconsistentDB
		}
		sanity[tech] = level

		p.technologies[tech] = level
	}

	return nil
}

// fetchShips :
// Similar to the `fetchGeneralInfo` but handles the
// retrieval of the planet's ships data.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (p *Planet) fetchShips(data Instance) error {
	p.Ships = make(map[string]ShipInfo, 0)

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
	var ID string
	var s ShipInfo

	sanity := make(map[string]bool)

	for dbRes.Next() {
		err = dbRes.Scan(
			&ID,
			&s.Amount,
		)

		if err != nil {
			return err
		}

		_, ok := sanity[ID]
		if ok {
			return model.ErrInconsistentDB
		}
		sanity[ID] = true

		desc, err := data.Ships.GetShipFromID(ID)
		if err != nil {
			return err
		}

		s.ShipDesc = desc

		p.Ships[ID] = s
	}

	return nil
}

// fetchShips :
// Similar to the `fetchGeneralInfo` but handles the
// retrieval of the planet's defenses data.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (p *Planet) fetchDefenses(data Instance) error {
	p.Defenses = make(map[string]DefenseInfo, 0)

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
	var ID string
	var d DefenseInfo

	sanity := make(map[string]bool)

	for dbRes.Next() {
		err = dbRes.Scan(
			&ID,
			&d.Amount,
		)

		if err != nil {
			return err
		}

		_, ok := sanity[ID]
		if ok {
			return model.ErrInconsistentDB
		}
		sanity[ID] = true

		desc, err := data.Defenses.GetDefenseFromID(ID)
		if err != nil {
			return err
		}

		d.DefenseDesc = desc

		p.Defenses[ID] = d
	}

	return nil
}

// SaveToDB :
// Used to save the content of this planet to
// the DB. In case an error is raised during
// the operation a comprehensive error is
// returned.
//
// The `proxy` allows to access to the DB.
//
// Returns any error.
func (p *Planet) SaveToDB(proxy db.Proxy) error {
	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_planet",
		Args: []interface{}{
			p,
			p.Resources,
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
		case "planets_pkey":
			return ErrDuplicatedElement
		}

		return dee
	}

	fkve, ok := dbe.Err.(db.ForeignKeyViolationError)
	if ok {
		switch fkve.ForeignKey {
		case "player":
			return ErrNonExistingPlayer
		}

		return fkve
	}

	return dbe
}

// Convert :
// Implementation of the `db.Convertible` interface
// from the DB package in order to only include fields
// that need to be marshalled in the planet's creation.
//
// Returns the converted version of the planet which
// only includes relevant fields.
func (p *Planet) Convert() interface{} {
	return struct {
		ID       string `json:"id"`
		Player   string `json:"player"`
		Name     string `json:"name"`
		MinTemp  int    `json:"min_temperature"`
		MaxTemp  int    `json:"max_temperature"`
		Fields   int    `json:"fields"`
		Galaxy   int    `json:"galaxy"`
		System   int    `json:"solar_system"`
		Position int    `json:"position"`
		Diameter int    `json:"diameter"`
	}{
		ID:       p.ID,
		Player:   p.Player,
		Name:     p.Name,
		MinTemp:  p.MinTemp,
		MaxTemp:  p.MaxTemp,
		Fields:   p.Fields,
		Galaxy:   p.Coordinates.Galaxy,
		System:   p.Coordinates.System,
		Position: p.Coordinates.Position,
		Diameter: p.Diameter,
	}
}

// MarshalJSON :
// Implementation of the `Marshaler` interface to allow
// only specific information to be marshalled when the
// planet needs to be exported. Indeed we don't want to
// export all the fields of all the elements defined as
// members of the planet.
// Most of the info will be marshalled except for deps
// on various buildings/ships/defenses built on this
// planet as it's not the place of this struct to be
// defining that.
// The approach we follow is to define a similar struct
// to the planet but do not include the tech deps.
//
// Returns the marshalled bytes for this planet along
// with any error.
func (p *Planet) MarshalJSON() ([]byte, error) {
	type lightInfo struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Level int    `json:"level"`
	}

	type lightCount struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Amount int    `json:"amount"`
	}

	type lightPlanet struct {
		ID                   string             `json:"id"`
		Player               string             `json:"player"`
		Coordinates          Coordinate         `json:"coordinate"`
		Name                 string             `json:"name"`
		Fields               int                `json:"fields"`
		MinTemp              int                `json:"min_temperature"`
		MaxTemp              int                `json:"max_temperature"`
		Diameter             int                `json:"diameter"`
		Resources            []ResourceInfo     `json:"resources"`
		Buildings            []lightInfo        `json:"buildings,omitempty"`
		Ships                []lightCount       `json:"ships,omitempty"`
		Defenses             []lightCount       `json:"defenses,omitempty"`
		BuildingsUpgrade     []BuildingAction   `json:"buildings_upgrade"`
		TechnologiesUpgrade  []TechnologyAction `json:"technologies_upgrade"`
		ShipsConstruction    []ShipAction       `json:"ships_construction"`
		DefensesConstruction []DefenseAction    `json:"defenses_construction"`
		SourceFleets         []string           `json:"source_fleets"`
		IncomingFleets       []string           `json:"incoming_fleets"`
	}

	// Copy the planet's data.
	lp := lightPlanet{
		ID:                   p.ID,
		Player:               p.Player,
		Coordinates:          p.Coordinates,
		Name:                 p.Name,
		Fields:               p.Fields,
		MinTemp:              p.MinTemp,
		MaxTemp:              p.MaxTemp,
		Diameter:             p.Diameter,
		BuildingsUpgrade:     p.BuildingsUpgrade,
		TechnologiesUpgrade:  p.TechnologiesUpgrade,
		ShipsConstruction:    p.ShipsConstruction,
		DefensesConstruction: p.DefensesConstruction,
		SourceFleets:         p.SourceFleets,
		IncomingFleets:       p.IncomingFleets,
	}

	// Copy resources from map to slice.
	for _, r := range p.Resources {
		lp.Resources = append(lp.Resources, r)
	}

	// Make shallow copy of the buildings, ships and
	// defenses without including the tech deps.
	for _, b := range p.Buildings {
		lb := lightInfo{
			ID:    b.ID,
			Name:  b.Name,
			Level: b.Level,
		}

		lp.Buildings = append(lp.Buildings, lb)
	}

	for _, s := range p.Ships {
		ls := lightCount{
			ID:     s.ID,
			Name:   s.Name,
			Amount: s.Amount,
		}

		lp.Ships = append(lp.Ships, ls)
	}

	for _, d := range p.Defenses {
		ld := lightCount{
			ID:     d.ID,
			Name:   d.Name,
			Amount: d.Amount,
		}

		lp.Defenses = append(lp.Defenses, ld)
	}

	return json.Marshal(lp)
}

// validateAction :
// Used to make sure that the action described by
// the costs and tech dependencies in input can be
// performed given the infrastructure and resources
// available on the planet.
//
// The `costs` defines the costs associated to the
// action.
//
// The `desc` defines a base description of the
// element attached to the action: it mainly is
// used to get an idea of the dependencies that
// need to be met for the element to be built on
// this planet.
//
// The `data` allows to access to the DB if needed.
//
// Returns any error. In case the return value is
// `nil` it means that the action can be performed
// on this planet.
func (p *Planet) validateAction(costs map[string]int, desc model.UpgradableDesc, data Instance) error {
	// Make sure that there are enough resources on the planet.
	if len(costs) == 0 {
		return ErrNoCost
	}

	for res, amount := range costs {
		// Find the amount existing on the planet. We
		// will consider that if the resource does not
		// exist on the internal array it means that
		// no amount is present.
		desc, ok := p.Resources[res]

		if !ok || desc.Amount < float32(amount) {
			return ErrNotEnoughResources
		}
	}

	// Make sure that the tech tree is consistent with the
	// expectations.
	for _, bDep := range desc.BuildingsDeps {
		bi, ok := p.Buildings[bDep.ID]

		if !ok || bi.Level < bDep.Level {
			return ErrTechDepsNotMet
		}
	}

	for _, tDep := range desc.TechnologiesDeps {
		level, ok := p.technologies[tDep.ID]

		if !ok || level < tDep.Level {
			return ErrTechDepsNotMet
		}
	}

	// Seems like all conditions are valid.
	return nil
}

// validateFleet :
// Used to make sure that the data described by
// the component in input is consistent with the
// data actually contained in the planet. It is
// meant to check that the requested resources
// can be provided by the planet and that there
// is enough fuel to make the ships move.
//
// The `fuels` defines the amount of fuel that
// is needed by the component to move.
//
// The `cargos` defines the amount of resources
// that should be taken by the component.
//
// The `ships` defines the list of ships that
// should be subtracted to the current fleet
// deployed at this planet. Each value should
// be at most equal to the total number of
// ships of this kind on the planet.
//
// The `data` allows to access to the DB in
// case it's needed.
//
// Returns any error. The behavior is similar
// to the `validateAction` where no error is
// meant to indicate that the component is a
// valid one compared to the planet's data.
func (p *Planet) validateFleet(fuels []model.ResourceAmount, cargos []model.ResourceAmount, ships []ShipInFleet, data Instance) error {
	// Gather existing resources.
	available := make(map[string]float32)

	for _, res := range p.Resources {
		ex := available[res.Resource]
		ex += res.Amount
		available[res.Resource] = ex
	}

	// Make sure that there's enough fuel.
	for _, fuel := range fuels {
		res := available[fuel.Resource]

		if res < fuel.Amount {
			return ErrNotEnoughFuel
		}

		res -= fuel.Amount
		available[fuel.Resource] = res
	}

	// Make sure that there's enough resources.
	for _, cargo := range cargos {
		res := available[cargo.Resource]

		if res < cargo.Amount {
			return ErrNotEnoughResources
		}

		res -= cargo.Amount
		available[cargo.Resource] = res
	}

	// Make sure that there's enough ships.
	for _, ship := range ships {
		s, ok := p.Ships[ship.ID]

		if !ok || s.Amount < ship.Count {
			return ErrNotEnoughShips
		}
	}

	return nil
}
