package data

import (
	"encoding/json"
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"

	"github.com/google/uuid"
)

// PlayersProxy :
// Intended as a wrapper to access properties of players
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
type PlayersProxy struct {
	dbase *db.DB
	log   logger.Logger
}

// NewPlayersProxy :
// Create a new proxy on the input `dbase` to access the
// properties of players as registered in the DB.
// In case the provided DB is `nil` a panic is issued.
//
// The `dbase` represents the database to use to fetch
// data related to players.
//
// The `log` will be used to notify information so that
// we can have an idea of the activity of this component.
// One possible example is for timing the requests.
//
// Returns the created proxy.
func NewPlayersProxy(dbase *db.DB, log logger.Logger) PlayersProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create players proxy from invalid DB"))
	}

	return PlayersProxy{dbase, log}
}

// Players :
// Allows to fetch the list of players currently registered
// in the DB. This defines how many unique players already
// have created at least an player in a universe.
// The user can choose to filter parts of the players using
// an array of filters that will be applied to the SQL query.
// No controls is enforced on the filters so one should make
// sure that it's consistent with the underlying table.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// players available. Each one is appended `as-is` to the
// query.
//
// Returns the list of players along with any errors. Note
// that in case the error is not `nil` the returned list is
// to be ignored.
func (p *PlayersProxy) Players(filters []DBFilter) ([]Player, error) {
	// Create the query and execute it.
	query := fmt.Sprintf("select id, uni, account, name from players")
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

		// Populate the technologies researched by this player.
		err = p.fetchPlayerData(&player)
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch data for player \"%s\" (err: %v)", player.ID, err))
			continue
		}

		players = append(players, player)
	}

	return players, nil
}

// fetchPlayerData :
// Used to fetch data related to the player in argument. It
// mainly consists in the list of technologies researched by
// the player.
//
// The `player` references the player for which data should
// be fetched. We assume that the internal fields (and more
// specifically the identifier) are already populated.
//
// Returns any error.
func (p *PlayersProxy) fetchPlayerData(player *Player) error {
	// Check whether the player has an identifier assigned.
	if player.ID == "" {
		return fmt.Errorf("Unable to fetch data from player with invalid identifier")
	}

	// Fetch technologies.
	query := fmt.Sprintf("select technology, level from player_technologies where player='%s'", player.ID)
	rows, err := p.dbase.DBQuery(query)

	if err != nil {
		return fmt.Errorf("Could not fetch technologies for player \"%s\" (err: %v)", player.ID, err)
	}

	player.Technologies = make([]Technology, 0)
	var tech Technology

	for rows.Next() {
		err = rows.Scan(
			&tech.ID,
			&tech.Level,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve technology for player \"%s\" (err: %v)", player.ID, err))
			continue
		}

		player.Technologies = append(player.Technologies, tech)
	}

	return nil
}

// Create :
// Used to perform the creation of the player described
// by the input structure in the DB. The player is both
// associated to an account and a universe. The database
// guarantees that no two players can exist in the same
// universe belonging to the same account.
// This method also handles the creation of the needed
// data for a player to truly be ready to start in a new
// universe (which means creating a homeworld).
//
// The `player` describes the element to create in `DB`.
// This value may be modified by the function in case it
// does not define a valid identifier.
//
// The return status indicates whether the player could
// be created or not (in which case an error describes
// the failure reason).
func (p *PlayersProxy) Create(player *Player) error {
	// Assign a valid identifier if this is not already the case.
	if player.ID == "" {
		player.ID = uuid.New().String()
	}

	// Validate that the input data describe a valid player.
	if !player.valid() {
		return fmt.Errorf("Could not create player \"%s\", some properties are invalid", player.Name)
	}

	// Marshal the input player to pass it to the import script.
	data, err := json.Marshal(player)
	if err != nil {
		return fmt.Errorf("Could not import player \"%s\" for \"%s\" (err: %v)", player.Name, player.AccountID, err)
	}
	jsonToSend := string(data)

	fmt.Println(fmt.Sprintf("Trying to insert \"%v\"", *player))

	query := fmt.Sprintf("select * from create_player('%s')", jsonToSend)
	_, err = p.dbase.DBExecute(query)

	// Check for errors. We will refine this process a bit to try
	// to detect cases where the user tries to create a player and
	// there's already an entry for this account in the same uni.
	// In this case we should get an error indicating a `23505` as
	// return code. We will refine the error in this case.
	if err != nil {
		// Check for duplicated key error.
		msg := fmt.Sprintf("%v", err)

		if strings.Contains(msg, getDuplicatedElementErrorKey()) {
			return fmt.Errorf("Could not import player \"%s\", account \"%s\" already exists in universe \"%s\" (err: %s)", player.Name, player.AccountID, player.UniverseID, msg)
		}

		// Check for foreign key violation error.
		if strings.Contains(msg, getForeignKeyViolationErrorKey()) {
			return fmt.Errorf("Could not import player \"%s\", account \"%s\" or universe \"%s\" does not exist (err: %s)", player.Name, player.AccountID, player.UniverseID, msg)
		}

		return fmt.Errorf("Could not import player \"%s\" for \"%s\" (err: %s)", player.Name, player.AccountID, msg)
	}

	// All is well.
	return nil
}
