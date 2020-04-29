package data

import (
	"fmt"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"

	"github.com/google/uuid"
)

// ActionProxy :
// Provides a way to handle the upgrade actions for the
// server. Upgrade actions can be attached to buildings
// technologies, ships or defenses. They are also linked
// to a planet where the action is actually performed.
// This proxy intends at providing the necessary checks
// to make sure that a request to register a new action
// is possible given the resources and infrastructure
// currently existing on the planet.
type ActionProxy struct {
	commonProxy
}

// ErrInvalidAction :
// Used to indicate that the action provided in input
// could not be analyzed in some way: typically the
// fetching of the planet attached to this action has
// failed, etc.
var ErrInvalidAction = fmt.Errorf("Unable to analyze invalid action")

// ErrImpossibleAction :
// Used to indicate that the action provided in input
// cannot be performed in the specified planet. This
// might mean that there are not enough resources, or
// an inconsistent tech tree etc.
var ErrImpossibleAction = fmt.Errorf("Cannot perform action on planet")

// NewActionProxy :
// Create a new proxy allowing to serve the requests
// related to actions.
//
// The `data` defines the data model to use to fetch
// information and verify requests.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewActionProxy(data model.Instance, log logger.Logger) ActionProxy {
	return ActionProxy{
		commonProxy: newCommonProxy(data, log, "actions"),
	}
}

// CreateBuildingAction :
// Used to perform the creation of the building upgrade
// action described by the input data to the DB. In case
// the creation can not be performed an error is sent
// back to the client.
//
// The `a` describes the element to create in DB. It is
// the representation of the desired action to perform
// on the specified planet.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`. It also provides the identifier of the action
// that was created by this method.
func (p *ActionProxy) CreateBuildingAction(a model.BuildingAction) (string, error) {
	// Assign a valid identifier if this is not already the case.
	if a.ID == "" {
		a.ID = uuid.New().String()
	}

	// Fetch the planet related to this action and use it
	// as read write access.
	planet, err := model.NewReadWritePlanet(a.Planet, p.data)
	defer func() {
		err := planet.Close()
		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Could not release lock on planet \"%s\" (err: %v)", planet.ID, err))
		}
	}()

	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not fetch planet related to building action (err: %v)", err))
		return a.ID, ErrInvalidAction
	}

	// Consolidate the action (typically completion time
	// and effects).
	err = a.ConsolidateEffects(p.data, &planet)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not consolidate building action effects (err: %v)", err))
		return a.ID, ErrInvalidAction
	}

	err = a.ConsolidateCompletionTime(p.data, &planet)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not consolidate completion time for building action on \"%s\" (err: %v)", planet.ID, err))
		return a.ID, ErrInvalidAction
	}

	// Validate the action's data against its parent planet
	err = a.Validate(p.data, &planet)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Cannot perform building action for \"%s\" on \"%s\" (err: %v)", a.Element, planet.ID, err))
		return a.ID, ErrImpossibleAction
	}

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_building_upgrade_action",
		Args: []interface{}{
			a,
			a.Production,
			a.Storage,
		},
	}

	err = p.proxy.InsertToDB(query)

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not create building action on \"%s\" (err: %v)", planet.ID, err))
		return a.ID, err
	}

	p.trace(logger.Notice, fmt.Sprintf("Registered action to upgrade \"%s\" to level %d on \"%s\"", a.Element, a.DesiredLevel, a.Planet))

	// All is well.
	return a.ID, nil
}

// CreateTechnologyAction :
// Used to perform the creation of the technology upgrade
// action described by the input data to the DB. In case
// the creation can not be performed an error is returned
// to the client.
//
// The `a` describes the element to create in DB. It is
// the representation of the desired action to perform
// on the specified planet.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`. It also indicates the identifier of the action
// that was created.
func (p *ActionProxy) CreateTechnologyAction(a model.TechnologyAction) (string, error) {
	// Assign a valid identifier if this is not already the case.
	if a.ID == "" {
		a.ID = uuid.New().String()
	}

	// Fetch the planet related to this action and use it
	// as read write access.
	planet, err := model.NewReadWritePlanet(a.Planet, p.data)
	defer func() {
		err := planet.Close()
		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Could not release lock on planet \"%s\" (err: %v)", planet.ID, err))
		}
	}()

	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not fetch planet related to technology action (err: %v)", err))
		return a.ID, ErrInvalidAction
	}

	// Consolidate the completion time for this technology.
	err = a.ConsolidateCompletionTime(p.data, &planet)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not consolidate completion time for technology action on \"%s\" (err: %v)", planet.ID, err))
		return a.ID, ErrInvalidAction
	}

	// Validate the action's data against its parent planet
	err = a.Validate(p.data, &planet)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Cannot perform technology action for \"%s\" on \"%s\" (err: %v)", a.Element, planet.ID, err))
		return a.ID, ErrImpossibleAction
	}

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_technology_upgrade_action",
		Args:   []interface{}{a},
	}

	err = p.proxy.InsertToDB(query)

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not create technology action on \"%s\" (err: %v)", planet.ID, err))
		return a.ID, err
	}

	p.trace(logger.Notice, fmt.Sprintf("Registered action to upgrade \"%s\" to level %d on \"%s\"", a.Element, a.DesiredLevel, a.Planet))

	return a.ID, err
}

