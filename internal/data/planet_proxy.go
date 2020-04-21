package data

import (
	"fmt"
	"math/rand"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// PlanetProxy :
// Intended as a wrapper to access properties of planets
// and retrieve data from the database. In most cases we
// need to access some properties of the planets for a
// given identifier. A planet is a set of resources, a
// set of ships and defenses that are currently built on
// it and some upgrade actions (meaning that a building
// will have one more level soon or that there will be
// more ships deployed on this planet).
type PlanetProxy struct {
	commonProxy
}

// NewPlanetProxy :
// Create a new proxy allowing to serve the requests
// related to planets.
//
// The `dbase` represents the database to use to fetch
// data related to planets.
//
// The `data` defines the data model to use to fetch
// information and verify actions.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewPlanetProxy(dbase *db.DB, data model.Instance, log logger.Logger) PlanetProxy {
	return PlanetProxy{
		commonProxy: newCommonProxy(dbase, data, log, "planets"),
	}
}

// buildQuery :
// Used to assemble a query description struct from
// the input properties.
//
// The `props` define the properties that should be
// Used for the query.
//
// The `table` defines the table in which the query
// should be executed.
//
// The `filterName` defines the name of the column
// to filter.
//
// The `filter` defines the value which should be
// kept in the `filterName` column.
//
// Returns the description of the query built from
// the input properties.
func (p *PlanetProxy) buildQuery(props []string, table string, filterName string, filter string) db.QueryDesc {
	return db.QueryDesc{
		Props: props,
		Table: table,
		Filters: []db.Filter{
			{
				Key:    filterName,
				Values: []string{filter},
			},
		},
	}
}

// Planets :
// Return a list of planets registered so far in all the planets
// defined in the DB. The input filters might help to narrow the
// search a bit by providing coordinates to look for and a uni
// to look into.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// accounts available. Each one is appended `as-is` to the
// query.
//
// Returns the list of planets registered in the DB and matching
// the input list of filters. In case the error is not `nil` the
// value of the array should be ignored.
func (p *PlanetProxy) Planets(filters []db.Filter) ([]model.Planet, error) {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"p.id",
		},
		Table:   "planets p inner join players pl on p.player=pl.id",
		Filters: filters,
	}

	res, err := p.proxy.FetchFromDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not query DB to fetch planets (err: %v)", err))
		return []model.Planet{}, err
	}

	// We now need to retrieve all the identifiers that matched
	// the input filters and then build the corresponding planets
	// object for each one of them.
	var ID string
	IDs := make([]string, 0)

	for res.Next() {
		err = res.Scan(&ID)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Error while fetching planet ID (err: %v)", err))
			continue
		}

		IDs = append(IDs, ID)
	}

	planets := make([]model.Planet, 0)

	for _, ID = range IDs {
		pla, err := model.NewPlanetFromDB(ID, p.proxy)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Unable to fetch planet \"%s\" data from DB (err: %v)", ID, err))
			continue
		}

		planets = append(planets, pla)
	}

	return planets, nil
}

// generateResources :
// Used to perform the creation of the default resources for
// a planet when it's being created. This translate the fact
// that a planet is never really `empty` in the game.
// This function will create all the necessary entries in the
// planet input object but does not create anything in the DB.
//
// The `planet` defines the element for which resources need
// to be generated.
//
// Returns any error.
func (p *PlanetProxy) generateResources(planet *model.Planet) error {
	// A planet always has the base amount defined in the DB.
	resources, err := p.data.Resources.Resources(p.proxy, []db.Filter{})
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Unable to generate resources for planet (err: %v)", err))
		return err
	}

	// We will consider that we have a certain number of each
	// resources readily available on each planet upon its
	// creation. The values are hard-coded there and we use
	// the identifier of the resources retrieved from the DB
	// to populate the planet.
	if planet.Resources == nil {
		planet.Resources = make([]model.ResourceInfo, 0)
	}

	for _, res := range resources {
		desc := model.ResourceInfo{
			Resource:   res.ID,
			Amount:     res.BaseAmount,
			Storage:    res.BaseStorage,
			Production: res.BaseProd,
		}

		planet.Resources = append(planet.Resources, desc)
	}

	return nil
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

	// Update upgrade actions.
	err := p.updateConstructionActions(planet.ID)
	if err != nil {
		return fmt.Errorf("Unable to update upgrade actions for planet \"%s\" (err: %v)", planet.ID, err)
	}

	// Fetch resources.
	err = p.fetchPlanetResources(planet)
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

