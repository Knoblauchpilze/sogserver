package model

import (
	"fmt"
	"oglike_server/pkg/db"
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

// ErrInvalidUniverse :
// Used to indicate that the universe provided in input is
// not valid.
var ErrInvalidUniverse = fmt.Errorf("Invalid universe with no identifier")

// ErrDuplicatedUniverse :
// Used to indicate that the universe's identifier provided
// input is not unique in the DB.
var ErrDuplicatedUniverse = fmt.Errorf("Invalid not unique universe")

// ErrDuplicatedCoordinates :
// Used to indicate that some coordinates used in a process
// was actually already existing.
var ErrDuplicatedCoordinates = fmt.Errorf("Invalid duplicated coordinates")

// Valid :
// Used to determine whether the parameters defined for this
// universe are consistent with what is expected. This will
// typically check that the ratios are in the range `[0; 1]`
// and some other common assumptions.
// Note that it requires that the `ID` is valid as well.
//
// Returns `true` if the universe is valid (i.e. all values
// are consistent with the expected ranges).
func (u *Universe) Valid() bool {
	return validUUID(u.ID) &&
		u.Name != "" &&
		u.EcoSpeed > 0 &&
		u.FleetSpeed > 0 &&
		u.ResearchSpeed > 0 &&
		u.FleetsToRuins >= 0.0 && u.FleetsToRuins <= 1.0 &&
		u.DefensesToRuins >= 0.0 && u.DefensesToRuins <= 1.0 &&
		u.FleetConsumption >= 0.0 && u.FleetConsumption <= 1.0 &&
		u.GalaxiesCount > 0 &&
		u.GalaxySize > 0 &&
		u.SolarSystemSize > 0
}

// String :
// Implementation of the `Stringer` interface to make
// sure displaying this universe is easy.
//
// Returns the corresponding string.
func (u Universe) String() string {
	return fmt.Sprintf("[id: %s, name: \"%s\"]", u.ID, u.Name)
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
func NewUniverseFromDB(ID string, data Instance) (Universe, error) {
	// Create the universe.
	u := Universe{
		ID: ID,
	}

	// Consistency.
	if u.ID == "" {
		return u, ErrInvalidUniverse
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
				Values: []string{u.ID},
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
		return u, ErrInvalidUniverse
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
		return u, ErrDuplicatedUniverse
	}

	return u, err
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
				Values: []string{u.ID},
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