// CreateShipAction :
// Used to perform the creation of the ship upgrade action
// described by the input data to the DB. In case the
// creation can not be performed an error is returned to
// the client.
//
// The `a` describes the element to create in DB. It is
// the representation of the desired action to perform
// on the specified planet.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`. It also indicates the identifier of the action
// that was created.
func (p *ActionProxy) CreateShipAction(a model.ShipAction) (string, error) {
	// Assign a valid identifier if this is not already the case.
	if a.ID == "" {
		a.ID = uuid.New().String()
	}

	// Fetch the planet related to this action and use it
	// as read write access.
	planet, err := model.NewReadWritePlanet(a.Planet, p.data)
	defer func() {
		err := planet.Close()
		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Could not release lock on planet \"%s\" (err: %v)", planet.ID, err))
		}
	}()

	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not fetch planet related to ship action (err: %v)", err))
		return a.ID, ErrInvalidAction
	}

	// Consolidate the completion time for this action and
	// the amount of units to produce.
	a.Remaining = a.Amount

	err = a.ConsolidateCompletionTime(p.data, &planet)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not consolidate completion time for ship action on \"%s\" (err: %v)", planet.ID, err))
		return a.ID, ErrInvalidAction
	}

	// Validate the action's data against its parent planet
	err = a.Validate(p.data, &planet)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Cannot perform ship action for \"%s\" on \"%s\" (err: %v)", a.Element, planet.ID, err))
		return a.ID, ErrImpossibleAction
	}

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_ship_upgrade_action",
		Args:   []interface{}{a},
	}

	err = p.proxy.InsertToDB(query)

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not create ship action on \"%s\" (err: %v)", planet.ID, err))
		return a.ID, err
	}

	p.trace(logger.Notice, fmt.Sprintf("Registered action to create %d \"%s\" on \"%s\"", a.Remaining, a.Element, a.Planet))

	return a.ID, err
}

// CreateDefenseAction :
// Used to perform the creation of the defense upgrade
// action described by the input data to the DB. In case
// the creation can not be performed an error is returned
// to the client.
//
// The `a` describes the element to create in DB. It is
// the representation of the desired action to perform
// on the specified planet.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`. It also indicates the identifier of the action
// that was created.
func (p *ActionProxy) CreateDefenseAction(a model.DefenseAction) (string, error) {
	// Assign a valid identifier if this is not already the case.
	if a.ID == "" {
		a.ID = uuid.New().String()
	}

	// Fetch the planet related to this action and use it
	// as read write access.
	planet, err := model.NewReadWritePlanet(a.Planet, p.data)
	defer func() {
		err := planet.Close()
		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Could not release lock on planet \"%s\" (err: %v)", planet.ID, err))
		}
	}()

	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not fetch planet related to defense action (err: %v)", err))
		return a.ID, ErrInvalidAction
	}

	// Consolidate the completion time for this action and
	// the amount of units to produce.
	a.Remaining = a.Amount

	err = a.ConsolidateCompletionTime(p.data, &planet)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not consolidate completion time for defense action on \"%s\" (err: %v)", planet.ID, err))
		return a.ID, ErrInvalidAction
	}

	// Validate the action's data against its parent planet
	err = a.Validate(p.data, &planet)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Cannot perform defense action for \"%s\" on \"%s\" (err: %v)", a.Element, planet.ID, err))
		return a.ID, ErrImpossibleAction
	}

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_defense_upgrade_action",
		Args:   []interface{}{a},
	}

	err = p.proxy.InsertToDB(query)

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not create defense action on \"%s\" (err: %v)", planet.ID, err))
		return a.ID, err
	}

	p.trace(logger.Notice, fmt.Sprintf("Registered action to create %d \"%s\" on \"%s\"", a.Remaining, a.Element, a.Planet))

	return a.ID, err
}
