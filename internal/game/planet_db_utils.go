package game

import (
	"encoding/json"
	"fmt"
	"oglike_server/pkg/db"
	"time"
)

// ErrActionStillInProgress : Indicates that an action is still running on a planet
// hence preventing its deletion.
var ErrActionStillInProgress = fmt.Errorf("Unable to delete planet due to an action still being in progress")

// ErrFleetNotYetReturned : Indicates that a fleet did not yet returned to the
// planet hence preventing its deletion.
var ErrFleetNotYetReturned = fmt.Errorf("Unable to delete planet due to fleet not yet returned")

// ErrFleetNotYetArrived : Indicates that a fleet did not yet arrived to the planet
// hence preventing its deletion.
var ErrFleetNotYetArrived = fmt.Errorf("Unable to delete planet due to fleet not yet arrived")

// ErrCannotDeleteMoon : Indicates that it is not possible to voluntarily delete a moon.
var ErrCannotDeleteMoon = fmt.Errorf("Cannot voluntarily delete a moon")

// ErrHomeworldCannotBeDeleted : Indicates that a deletion of the homeworld is not possible.
var ErrHomeworldCannotBeDeleted = fmt.Errorf("Cannot delete homeworld for player")

// SaveToDB :
// Used to save the content of this planet to
// the DB. In case an error is raised during
// the operation a comprehensive error is
// returned.
//
// The `proxy` allows to access to the DB.
//
// Returns any error.
func (p *Planet) SaveToDB(proxy db.Proxy) error {
	// Create the query and execute it: we will
	// use the dedicated handler to provide a
	// comprehensive error.
	query := db.InsertReq{
		Script: "create_planet",
		Args: []interface{}{
			p,
			p.Resources,
			time.Now(),
		},
	}

	err := proxy.InsertToDB(query)

	return p.analyzeDBError(err)
}

// UpdateInDB :
// used to update the content of the planet in
// the DB. Only part of the planet's data can
// be updated as specified by this function.
//
// The `proxy` allows to access to the DB.
//
// Returns any error.
func (p *Planet) UpdateInDB(proxy db.Proxy) error {
	// Make sure that the name of the planet is
	// valid.
	if p.Name == "" {
		return ErrInvalidUpdateData
	}

	// Check whether this planet is a real planet
	// or a moon: this will change the script to
	// use to perform the update.
	script := "update_planet"
	if p.Moon {
		script = "update_moon"
	}

	// Create the query and execute it. In a
	// similar way we need to provide some
	// analysis of any error.
	query := db.InsertReq{
		Script: script,
		Args: []interface{}{
			p.ID,
			struct {
				Name string `json:"name"`
			}{
				Name: p.Name,
			},
		},
	}

	err := proxy.InsertToDB(query)

	return p.analyzeDBError(err)
}

// UpdateProduction :
// used to update the production of resources on a planet in
// the DB.
//
// The `production` defines the list of buildings whose factor
// of production should be updated.
//
// The `proxy` allows to access to the DB.
//
// Returns any error.
func (p *Planet) UpdateProduction(production []BuildingInfo, proxy db.Proxy) error {
	// Make sure that the identifier of the planet is valid.
	if p.ID == "" {
		return ErrInvalidUpdateData
	}

	// Create the query and execute it. In a similar way we
	// need to provide some analysis of any error.
	query := db.InsertReq{
		Script: "update_planet_production",
		Args: []interface{}{
			p.ID,
			production,
		},
	}

	err := proxy.InsertToDB(query)

	return p.analyzeDBError(err)
}

