package data

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"oglike_server/internal/locker"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"

	"github.com/google/uuid"
)

// PlanetProxy :
// Intended as a wrapper to access properties of planets
// and retrieve data from the database. This helps hiding
// the complexity of how the data is laid out in the `DB`
// and the precise name of tables from the exterior world.
// Note that as this proxy uses some functionalities to
// fetch universes information we figured that it would
// be more interesting to factorize the behavior and reuse
// the functions through composition.
//
// The `dbase` is the database that is wrapped by this
// object. It is checked for consistency upon building the
// wrapper.
//
// The `log` allows to perform display to the user so as
// to inform of potential issues and debug information to
// the outside world.
//
// The `resources` is a map populated from the DB which
// keeps track of the identifier in the DB for each of
// the resources used by the game. It allows to be much
// more efficient in the event of a creation of a planet
// as we only need to query the local information about
// resources rather than contacting the DB each time.
// In addition to the name this table also defines the
// base production for each resource and the initial
// storage available for it.
//
// The `uniProxy` defines a proxy allowing to access a
// part of the behavior related to universes typically
// when fetching a universe into which a planet should
// be created.
//
// The `lock` allows to lock specific resources when some
// planet data should be retrieved. A planet can indeed
// have some upgrade actions attached to it to define the
// upgrade of a building, or the construction of ships or
// defense systems.
// Each of these actions are executed in a lazy update
// fashion where the action is created and then performed
// only when needed: we store the completion time and if
// the data needs to be accessed we upgrade it.
// This mechanism requires that when the data needs to be
// fetched for a planet an update operation should first
// be performed to ensure that the data is up-to-date.
// As no data is shared across planets we don't see the
// need to lock all the planets when a single one should
// be updated. Using the `ConcurrentLock` we have a way
// to lock only some planets which is exactly what we
// need.
//
// The `buildingCosts` is used when the data for a planet
// is needed in order to compute the costs that are linked
// to the next level of a building based on the level that
// has been reached so far. It is initialized upon creating
// the proxy so that the information is readily available
// when needed.
//
// The `prodRules` is used in a similar way as the other
// `buildingCosts` attribute. It regroups the information
// about the production of each resource associated to
// some building. Typically mines are able to generate a
// certain quantity of resources for each level: this is
// an information that hardly changes throughout the life
// of the server and we prefer to load it in memory so as
// to make it easily available.
type PlanetProxy struct {
	dbase         *db.DB
	log           logger.Logger
	resources     map[string]ResourceDesc
	uniProxy      UniverseProxy
	lock          *locker.ConcurrentLocker
	buildingCosts map[string]ConstructionCost
	prodRules     map[string][]ProductionRule
}

// getDefaultPlanetName :
// Used to retrieve a default name for a planet. The
// generated name will be different based on whether
// the planet is a homeworld or no.
//
// The `isHomeworld` is `true` if we should generate
// a name for the homeworld and `false` otherwise.
//
// Returns a string corresponding to the name of the
// planet.
func getDefaultPlanetName(isHomeWorld bool) string {
	if isHomeWorld {
		return "homeworld"
	}

	return "planet"
}

// getPlanetTemperatureAmplitude :
// Used to retrieve the default planet temperature's
// amplitude. Basically the interval between the min
// and max temperature will always be equal to this
// value.
//
// Returns the default temperature amplitude for the
// planets.
func getPlanetTemperatureAmplitude() int {
	return 50
}

// NewPlanetProxy :
// Create a new proxy on the input `dbase` to access the
// properties of planets as registered in the DB. In
// case the provided DB is `nil` a panic is issued.
//
// The `dbase` represents the database to use to fetch
// data related to planets.
//
// The `log` will be used to notify information so that
// we can have an idea of the activity of this component.
// One possible example is for timing the requests.
//
// The `unis` defines a proxy that can be used to fetch
// information about the universes when creating planets.
//
// Returns the created proxy.
func NewPlanetProxy(dbase *db.DB, log logger.Logger, unis UniverseProxy) PlanetProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create planets proxy from invalid DB"))
	}

	// Fetch resources from the DB to populate the internal
	// map. We will use the dedicated handler which is used
	// to actually fetch the data and always return a valid
	// value.
	resources, err := initResourcesFromDB(dbase, log)
	if err != nil {
		log.Trace(logger.Error, fmt.Sprintf("Could not fetch resources' identifiers from DB (err: %v)", err))
	}

	// In a similar way, fetch the initial costs and the
	// progression rules for each building.
	buildingCosts, err := initProgressCostsFromDB(dbase, log, "building", "buildings_costs_progress", "buildings_costs")
	if err != nil {
		log.Trace(logger.Error, fmt.Sprintf("Could not fetch buildings costs from DB (err: %v)", err))
	}

	// Finally fetch the production rules for each building.
	prodRules, err := initBuildingsProductionRulesFromDB(dbase, log)
	if err != nil {
		log.Trace(logger.Error, fmt.Sprintf("Could not fetch buildings production rules from DB (err: %v)", err))
	}

	return PlanetProxy{
		dbase,
		log,
		resources,
		unis,
		locker.NewConcurrentLocker(log),
		buildingCosts,
		prodRules,
	}
}

