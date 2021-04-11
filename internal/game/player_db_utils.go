package game

import (
	"encoding/json"
	"oglike_server/pkg/db"
)

// SaveToDB :
// Used to save the content of this player to
// the DB. In case an error is raised during
// the operation a comprehensive error is
// returned.
//
// The `proxy` allows to access to the DB.
//
// Returns any error.
func (p *Player) SaveToDB(proxy db.Proxy) error {
	// Check consistency.
	if err := p.valid(); err != nil {
		return err
	}

	// Create the query and execute it: we will
	// use the dedicated handler to provide a
	// comprehensive error.
	query := db.InsertReq{
		Script: "create_player",
		Args:   []interface{}{p},
	}

	err := proxy.InsertToDB(query)

	return p.analyzeDBError(err)
}

// UpdateInDB :
// Used to update the content of the player in
// the DB. Only part of the player's data can
// be updated as specified by this function.
//
// The `proxy` allows to access to the DB.
//
// Returns any error.
func (p *Player) UpdateInDB(proxy db.Proxy) error {
	// Make sure that the name of the player is
	// valid.
	if p.Name == "" {
		return ErrInvalidUpdateData
	}

	// Create the query and execute it. In a
	// similar way we need to provide some
	// analysis of any error.
	query := db.InsertReq{
		Script: "update_player",
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

// DeleteFromDB :
// Used to perform the deletion of the player from the
// DB if the conditions are matched. It will make sure
// that nothing prevents the deletion of the player in
// the DB before doing so.
//
// The `data` defines a way to access to the DB.
//
// Returns any error.
func (p *Player) DeleteFromDB(data Instance) error {
	// A player can only be deleted if none of its colonized
	// planets are active in any way (no upgrades, no fleets,
	// etc.).
	for _, plaID := range p.Planets {
		pla, err := NewPlanetFromDB(plaID, data)
		if err != nil {
			return err
		}

		// If the planet is not eligible for deletion this is
		// also the case of the player.
		if err = pla.isEligibleForDeletion(); err != nil {
			return err
		}
	}

	// If we reached this point it means that the player can
	// be deleted.
	query := db.InsertReq{
		Script: "delete_player",
		Args: []interface{}{
			p.ID,
		},
	}

	err := data.Proxy.InsertToDB(query)

	return err
}

// analyzeDBError :
// used to perform the analysis of a DB error based
// on the structure of the players' table to produce
// a comprehensive error of what went wrong.
//
// The `err` defines the error to analyze.
//
// Returns a comprehensive error or the input error
// if nothing can be extracted from the input data.
func (p *Player) analyzeDBError(err error) error {
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
		case "players_pkey":
			return ErrDuplicatedElement
		case "players_universe_account_key":
			return ErrMultipleAccountInUniverse
		case "players_universe_name_key":
			return ErrNameAlreadyInUse
		}

		return dee
	}

	fkve, ok := dbe.Err.(db.ForeignKeyViolationError)
	if ok {
		switch fkve.ForeignKey {
		case "account":
			return ErrNonExistingAccount
		case "universe":
			return ErrNonExistingUniverse
		}

		return fkve
	}

	return dbe
}

// MarshalJSON :
// Implementation of the `Marshaler` interface to allow
// only specific information to be marshalled when the
// player needs to be exported. It fills a similar role
// to the `Convert` method but only to provide a clean
// interface to the outside world where only relevant
// info is provided.
//
// Returns the marshalled bytes for this player along
// with any error.
func (p *Player) MarshalJSON() ([]byte, error) {
	type lightInfo struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Level int    `json:"level"`
	}

	type lightPlayer struct {
		ID           string      `json:"id"`
		Account      string      `json:"account"`
		Universe     string      `json:"universe"`
		Name         string      `json:"name"`
		Technologies []lightInfo `json:"technologies"`
		Planets      []string    `json:"planets"`
		Score        Points      `json:"score"`
	}

	// Copy the planet's data.
	lp := lightPlayer{
		ID:       p.ID,
		Account:  p.Account,
		Universe: p.Universe,
		Name:     p.Name,
		Planets:  p.Planets,
		Score:    p.Score,
	}

	// Make shallow copy of the buildings, ships and
	// defenses without including the tech deps.
	for _, t := range p.Technologies {
		lt := lightInfo{
			ID:    t.ID,
			Name:  t.Name,
			Level: t.Level,
		}

		lp.Technologies = append(lp.Technologies, lt)
	}

	return json.Marshal(lp)
}
