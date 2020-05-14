package data

import (
	"fmt"
	"oglike_server/internal/game"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"

	"github.com/google/uuid"
)

// PlayerProxy :
// Intended as a wrapper to access properties of players
// and retrieve data from the database. In most cases we
// need to access some properties of the players for a
// given identifier. A player is the instance of some
// account in a given universe. It is usually linked to
// a list of planets and fleets which are the way the
// player interacts in the universe.
type PlayerProxy struct {
	commonProxy
}

// NewPlayerProxy :
// Create a new proxy allowing to serve the requests
// related to players.
//
// The `data` defines the data model to use to fetch
// information and verify requests.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewPlayerProxy(data game.Instance, log logger.Logger) PlayerProxy {
	return PlayerProxy{
		commonProxy: newCommonProxy(data, log, "players"),
	}
}

// Players :
// Return a list of players registered so far in all the
// players defined in the DB. The input filters might help
// to narrow the search a bit by providing an account to
// look for and a uni to look into.
//
// The `filters` define some filtering properties that can
// be applied to the SQL query to only select part of all
// the players available. Each one is appended `as-is` to
// the query.
//
// Returns the list of players registered in the DB and
// matching the input list of filters. In case the error
// is not `nil` the value of the array should be ignored.
func (p *PlayerProxy) Players(filters []db.Filter) ([]game.Player, error) {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table:   "players",
		Filters: filters,
	}

	dbRes, err := p.data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not query DB to fetch players (err: %v)", err))
		return []game.Player{}, err
	}
	if dbRes.Err != nil {
		p.trace(logger.Error, fmt.Sprintf("Invalid query to fetch players (err: %v)", dbRes.Err))
		return []game.Player{}, dbRes.Err
	}

	// We now need to retrieve all the identifiers that matched
	// the input filters and then build the corresponding players
	// object for each one of them.
	var ID string
	IDs := make([]string, 0)

	for dbRes.Next() {
		err = dbRes.Scan(&ID)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Error while fetching player ID (err: %v)", err))
			continue
		}

		IDs = append(IDs, ID)
	}

	players := make([]game.Player, 0)

	for _, ID = range IDs {
		pla, err := game.NewPlayerFromDB(ID, p.data)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Unable to fetch player \"%s\" data from DB (err: %v)", ID, err))
			continue
		}

		players = append(players, pla)
	}

	if len(players) == 0 {
		return players, game.ErrElementNotFound
	}

	return players, nil
}

// Messages :
// Return a list of the messages for a given player and
// matching the input filters. The messages are returned
// along with their types and arguments.
//
// The `filters` define some filtering properties to be
// applied when querying the messages.
//
// Returns the list of messages for this player as the
// DB defines it.
func (p *PlayerProxy) Messages(filters []db.Filter) ([]game.Message, error) {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"mp.id",
		},
		Table:   "messages_players mp inner join messages_ids mi on mp.message = mi.id",
		Filters: filters,
	}

	dbRes, err := p.data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not query DB to fetch messages (err: %v)", err))
		return []game.Message{}, err
	}
	if dbRes.Err != nil {
		p.trace(logger.Error, fmt.Sprintf("Invalid query to fetch messages (err: %v)", dbRes.Err))
		return []game.Message{}, dbRes.Err
	}

	// We now need to retrieve all the identifiers that matched
	// the input filters and then build the corresponding players
	// object for each one of them.
	var ID string
	IDs := make([]string, 0)

	for dbRes.Next() {
		err = dbRes.Scan(&ID)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Error while fetching message ID (err: %v)", err))
			continue
		}

		IDs = append(IDs, ID)
	}

	messages := make([]game.Message, 0)

	for _, ID = range IDs {
		msg, err := game.NewMessageFromDB(ID, p.data)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Unable to fetch message \"%s\" data from DB (err: %v)", ID, err))
			continue
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// Create :
// Used to perform the creation of the player described
// by the input structure in the DB. The player is both
// associated to an account and a universe. The database
// guarantees that no two players can exist in the same
// universe belonging to the same account.
//
// The `player` describes the element to create in `DB`.
// This value may be modified by the function in case it
// does not define a valid identifier.
//
// The return status indicates whether the player could
// be created or not (in which case an error describes
// the failure reason). Also returns the identifier of
// the player that was created.
func (p *PlayerProxy) Create(player game.Player) (string, error) {
	// Assign a valid identifier if this is not already the case.
	if player.ID == "" {
		player.ID = uuid.New().String()
	}

	err := player.SaveToDB(p.data.Proxy)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not create player \"%s\" (err: %v)", player.Name, err))
		return player.ID, err
	}

	p.trace(logger.Notice, fmt.Sprintf("Created new player \"%s\" with id \"%s\"", player.Name, player.ID))

	return player.ID, nil
}