// Planets :
// Return a list of planets registered so far in all the planets
// defined in the DB. The input filters might help to narrow the
// search a bit by providing coordinates to look for and a uni to
// look into.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// accounts available. Each one is appended `as-is` to the
// query.
//
// Returns the list of planets registered in the DB and matching
// the input list of filters. In case the error is not `nil` the
// value of the array should be ignored.
func (p *PlanetProxy) Planets(filters []DBFilter) ([]Planet, error) {
	// Create the query and execute it.
	props := []string{
		"p.id",
		"p.player",
		"p.name",
		"p.fields",
		"p.min_temperature",
		"p.max_temperature",
		"p.diameter",
		"p.galaxy",
		"p.solar_system",
		"p.position",
	}

	table := "planets p inner join players pl"
	joinCond := "p.player=pl.id"

	query := fmt.Sprintf("select %s from %s on %s", strings.Join(props, ", "), table, joinCond)

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
		return nil, fmt.Errorf("Could not query DB to fetch planets (err: %v)", err)
	}

	// Populate the return value.
	planets := make([]Planet, 0)
	var planet Planet

	for rows.Next() {
		err = rows.Scan(
			&planet.ID,
			&planet.PlayerID,
			&planet.Name,
			&planet.Fields,
			&planet.MinTemp,
			&planet.MaxTemp,
			&planet.Diameter,
			&planet.Galaxy,
			&planet.System,
			&planet.Position,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for planet (err: %v)", err))
			continue
		}

		// Fetch buildings, ships and defenses for this planet.
		err = p.fetchPlanetData(&planet)
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch data for planet \"%s\" (err: %v)", planet.ID, err))
			continue
		}

		planets = append(planets, planet)
	}

	return planets, nil
}

// fetchPlanetData :
// Used to fetch data built on the planet provided in input.
// This typically include the buildings, the ships deployed
// and the defenses installed.
//
// The `planet` references the planet for which data should
// be fetched. We assume that the internal fields (and more
// specifically the identifier) are already populated.
//
// Returns any error.
func (p *PlanetProxy) fetchPlanetData(planet *Planet) error {
	// Check whether the planet has an identifier assigned.
	if planet.ID == "" {
		return fmt.Errorf("Unable to fetch data from planet with invalid identifier")
	}

	// Fetch resources.
	err := p.fetchPlanetResources(planet)
	if err != nil {
		return fmt.Errorf("Could not fetch resources for planet \"%s\" (err: %v)", planet.ID, err)
	}

	// Fetch buildings.
	err = p.fetchPlanetBuildings(planet)
	if err != nil {
		return fmt.Errorf("Could not fetch buildings for planet \"%s\" (err: %v)", planet.ID, err)
	}

	// Fetch ships.
	err = p.fetchPlanetShips(planet)
	if err != nil {
		return fmt.Errorf("Could not fetch ships for planet \"%s\" (err: %v)", planet.ID, err)
	}

	// Fetch defenses.
	err = p.fetchPlanetDefenses(planet)
	if err != nil {
		return fmt.Errorf("Could not fetch defenses for planet \"%s\" (err: %v)", planet.ID, err)
	}

	return nil
}

// fetchPlanetResources :
// Used to fetch the resources currently present on a planet.
// We need to execute a script to update the production of a
// planet since the last actualization.
//
// The `planet` defines the planet for which resources should
// be updated. An invalie value will return an error.
//
// Returns any error.
func (p *PlanetProxy) fetchPlanetResources(planet *Planet) error {
	// Check consistency.
	if planet == nil || planet.ID == "" {
		return fmt.Errorf("Unable to fetch resources from planet with invalid identifier")
	}

	// We need to update the resources existing on the planet
	// before fetching the data.
	err := p.updateResourcesFromProduction(planet.ID)
	if err != nil {
		return fmt.Errorf("Could not update resources production for planet \"%s\" (err: %v)", planet.ID, err)
	}

	query := fmt.Sprintf("select res, amount, production, storage_capacity from planets_resources where planet='%s'", planet.ID)
	rows, err := p.dbase.DBQuery(query)

	if err != nil {
		return fmt.Errorf("Could not fetch resources for planet \"%s\" (err: %v)", planet.ID, err)
	}

	planet.Resources = make([]Resource, 0)
	var resource Resource

	for rows.Next() {
		err = rows.Scan(
			&resource.ID,
			&resource.Amount,
			&resource.Production,
			&resource.Storage,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve resource for planet \"%s\" (err: %v)", planet.ID, err))
			continue
		}

		planet.Resources = append(planet.Resources, resource)
	}

	return nil
}

