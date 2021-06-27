package game

import (
	"fmt"
	"math"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
)

// Player :
// Define a player which is the instance of an account in
// a particular universe. We can access to the nickname of
// the player in this universe along with the account it
// belongs to and the universe associated to it.
type Player struct {
	// The `ID` represents the unique ID for this player.
	ID string `json:"id"`

	// The `Account` represents the identifier of the main
	// account associated with this player. An account can be
	// registered on any number of universes (with a limit of
	// `1` character per universe).
	Account string `json:"account"`

	// The `Universe` is the identifier of the universe in which
	// this player is registered. This determines where it can
	// perform actions.
	Universe string `json:"universe"`

	// The `Name` represents the in-game display name for this
	// player. It is distinct from the account's name.
	Name string `json:"name"`

	// The `Technologies` defines each technology that this
	// player has already researched with their associated
	// level.
	Technologies map[string]TechnologyInfo `json:"-"`

	// The `Planets` defines the list of the identifiers of
	// the planets owned by this player.
	Planets []string `json:"planets"`

	// The `planetsCount` allows to count how many planets
	// this player already possesses.
	planetsCount int

	// The `fleetsCount` allows to count how many fleets are
	// already sent by this player. This accounts for both
	// the regular fleets and the expeditions.
	fleetsCount int

	// The `expCount` allows to count how many expedition
	// fleets are already sent by this player.
	expCount int

	// The `Points` defines the accumulated score of the
	// player.
	Score Points `json:"score"`
}

// Points :
// Define the score of a player by counting the points
// accumulated in various categories.
type Points struct {
	// The number of economy points gained by the player.
	Economy float32 `json:"economy"`

	// The number of research points gained by the player.
	Research float32 `json:"research"`

	// The current amount of military points owned by this
	// player.
	Military float32 `json:"military"`

	// The number of military points built by the
	// player.
	MilitaryBuilt float32 `json:"military_built"`

	// The number of military points lost by the player.
	MilitaryLost float32 `json:"military_lost"`

	// The number of military points destroyed by the
	// player.
	MilitaryDestroyed float32 `json:"military_destroyed"`
}

// TechnologyInfo :
// Defines the information about a technology of a
// player. It reuses most of the base description
// for a technology with the addition of a level to
// indicate the current state reached by the player.
type TechnologyInfo struct {
	// Reuses most of the base description for a
	// technology with the addition of a level to
	// indicate the current state reached by the
	// player.
	model.TechnologyDesc

	// The `Level` defines the level reached by this
	// technology for a given player.
	Level int `json:"level"`
}

// ErrInvalidUniverseForPlayer : Indicates that the universe for a player is not valid.
var ErrInvalidUniverseForPlayer = fmt.Errorf("invalid universe identifier for player")

// ErrInvalidAccountForPlayer : Indicates that the account for a player is not valid.
var ErrInvalidAccountForPlayer = fmt.Errorf("invalid account identifier for player")

// ErrMultipleAccountInUniverse : Indicates that a player tries to register multiple
// times in a single universe.
var ErrMultipleAccountInUniverse = fmt.Errorf("cannot register account twice in a universe")

// ErrNameAlreadyInUse : Indicates that the name for a player is already in use.
var ErrNameAlreadyInUse = fmt.Errorf("name is already in use in universe")

// ErrNonExistingAccount : Indicates that the account does not exist for this player.
var ErrNonExistingAccount = fmt.Errorf("inexisting parent account")

// ErrNonExistingUniverse : Indicates that the universe does not exist for this player.
var ErrNonExistingUniverse = fmt.Errorf("inexisting parent universe")