// updateConstructionActions :
// Used to perform the update of the construction actions for
// the planet described by the input identifier. It will call
// the corresponding DB script to get up-to-date values for
// the planet.
//
// The `planetID` defines the identifier of the planet which
// should be updated.
//
// Returns any error that occurred during the update.
func (p *PlanetProxy) updateConstructionActions(planetID string) error {
	// Update resources.
	query := fmt.Sprintf("SELECT update_resources_for_planet('%s')", planetID)
	err := p.performWithLock(planetID, query)
	if err != nil {
		return fmt.Errorf("Could not update resources for \"%s\" (err: %v)", planetID, err)
	}

	query = fmt.Sprintf("SELECT update_building_upgrade_action('%s')", planetID)
	err = p.performWithLock(planetID, query)
	if err != nil {
		return fmt.Errorf("Could not update buildings upgrade actions for \"%s\" (err: %v)", planetID, err)
	}

	query = fmt.Sprintf("SELECT update_ship_upgrade_action('%s')", planetID)
	err = p.performWithLock(planetID, query)
	if err != nil {
		return fmt.Errorf("Could not update ships upgrade actions for \"%s\" (err: %v)", planetID, err)
	}

	query = fmt.Sprintf("SELECT update_defense_upgrade_action('%s')", planetID)
	err = p.performWithLock(planetID, query)
	if err != nil {
		return fmt.Errorf("Could not update defenses upgrade actions for \"%s\" (err: %v)", planetID, err)
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

	planet.Resources = make([]Resource, 0)

	// Create the query and execute it.
	query := p.buildQuery(
		[]string{
			"res",
			"amount",
			"production",
			"storage_capacity",
		},
		"planets_resources",
		"planet",
		planet.ID,
	)

	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return fmt.Errorf("Could not query DB to fetch resources for planet \"%s\" (err: %v)", planet.ID, err)
	}

	// Populate the return value.
	var resource Resource

	for res.next() {
		err = res.scan(
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

	// Create the query and execute it.
	query := p.buildQuery(
		[]string{
			"building",
			"level",
		},
		"planets_buildings",
		"planet",
		planet.ID,
	)

	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return fmt.Errorf("Could not query DB to fetch buildings for planet \"%s\" (err: %v)", planet.ID, err)
	}

	// Populate the return value.
	var building Building

	for res.next() {
		err = res.scan(
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

	// Create the query and execute it.
	query := p.buildQuery(
		[]string{
			"ship",
			"count",
		},
		"planets_ships",
		"planet",
		planet.ID,
	)

	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return fmt.Errorf("Could not query DB to fetch ships for planet \"%s\" (err: %v)", planet.ID, err)
	}

	// Populate the return value.
	var ship Ship

	for res.next() {
		err = res.scan(
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

	// Create the query and execute it.
	query := p.buildQuery(
		[]string{
			"defense",
			"count",
		},
		"planets_defenses",
		"planet",
		planet.ID,
	)

	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return fmt.Errorf("Could not query DB to fetch defenses for planet \"%s\" (err: %v)", planet.ID, err)
	}

	// Populate the return value.
	var defense Defense

	for res.next() {
		err = res.scan(
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
	if len(p.bCosts) == 0 {
		err := p.init()
		if err != nil {
			return fmt.Errorf("Unable to generate buildings costs for building \"%s\", none defined", building.ID)
		}
	}

	// Find the building in the costs table.
	info, ok := p.bCosts[building.ID]
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
	if len(p.pRules) == 0 {
		err := p.init()
		if err != nil {
			return fmt.Errorf("Unable to generate buildings production rules for building \"%s\", none defined", building.ID)
		}
	}

	// Find the building in the production table.
	rules, ok := p.pRules[building.ID]
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
	query := p.buildQuery(
		[]string{
			"p.galaxy",
			"p.solar_system",
			"p.position",
		},
		"planets p inner join players pl on p.player=pl.id",
		"pl.uni",
		uni.ID,
	)

	// Create the query and execute it.
	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return map[int]Coordinate{}, fmt.Errorf("Could not fetch used coordinates for universe \"%s\" (err: %v)", uni.ID, err)
	}

	// Traverse all the coordinates and populate the list.
	coords := make(map[int]Coordinate)
	var coord Coordinate

	for res.next() {
		err = res.scan(
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
func (p *PlanetProxy) createPlanet(planet *model.Planet) error {
	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_planet",
		Args: []interface{}{
			planet,
			planet.Resources,
		},
	}

	err := p.proxy.InsertToDB(query)

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not import planet \"%s\" (err: %v)", planet.Name, err))
	}

	return err
}