// fetchPlanetBuildings :
// Used to fetch the buildings currently present on a planet.
// In addition to fetching the data this method will also set
// the upgrade actions for the planet and execute the upgrades
// if needed.
//
// The `planet` defines the planet for which buildings should
// be fetched. An invalid value will return an error.
//
// Returns any error that happended while fetching buildings.
func (p *PlanetProxy) fetchPlanetBuildings(planet *Planet) error {
	// Check consistency.
	if planet == nil || planet.ID == "" {
		return fmt.Errorf("Unable to fetch buildings from planet with invalid identifier")
	}

	planet.Buildings = make([]Building, 0)

	// We need to update the construction actions for buildings
	// first.
	err := p.updateBuildingsConstructionActions(planet.ID)
	if err != nil {
		return fmt.Errorf("Could not update buildings upgrade actions for planet \"%s\" (err: %v)", planet.ID, err)
	}

	query := fmt.Sprintf("select building, level from planets_buildings where planet='%s'", planet.ID)
	rows, err := p.dbase.DBQuery(query)

	if err != nil {
		return fmt.Errorf("Could not fetch buildings for planet \"%s\" (err: %v)", planet.ID, err)
	}

	var building Building

	for rows.Next() {
		err = rows.Scan(
			&building.ID,
			&building.Level,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve building for planet \"%s\" (err: %v)", planet.ID, err))
			continue
		}

		// Update the costs for this building.
		err = p.updateBuildingCosts(&building)
		if err != nil {
			building.Cost = make([]ResourceAmount, 0)

			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve costs for building \"%s\" for planet \"%s\" (err: %v)", building.ID, planet.ID, err))
		}

		// Update the production for this building.
		err = p.updateBuildingProduction(&building, planet)
		if err != nil {
			building.Production = make([]ResourceAmount, 0)
			building.ProductionIncrease = make([]ResourceAmount, 0)

			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve production for building \"%s\" for planet \"%s\" (err: %v)", building.ID, planet.ID, err))
		}

		planet.Buildings = append(planet.Buildings, building)
	}

	return nil
}

// fetchPlanetShips :
// Fills a similar role to `fetchPlanetBuildings` but handles
// the ships associated to a planet. Note that this does not
// handle the ships currently directed towards the planet but
// which do not have reached it yet.
//
// The `planet` defines the planet for which ships should be
// fetched. An invalid value will return an error.
//
// Returns any error that happended while fetching ships.
func (p *PlanetProxy) fetchPlanetShips(planet *Planet) error {
	// Check consistency.
	if planet == nil || planet.ID == "" {
		return fmt.Errorf("Unable to fetch ships from planet with invalid identifier")
	}

	planet.Ships = make([]Ship, 0)

	// We need to update the construction actions for defenses
	// first.
	err := p.updateShipsConstructionActions(planet.ID)
	if err != nil {
		return fmt.Errorf("Could not update ships construction actions for planet \"%s\" (err: %v)", planet.ID, err)
	}

	query := fmt.Sprintf("select ship, count from planets_ships where planet='%s'", planet.ID)
	rows, err := p.dbase.DBQuery(query)

	if err != nil {
		return fmt.Errorf("Could not fetch ships for planet \"%s\" (err: %v)", planet.ID, err)
	}

	var ship Ship

	for rows.Next() {
		err = rows.Scan(
			&ship.ID,
			&ship.Count,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve ship for planet \"%s\" (err: %v)", planet.ID, err))
			continue
		}

		planet.Ships = append(planet.Ships, ship)
	}

	return nil
}