// ErrInconsistentPlanetFound : Indicates that inconsistencies were found for the
// planets of a player.
var ErrInconsistentPlanetFound = fmt.Errorf("inconsistencies found for planets of a player")

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
func NewPlayerFromDB(ID string, data Instance) (Player, error) {
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

	err = p.fetchFleetsCount(data)
	if err != nil {
		return p, err
	}

	err = p.fetchTechnologies(data)
	if err != nil {
		return p, err
	}

	err = p.fetchPlanets(data)

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
func (p *Player) fetchGeneralInfo(data Instance) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"p.account",
			"p.universe",
			"p.name",
			"pl.economy_points",
			"pl.research_points",
			"pl.military_points",
			"pl.military_points_built",
			"pl.military_points_lost",
			"pl.military_points_destroyed",
		},
		Table: "players p inner join players_points pl on p.id = pl.player",
		Filters: []db.Filter{
			{
				Key:    "p.id",
				Values: []interface{}{p.ID},
			},
		},
		Verbose: true,
	}

	dbRes, err := data.Proxy.FetchFromDB(query)

	// Check for errors.
	if err != nil {
		return err
	}
	defer dbRes.Close()

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
		&p.Score.Economy,
		&p.Score.Research,
		&p.Score.Military,
		&p.Score.MilitaryBuilt,
		&p.Score.MilitaryLost,
		&p.Score.MilitaryDestroyed,
	)

	// Make sure that it's the only player.
	if dbRes.Next() {
		return ErrDuplicatedElement
	}

	return err
}