// isEligibleForDeletion :
// Defines whether this planet is eligible to be
// deleted from the DB. This only include the
// *planet*'s status and does not verify that the
// parent player has more than one planet. So keep
// in mind that calling this function is *not*
// enough to check whether this planet can be
// deleted.
//
// The return value indicates whether the planet
// can be deleted. If it is `nil` it means that
// there are no obvious obstacles to the deletion
// of the planet.
func (p *Planet) isEligibleForDeletion() error {
	// Make sure that there are no upgrade actions.
	if len(p.BuildingsUpgrade) > 0 {
		return ErrActionStillInProgress
	}
	if len(p.TechnologiesUpgrade) > 0 {
		return ErrActionStillInProgress
	}
	if len(p.ShipsConstruction) > 0 {
		return ErrActionStillInProgress
	}
	if len(p.DefensesConstruction) > 0 {
		return ErrActionStillInProgress
	}

	// Make sure that there are no incoming fleets or fleets
	// that started from this planet.
	if len(p.SourceFleets) > 0 {
		return ErrFleetNotYetReturned
	}
	if len(p.IncomingFleets) > 0 {
		return ErrFleetNotYetArrived
	}

	// Finally make sure that we're not deleting a moon.
	if p.Moon {
		return ErrCannotDeleteMoon
	}

	return nil
}

// DeleteFromDB :
// Used to perform the deletion of the planet from the
// DB if the conditions are matched. It will make sure
// that nothing prevents the deletion of the planet in
// the DB before doing so.
//
// The `data` defines a way to access to the DB.
//
// Returns any error.
func (p *Planet) DeleteFromDB(data Instance) error {
	// A planet can only be deleted if it is not referenced
	// by any fleet and does not have any related action to
	// be completed.
	// Part of the checks are handled by the above method.
	if err := p.isEligibleForDeletion(); err != nil {
		return err
	}

	// We should also make sure that it's not the last planet
	// of the player associated to the planet.
	player, err := NewPlayerFromDB(p.Player, data)
	if err != nil {
		return err
	}

	if player.planetsCount == 1 {
		return ErrHomeworldCannotBeDeleted
	}

	// No obvious reasons prevent the deletion of the planet
	// we can proceed.
	query := db.InsertReq{
		Script: "delete_planet",
		Args: []interface{}{
			p.ID,
		},
	}

	err = data.Proxy.InsertToDB(query)

	return err
}

// analyzeDBError :
// used to perform the analysis of a DB error
// based on the structure of the planets' table
// to produce a comprehensive error of what
// went wrong.
//
// The `err` defines the error to analyze.
//
// Returns a comprehensive error or the input
// error if nothing can be extracted from the
// input data.
func (p *Planet) analyzeDBError(err error) error {
	// In case the error is not a `db.Error` we can't do
	// anything, so just return the input error.
	dbe, ok := err.(db.Error)
	if !ok {
		return err
	}

	// Otherwise we can try to make some sense of it.
	dee, ok := dbe.Err.(db.DuplicatedElementError)
	if ok {
		switch dee.Constraint {
		case "planets_pkey":
			return ErrDuplicatedElement
		}

		return dee
	}

	fkve, ok := dbe.Err.(db.ForeignKeyViolationError)
	if ok {
		switch fkve.ForeignKey {
		case "player":
			return ErrNonExistingPlayer
		}

		return fkve
	}

	return dbe
}

// Convert :
// Implementation of the `db.Convertible` interface
// from the DB package in order to only include fields
// that need to be marshalled in the planet's creation.
//
// Returns the converted version of the planet which
// only includes relevant fields.
func (p *Planet) Convert() interface{} {
	// We convert the this object differently based on
	// whether it is a planet or a moon.
	if p.Moon {
		return struct {
			ID       string `json:"id"`
			Planet   string `json:"planet"`
			Name     string `json:"name"`
			Fields   int    `json:"fields"`
			Diameter int    `json:"diameter"`
		}{
			ID:       p.ID,
			Planet:   p.planet,
			Name:     p.Name,
			Fields:   p.Fields,
			Diameter: p.Diameter,
		}
	}

	return struct {
		ID           string    `json:"id"`
		Player       string    `json:"player"`
		Name         string    `json:"name"`
		MinTemp      int       `json:"min_temperature"`
		MaxTemp      int       `json:"max_temperature"`
		Fields       int       `json:"fields"`
		Galaxy       int       `json:"galaxy"`
		System       int       `json:"solar_system"`
		Position     int       `json:"position"`
		Diameter     int       `json:"diameter"`
		LastActivity time.Time `json:"last_activity"`
	}{
		ID:           p.ID,
		Player:       p.Player,
		Name:         p.Name,
		MinTemp:      p.MinTemp,
		MaxTemp:      p.MaxTemp,
		Fields:       p.Fields,
		Galaxy:       p.Coordinates.Galaxy,
		System:       p.Coordinates.System,
		Position:     p.Coordinates.Position,
		Diameter:     p.Diameter,
		LastActivity: p.LastActivity,
	}
}