// fetchPlanetDefenses :
// Fills a similar role to `fetchPlanetBuildings` but handles
// the defenses associated to a planet. Note that this does
// not handle the defenses that are currently being built but
// nontetheless provide an update of the current construction
// actions running on the planet so that the count is as close
// as possible from the current situation.
//
// The `planet` defines the planet for which defenses should
// be fetched. An invalid value will return an error.
//
// Returns any error that happended while fetching defenses.
func (p *PlanetProxy) fetchPlanetDefenses(planet *Planet) error {
	// Check consistency.
	if planet == nil || planet.ID == "" {
		return fmt.Errorf("Unable to fetch defenses from planet with invalid identifier")
	}

	planet.Defenses = make([]Defense, 0)

	// We need to update the construction actions for ships first.
	err := p.updateDefensesConstructionActions(planet.ID)
	if err != nil {
		return fmt.Errorf("Could not update defenses construction actions for planet \"%s\" (err: %v)", planet.ID, err)
	}

	// Fetch the data once it is up-to-date.
	query := fmt.Sprintf("select defense, count from planets_defenses where planet='%s'", planet.ID)
	rows, err := p.dbase.DBQuery(query)

	if err != nil {
		return fmt.Errorf("Could not fetch defenses for planet \"%s\" (err: %v)", planet.ID, err)
	}

	var defense Defense

	for rows.Next() {
		err = rows.Scan(
			&defense.ID,
			&defense.Count,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve defense for planet \"%s\" (err: %v)", planet.ID, err))
			continue
		}

		planet.Defenses = append(planet.Defenses, defense)
	}

	return nil
}

// updateResourcesFromProduction :
// Performs the update of the resources currently accumulated
// on a planet since the last update. This include running some
// scripts which will automatically compute the amount produced
// during the elapsed period.
// We must also take care of actions that potentially changed
// the amount of resources existing such as fleets interactions.
//
// The `planet` defines the planet for which resources should
// be updated since the last actualization.
//
// Returns any error generated by the process.
func (p *PlanetProxy) updateResourcesFromProduction(planet string) error {
	query := fmt.Sprintf("SELECT update_resources_for_planet('%s')", planet)

	return performWithLock(planet, p.dbase, query, p.lock)
}

// updateBuildingsConstructionActions :
// Performs the update of the building construction actions
// related to a specific planet provided in input. It will
// analyze the actions registered for the planet (at most
// a single one) and perform the update if the action has
// finished already.
// This comes handy to make sure that the data always is
// an accurate reflection of the buildings present on each
// planet.
//
// The `planet` defines the identifier of the planet for
// which the buildings upgrade actions should be executed.
//
// Returns any error that may have occurred while updating.
func (p *PlanetProxy) updateBuildingsConstructionActions(planet string) error {
	query := fmt.Sprintf("SELECT update_building_upgrade_action('%s')", planet)

	return performWithLock(planet, p.dbase, query, p.lock)
}

// updateBuildingCosts :
// Used to perform the computation of the costs for the
// next level of the building described in argument. The
// output values will be saved directly in the input
// object.
//
// The `building` defines the object for which the costs
// should be computed. A `nil` value will raise an error.
//
// Returns any error.
func (p *PlanetProxy) updateBuildingCosts(building *Building) error {
	// Check consistency.
	if building == nil || building.ID == "" {
		return fmt.Errorf("Cannot update building costs from invalid building")
	}

	// In case the costs for building are not populated try
	// to update it.
	if len(p.buildingCosts) == 0 {
		costs, err := initProgressCostsFromDB(p.dbase, p.log, "building", "buildings_costs_progress", "buildings_costs")
		if err != nil {
			return fmt.Errorf("Unable to generate buildings costs for building \"%s\", none defined", building.ID)
		}

		p.buildingCosts = costs
	}

	// Find the building in the costs table.
	info, ok := p.buildingCosts[building.ID]
	if !ok {
		return fmt.Errorf("Could not compute costs for unknown building \"%s\"", building.ID)
	}

	// Compute the cost for each resource.
	building.Cost = info.ComputeCosts(building.Level)

	return nil
}

// updateBuildingProduction :
// Used to perform the computation of the production for
// the current level of the building. We will also update
// the production increase for the next level. The output
// values will be saved directly in the input building.
//
// The `building` defines the object for which the prod
// should be computed. A `nil` value will raise an error.
//
// The `planet` is used to provide information relative
// to the temperature on the place of production as some
// buildings are dependent on the temperature to provide
// the final amount produced.
//
// Returns any error.
func (p *PlanetProxy) updateBuildingProduction(building *Building, planet *Planet) error {
	// Check consistency.
	if building == nil || building.ID == "" {
		return fmt.Errorf("Cannot update building production from invalid building")
	}

	// In case the production rules for buildings are not
	// populated try to update it.
	if len(p.prodRules) == 0 {
		rules, err := initBuildingsProductionRulesFromDB(p.dbase, p.log)
		if err != nil {
			return fmt.Errorf("Unable to generate buildings production rules for building \"%s\", none defined", building.ID)
		}

		p.prodRules = rules
	}

	// Find the building in the production table.
	rules, ok := p.prodRules[building.ID]
	if !ok {
		// The building does not seem to produce any resource. That
		// or the rules do not have been updated but we can't do
		// much about this anyways so we can safely return an empty
		// production for this building.
		building.Production = make([]ResourceAmount, 0)
		building.ProductionIncrease = make([]ResourceAmount, 0)

		return nil
	}

	// Compute the production and production increase for
	// each resource this building is associated to.
	building.Production = make([]ResourceAmount, 0)
	building.ProductionIncrease = make([]ResourceAmount, 0)

	for _, rule := range rules {
		prodCurLevel := rule.ComputeProduction(building.Level, planet.averageTemp())
		prodNextLevel := rule.ComputeProduction(building.Level+1, planet.averageTemp())

		prodIncrease := ResourceAmount{
			Resource: rule.Resource,
			Amount:   prodNextLevel.Amount - prodCurLevel.Amount,
		}

		building.Production = append(building.Production, prodCurLevel)
		building.ProductionIncrease = append(building.ProductionIncrease, prodIncrease)
	}

	return nil
}

