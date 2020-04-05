package data

import (
	"encoding/json"
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"

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
		p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to upgrade \"%s\" to level %d on \"%s\"", action.BuildingID, action.Level, action.PlanetID))
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
		p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to upgrade \"%s\" to level %d for \"%s\"", action.TechnologyID, action.Level, action.PlayerID))
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

	err := p.createAction(action, "create_defense_upgrade_action")

	if err == nil {
		p.log.Trace(logger.Notice, fmt.Sprintf("Registered action to build \"%s\" on \"%s\"", action.DefenseID, action.PlanetID))
	}

	return err
}