// fetchFleetsCount :
// Simialr to the `fetchGeneralInfo` but handles
// the retrieval of the player's fleets count.
//
// The `data` defines the object to access the
// DB.
//
// Returns any error.
func (p *Player) fetchFleetsCount(data Instance) error {
	// Use a function to queue both requests after one another.
	var gErr error

	func() {
		// Create the query and execute it.
		query := db.QueryDesc{
			Props: []string{
				"count(*)",
			},
			Table: "fleets",
			Filters: []db.Filter{
				{
					Key:    "player",
					Values: []interface{}{p.ID},
				},
			},
		}

		dbRes, err := data.Proxy.FetchFromDB(query)

		// Check for errors.
		if err != nil {
			gErr = err
			return
		}
		defer dbRes.Close()

		if dbRes.Err != nil {
			gErr = dbRes.Err
			return
		}

		// Scan the player's data.
		atLeastOne := dbRes.Next()
		if !atLeastOne {
			gErr = ErrElementNotFound
			return
		}

		gErr = dbRes.Scan(
			&p.fleetsCount,
		)

		// Make sure that it's the only player.
		if dbRes.Next() {
			gErr = ErrDuplicatedElement
			return
		}
	}()

	if gErr != nil {
		return gErr
	}

	// Fetch expeditions count.
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"count(*)",
		},
		Table: "fleets f inner join fleets_objectives fo on f.objective = fo.id",
		Filters: []db.Filter{
			{
				Key:    "f.player",
				Values: []interface{}{p.ID},
			},
			{
				Key:    "fo.name",
				Values: []interface{}{"expedition"},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)

	// Check for errors.
	if err != nil {
		return err
	}
	defer dbRes.Close()

	if dbRes.Err != nil {
		return dbRes.Err
	}

	// Scan the player's data.
	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return ErrElementNotFound
	}

	err = dbRes.Scan(
		&p.expCount,
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
func (p *Player) fetchTechnologies(data Instance) error {
	p.Technologies = make(map[string]TechnologyInfo)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"technology",
			"level",
		},
		Table: "players_technologies",
		Filters: []db.Filter{
			{
				Key:    "player",
				Values: []interface{}{p.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)

	// Check for errors.
	if err != nil {
		return err
	}
	defer dbRes.Close()

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

// fetchPlanets :
// Similar to `fetchGeneralInfo` but allows to
// retrieve the identifiers of the planets owned
// by a player.
//
// The `data` defines the object to access the
// DB.
//
// Returns any error.
func (p *Player) fetchPlanets(data Instance) error {
	p.planetsCount = 0
	p.Planets = make([]string, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table: "planets",
		Filters: []db.Filter{
			{
				Key:    "player",
				Values: []interface{}{p.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)

	// Check for errors.
	if err != nil {
		return err
	}
	defer dbRes.Close()

	if dbRes.Err != nil {
		return dbRes.Err
	}

	var ID string

	for dbRes.Next() {
		err = dbRes.Scan(&ID)

		if err != nil {
			return err
		}

		p.Planets = append(p.Planets, ID)
		p.planetsCount++
	}

	return nil
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

// CanColonize :
// Used to determine whether this player is able to
// perform a colonization operation given its level
// of astrophysics research and the count of planets
// he already owns.
//
// The `data` allows to fetch information about the
// colonization process.
//
// Returns `true` if the player can colonize a new
// planet along with any error.
func (p *Player) CanColonize(data Instance) (bool, error) {
	// We will compare the level of the astrophysics
	// research against the number of planets already
	// colonized by the player.
	astroID, err := data.Technologies.GetIDFromName("astrophysics")
	if err != nil {
		return false, err
	}

	astro, ok := p.Technologies[astroID]

	if !ok {
		// The astrophysics technology is not researched,
		// the player cannot colonize beyond the homeworld.
		return false, nil
	}

	// Every two levels of astrophysics, a new colony
	// can be added to a player's empire. So the level
	// `0` means no colony, the `1`st level allows the
	// first colony, the `3`rd level allows the second
	// etc.
	// This gives the array below:
	//
	// +-------+----------+---------+
	// | Level | Colonies | Planets |
	// +-------+----------+---------+
	// |   0   |    0     |    1    |
	// +-------+----------+---------+
	// |   1   |    1     |    2    |
	// +-------+----------+---------+
	// |   2   |    1     |    2    |
	// +-------+----------+---------+
	// |   3   |    2     |    3    |
	// +-------+----------+---------+
	// |   4   |    2     |    3    |
	// +-------+----------+---------+
	// |   5   |    3     |    4    |
	// +-------+----------+---------+
	//
	// So the general formula seems to be the following:
	//
	// planets = 2 + (level - 1) / 2
	//
	// It works everywhere except when the level of the
	// astrophysics research is `0`. In which case we
	// forcibly return `false` as there's no colonization
	// possible.
	if astro.Level == 0 {
		return false, nil
	}

	maxPlanetsCount := 2 + (astro.Level-1)/2

	return p.planetsCount < maxPlanetsCount, nil
}

// CanSendFleet :
// Used to determine whether this player is able to
// send a new fleet based on the current level of
// the computers technology and the number of fleets
// already created.
//
// The `data` allows to fetch information about the
// fleets already available.
//
// `objective` the objective of the fleet. Allows to
// potentially differientate and allow certain types
// of fleets to be send.
//
// Returns `true` if the player can send a new fleet
// along with any errors.
func (p *Player) CanSendFleet(data Instance, objective string) (bool, error) {
	// We will compare the level of the computers
	// research against the number of fleets already
	// send by the player.
	compID, err := data.Technologies.GetIDFromName("computers")
	if err != nil {
		return false, err
	}

	comp, ok := p.Technologies[compID]

	if !ok {
		// The computers technology is not researched,
		// the player can send only a single fleet.
		return p.fleetsCount == 0, nil
	}

	astroID, err := data.Technologies.GetIDFromName("astrophysics")
	if err != nil {
		return false, err
	}

	astro, ok := p.Technologies[astroID]

	if !ok {
		// The astrophysics technology is not researched,
		// the player cannot colonize beyond the homeworld.
		return false, nil
	}

	// Compute the number of fleets that can be sent
	// based on the following formulas:
	// https://ogame.fandom.com/wiki/Computer_Technology
	// https://ogame.fandom.com/wiki/Astrophysics
	fleetsMax := comp.Level + 1
	expMax := int(math.Floor(math.Sqrt(float64(astro.Level))))

	// Determine whether the objective defines an expedition
	// or a regular fleet.
	expID, err := data.Objectives.GetIDFromName("expedition")
	if err != nil {
		return false, err
	}

	// Verify that not too many expeditions have been
	// sent already.
	if objective == expID && p.expCount+1 > expMax {
		// Too many expeditions.
		return false, nil
	}

	// The total number of fleets should be smaller or
	// equal to the authorized amount.
	return p.fleetsCount+1 <= fleetsMax, nil
}
