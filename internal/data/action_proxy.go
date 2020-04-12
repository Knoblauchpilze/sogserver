package data

import (
	"encoding/json"
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"

	"github.com/google/uuid"
)

// ActionProxy :
// Intended as a wrapper to access and register new props
// for actions that aim at upgrading buildings, technos
// or building new ships or defenses on planets.
// The goal of this proxy is to hide the real layout of
// the DB to the exterior world and provide a sanitized
// environment to perform these interactions.
//
// The `dbase` is the database that is wrapped by this
// object. It is checked for consistency upon building the
// wrapper.
//
// The `log` allows to perform display to the user so as
// to inform of potential issues and debug information to
// the outside world.
type ActionProxy struct {
	dbase *db.DB
	log   logger.Logger
}

// NewActionProxy :
// Create a new proxy on the input `dbase` to access the
// properties of upgrade actions registered in the DB. It
// includes all kind of actions (for buildings, ships,
// defenses or technologies).
// In case the provided DB is `nil` a panic is issued.
//
// The `dbase` represents the database to use to fetch
// data related to upgrade actions.
//
// The `log` will be used to notify information so that
// we can have an idea of the activity of this component.
// One possible example is for timing the requests.
//
// Returns the created proxy.
func NewActionProxy(dbase *db.DB, log logger.Logger) ActionProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create actions proxy from invalid DB"))
	}

	return ActionProxy{dbase, log}
}

// Buildings :
// Allows to fetch the list of upgrade action currently
// registered in the DB given the filters parameters. It
// can be used to get an idea of the actions pending for
// a planet regarding the buildings.
// The user can choose to filter parts of the buildings
// actions using an array of filters that will be applied
// to the SQL query.
// No controls is enforced on the filters so one should
// make sure that it's consistent with the underlying
// table.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// upgrade actions available. Each one is appended `as-is`
// to the query.
//
// Returns the list of building upgrade actions along with
// any errors. Note that in case the error is not `nil` the
// returned list is to be ignored.
func (p *ActionProxy) Buildings(filters []DBFilter) ([]BuildingUpgradeAction, error) {
	// Create the query and execute it.
	props := []string{
		"id",
		"planet",
		"building",
		"current_level",
		"desired_level",
		"completion_time",
	}

	table := "construction_actions_buildings"

	query := fmt.Sprintf("select %s from %s", strings.Join(props, ", "), table)

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
		return nil, fmt.Errorf("Could not query DB to fetch buildings upgrade actions (err: %v)", err)
	}

	// Populate the return value.
	actions := make([]BuildingUpgradeAction, 0)
	var act BuildingUpgradeAction

	for rows.Next() {
		err = rows.Scan(
			&act.ID,
			&act.PlanetID,
			&act.BuildingID,
			&act.CurrentLevel,
			&act.DesiredLevel,
			&act.CompletionTime,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for building upgrade action (err: %v)", err))
			continue
		}

		actions = append(actions, act)
	}

	return actions, nil
}

// Technologies :
// Similar to the `Buildings` feature but can be used to get
// the list of technology upgrade actions. Instead of fetching
// by planet (which would not make much sense) the result is
// fetched at a player level.
// The input filters describe some additional criteria that
// should be matched by the upgrade actions.
//
// The `filters` define some filtering properties to apply
// to the selected upgrade actions. Each one is used directly
// agains the columns of the related table.
//
// Returns the list of technology upgrade actions matching
// the input filters. This list should be ignored if the
// error is not `nil`.
func (p *ActionProxy) Technologies(filters []DBFilter) ([]TechnologyUpgradeAction, error) {
	// Create the query and execute it.
	props := []string{
		"id",
		"player",
		"technology",
		"planet",
		"current_level",
		"desired_level",
		"completion_time",
	}

	table := "construction_actions_technologies"

	query := fmt.Sprintf("select %s from %s", strings.Join(props, ", "), table)

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
		return nil, fmt.Errorf("Could not query DB to fetch technologies upgrade actions (err: %v)", err)
	}

	// Populate the return value.
	actions := make([]TechnologyUpgradeAction, 0)
	var act TechnologyUpgradeAction

	for rows.Next() {
		err = rows.Scan(
			&act.ID,
			&act.PlayerID,
			&act.TechnologyID,
			&act.PlanetID,
			&act.CurrentLevel,
			&act.DesiredLevel,
			&act.CompletionTime,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for technology upgrade action (err: %v)", err))
			continue
		}

		actions = append(actions, act)
	}

	return actions, nil
}

// Ships :
// Similar to the `Buildings` feature but can be used to get
// the list of ships construction actions. The input filters
// describe some additional criteria that should be matched
// by the construction actions (typically actions related to
// a specific ship, etc.).
//
// The `filters` define some filtering properties to apply
// to the selected upgrade actions. Each one is used directly
// agains the columns of the related table.
//
// Returns the list of ships being built that match the input
// filters. This list should be ignored if the error is not
// `nil`.
func (p *ActionProxy) Ships(filters []DBFilter) ([]ShipUpgradeAction, error) {
	// Create the query and execute it.
	props := []string{
		"id",
		"planet",
		"ship",
		"amount",
		"remaining",
		"completion_time",
	}

	table := "construction_actions_ships"

	query := fmt.Sprintf("select %s from %s", strings.Join(props, ", "), table)

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
		return nil, fmt.Errorf("Could not query DB to fetch ships construction actions (err: %v)", err)
	}

	// Populate the return value.
	actions := make([]ShipUpgradeAction, 0)
	var act ShipUpgradeAction

	for rows.Next() {
		err = rows.Scan(
			&act.ID,
			&act.PlanetID,
			&act.ShipID,
			&act.Amount,
			&act.Remaining,
			&act.CompletionTime,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for ship construction action (err: %v)", err))
			continue
		}

		actions = append(actions, act)
	}

	return actions, nil
}

