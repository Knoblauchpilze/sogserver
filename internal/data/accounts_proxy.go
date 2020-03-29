package data

import (
	"encoding/json"
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"

	"github.com/google/uuid"
)

// getDuplicatedElementErrorKey :
// Used to retrieve a string describing part of the error
// message issued by the database when trying to insert a
// duplicated element on a unique column. Can be used to
// standardize the definition of this error.
//
// Return part of the error string issued when inserting
// an already existing key.
func getDuplicatedElementErrorKey() string {
	return "SQLSTATE 23505"
}

// AccountProxy :
// Intended as a wrapper to access properties of accounts
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
type AccountProxy struct {
	dbase *db.DB
	log   logger.Logger
}

// NewAccountProxy :
// Create a new proxy on the input `dbase` to access the
// properties of accounts as registered in the DB.
// In case the provided DB is `nil` a panic is issued.
//
// The `dbase` represents the database to use to fetch
// data related to accounts.
//
// The `log` will be used to notify information so that
// we can have an idea of the activity of this component.
// One possible example is for timing the requests.
//
// Returns the created proxy.
func NewAccountProxy(dbase *db.DB, log logger.Logger) AccountProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create accounts proxy from invalid DB"))
	}

	return AccountProxy{dbase, log}
}

// Accounts :
// Allows to fetch the list of accounts currently registered
// in the DB. This defines how many unique players already
// have created at least an account in a universe.
// The user can choose to filter parts of the accounts using
// an array of filters that will be applied to the SQL query.
// No controls is enforced on the filters so one should make
// sure that it's consistent with the underlying table.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// accounts available. Each one is appended `as-is` to the
// query.
//
// Returns the list of accounts along with any errors. Note
// that in case the error is not `nil` the returned list is
// to be ignored.
func (p *AccountProxy) Accounts(filters []Filter) ([]Account, error) {
	// Create the query and execute it.
	query := fmt.Sprintf("select id, mail, name from accounts")
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
		return nil, fmt.Errorf("Could not query DB to fetch accounts (err: %v)", err)
	}

	// Populate the return value.
	accounts := make([]Account, 0)
	var acc Account

	for rows.Next() {
		err = rows.Scan(
			&acc.ID,
			&acc.Mail,
			&acc.Name,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for account (err: %v)", err))
			continue
		}

		accounts = append(accounts, acc)
	}

	return accounts, nil
}

// Characters :
// Return the list of universes into which the input user
// is registered. The input string is interpreted as the
// identifier of a player's account and use to query the
// corresponding information in the database.
//
// The `user` is a string representing the identifier of
// the account for this user.
//
// Returns the list of players' data registered for this
// account along with any error. In case the error is not
// `nil` the value of the array should be ignored.
func (p *AccountProxy) Characters(user string) ([]Player, error) {
	// Create the query and execute it.
	props := []string{
		"id",
		"uni",
		"player",
		"name",
	}

	query := fmt.Sprintf("select %s from players", strings.Join(props, ", "))
	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch players (err: %v)", err)
	}

	// Populate the return value.
	players := make([]Player, 0)
	var player Player

	for rows.Next() {
		err = rows.Scan(
			&player.ID,
			&player.UniverseID,
			&player.AccountID,
			&player.Name,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for player (err: %v)", err))
			continue
		}

		players = append(players, player)
	}

	return players, nil
}

// Planets :
// Return a list of planets associated to the player for
// the relevant universe and account. It queries the DB
// to fetch the relevant data and return it through an
// array of planets.
//
// The `player` describes the identifier of the account
// for which planets should be fetched. We assume that
// it contains valid data. If this is not the case no
// planets will likely be retrieved.
//
// Returns the list of planets for this account along
// with any error. In case the error is not `nil` the
// value of the array should be ignored.
func (p *AccountProxy) Planets(player string) ([]Planet, error) {
	// Create the query and execute it.
	props := []string{
		"id",
		"player",
		"name",
		"fields",
		"min_temperature",
		"max_temperature",
		"diameter",
		"galaxy",
		"solar_system",
		"position",
	}

	query := fmt.Sprintf("select %s from planets where id='%s'", strings.Join(props, ", "), player)
	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch planets for player \"%s\" (err: %v)", player, err)
	}

	// Populate the return value.
	planets := make([]Planet, 0)
	var planet Planet

	galaxy := 0
	system := 0
	position := 0

	for rows.Next() {
		err = rows.Scan(
			&planet.ID,
			&planet.PlayerID,
			&planet.Name,
			&planet.Fields,
			&planet.MinTemp,
			&planet.MaxTemp,
			&planet.Diameter,
			&galaxy,
			&system,
			&position,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve planet for player \"%s\" (err: %v)", player, err))
			continue
		}

		planet.Coords = Coordinate{
			galaxy,
			system,
			position,
		}

		planets = append(planets, planet)
	}

	return planets, nil
}

