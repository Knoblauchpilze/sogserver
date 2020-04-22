package data

import (
	"fmt"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
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
// The `dbase` represents the database to use to fetch
// data related to players.
//
// The `data` defines the data model to use to fetch
// information and verify actions.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewPlayerProxy(dbase *db.DB, data model.Instance, log logger.Logger) PlayerProxy {
	return PlayerProxy{
		commonProxy: newCommonProxy(dbase, data, log, "players"),
	}
}

// Players :
// Return a list of players registered so far in all the
// players defined in the DB. The input filters might help
// to narrow the search a bit by providing an account to
// look for and a uni to look into.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// players available. Each one is appended `as-is` to the
// query.
//
// Returns the list of players registered in the DB and
// matching the input list of filters. In case the error
// is not `nil` the value of the array should be ignored.
func (p *PlayerProxy) Players(filters []db.Filter) ([]model.Player, error) {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
			"uni",
			"account",
			"name",
		},
		Table:   "players",
		Filters: filters,
	}

	res, err := p.proxy.FetchFromDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not query DB to fetch players (err: %v)", err))
		return []model.Player{}, err
	}

	// We now need to retrieve all the identifiers that matched
	// the input filters and then build the corresponding players
	// object for each one of them.
	var ID string
	IDs := make([]string, 0)

	for res.Next() {
		err = res.Scan(&ID)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Error while fetching player ID (err: %v)", err))
			continue
		}

		IDs = append(IDs, ID)
	}

	players := make([]model.Player, 0)

	for _, ID = range IDs {
		pla, err := model.NewPlayerFromDB(ID, p.data)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Unable to fetch player \"%s\" data from DB (err: %v)", ID, err))
			continue
		}

		players = append(players, pla)
	}

	return players, nil
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
func (p *PlayerProxy) Create(player model.Player) error {
	// Check consistency.
	if player.Valid() {
		return model.ErrInvalidPlayer
	}

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_player",
		Args:   []interface{}{player},
	}

	err := p.proxy.InsertToDB(query)

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not import player in \"%s\" for \"%s\" (err: %v)", player.Universe, player.Account, err))
		return err
	}

	p.trace(logger.Notice, fmt.Sprintf("Created new player \"%s\" with id \"%s\"", player.Name, player.ID))

	return nil
}