// Defenses :
// Similar to the `Buildings` feature but can be used to get
// the list of defenses construction actions. The provided
// filters describe some additional criteria that should be
// matched by the construction actions (typically actions
// related to a specific defense system, etc.).
//
// The `filters` define some filtering properties to apply
// to the selected upgrade actions. Each one is used directly
// agains the columns of the related table.
//
// Returns the list of defenses being built that match the
// input filters. This list should be ignored if the error
// is not `nil`.
func (p *ActionProxy) Defenses(filters []DBFilter) ([]DefenseUpgradeAction, error) {
	// Create the query and execute it.
	props := []string{
		"id",
		"planet",
		"defense",
		"amount",
		"remaining",
		"completion_time",
	}

	table := "construction_actions_defenses"

	query := fmt.Sprintf("select %s from %s", strings.Join(props, ", "), table)

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
		return nil, fmt.Errorf("Could not query DB to fetch defenses construction actions (err: %v)", err)
	}

	// Populate the return value.
	actions := make([]DefenseUpgradeAction, 0)
	var act DefenseUpgradeAction

	for rows.Next() {
		err = rows.Scan(
			&act.ID,
			&act.PlanetID,
			&act.DefenseID,
			&act.Amount,
			&act.Remaining,
			&act.CompletionTime,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for defense construction action (err: %v)", err))
			continue
		}

		actions = append(actions, act)
	}

	return actions, nil
}

// createAction :
// Used to mutualize the creation of upgrade action by
// considering that what differs between an action that
// aims at upgrading a building, a technology, a ship
// or a defense is mostly the action itself and the
// script that will be used to perform the insertion.
//
// The `action` defines the upgrade action itself that
// should be inserted in the DB. It will be marshalled
// and passed on to the insertion script.
//
// The `script` represents the name of the script to
// use to perform the insertion of the input upgrade
// action into the DB.
//
// Returns any error that occurred during the insertion
// of the upgrade action in the DB.
func (p *ActionProxy) createAction(action UpgradeAction, script string) error {
	// Check whether the action is valid.
	if !action.valid() {
		return fmt.Errorf("Could not create upgrade action, some properties are invalid")
	}

	// Marshal the input account to pass it to the import script.
	data, err := json.Marshal(action)
	if err != nil {
		return fmt.Errorf("Could not import upgrade action for %s (err: %v)", action, err)
	}
	jsonToSend := string(data)

	query := fmt.Sprintf("select * from %s('%s')", script, jsonToSend)
	_, err = p.dbase.DBExecute(query)

	// Check for errors during the insertion process.
	if err != nil {
		return fmt.Errorf("Could not import upgrade action %s (err: %s)", action, err)
	}

	// All is well.
	return nil
}

// CreateBuildingAction :
// Used to perform the creation of the building upgrade
// action described by the input data to the DB. In case
// the creation can not be performed an error is sent
// back to the client.
//
// The `action` describes the element to create in DB.
// It corresponds to the desired upgrade action requested
// by the user.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`.
func (p *ActionProxy) CreateBuildingAction(action *BuildingUpgradeAction) error {
	// Assign a valid identifier if this is not already the case.
	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	err := p.createAction(action, "create_building_upgrade_action")

	if err == nil {
		p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to upgrade \"%s\" to level %d on \"%s\"", action.BuildingID, action.DesiredLevel, action.PlanetID))
	}

	return err
}

// CreateTechnologyAction :
// Used to perform the creation of the technology upgrade
// action described by the input data to the DB. In case
// the creation can not be performed an error is returned
// to the client.
//
// The `action` describes the element to create in DB.
// It corresponds to the desired upgrade action requested
// by the user.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`.
func (p *ActionProxy) CreateTechnologyAction(action *TechnologyUpgradeAction) error {
	// Assign a valid identifier if this is not already the case.
	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	err := p.createAction(action, "create_technology_upgrade_action")

	if err == nil {
		p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to upgrade \"%s\" to level %d for \"%s\"", action.TechnologyID, action.DesiredLevel, action.PlayerID))
	}

	return err
}

// CreateShipAction :
// Used to perform the creation of the ship upgrade action
// described by the input data to the DB. In case the
// creation can not be performed an error is returned to
// the client.
//
// The `action` describes the element to create in DB.
// It corresponds to the desired upgrade action requested
// by the user.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`.
func (p *ActionProxy) CreateShipAction(action *ShipUpgradeAction) error {
	// Assign a valid identifier if this is not already the case.
	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	// Make sure that the `remaining` count is identical to
	// the initial amount.
	action.Remaining = action.Amount

	err := p.createAction(action, "create_ship_upgrade_action")

	if err == nil {
		p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to build \"%s\" on \"%s\"", action.ShipID, action.PlanetID))
	}

	return err
}

// CreateDefenseAction :
// Used to perform the creation of the defense upgrade
// action described by the input data to the DB. In case
// the creation can not be performed an error is returned
// to the client.
//
// The `action` describes the element to create in DB.
// It corresponds to the desired upgrade action requested
// by the user.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`.
func (p *ActionProxy) CreateDefenseAction(action *DefenseUpgradeAction) error {
	// Assign a valid identifier if this is not already the case.
	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	// Make sure that the `remaining` count is identical to
	// the initial amount.
	action.Remaining = action.Amount

	err := p.createAction(action, "create_defense_upgrade_action")

	if err == nil {
		p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to build \"%s\" on \"%s\"", action.DefenseID, action.PlanetID))
	}

	return err
}