// updateShipsConstructionActions :
// Similar to the `updateBuildingsConstructionActions` but
// performs the update of the ships construction actions.
// We will follow a very similar process where the actions
// are analyzed to update the ships count on the planet and
// removed if necessary.
//
// The `planet` defines the identifier of the planet for
// which the ships construction actions should be executed.
//
// Returns any error that may have occurred while updating.
func (p *PlanetProxy) updateShipsConstructionActions(planet string) error {
	query := fmt.Sprintf("SELECT update_ship_upgrade_action('%s')", planet)

	return performWithLock(planet, p.dbase, query, p.lock)
}

// updateDefensesConstructionActions :
// Similar to the `updateBuildingsConstructionActions` but
// performs the update of the defenses construction actions.
// The defenses count will be updated and actions removed if
// all their related defenses have been built.
//
// The `planet` defines the identifier of the planet for
// which the defenses construction actions should be executed.
//
// Returns any error that may have occurred while updating.
func (p *PlanetProxy) updateDefensesConstructionActions(planet string) error {
	query := fmt.Sprintf("SELECT update_defense_upgrade_action('%s')", planet)

	return performWithLock(planet, p.dbase, query, p.lock)
}

// CreateFor :
// Used to handle the creation of a planet for the specified
// player at the input coordinate. In case the coordinates are
// `nil` we will assume that we are creating the homeworld for
// the player and thus we can choose the coordinates randomly.
// Otherwise we will try to create the planet at the specified
// coordinates and fail if the coordinates are not available.
// The universe to create the planet in is described by the
// `UniverseID` of the player.
//
// The `player` represents the account for which the planet is
// to be created. We assume that the universe and the player's
// identifiers are valid (otherwise we won't be able to attach
// the planet to a valid account).
//
// The `coord` represents the desired coordinates where the
// planet should be created. In case this value is `nil` we
// assume that the homeworld of the player should be created
// and thus we will choose randomly some coordinates among
// the available locations.
//
// Returns any error arised during the creation process.
func (p *PlanetProxy) CreateFor(player Player, coord *Coordinate) error {
	// Check consistency.
	if player.ID == "" {
		return fmt.Errorf("Cannot create planet for invalid player")
	}
	if player.UniverseID == "" {
		return fmt.Errorf("Cannot create planet for player \"%s\" in invalid universe", player.ID)
	}

	// First we need to fetch the universe related to the
	// planet to create.
	uni, err := p.fetchUniverse(player.UniverseID)
	if err != nil {
		return fmt.Errorf("Could not create planet for \"%s\" (err: %v)", player.ID, err)
	}

	// Create the planet from the available data.
	planet, err := p.generatePlanet(player.ID, coord, uni)
	if err != nil {
		return fmt.Errorf("Could not create planet for \"%s\" (err: %v)", player.ID, err)
	}

	// We will now try to insert the planet into the DB if
	// we have valid coordinates. Note that the process is
	// quite different depending on whether we have a list
	// of coordinates to pick from or a single one. List
	// can occurs when we want to create the homeworld for
	// a player, in which case we want to select a random
	// location to insert it whereas when the coordinate
	// is provided by the user we want to create a planet
	// at *this* location and no where else.
	if coord != nil {
		return p.createPlanet(planet)
	}

	// Retrieve the list of coordinates that are already
	// used in the universe the player's in.
	usedCoords, err := p.generateUsedCoords(uni)
	totalPlanets := uni.GalaxiesCount * uni.GalaxySize * uni.SolarSystemSize

	// Try to insert the planet in the DB while we have some
	// untested coordinates and we didn't suceed in inserting
	// it.
	inserted := false
	trials := 0

	for !inserted && len(usedCoords) < totalPlanets && trials < 10 {
		// Pick a random coordinate and check whether it belongs
		// to the already used coordinates. If this is the case
		// we will try to pick a new one. Otherwise we will try
		// to perform the insertion of the planet in the DB.
		// In case the insertion fails we will add the selected
		// coordinates to the list of used one so as not to try
		// to use it again.
		coord := Coordinate{
			rand.Int() % uni.GalaxiesCount,
			rand.Int() % uni.GalaxySize,
			rand.Int() % uni.SolarSystemSize,
		}

		exists := true
		for exists {
			key := coord.Linearize(uni)

			if _, ok := usedCoords[key]; !ok {
				// We found a not yet used coordinate.
				exists = false
			} else {
				// Pick some new coordinates.
				coord = Coordinate{
					rand.Int() % uni.GalaxiesCount,
					rand.Int() % uni.GalaxySize,
					rand.Int() % uni.SolarSystemSize,
				}
			}
		}

		planet.Galaxy = coord.Galaxy
		planet.System = coord.System
		planet.Position = coord.Position

		// Whenever we update the coordinates of the planet we
		// need to generate new temperature and size.
		p.generatePlanetSize(&planet)

		// Try to create the planet at the specified coordinates.
		err = p.createPlanet(planet)

		// Check for errors.
		if err == nil {
			p.log.Trace(logger.Notice, fmt.Sprintf("Created planet at %v for \"%s\" in \"%s\" with %d field(s)", coord, player.ID, player.UniverseID, planet.Fields))
			inserted = true
		} else {
			p.log.Trace(logger.Warning, fmt.Sprintf("Could not import planet \"%s\" for \"%s\" (err: %v)", planet.Name, player.ID, err))

			// Register this coordinate as being used as we can't
			// successfully use it to create the planet anyways.
			usedCoords[coord.Linearize(uni)] = coord
		}

		trials++
	}

	// Check whether we could insert the element in the DB: if
	// this is not the case we couldn't create the planet.
	if !inserted {
		return fmt.Errorf("Could not insert planet for player \"%s\" in DB after %d trial(s)", player.ID, trials)
	}

	return nil
}

