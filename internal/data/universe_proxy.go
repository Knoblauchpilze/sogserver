package data

import (
	"encoding/json"
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"

	"github.com/google/uuid"
)

// UniverseProxy :
// Intended as a wrapper to access properties of universes
// and retrieve data from the database. This helps hiding
// the complexity of how the data is laid out in the `DB`
// and the precise name of tables from the exterior world.
//
// The `dbase` is the database that is wrapped by this
// object. It is checked for consistency upon building the
// wrapper.
//
// The `log` allows to perform display to the user so as
// to inform of potential issues and debug information to
// the outside world.
type UniverseProxy struct {
	dbase *db.DB
	log   logger.Logger
}

// NewUniverseProxy :
// Create a new proxy on the input `dbase` to access the
// properties of universes as registered in the DB. In
// case the provided DB is `nil` a panic is issued.
// Information in the following thread helped shape this
// component (and the similar ones in the package):
// https://www.reddit.com/r/golang/comments/9i5cpg/good_approach_to_interacting_with_databases/
//
// The `dbase` represents the database to use to fetch
// data related to universes.
//
// The `log` will be used to notify information so that
// we can have an idea of the activity of this component.
// One possible example is for timing the requests.
//
// Returns the created proxy.
func NewUniverseProxy(dbase *db.DB, log logger.Logger) UniverseProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create universes proxy from invalid DB"))
	}

	return UniverseProxy{dbase, log}
}

// GetIdentifierDBColumnName :
// Used to retrieve the string literal defining the name of the
// identifier column in the `universes` table in the database.
//
// Returns the name of the `identifier` column in the database.
func (p *UniverseProxy) GetIdentifierDBColumnName() string {
	return "id"
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
func (p *UniverseProxy) Universes(filters []DBFilter) ([]Universe, error) {
	// Create the query and execute it.
	props := []string{
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
	}

	query := fmt.Sprintf("select %s from universes", strings.Join(props, ", "))
	if len(filters) > 0 {
		query += " where"

		for id, filter := range filters {
			if id > 0 {
				query += " and"
			}
			query += fmt.Sprintf(" %s", filter)
		}
	}

	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch universes (err: %v)", err)
	}

	// Populate the return value.
	universes := make([]Universe, 0)
	var uni Universe

	for rows.Next() {
		err = rows.Scan(
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
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for universe (err: %v)", err))
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

	// Marshal the input universe to pass it to the import script.
	data, err := json.Marshal(uni)
	if err != nil {
		return fmt.Errorf("Could not import universe \"%s\" (err: %v)", uni.Name, err)
	}
	jsonToSend := string(data)

	query := fmt.Sprintf("select * from create_universe('%s')", jsonToSend)
	_, err = p.dbase.DBExecute(query)

	// Check for errors.
	if err != nil {
		return fmt.Errorf("Could not import universe \"%s\" (err: %v)", uni.Name, err)
	}

	// All is well.
	return nil
}
