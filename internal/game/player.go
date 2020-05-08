package game

import (
	"encoding/json"
	"fmt"
	"oglike_server/internal/model"
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
	ID           string                    `json:"id"`
	Account      string                    `json:"account"`
	Universe     string                    `json:"universe"`
	Name         string                    `json:"name"`
	Technologies map[string]TechnologyInfo `json:"-"`
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
	model.TechnologyDesc

	Level int `json:"level"`
}

// ErrInvalidUniverseForPlayer : Indicates that the universe for a player is not valid.
var ErrInvalidUniverseForPlayer = fmt.Errorf("Invalid universe identifier for player")

// ErrInvalidAccountForPlayer : Indicates that the account for a player is not valid.
var ErrInvalidAccountForPlayer = fmt.Errorf("Invalid account identifier for player")

// ErrMultipleAccountInUniverse : Indicates that a player tries to register multiple
// times in a single universe.
var ErrMultipleAccountInUniverse = fmt.Errorf("Cannot register account twice in a universe")

// ErrNameAlreadyInUse : Indicates that the name for a player is already in use.
var ErrNameAlreadyInUse = fmt.Errorf("Name is already in use in universe")

// ErrNonExistingAccount : Indicates that the account does not exist for this player.
var ErrNonExistingAccount = fmt.Errorf("Inexisting parent account")

// ErrNonExistingUniverse : Indicates that the universe does not exist for this player.
var ErrNonExistingUniverse = fmt.Errorf("Inexisting parent universe")

// valid :
// Determines whether the player is valid. By valid we only mean
// obvious syntax errors.
//
// Returns any error or `nil` if the player seems valid.
func (p *Player) valid() error {
	if !validUUID(p.ID) {
		return ErrInvalidElementID
	}
	if p.Name == "" {
		return ErrInvalidName
	}
	if !validUUID(p.Account) {
		return ErrInvalidAccountForPlayer
	}
	if !validUUID(p.Universe) {
		return ErrInvalidUniverseForPlayer
	}

	return nil
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
// The `mode` defines the reading mode for the data
// access for this planet.
//
// Returns the player as fetched from the DB along
// with any errors.
func NewPlayerFromDB(ID string, data model.Instance) (Player, error) {
	// Create the player.
	p := Player{
		ID: ID,
	}

	// Consistency.
	if !validUUID(p.ID) {
		return p, ErrInvalidElementID
	}

	// Fetch the player's data.
	err := p.fetchGeneralInfo(data)
	if err != nil {
		return p, err
	}

	err = p.fetchTechnologies(data)

	return p, err
}

// fetchGeneralInfo :
// Used internally when building a player from the
// DB to populate the general information about the
// player such as its associated account and pseudo.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (p *Player) fetchGeneralInfo(data model.Instance) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"account",
			"universe",
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
	if dbRes.Err != nil {
		return dbRes.Err
	}

	// Scan the player's data.
	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return ErrElementNotFound
	}

	err = dbRes.Scan(
		&p.Account,
		&p.Universe,
		&p.Name,
	)

	// Make sure that it's the only player.
	if dbRes.Next() {
		return ErrDuplicatedElement
	}

	return err
}

// fetchTechnologies :
// Similar to the `fetchGeneralInfo` but handles
// the retrieval of the player's technology data.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (p *Player) fetchTechnologies(data model.Instance) error {
	p.Technologies = make(map[string]TechnologyInfo, 0)

	// Before fetching the technologies we need to
	// perform the update of the upgrade actions if
	// any.
	err := data.UpdateTechnologiesForPlayer(p.ID)
	if err != nil {
		return err
	}

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
	if dbRes.Err != nil {
		return dbRes.Err
	}

	// Populate the return value.
	var ID string
	var t TechnologyInfo

	sanity := make(map[string]bool)

	for dbRes.Next() {
		err = dbRes.Scan(
			&ID,
			&t.Level,
		)

		if err != nil {
			return err
		}

		_, ok := sanity[ID]
		if ok {
			return model.ErrInconsistentDB
		}
		sanity[ID] = true

		desc, err := data.Technologies.GetTechnologyFromID(ID)
		if err != nil {
			return err
		}

		t.TechnologyDesc = desc

		p.Technologies[ID] = t
	}

	return nil
}

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

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_player",
		Args:   []interface{}{p},
	}

	err := proxy.InsertToDB(query)

	// Analyze the error in order to provide some
	// comprehensive message.
	dbe, ok := err.(db.Error)
	if !ok {
		return err
	}

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

// GetTechnology :
// Retrieves the technology from the input identifier.
//
// The `ID` defines the identifier of the technology
// to fetch from the player.
//
// Returns the technology description corresponding
// to the input identifier along with any error.
func (p *Player) GetTechnology(ID string) (TechnologyInfo, error) {
	tech, ok := p.Technologies[ID]

	if !ok {
		return TechnologyInfo{}, model.ErrInvalidID
	}

	return tech, nil
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
	}

	// Copy the planet's data.
	lp := lightPlayer{
		ID:       p.ID,
		Account:  p.Account,
		Universe: p.Universe,
		Name:     p.Name,
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
