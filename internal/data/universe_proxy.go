package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"

	"github.com/google/uuid"
)

// UniverseProxy :
// Intended as a wrapper to access properties of universes
// and retrieve data from the database. Internally uses the
// common proxy defined in this package.
type UniverseProxy struct {
	commonProxy
}

// NewUniverseProxy :
// Create a new proxy allowing to serve the requests
// related to universes.
//
// The `dbase` represents the database to use to fetch
// data related to universes.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewUniverseProxy(dbase *db.DB, log logger.Logger) UniverseProxy {
	return UniverseProxy{
		newCommonProxy(dbase, log, "universes"),
	}
}

// Universes :
// Allows to fetch the list of universes currently available
// for a player to create an account. Universes should only
// be created when needed and are not typically something a
// player can do.
// The user can choose to filter parts of the universes using
// an array of filters that will be applied to the SQL query.
// No controls is enforced on the filters so one should make
// sure that it's consistent with the underlying table.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// universes available. Each one is appended `as-is` to the
// query.
//
// Returns the list of universes along with any errors. Note
// that in case the error is not `nil` the returned list is
// to be ignored.
func (p *UniverseProxy) Universes(filters []db.Filter) ([]Universe, error) {
	// Create the query and execute it.
	query := queryDesc{
		props: []string{
			"id",
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
		table:   "universes",
		filters: filters,
	}

	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return []Universe{}, fmt.Errorf("Could not query DB to fetch universes (err: %v)", err)
	}

	// Populate the return value.
	universes := make([]Universe, 0)
	var uni Universe

	for res.next() {
		err = res.scan(
			&uni.ID,
			&uni.Name,
			&uni.EcoSpeed,
			&uni.FleetSpeed,
			&uni.ResearchSpeed,
			&uni.FleetsToRuins,
			&uni.DefensesToRuins,
			&uni.FleetConsumption,
			&uni.GalaxiesCount,
			&uni.GalaxySize,
			&uni.SolarSystemSize,
		)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Could not retrieve info for universe (err: %v)", err))
			continue
		}

		universes = append(universes, uni)
	}

	return universes, nil
}

// Create :
// Used to perform the creation of the universe described
// by the input data to the DB. In case the creation cannot
// be performed an error is returned.
//
// The `uni` describes the element to create in DB. This
// value may be modified by the function mainly to update
// the identifier of the universe if none have been set.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`.
func (p *UniverseProxy) Create(uni *Universe) error {
	// Assign a valid identifier if this is not already the case.
	if uni.ID == "" {
		uni.ID = uuid.New().String()
	}

	// Validate that the input data describe a valid universe.
	if !uni.valid() {
		return fmt.Errorf("Could not create universe \"%s\", some properties are invalid", uni.Name)
	}

	// Create the query and execute it.
	query := insertReq{
		script: "create_universe",
		args:   []interface{}{*uni},
	}

	err := p.insertToDB(query)

	// Check for errors.
	if err != nil {
		return fmt.Errorf("Could not import universe \"%s\" (err: %v)", uni.Name, err)
	}

	p.trace(logger.Notice, fmt.Sprintf("Created new universe \"%s\" with id \"%s\"", uni.Name, uni.ID))

	return nil
}
