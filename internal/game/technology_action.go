package game

import (
	"fmt"
	"oglike_server/pkg/db"
	"time"
)

// TechnologyAction :
// Used as a way to refine the `ProgressAction` for the
// specific case of technologies. It mostly add the info
// to compute the completion time for a technology.
type TechnologyAction struct {
	// Reuses the notion of a progress action.
	ProgressAction

	// The `Player` defines the player owning the planet on
	// which this action is performed.
	Player string `json:"player"`
}

// ErrInvalidPointsForAction : Indicates that the number of points brought by the
// action is not valid.
var ErrInvalidPointsForAction = fmt.Errorf("invalid completion points for action")

// ErrNonExistingPlayer : Indicates that the parent player for the action does not exist.
var ErrNonExistingPlayer = fmt.Errorf("parent player does not exist")

// valid :
// Determines whether this action is valid. By valid we
// only mean obvious syntax errors.
//
// Returns any error or `nil` if the action seems valid.
func (a *TechnologyAction) valid() error {
	if err := a.ProgressAction.valid(); err != nil {
		return err
	}

	if !validUUID(a.Player) {
		return ErrInvalidPlayerForAction
	}
	if a.DesiredLevel != a.CurrentLevel+1 {
		return ErrInvalidLevelForAction
	}

	if a.Points <= 0.0 {
		return ErrInvalidPointsForAction
	}

	return nil
}

// NewTechnologyActionFromDB :
// Used similarly to the `NewBuildingActionFromDB`
// element but to fetch the actions related to the
// research of a new technology by a player on a
// given planet.
//
// The `ID` defines the identifier of the action to
// fetch from the DB.
//
// The `data` allows to actually access to the
// data in the DB.
//
// Returns the corresponding technology action
// along with any error.
func NewTechnologyActionFromDB(ID string, data Instance) (TechnologyAction, error) {
	// Create the return value and fetch the base
	// data for this action.
	a := TechnologyAction{}

	// Create the action using the base handler.
	var err error
	a.ProgressAction, err = newProgressActionFromDB(ID, data, "construction_actions_technologies", false)

	// Consistency.
	if err != nil {
		return a, err
	}

	// Update the cost for this action. We will fetch
	// the tech related to the action and compute how
	// many resources are needed to build it.
	sd, err := data.Technologies.GetTechnologyFromID(a.Element)
	if err != nil {
		return a, err
	}

	costs := sd.Cost.ComputeCost(a.CurrentLevel)
	a.Costs = make([]Cost, 0)

	for res, amount := range costs {
		c := Cost{
			Resource: res,
			Cost:     float32(amount),
		}

		a.Costs = append(a.Costs, c)
	}

	return a, nil
}

