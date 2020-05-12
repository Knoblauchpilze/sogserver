package game

import (
	"fmt"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"strconv"
)

// Universe :
// Define a universe in terms of OG semantic. This is a set
// of planets gathered in a certain number of galaxies and
// a set of parameters that configure the economic, combat
// and technologies available in it.
//
// The `ID` defines the unique identifier for this universe.
//
// The `Name` defines a human-readable name for it.
//
// The `EcoSpeed` is a value in the range `[0; inf]` which
// defines a multiplication factor that is added to shorten
// the economy (i.e. building construction time, etc.).
//
// The `FleetSpeed` is similar to the `EcoSpeed` but controls
// the speed boost for fleets travel time.
//
// The `ResearchSpeed` controls how researches are shortened
// compared to the base value.
//
// The `FleetsToRuins` defines the percentage of resources
// that go into a debris fields when a ship is destroyed in
// a battle.
//
// The `DefensesToRuins` defines a similar percentage for
// defenses in the event of a battle.
//
// The `FleetConsumption` is a value in the range `[0; 1]`
// defining how the consumption is biased compared to the
// canonical value.
//
// The `GalaxiesCount` defines the number of galaxies in
// the universe.
//
// The `GalaxySize` defines the number of solar systems
// in a single galaxy.
//
// The `SolarSystemSize` defines the number of planets in
// each solar system of each galaxy.
type Universe struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	EcoSpeed         int     `json:"economic_speed"`
	FleetSpeed       int     `json:"fleet_speed"`
	ResearchSpeed    int     `json:"research_speed"`
	FleetsToRuins    float32 `json:"fleets_to_ruins_ratio"`
	DefensesToRuins  float32 `json:"defenses_to_ruins_ratio"`
	FleetConsumption float32 `json:"fleets_consumption_ratio"`
	GalaxiesCount    int     `json:"galaxies_count"`
	GalaxySize       int     `json:"galaxy_size"`
	SolarSystemSize  int     `json:"solar_system_size"`
}

// ErrDuplicatedCoordinates : Indicates that some coordites appeared twice.
var ErrDuplicatedCoordinates = fmt.Errorf("Invalid duplicated coordinates")

// ErrPlanetNotFound : No planet exists at the specified coordinates.
var ErrPlanetNotFound = fmt.Errorf("No planet at the specified coordinates")

// ErrInvalidCoordinates : Input coordinates are not valid given the universe structure.
var ErrInvalidCoordinates = fmt.Errorf("Invalid coordinates relative to universe structure")

// ErrDuplicatedPlanet : Indicates that there several planets share the same coordinates.
var ErrDuplicatedPlanet = fmt.Errorf("Several planets share the same coordinates")

// ErrInvalidEcoSpeed : The economic speed is not within admissible range.
var ErrInvalidEcoSpeed = fmt.Errorf("Economic speed is not within admissible range")

// ErrInvalidFleetSpeed : The fleet speed is not within admissible range.
var ErrInvalidFleetSpeed = fmt.Errorf("Fleet speed is not within admissible range")

// ErrInvalidResearchSpeed : The research speed is not within admissible range.
var ErrInvalidResearchSpeed = fmt.Errorf("Research speed is not within admissible range")

// ErrFleetsToRuins : The fleets to ruins ratio is not within admissible range.
var ErrFleetsToRuins = fmt.Errorf("Fleets to ruins ratio is not within admissible range")

// ErrDefensesToRuins : The defenses to ruins ratio is not within admissible range.
var ErrDefensesToRuins = fmt.Errorf("Defenses to ruins is not within admissible range")

// ErrFleetConsumption : The fleet consumption is not within admissible range.
var ErrFleetConsumption = fmt.Errorf("Fleet consumption is not within admissible range")

// ErrGalaxiesCount : The number of galaxies is not within admissible range.
var ErrGalaxiesCount = fmt.Errorf("Galaxies count is not within admissible range")