// fetchUniverse :
// Used to fetch the universe from the DB with an identifier
// matching the input one. If no such universe can be fetched
// an error is returned.
//
// The `id` defines the index of the universe to fetch.
//
// Returns the universe corresponding to the input identifier
// along with any errors.
func (p *PlanetProxy) fetchUniverse(id string) (Universe, error) {
	// Create the db filters from the input identifier.
	filters := make([]DBFilter, 1)

	filters[0] = DBFilter{
		"id",
		[]string{id},
	}

	unis, err := p.uniProxy.Universes(filters)

	// Check for errors and cases where we retrieve several
	// universes.
	if err != nil {
		return Universe{}, err
	}
	if len(unis) > 1 {
		err = fmt.Errorf("Retrieved %d universes for id \"%s\"", len(unis), id)
	}

	return unis[0], err
}

// generateUsedCoords :
// Used to find and generate a list of the used coordinates
// in the corresponding universe. Note that the list is only
// some snapshot of the state of the coordinates which can
// evolve through time. Typically if some pending requests to
// insert a planet are pending or some actions require some
// action to create/destroy a planet this list will be changed
// and might not be accurate.
// We figure it's not really a problem to insert elements in
// the DB as it's unlikely to ever failed a lot of times in
// a row. What can maybe happen is that the first try fails
// to insert a planet but the second one with a different set
// of coordinates it will most likely succeed.
//
// The `uni` defines the universe for which available coords
// should be fetched. This will be fetched from the DB.
//
// The return value includes all the user coordinates in the
// universe along with any errors.
func (p *PlanetProxy) generateUsedCoords(uni Universe) (map[int]Coordinate, error) {
	// Create the query allowing to fetch all the planets of
	// a specific universe. This will consistute the list of
	// used planets for this universe.
	props := []string{
		"p.galaxy",
		"p.solar_system",
		"p.position",
	}

	table := "planets p inner join players pl"
	joinCond := "p.player=pl.id"
	where := fmt.Sprintf("pl.uni='%s'", uni.ID)

	query := fmt.Sprintf("select %s from %s on %s where %s", strings.Join(props, ", "), table, joinCond, where)

	// Execute the query and check for errors.
	rows, err := p.dbase.DBQuery(query)
	if err != nil {
		return nil, fmt.Errorf("Could not fetch used coordinates for universe \"%s\" (err: %v)", uni.ID, err)
	}

	// Traverse all the coordinates and populate the list.
	coords := make(map[int]Coordinate)
	var coord Coordinate

	for rows.Next() {
		err = rows.Scan(
			&coord.Galaxy,
			&coord.System,
			&coord.Position,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for coordinate in universe \"%s\" (err: %v)", uni.ID, err))
			continue
		}

		key := coord.Linearize(uni)

		// Check whether it's the first time we encounter
		// this used location.
		if _, ok := coords[key]; ok {
			p.log.Trace(logger.Error, fmt.Sprintf("Overriding used coordinate %v in universe \"%s\"", coord, uni.ID))
		}

		coords[key] = coord
	}

	return coords, nil
}

