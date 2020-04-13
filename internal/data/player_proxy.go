package data

import (
	"encoding/json"
	"fmt"
	"oglike_server/internal/locker"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"

	"github.com/google/uuid"
)

// PlayerProxy :
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
//
// The `lock` allows to lock specific resources when some
// player data should be retrieved. Indeed the player's
// data include the technologies researched so far by the
// player which are potentially being upgraded through a
// `upgrade action` mechanism. In order to make sure that
// the data is up-to-date, we have to process these actions
// each time a player's data needs to be retrieved.
// This plays well with the mechanism we decided to use
// to have some kind of lazy processing of actions: the
// research of the player are not taken into account (i.e.
// registered in the DB) until it's really needed.
//
// The `techCosts` is used when the data for a player is
// needed in order to compute the costs that are linked
// to the next level of a technology based on the level
// reached so far. It is initialized when building the
// proxy so that the information is readily available
// when needed.
type PlayerProxy struct {
	dbase     *db.DB
	log       logger.Logger
	lock      *locker.ConcurrentLocker
	techCosts map[string]ConstructionCost
}

// initTechCostsFromDB :
// Used to query information from the DB and fetch info
// for technologies that will be used to compute costs
// of a technology for a given level. The information
// fetched in this function hardly changes during the
// life of the server and it's more interesting to load
// it in RAM so as to save time each time a request of
// a player is handled.
// In case the DB cannot be contacted an error is sent
// back but a valid map is still returned (it is empty
// though).
//
// The `dbase` represents the DB from which info about
// technologies costs should be fetched.
//
// The `log` allows to notify errors and info to the
// user in case of any failure.
//
// Returns a map representing for each technology its
// associated cost and rule of progression.
func initTechCostsFromDB(dbase *db.DB, log logger.Logger) (map[string]ConstructionCost, error) {
	techCosts := make(map[string]ConstructionCost)

	if dbase == nil {
		return techCosts, fmt.Errorf("Could not technologies costs from DB, no DB provided")
	}

	// First retrieve the technologies progression rule.
	props := []string{
		"technology",
		"progress",
	}
	table := "technologies_costs_progress"

	query := fmt.Sprintf("select %s from %s", strings.Join(props, ", "), table)

	rows, err := dbase.DBQuery(query)
	if err != nil {
		return techCosts, fmt.Errorf("Could not initialize technologies costs from DB (err: %v)", err)
	}

	var techID string
	var progress float32

	for rows.Next() {
		err = rows.Scan(
			&techID,
			&progress,
		)

		if err != nil {
			log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for technology cost (err: %v)", err))
			continue
		}

		existing, ok := techCosts[techID]
		if ok {
			log.Trace(logger.Error, fmt.Sprintf("Overriding progression rule for technology \"%s\" (existing was %f, new is %f)", techID, existing.ProgressionRule, progress))
		}

		techCosts[techID] = ConstructionCost{
			make(map[string]int),
			progress,
		}
	}

	// Now populate the costs for each technology.
	props = []string{
		"technology",
		"res",
		"cost",
	}
	table = "technologies_costs"

	query = fmt.Sprintf("select %s from %s", strings.Join(props, ", "), table)

	rows, err = dbase.DBQuery(query)
	if err != nil {
		return techCosts, fmt.Errorf("Could not initialize technologies costs from DB (err: %v)", err)
	}

	var res string
	var cost int

	for rows.Next() {
		err = rows.Scan(
			&techID,
			&res,
			&cost,
		)

		if err != nil {
			log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for technology cost (err: %v)", err))
			continue
		}

		tech, ok := techCosts[techID]

		if !ok {
			log.Trace(logger.Error, fmt.Sprintf("Cannot define cost of %d of resource \"%s\" for technology \"%s\" not found in progression rules", cost, res, techID))
			continue
		}

		existing, ok := tech.InitCosts[res]
		if ok {
			log.Trace(logger.Error, fmt.Sprintf("Overriding cost for resource \"%s\" in technology \"%s\" (existing was %d, new is %d)", res, techID, existing, cost))
		}

		tech.InitCosts[res] = cost

		techCosts[techID] = tech
	}

	return techCosts, nil
}

// NewPlayerProxy :
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
func NewPlayerProxy(dbase *db.DB, log logger.Logger) PlayerProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create players proxy from invalid DB"))
	}

	// Fetch the information related to technlogies costs
	// from the DB to populate the internal map. We will
	// use the dedicated handler which is used to actually
	// fetch the data and always return a valid value.
	techCosts, err := initTechCostsFromDB(dbase, log)
	if err != nil {
		log.Trace(logger.Error, fmt.Sprintf("Could not fetch technologies costs from DB (err: %v)", err))
	}

	return PlayerProxy{
		dbase,
		log,
		locker.NewConcurrentLocker(log),
		techCosts,
	}
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

	// Fetch technologies.
	query := fmt.Sprintf("select technology, level from player_technologies where player='%s'", player.ID)
	rows, err := p.dbase.DBQuery(query)

	if err != nil {
		return fmt.Errorf("Could not fetch technologies for player \"%s\" (err: %v)", player.ID, err)
	}

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
	// Prevent invalid player identifier.
	if player == "" {
		return fmt.Errorf("Cannot update technologies upgrade actions for invalid player")
	}

	// Acquire a lock on this player.
	lock := p.lock.Acquire(player)
	defer p.lock.Release(lock)

	// Perform the update: we will wrap the function inside
	// a dedicated handler to make sure that we don't lock
	// the resource more than necessary.
	var err error
	var errExec error

	func() {
		lock.Lock()
		defer func() {
			err = lock.Release()
		}()

		// Perform the update.
		query := fmt.Sprintf("SELECT update_technology_upgrade_action('%s')", player)
		_, errExec = p.dbase.DBExecute(query)
	}()

	// Return any error.
	if errExec != nil {
		return fmt.Errorf("Could not update technologies upgrade actions for \"%s\" (err: %v)", player, errExec)
	}
	if err != nil {
		return fmt.Errorf("Could not release locker when updading technologies upgrade actions for \"%s\" (err: %v)", player, err)
	}

	return nil
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
		costs, err := initTechCostsFromDB(p.dbase, p.log)
		if err != nil {
			return fmt.Errorf("Unable to generate technologies costs for technology \"%s\", none defined", tech.ID)
		}

		p.techCosts = costs
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

	// Marshal the input player to pass it to the import script.
	data, err := json.Marshal(player)
	if err != nil {
		return fmt.Errorf("Could not import player \"%s\" for \"%s\" (err: %v)", player.Name, player.AccountID, err)
	}
	jsonToSend := string(data)

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