// ErrGalaxySize : The size of a galaxy is not within admissible range.
var ErrGalaxySize = fmt.Errorf("Galaxy size is not within admissible range")

// ErrSolarSystemSize : The size of a solar system is not within admissible range.
var ErrSolarSystemSize = fmt.Errorf("Solar system size is not within admissible range")

// valid :
// Determines whether the universe is valid. By valid we only
// mean obvious syntax errors.
//
// Returns any error or `nil` if the universe seems valid.
func (u *Universe) valid() error {
	if !validUUID(u.ID) {
		return ErrInvalidElementID
	}
	if u.Name == "" {
		return ErrInvalidName
	}
	if u.EcoSpeed <= 0 {
		return ErrInvalidEcoSpeed
	}
	if u.FleetSpeed <= 0 {
		return ErrInvalidFleetSpeed
	}
	if u.ResearchSpeed <= 0 {
		return ErrInvalidResearchSpeed
	}
	if u.FleetsToRuins < 0.0 && u.FleetsToRuins > 1.0 {
		return ErrFleetsToRuins
	}
	if u.DefensesToRuins < 0.0 && u.DefensesToRuins > 1.0 {
		return ErrDefensesToRuins
	}
	if u.FleetConsumption < 0.0 && u.FleetConsumption > 1.0 {
		return ErrFleetConsumption
	}
	if u.GalaxiesCount <= 0 {
		return ErrGalaxiesCount
	}
	if u.GalaxySize <= 0 {
		return ErrGalaxySize
	}
	if u.SolarSystemSize <= 0 {
		return ErrSolarSystemSize
	}

	return nil
}

// NewUniverseFromDB :
// Used to fetch the content of the universe from
// the input DB and populate all internal fields
// from it. In case the DB cannot be fetched or
// some errors are encoutered, the return value
// will include a description of the error.
//
// The `ID` defines the ID of the universe to get.
// It should be fetched from the DB and is assumed
// to refer to an existing universe.
//
// The `data` allows to actually perform the DB
// requests to fetch the universe's data.
//
// Returns the universe as fetched from the DB
// along with any errors.
func NewUniverseFromDB(ID string, data model.Instance) (Universe, error) {
	// Create the universe.
	u := Universe{
		ID: ID,
	}

	// Consistency.
	if !validUUID(u.ID) {
		return u, ErrInvalidElementID
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"name",
			"economic_speed",
			"fleet_speed",
			"research_speed",
			"fleets_to_ruins_ratio",
			"defenses_to_ruins_ratio",
			"fleets_consumption_ratio",
			"galaxies_count",
			"galaxy_size",
			"solar_system_size",
		},
		Table: "universes",
		Filters: []db.Filter{
			{
				Key:    "id",
				Values: []interface{}{u.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return u, err
	}
	if dbRes.Err != nil {
		return u, dbRes.Err
	}

	// Scan the universe's data.
	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return u, ErrElementNotFound
	}

	err = dbRes.Scan(
		&u.Name,
		&u.EcoSpeed,
		&u.FleetSpeed,
		&u.ResearchSpeed,
		&u.FleetsToRuins,
		&u.DefensesToRuins,
		&u.FleetConsumption,
		&u.GalaxiesCount,
		&u.GalaxySize,
		&u.SolarSystemSize,
	)

	// Make sure that it's the only universe.
	if dbRes.Next() {
		return u, ErrDuplicatedElement
	}

	return u, err
}

// SaveToDB :
// Used to save the content of this universe to
// the DB. In case an error is raised during the
// operation a comprehensive error is returned.
//
// The `proxy` allows to access to the DB.
//
// Returns any error.
func (u *Universe) SaveToDB(proxy db.Proxy) error {
	// Check consistency.
	if err := u.valid(); err != nil {
		return err
	}

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_universe",
		Args:   []interface{}{u},
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
		case "universes_pkey":
			return ErrDuplicatedElement
		case "universes_name_key":
			return ErrInvalidName
		}
	}

	return dbe
}