// SaveToDB :
// Used to save the content of this action to
// the DB. In case an error is raised during
// the operation a comprehensive error is
// returned.
//
// The `proxy` allows to access to the DB.
//
// Returns any error.
func (a *TechnologyAction) SaveToDB(proxy db.Proxy) error {
	// Check consistency.
	if err := a.valid(); err != nil {
		return err
	}

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_technology_upgrade_action",
		Args: []interface{}{
			a,
			a.Costs,
		},
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
		case "construction_actions_technologies_player_key":
			return ErrOnlyOneActionAuthorized
		}

		return dee
	}

	fkve, ok := dbe.Err.(db.ForeignKeyViolationError)
	if ok {
		switch fkve.ForeignKey {
		case "planet":
			return ErrNonExistingPlanet
		case "element":
			return ErrNonExistingElement
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
// that need to be marshalled in the fleet's creation.
//
// Returns the converted version of this action which
// only includes relevant fields.
func (a *TechnologyAction) Convert() interface{} {
	return struct {
		ID             string    `json:"id"`
		Planet         string    `json:"planet"`
		Player         string    `json:"player"`
		Element        string    `json:"element"`
		CurrentLevel   int       `json:"current_level"`
		DesiredLevel   int       `json:"desired_level"`
		Points         float32   `json:"points"`
		CompletionTime time.Time `json:"completion_time"`
		CreatedAt      time.Time `json:"created_at"`
	}{
		ID:             a.ID,
		Planet:         a.Planet,
		Player:         a.Player,
		Element:        a.Element,
		CurrentLevel:   a.CurrentLevel,
		DesiredLevel:   a.DesiredLevel,
		Points:         a.Points,
		CompletionTime: a.CompletionTime,
		CreatedAt:      a.creationTime,
	}
}

// ConsolidateEffects :
// Used to make sure that the number of points granted
// when completing this action are consistent with the
// desired level of the building.
//
// The `data` defines a way to access to the duration
// and costs of technologies.
//
// The `p` defines the parent planet where the action
// is meant to be performed.
//
// Returns any error.
func (a *TechnologyAction) ConsolidateEffects(data Instance, p *Planet) error {
	// Consistency.
	if a.Planet != p.ID {
		return ErrMismatchInVerification
	}

	// We need to retrieve the technology related to this
	// action.
	td, err := data.Technologies.GetTechnologyFromID(a.Element)
	if err != nil {
		return err
	}

	// Now we can compute the additional points that will
	// be brought by this action upon completing it.
	costs := td.Cost.ComputeCost(a.CurrentLevel)

	a.Points = 0.0
	for _, cost := range costs {
		a.Points += float32(cost)
	}
	a.Points /= 1000

	return nil
}

// consolidateCompletionTime :
// Used to update the completion time required for this
// action to complete based on the amount of resources
// needed by the next level of the technology level. It
// also uses the level of research labs for the player
// researching the technology.
//
// The `data` allows to get information on the data
// that will be used to compute the completion time.
//
// The `p` argument defines the planet onto which the
// action should be performed. Note that we assume it
// corresponds to the actual planet defined by this
// action. It is used in order not to dead lock with
// the planet that has probably already been acquired
// by the action creation process.
//
// The `ratio` defines a flat multiplier to apply to
// the completion time of the action to take the parent
// universe properties into consideration.
//
// Returns any error.
func (a *TechnologyAction) consolidateCompletionTime(data Instance, p *Planet, ratio float32) error {
	// Consistency.
	if a.Planet != p.ID {
		return ErrMismatchInVerification
	}

	// First, we need to determine the cost for each of
	// the individual unit to produce.
	td, err := data.Technologies.GetTechnologyFromID(a.Element)
	if err != nil {
		return err
	}

	costs := td.Cost.ComputeCost(a.CurrentLevel)

	// Populate the cost.
	a.Costs = make([]Cost, 0)

	for res, amount := range costs {
		c := Cost{
			Resource: res,
			Cost:     float32(amount),
		}

		a.Costs = append(a.Costs, c)
	}

	// Fetch the total research power available for this
	// action. It will not account for the current planet
	// research lab so we still have to use it.
	power, err := a.fetchResearchPower(data, p)
	if err != nil {
		return err
	}

	// Retrieve the cost in metal and crystal as it is
	// the only costs that matters.
	metalDesc, err := data.Resources.GetResourceFromName("metal")
	if err != nil {
		return err
	}
	crystalDesc, err := data.Resources.GetResourceFromName("crystal")
	if err != nil {
		return err
	}

	m := costs[metalDesc.ID]
	c := costs[crystalDesc.ID]

	hours := float64(m+c) / (1000.0 * (1.0 + float64(power)))
	hours *= float64(ratio)

	t, err := time.ParseDuration(fmt.Sprintf("%fh", hours))
	if err != nil {
		return ErrInvalidDuration
	}

	a.creationTime = time.Now()
	a.CompletionTime = a.creationTime.Add(t)

	return nil
}

// fetchResearchPower :
// Used to fetch the research power available for the input
// planet. It will query the list of research labs on all
// planets of the player and select the required amount as
// defined by the level of the galactic research network.
// It *will* include the level of the planet linked to this
// action.
//
// The `data` allows to access to the DB.
//
// The `planet` defines the planet for which the research
// power should be computed.
//
// Returns the research power available including the
// power brought by this planet along with any error.
func (a *TechnologyAction) fetchResearchPower(data Instance, planet *Planet) (int, error) {
	// First, fetch the level of the research lab on the
	// planet associated to this action: this will be the
	// base of the research.
	labID, err := data.Buildings.GetIDFromName("research lab")
	if err != nil {
		return 0, err
	}
	lab := planet.Buildings[labID]

	// Get the level of the intergalactic research network
	// technology reached by the player owning this planet.
	// It will indicate how many elements we should keep
	// in the list of reserch labs.
	igrn, err := data.Technologies.GetIDFromName("intergalactic research network")
	if err != nil {
		return lab.Level, err
	}

	labCount := planet.technologies[igrn]

	// Perform the query to get the levels of the labs on
	// each planet owned by this player.
	query := db.QueryDesc{
		Props: []string{
			"pb.planet",
			"pb.level",
		},
		Table: "planets_buildings pb inner join planets p on pb.planet=p.id",
		Filters: []db.Filter{
			{
				Key:    "p.player",
				Values: []interface{}{planet.Player},
			},
		},
		// Note that we add `1` to the number of research labs in order
		// to account for the lab doing the research. Level `1` actually
		// tells that 1 lab can research the same techno at the same
		// time than the one launching the research.
		Ordering: fmt.Sprintf("order by level desc limit %d", labCount+1),
	}

	dbRes, err := data.Proxy.FetchFromDB(query)

	// Check for errors.
	if err != nil {
		return lab.Level, err
	}
	defer dbRes.Close()

	if dbRes.Err != nil {
		return lab.Level, dbRes.Err
	}

	var pID string
	var labLevel int
	power := 0
	processedLabs := 0
	planetBelongsToTopLabs := false

	for dbRes.Next() && ((!planetBelongsToTopLabs && processedLabs < labCount) || planetBelongsToTopLabs) {
		err = dbRes.Scan(
			&pID,
			&labLevel,
		)

		if err != nil {
			return lab.Level, err
		}

		if pID == planet.ID {
			planetBelongsToTopLabs = true
		} else {
			power += labLevel
		}

		processedLabs++
	}

	return lab.Level + power, nil
}

// Validate :
// Used to make sure that the action can be performed on
// the planet it is linked to. This will check that the
// tech tree is consistent with what's expected from the
// ship, that resources are available etc.
//
// The `data` allows to access to the DB if needed.
//
// The `p` defines the planet attached to this action:
// it needs to be provided as input so that resource
// locking is easier.
//
// The `ratio` defines a flat multiplier to apply to
// the result of the validation and more specifically
// to the computation of the completion time. It helps
// taking into account the properties of the parent's
// universe.
//
// Returns any error.
func (a *TechnologyAction) Validate(data Instance, p *Planet, ratio float32) error {
	// Consistency.
	if a.Planet != p.ID || a.Player != p.Player {
		return ErrMismatchInVerification
	}

	// Update completion time and costs.
	err := a.consolidateCompletionTime(data, p, ratio)
	if err != nil {
		return err
	}

	// Make sure that the current level of the technology
	// is consistent with what's desired.
	td, err := data.Technologies.GetTechnologyFromID(a.Element)
	if err != nil {
		return err
	}

	tLevel, ok := p.technologies[td.ID]
	if !ok && a.CurrentLevel > 0 {
		return ErrLevelIncorrect
	}
	if tLevel != a.CurrentLevel {
		return ErrLevelIncorrect
	}

	// Validate against planet's data.
	costs := td.Cost.ComputeCost(a.CurrentLevel)

	return p.validateAction(costs, td.UpgradableDesc, data)
}
