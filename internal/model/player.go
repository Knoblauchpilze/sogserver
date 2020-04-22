package model

import (
	"fmt"
	"oglike_server/pkg/db"
)

// Player :
// Define a player which is the instance of an account in
// a particular universe. We can access to the nickname of
// the player in this universe along with the account it
// belongs to and the universe associated to it.
//
//
// The `ID` represents the unique ID for this player.
//
// The `Account` represents the identifier of the main
// account associated with this player. An account can be
// registered on any number of universes (with a limit of
// `1` character per universe).
//
// The `Universe` is the identifier of the universe in which
// this player is registered. This determines where it can
// perform actions.
//
// The `Name` represents the in-game display name for this
// player. It is distinct from the account's name.
//
// The `Technologies` defines each technology that this
// player has already researched with their associated
// level.
type Player struct {
	ID           string           `json:"id"`
	Account      string           `json:"account"`
	Universe     string           `json:"uni"`
	Name         string           `json:"name"`
	Technologies []TechnologyInfo `json:"technologies"`
}

// TechnologyInfo :
// Defines the information about a technology of a
// player. It reuses most of the base description
// for a technology with the addition of a level to
// indicate the current state reached by the player.
//
// The `Level` defines the level reached by this
// technology for a given player.
type TechnologyInfo struct {
	TechnologyDesc

	Level int `json:"level"`
}

// ErrInvalidPlayer :
// Used to indicate that the player provided in input is
// not valid.
var ErrInvalidPlayer = fmt.Errorf("Invalid player with no identifier")

// ErrDuplicatedPlayer :
// Used to indicate that the player's identifier provided
// input is not unique in the DB.
var ErrDuplicatedPlayer = fmt.Errorf("Invalid not unique player")

// Valid :
// Used to determine whether this player is valid. We will
// check that it does not contain obvious errors such as
// wrong identifiers etc. No check is performed to make
// sure that the player actually exists in the DB.
//
// Returns `true` if this player is valid.
func (p *Player) Valid() bool {
	return validUUID(p.ID) && validUUID(p.Account) && validUUID(p.Universe) && len(p.Name) > 0
}

// NewPlayerFromDB :
// Used to fetch the content of the player from the
// input DB and populate all internal fields from it.
// In case the DB cannot be fetched or some errors
// are encoutered, the return value will include a
// description of the error.
//
// The `ID` defines the identifier of the player to
// create. It should be fetched from the DB and is
// assumed to refer to an existing player.
//
// The `data` allows to actually perform the DB
// requests to fetch the player's data.
//
// Returns the player as fetched from the DB along
// with any errors.
func NewPlayerFromDB(ID string, data Instance) (Player, error) {
	// Create the player.
	p := Player{
		ID: ID,
	}

	// TODO: Should upgrade technology actions probably ?
	// It was done this way previously:
	// query := fmt.Sprintf("SELECT update_technology_upgrade_action('%s')", player)

	// Fetch the player's data.
	err := p.fetchGeneralInfo(data)
	if err != nil {
		return p, err
	}

	err = p.fetchTechnologies(data)
	if err != nil {
		return p, err
	}

	return p, nil
}

// fetchGeneralInfo :
// Used internally when building a player from the
// DB to populate the general information about the
// player such as its associated account and pseudo.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (p *Player) fetchGeneralInfo(data Instance) error {
	// Consistency.
	if p.ID == "" {
		return ErrInvalidPlayer
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"account",
			"uni",
			"name",
		},
		Table: "players",
		Filters: []db.Filter{
			{
				Key:    "id",
				Values: []string{p.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}

	// Scan the player's data.
	err = dbRes.Scan(
		&p.Account,
		&p.Universe,
		&p.Name,
	)

	// Make sure that it's the only player.
	if dbRes.Next() {
		return ErrDuplicatedPlayer
	}

	return nil
}

// fetchTechnologies :
// Similar to the `fetchGeneralInfo` but handles
// the retrieval of the player's technology data.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (p *Player) fetchTechnologies(data Instance) error {
	// Consistency.
	if p.ID == "" {
		return ErrInvalidPlayer
	}

	p.Technologies = make([]TechnologyInfo, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"technology",
			"level",
		},
		Table: "player_technologies",
		Filters: []db.Filter{
			{
				Key:    "player",
				Values: []string{p.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}

	// Populate the return value.
	var ID string
	var t TechnologyInfo

	for dbRes.Next() {
		err = dbRes.Scan(
			&ID,
			&t.Level,
		)

		if err != nil {
			return err
		}

		desc, err := data.Technologies.getTechnologyFromID(ID)
		if err != nil {
			return err
		}

		t.TechnologyDesc = desc

		// TODO: Update cost, production and storage.

		p.Technologies = append(p.Technologies, t)
	}

	return nil
}