// createPlanet :
// Used to attempt to create the planet with the specified
// data. We only try to perform the insertion of the input
// data into the DB and report in error if it cannot be
// performed for some reasons.
//
// The `planet` defines the data to insert in the DB.
//
// Returns an error in case the planet could not be used
// in the DB (for example because the coordinates already
// are used).
func (p *PlanetProxy) createPlanet(planet Planet) error {
	// Marshal the planet.
	data, err := json.Marshal(planet)
	if err != nil {
		return fmt.Errorf("Could not import planet \"%s\" for \"%s\" in DB (err: %v)", planet.Name, planet.PlayerID, err)
	}
	jsonForPlanet := string(data)

	// Marshal the resources for this planet.
	data, err = json.Marshal(planet.Resources)
	if err != nil {
		return fmt.Errorf("Could not import planet \"%s\" for \"%s\" in DB (err: %v)", planet.Name, planet.PlayerID, err)
	}
	jsonForResources := string(data)

	query := fmt.Sprintf("select * from create_planet('%s', '%s')", jsonForPlanet, jsonForResources)
	_, err = p.dbase.DBExecute(query)

	// Check for errors.
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Could not import planet \"%s\" (err: %v)", planet.Name, err))
	}

	return nil
}

// generatePlanet :
// Used to perform the generation of the properties of a planet
// based on the input player and coordinates. All the info to
// actually define the planet will be generated including the
// resources.
// The universe in which the planet should be provided in case
// the coordinates are to be determined by this function. As we
// will need information as to the positions that are possible
// for a planet in this case.
//
// The `player` defines the identifier of the player to which
// the planet belongs.
//
// The `coord` defines the coordinates of the planet to create.
// If the value is `nil` no data is generated.
//
// The `uni` argument defines the universe in which the planet
// is to be created. This helps with defining valid coordinates
// in case none are provided.
//
// Returns the created planet.
func (p *PlanetProxy) generatePlanet(player string, coord *Coordinate, uni Universe) (Planet, error) {
	trueCoords := Coordinate{0, 0, 0}
	if coord != nil {
		trueCoords = *coord
	}

	// Create the planet and generate base information.
	planet := Planet{
		player,
		uuid.New().String(),
		trueCoords.Galaxy,
		trueCoords.System,
		trueCoords.Position,
		getDefaultPlanetName(coord == nil),
		0,
		0,
		0,
		0,
		make([]Resource, 0),
		make([]Building, 0),
		make([]Ship, 0),
		make([]Defense, 0),
	}

	err := p.generateResources(&planet)

	return planet, err
}