// Technologies :
// Return a list of the current technologies developed
// on the account of the specified player. It queries
// the DB to fetch the relevant data and return it
// through an array.
//
// The `player` describes the account for which techno
// should be fetched. We assume that it contains valid
// data. If this is not the case no technologies will
// likely be retrieved.
//
// Returns the list of technologies for this account
// along with any error. In case the error is not `nil`
// the value of the array should be ignored.
func (p *AccountProxy) Technologies(player Player) ([]Technology, error) {
	// Create the query and execute it.
	props := []string{
		"pt.player",
		"pt.level",
		"t.name",
	}

	table := "player_technologies pt inner join technologies t"
	joinCond := "pt.technology=t.id"

	query := fmt.Sprintf("select %s from %s on %s", strings.Join(props, ", "), table, joinCond)
	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch technologies for player \"%s\" (err: %v)", player.ID, err)
	}

	// Populate the return value.
	technologies := make([]Technology, 0)
	var tech Technology

	for rows.Next() {
		err = rows.Scan(
			&tech.ID,
			&tech.Level,
			&tech.Name,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve technology for player \"%s\" (err: %v)", player.ID, err))
			continue
		}

		technologies = append(technologies, tech)
	}

	return technologies, nil
}

// Fleets :
// Return a list of the current fleets deployed on the
// account of the player. It is a convenience method
// compared to fetching the fleet of a single planet
// and looping on all the planets of the account.
// The internal DB is queried to fetch the relevant
// information.
//
// The `player` describes the account for which fleets
// should be fetched. We assume that it contains valid
// data. If this is not the case no fleets will likely
// be retrieved.
//
// Returns the list of fleets for this account along
// with any error. In case the error is not `nil` the
// value of the array should be ignored.
func (p *AccountProxy) Fleets(player Player) ([]Fleet, error) {
	// Create the query and execute it.
	props := []string{
		"f.id",
		"f.name",
		"fo.name",
		"f.arrival_time",
		"f.galaxy",
		"f.solar_system",
		"f.position",
	}

	tables := "fleet_ships fs inner join fleets f on fs.fleet=f.id inner join fleet_objectives fo on f.objective=fo.id"
	where := fmt.Sprintf("fs.player='%s'", player.ID)

	query := fmt.Sprintf("select %s from %s where %s", strings.Join(props, ", "), tables, where)
	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch fleets for player \"%s\" (err: %v)", player.ID, err)
	}

	// Populate the return value.
	fleets := make([]Fleet, 0)
	var fleet Fleet

	galaxy := 0
	system := 0
	position := 0

	for rows.Next() {
		err = rows.Scan(
			&fleet.ID,
			&fleet.Name,
			&fleet.Objective,
			&fleet.ArrivalTime,
			&galaxy,
			&system,
			&position,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve fleet for player \"%s\" (err: %v)", player.ID, err))
			continue
		}

		fleet.Coords = Coordinate{
			galaxy,
			system,
			position,
		}

		fleets = append(fleets, fleet)
	}

	return fleets, nil
}

// Create :
// Used to perform the creation of the account described
// by the input data to the DB. In case the creation can
// not be performed an error is returned.
//
// The `acc` describes the element to create in DB. This
// value may be modified by the function mainly to update
// the identifier of the account if none have been set.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`.
func (p *AccountProxy) Create(acc *Account) error {
	// Assign a valid identifier if this is not already the case.
	if acc.ID == "" {
		acc.ID = uuid.New().String()
	}

	// Validate that the input data describe a valid universe.
	if !acc.valid() {
		return fmt.Errorf("Could not create account \"%s\", some properties are invalid", acc.Name)
	}

	// Marshal the input universe to pass it to the import script.
	data, err := json.Marshal(acc)
	if err != nil {
		return fmt.Errorf("Could not import account \"%s\" (err: %v)", acc.Name, err)
	}
	jsonToSend := string(data)

	query := fmt.Sprintf("select * from create_account('%s')", jsonToSend)
	_, err = p.dbase.DBExecute(query)

	// Check for errors. We will refine this process a bit to try
	// to detect cases where the user tries to insert an account
	// with an already existing e-mail.
	// In this case we should get an error indicating a `23505` as
	// return code. We will refine the error in this case.
	if err != nil {
		// Check for duplicated key error.
		msg := fmt.Sprintf("%v", err)

		if strings.Contains(msg, getDuplicatedElementErrorKey()) {
			return fmt.Errorf("Could not import account \"%s\", mail \"%s\" already exists (err: %s)", acc.Name, acc.Mail, msg)
		}

		return fmt.Errorf("Could not import account \"%s\" (err: %s)", acc.Name, msg)
	}

	// Successfully created an account.
	p.log.Trace(logger.Notice, fmt.Sprintf("Created new account \"%s\" with id \"%s\"", acc.Name, acc.ID))

	// All is well.
	return nil
}