// MarshalJSON :
// Implementation of the `Marshaler` interface to allow
// only specific information to be marshalled when the
// planet needs to be exported. Indeed we don't want to
// export all the fields of all the elements defined as
// members of the planet.
// Most of the info will be marshalled except for deps
// on various buildings/ships/defenses built on this
// planet as it's not the place of this struct to be
// defining that.
// The approach we follow is to define a similar struct
// to the planet but do not include the tech deps.
//
// Returns the marshalled bytes for this planet along
// with any error.
func (p *Planet) MarshalJSON() ([]byte, error) {
	type lightInfo struct {
		ID               string  `json:"id"`
		Name             string  `json:"name"`
		Level            int     `json:"level"`
		ProductionFactor float32 `json:"production_factor"`
	}

	type lightCount struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Amount int    `json:"amount"`
	}

	type lightPlanet struct {
		ID                   string             `json:"id"`
		Player               string             `json:"player"`
		Coordinates          Coordinate         `json:"coordinate"`
		Name                 string             `json:"name"`
		Fields               int                `json:"fields"`
		MinTemp              int                `json:"min_temperature"`
		MaxTemp              int                `json:"max_temperature"`
		Diameter             int                `json:"diameter"`
		Resources            []ResourceInfo     `json:"resources"`
		Buildings            []lightInfo        `json:"buildings"`
		Ships                []lightCount       `json:"ships"`
		Defenses             []lightCount       `json:"defenses"`
		BuildingsUpgrade     []BuildingAction   `json:"buildings_upgrade"`
		TechnologiesUpgrade  []TechnologyAction `json:"technologies_upgrade"`
		ShipsConstruction    []ShipAction       `json:"ships_construction"`
		DefensesConstruction []DefenseAction    `json:"defenses_construction"`
		SourceFleets         []string           `json:"source_fleets"`
		IncomingFleets       []string           `json:"incoming_fleets"`
		CreatedAt            time.Time          `json:"created_at"`
		LastActivity         time.Time          `json:"last_activity"`
	}

	// Copy the planet's data.
	lp := lightPlanet{
		ID:                   p.ID,
		Player:               p.Player,
		Coordinates:          p.Coordinates,
		Name:                 p.Name,
		Fields:               p.Fields,
		MinTemp:              p.MinTemp,
		MaxTemp:              p.MaxTemp,
		Diameter:             p.Diameter,
		BuildingsUpgrade:     p.BuildingsUpgrade,
		TechnologiesUpgrade:  p.TechnologiesUpgrade,
		ShipsConstruction:    p.ShipsConstruction,
		DefensesConstruction: p.DefensesConstruction,
		SourceFleets:         p.SourceFleets,
		IncomingFleets:       p.IncomingFleets,
		CreatedAt:            p.CreatedAt,
		LastActivity:         p.LastActivity,
	}

	// Copy resources from map to slice.
	for _, r := range p.Resources {
		lp.Resources = append(lp.Resources, r)
	}

	// Make shallow copy of the buildings, ships and
	// defenses without including the tech deps.
	for _, b := range p.Buildings {
		lb := lightInfo{
			ID:               b.ID,
			Name:             b.Name,
			Level:            b.Level,
			ProductionFactor: b.ProductionFactor,
		}

		lp.Buildings = append(lp.Buildings, lb)
	}

	for _, s := range p.Ships {
		ls := lightCount{
			ID:     s.ID,
			Name:   s.Name,
			Amount: s.Amount,
		}

		lp.Ships = append(lp.Ships, ls)
	}

	for _, d := range p.Defenses {
		ld := lightCount{
			ID:     d.ID,
			Name:   d.Name,
			Amount: d.Amount,
		}

		lp.Defenses = append(lp.Defenses, ld)
	}

	return json.Marshal(lp)
}