// generatePlanetSize :
// Used to generate the size associated to a planet. The size
// is a general notion including both its actual diameter and
// also the temperature on the surface of the planet. Both
// values depend on the actual position of the planet in the
// parent solar system.
//
// The `planet` defines the planet for which the size should
// be generated.
func (p *PlanetProxy) generatePlanetSize(planet *Planet) {
	// Check whether the planet and its coordinates are valid.
	if planet == nil {
		return
	}

	// Create a random source to be used for the generation of
	// the planet's properties. We will use a procedural algo
	// which will be based on the position of the planet in its
	// parent universe.
	source := rand.NewSource(
		int64(
			NewCoordinate(
				planet.Galaxy,
				planet.System,
				planet.Position,
			).generateSeed(),
		),
	)
	rng := rand.New(source)

	// The table of the dimensions of the planet are inspired
	// from this link:
	// https://ogame.fandom.com/wiki/Colonizing_in_Redesigned_Universes
	var min int
	var max int
	var stdDev int

	switch planet.Position {
	case 0:
		// Range [96; 172], average 134.
		min = 96
		max = 172
		stdDev = max - min
	case 1:
		// Range [104; 176], average 140.
		min = 104
		max = 176
		stdDev = max - min
	case 2:
		// Range [112; 182], average 147.
		min = 112
		max = 182
		stdDev = max - min
	case 3:
		// Range [118; 208], average 163.
		min = 118
		max = 208
		stdDev = max - min
	case 4:
		// Range [133; 232], average 182.
		min = 133
		max = 232
		stdDev = max - min
	case 5:
		// Range [152; 248], average 200.
		min = 152
		max = 248
		stdDev = max - min
	case 6:
		// Range [156; 262], average 204.
		min = 156
		max = 262
		stdDev = max - min
	case 7:
		// Range [150; 246], average 198.
		min = 150
		max = 246
		stdDev = max - min
	case 8:
		// Range [142; 232], average 187.
		min = 142
		max = 232
		stdDev = max - min
	case 9:
		// Range [136; 210], average 173.
		min = 136
		max = 210
		stdDev = max - min
	case 10:
		// Range [125; 186], average 156.
		min = 125
		max = 186
		stdDev = max - min
	case 11:
		// Range [114; 172], average 143.
		min = 114
		max = 172
		stdDev = max - min
	case 12:
		// Range [100; 168], average 134.
		min = 100
		max = 168
		stdDev = max - min
	case 13:
		// Range [90; 164], average 127.
		min = 96
		max = 164
		stdDev = max - min
	case 14:
		fallthrough
	default:
		// Assume default case if the `15th` position
		// Range [90; 164], average 134.
		min = 90
		max = 164
		stdDev = max - min
	}

	mean := (max + min) / 2
	planet.Fields = mean + int(math.Round(rng.NormFloat64()*float64(stdDev)))

	// The diameter is derived from the fields count with a random part.
	planet.Diameter = 100*planet.Fields + int(math.Round(float64(100.0*rand.Float32())))

	// The temperatures are described in the following link:
	// https://ogame.fandom.com/wiki/Temperature
	switch planet.Position {
	case 0:
		// Range [220; 260], average 240.
		min = 220
		max = 260
		stdDev = max - min
	case 1:
		// Range [170; 210], average 190.
		min = 170
		max = 210
		stdDev = max - min
	case 2:
		// Range [120; 160], average 140.
		min = 120
		max = 160
		stdDev = max - min
	case 3:
		// Range [70; 110], average 90.
		min = 70
		max = 110
		stdDev = max - min
	case 4:
		// Range [60; 100], average 80.
		min = 60
		max = 100
		stdDev = max - min
	case 5:
		// Range [50; 90], average 70.
		min = 50
		max = 90
		stdDev = max - min
	case 6:
		// Range [40; 80], average 60.
		min = 40
		max = 80
		stdDev = max - min
	case 7:
		// Range [30; 70], average 50.
		min = 30
		max = 70
		stdDev = max - min
	case 8:
		// Range [20; 60], average 40.
		min = 20
		max = 60
		stdDev = max - min
	case 9:
		// Range [10; 50], average 30.
		min = 10
		max = 50
		stdDev = max - min
	case 10:
		// Range [0; 40], average 20.
		min = 0
		max = 40
		stdDev = max - min
	case 11:
		// Range [-10; 30], average 10.
		min = -10
		max = 30
		stdDev = max - min
	case 12:
		// Range [-50; -10], average -30.
		min = -50
		max = -10
		stdDev = max - min
	case 13:
		// Range [-90; -50], average -70.
		min = -90
		max = -50
		stdDev = max - min
	case 14:
		fallthrough
	default:
		// Assume default case if the `15th` position
		// Range [-130; -90], average -110.
		min = -130
		max = -90
		stdDev = max - min
	}

	mean = (max + min) / 2
	planet.MaxTemp = mean + int(math.Round(rng.NormFloat64()*float64(stdDev)))
	planet.MinTemp = planet.MaxTemp - getPlanetTemperatureAmplitude()
}

// generateResources :
// Used to perform the creation of the default resources for
// a planet when it's being created. This translate the fact
// that a planet is never really `empty` in the game.
// This function will create all the necessary entries in the
// planet input object but does not create anything in the DB.
//
// The `planet` defines the planet for which resources should
// be generated.
//
// Returns any error.
func (p *PlanetProxy) generateResources(planet *Planet) error {
	// Discard empty planets.
	if planet == nil {
		return fmt.Errorf("Unable to generate resources for invalid planet")
	}

	// Prevent creation of planets in case no resources are
	// available (because none have been retrieved from the
	// DB).
	// If this is the case we will first attempt to fetch
	// the resources from the DB and return an error is it
	// fails.
	if len(p.resources) == 0 {
		resources, err := initResourcesFromDB(p.dbase, p.log)
		if err != nil {
			return fmt.Errorf("Unable to generate resources for planet, none defined")
		}

		p.resources = resources
	}

	// We will consider that we have a certain number of each
	// resources readily available on each planet upon its
	// creation. The values are hard-coded there and we use
	// the identifier of the resources retrieved from the DB
	// to populate the planet.
	if planet.Resources == nil {
		planet.Resources = make([]Resource, 0)
	}

	for _, res := range p.resources {
		desc := Resource{
			ID:         res.ID,
			Planet:     planet.ID,
			Amount:     float32(res.BaseAmount),
			Production: float32(res.BaseProd),
			Storage:    float32(res.BaseStorage),
		}

		planet.Resources = append(planet.Resources, desc)
	}

	return nil
}
