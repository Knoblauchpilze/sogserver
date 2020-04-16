package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"

	"github.com/google/uuid"
)

// PlayerProxy :
// Intended as a wrapper to access properties of players and
// retrieve data from the database. Internally uses the common
// proxy defined in this package.
// The additional information used by this proxy is the fact
// that the technologies associated to a player should be
// assigned their costs to research the next level. In order
// to do that we will fetch the information from the tables
// in the DB and then use this whenever a request is issued.
// This has the advantage of only necessiting the query of
// the DB once: it is made possible by the fact that these
// info almost never change (and thus it makes sense to cache
// it).
//
// The `techCosts` are fetched from the DB and represent a
// way to compute the costs in resources for each technology
// for any level.
type PlayerProxy struct {
	techCosts map[string]ConstructionCost

	commonProxy
}

// NewPlayerProxy :
// Create a new proxy allowing to serve the requests
// related to players.
//
// The `dbase` represents the database to use to fetch
// data related to players.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewPlayerProxy(dbase *db.DB, log logger.Logger) PlayerProxy {
	proxy := PlayerProxy{
		make(map[string]ConstructionCost),

		newCommonProxy(dbase, log),
	}

	err := proxy.init()
	if err != nil {
		log.Trace(logger.Error, fmt.Sprintf("Could not fetch technologies costs from DB (err: %v)", err))
	}

	return proxy
}

// init :
// Used to perform the initialziation of the needed
// DB variables for this proxy. This typically means
// fetching technologies costs from the DB.
//
// Returns `nil` if the technologies could be fetched
// from the DB successfully.
func (p PlayerProxy) init() error {
	var err error

	// Fetch from DB.
	p.techCosts, err = initProgressCostsFromDB(
		p.dbase,
		p.log,
		"technology",
		"technologies_costs_progress",
		"technologies_costs",
	)

	return err
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
func (p *PlayerProxy) Players(filters []DBFilter) ([]Player, error) {
	// Create the query and execute it.
	query := queryDesc{
		props: []string{
			"id",
			"uni",
			"account",
			"name",
		},
		table:   "players",
		filters: filters,
	}

	// Create the query and execute it.
	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch players (err: %v)", err)
	}

	// Populate the return value.
	players := make([]Player, 0)
	var player Player

	for res.next() {
		err = res.scan(
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
		err = p.fetchPlayerTechnologies(&player)
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch data for player \"%s\" (err: %v)", player.ID, err))
			continue
		}

		players = append(players, player)
	}

	return players, nil
}

// fetchPlayerTechnologies :
// Used to fetch data related to the player in argument. It
// mainly consists in the list of technologies researched by
// the player.
//
// The `player` references the player for which data should
// be fetched. We assume that the internal fields (and more
// specifically the identifier) are already populated.
//
// Returns any error.
func (p *PlayerProxy) fetchPlayerTechnologies(player *Player) error {
	// Check whether the player has an identifier assigned.
	if player.ID == "" {
		return fmt.Errorf("Unable to fetch data from player with invalid identifier")
	}

	player.Technologies = make([]Technology, 0)

	// We need to update the technology upgrade actions that
	// might be registered for this player first.
	err := p.updateTechnologyUpgradeActions(player.ID)
	if err != nil {
		return fmt.Errorf("Could not update technology upgrade actions for player \"%s\" (err: %v)", player.ID, err)
	}

	// Create the query and execute it.
	query := queryDesc{
		props: []string{
			"technology, level",
			"level",
		},
		table: "player_technologies",
		filters: []DBFilter{
			DBFilter{
				Key:    "player",
				Values: []string{player.ID},
			},
		},
	}

	// Create the query and execute it.
	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return fmt.Errorf("Could not query DB to fetch technologies for player \"%s\" (err: %v)", player.ID, err)
	}

	// Populate the return value.
	var tech Technology

	for res.next() {
		err = res.scan(
			&tech.ID,
			&tech.Level,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for universe (err: %v)", err))
			continue
		}

		// Update the costs for this technology.
		err = p.updateTechnologyCosts(&tech)
		if err != nil {
			tech.Cost = make([]ResourceAmount, 0)

			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve costs for technology \"%s\" for player \"%s\" (err: %v)", tech.ID, player.ID, err))
		}

		player.Technologies = append(player.Technologies, tech)
	}

	return nil
}

// updateTechnologyUpgradeActions :
// Used to perform the update of the technology upgrade
// actions registered for the input player. This means
// parsing the actions registered for this player and
// see whether the corresponding actions can be applied
// to the DB. Basically it will check for each research
// action (should be at most 1 for any player) whether
// the completion time is in the past and if so update
// the corresponding current technology level in the
// corresponding table.
// The update action will also be deleted.
//
// The `player` defines the identifier of the player
// for which the technology upgrade action should be
// updated.
//
// Returns any error that may have occurred during the
// process.
func (p *PlayerProxy) updateTechnologyUpgradeActions(player string) error {
	query := fmt.Sprintf("SELECT update_technology_upgrade_action('%s')", player)

	return p.performWithLock(player, query)
}

// updateTechnologyCosts :
// Used to perform the computation of the costs for the
// next level of the technology described in argument.
// The output values will be saved directly in the input
// object.
//
// The `tech` defines the object for which the costs
// should be computed. A `nil` value will generate an
// error.
//
// Returns any error.
func (p *PlayerProxy) updateTechnologyCosts(tech *Technology) error {
	// Check consistency.
	if tech == nil || tech.ID == "" {
		return fmt.Errorf("Cannot update technology costs from invalid technology")
	}

	// In case the costs for technology are not populated
	// try to update it.
	if len(p.techCosts) == 0 {
		err := p.init()
		if err != nil {
			return fmt.Errorf("Unable to generate technologies costs for technology \"%s\", none defined", tech.ID)
		}
	}

	// Find the technology in the costs table.
	info, ok := p.techCosts[tech.ID]
	if !ok {
		return fmt.Errorf("Could not compute costs for unknown technology \"%s\"", tech.ID)
	}

	// Compute the cost for each resource.
	tech.Cost = info.ComputeCosts(tech.Level)

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
func (p *PlayerProxy) Create(player *Player) error {
	// Assign a valid identifier if this is not already the case.
	if player.ID == "" {
		player.ID = uuid.New().String()
	}

	// Validate that the input data describe a valid player.
	if !player.valid() {
		return fmt.Errorf("Could not create player \"%s\", some properties are invalid", player.Name)
	}

	// Create the query and execute it.
	query := insertReq{
		script: "create_player",
		args:   []interface{}{*player},
	}

	err := p.insertToDB(query)

	// Check for errors.
	if err != nil {
		// Check for duplicated key error.
		msg := fmt.Sprintf("%v", err)

		// TODO: This could maybe factorized in some sort of `personalizeErrorMessage` methods
		// that would be added to the `db_error_utils`.
		if strings.Contains(msg, getDuplicatedElementErrorKey()) {
			return fmt.Errorf("Could not import player \"%s\", account \"%s\" already exists in universe \"%s\" (err: %s)", player.Name, player.AccountID, player.UniverseID, msg)
		}

		// Check for foreign key violation error.
		if strings.Contains(msg, getForeignKeyViolationErrorKey()) {
			return fmt.Errorf("Could not import player \"%s\", account \"%s\" or universe \"%s\" does not exist (err: %s)", player.Name, player.AccountID, player.UniverseID, msg)
		}

		return fmt.Errorf("Could not import player \"%s\" for \"%s\" (err: %s)", player.Name, player.AccountID, msg)
	}

	return nil
}