// UsedCoords :
// Used to find and generate a list of the used coordinates
// in this universe. Note that the list is only some snapshot
// of the state of the coordinates which can evolve through
// time. Typically if some pending requests to insert a planet
// are outstanding or some fleets are registered which will
// ultimately lead to the colonization/destruction of some
// planet this list will change.
// In order to be practical the list of used coordinates is
// returned using a map. The keys correspond to the coords
// where the `Linearize` method with this universe as param
// as been called and the values are the raw coordinates.
//
// The `proxy` defines a way to access the DB to fetch the
// used coordinates.
//
// Returns the list of used coordinates along with any error.
func (u *Universe) UsedCoords(proxy db.Proxy) (map[int]Coordinate, error) {
	// Create the query allowing to fetch all the planets of
	// a specific universe. This will consistute the list of
	// used planets for this universe.
	query := db.QueryDesc{
		Props: []string{
			"p.galaxy",
			"p.solar_system",
			"p.position",
		},
		Table: "planets p inner join players pl on p.player=pl.id",
		Filters: []db.Filter{
			{
				Key:    "pl.universe",
				Values: []interface{}{u.ID},
			},
		},
	}

	dbRes, err := proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return map[int]Coordinate{}, err
	}
	if dbRes.Err != nil {
		return map[int]Coordinate{}, dbRes.Err
	}

	// Traverse all the coordinates and populate the list.
	coords := make(map[int]Coordinate)
	var coord Coordinate

	// Only planets are fetched by this function.
	coord.Type = World

	for dbRes.Next() {
		err = dbRes.Scan(
			&coord.Galaxy,
			&coord.System,
			&coord.Position,
		)

		if err != nil {
			return coords, db.ErrInvalidScan
		}

		key := coord.Linearize(u.GalaxySize, u.SolarSystemSize)

		// In case these coordinates already exist this is an issue.
		if _, ok := coords[key]; ok {
			return coords, ErrDuplicatedCoordinates
		}

		coords[key] = coord
	}

	return coords, nil
}

// GetPlanetAt :
// Used to attempt to retrieve the planet that exists at
// the specified coordinates. In case no planet exists
// a `ErrPlanetNotFound` error will be returned.
//
// The `coord` defines the coordinates from which a planet
// should be fetched.
//
// The `data` allows to access to the DB to fetch the
// planet's data.
//
// Returns the planet at the specified coordinates (or
// `nil` in case no planet exists) along with any error.
func (u *Universe) GetPlanetAt(coord Coordinate, data model.Instance) (*Planet, error) {
	// Make sure that the coordinate are valid for this universe.
	if !coord.valid(u.GalaxiesCount, u.GalaxySize, u.SolarSystemSize) {
		return nil, ErrInvalidCoordinates
	}

	// Create the query to fetch the planet from the coordinates.
	// Create the query and execute it.
	gas := strconv.Itoa(coord.Galaxy)
	sas := strconv.Itoa(coord.System)
	pas := strconv.Itoa(coord.Position)

	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table: "planets",
		Filters: []db.Filter{
			{
				Key:    "galaxy",
				Values: []interface{}{gas},
			},
			{
				Key:    "solar_system",
				Values: []interface{}{sas},
			},
			{
				Key:    "position",
				Values: []interface{}{pas},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return nil, err
	}
	if dbRes.Err != nil {
		return nil, dbRes.Err
	}

	// Scan the planet's data.
	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return nil, ErrPlanetNotFound
	}

	var ID string

	err = dbRes.Scan(
		&ID,
	)

	// Make sure that it's the only universe.
	if dbRes.Next() {
		return nil, ErrDuplicatedPlanet
	}

	// Fetch the planet using read write semantic.
	p, err := NewPlanetFromDB(ID, data)

	return &p, err
}
