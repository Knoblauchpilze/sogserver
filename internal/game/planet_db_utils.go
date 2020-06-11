package game

import (
	"encoding/json"
	"oglike_server/pkg/db"
	"time"
)

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
	if p.moon {
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
		Verbose: true,
	}

	err := proxy.InsertToDB(query)

	return p.analyzeDBError(err)
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
	if p.moon {
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
		ID       string `json:"id"`
		Player   string `json:"player"`
		Name     string `json:"name"`
		MinTemp  int    `json:"min_temperature"`
		MaxTemp  int    `json:"max_temperature"`
		Fields   int    `json:"fields"`
		Galaxy   int    `json:"galaxy"`
		System   int    `json:"solar_system"`
		Position int    `json:"position"`
		Diameter int    `json:"diameter"`
	}{
		ID:       p.ID,
		Player:   p.Player,
		Name:     p.Name,
		MinTemp:  p.MinTemp,
		MaxTemp:  p.MaxTemp,
		Fields:   p.Fields,
		Galaxy:   p.Coordinates.Galaxy,
		System:   p.Coordinates.System,
		Position: p.Coordinates.Position,
		Diameter: p.Diameter,
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
		ID    string `json:"id"`
		Name  string `json:"name"`
		Level int    `json:"level"`
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
		Buildings            []lightInfo        `json:"buildings,omitempty"`
		Ships                []lightCount       `json:"ships,omitempty"`
		Defenses             []lightCount       `json:"defenses,omitempty"`
		BuildingsUpgrade     []BuildingAction   `json:"buildings_upgrade,omitempty"`
		TechnologiesUpgrade  []TechnologyAction `json:"technologies_upgrade,omitempty"`
		ShipsConstruction    []ShipAction       `json:"ships_construction,omitempty"`
		DefensesConstruction []DefenseAction    `json:"defenses_construction,omitempty"`
		SourceFleets         []string           `json:"source_fleets,omitempty"`
		IncomingFleets       []string           `json:"incoming_fleets,omitempty"`
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
	}

	// Copy resources from map to slice.
	for _, r := range p.Resources {
		lp.Resources = append(lp.Resources, r)
	}

	// Make shallow copy of the buildings, ships and
	// defenses without including the tech deps.
	for _, b := range p.Buildings {
		lb := lightInfo{
			ID:    b.ID,
			Name:  b.Name,
			Level: b.Level,
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
