package data

import (
	"encoding/json"
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"

	"github.com/google/uuid"
)

// FleetProxy :
// Intended as a wrapper to access properties of fleets and
// retrieve data from the database. This helps hiding the
// complexity of how the data is laid out in the `DB` and
// the precise name of tables from the exterior world.
//
// The `dbase` is the database that is wrapped by this
// object. It is checked for consistency upon creating the
// wrapper.
//
// The `log` allows to perform display to the user so as
// to inform of potential issues and debug information to
// the outside world.
//
// The `uniProxy` defines a proxy allowing to access a
// part of the behavior related to universes typically
// when fetching a universe into which a fleet should
// be created.
type FleetProxy struct {
	dbase    *db.DB
	log      logger.Logger
	uniProxy UniverseProxy
}

// NewFleetProxy :
// Create a new proxy on the input `dbase` to access the
// properties of fleets as registered in the DB.
// In case the provided DB is `nil` a panic is issued.
//
// The `dbase` represents the database to use to fetch
// data related to fleets.
//
// The `log` will be used to notify information so that
// we can have an idea of the activity of this component.
// One possible example is for timing the requests.
//
// The `uniProxy` defines a proxy that can be used to
// fetch information about the universes when creating
// fleets.
//
// Returns the created proxy.
func NewFleetProxy(dbase *db.DB, log logger.Logger, unis UniverseProxy) FleetProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create fleets proxy from invalid DB"))
	}

	return FleetProxy{
		dbase,
		log,
		unis,
	}
}

// Fleets :
// Allows to fetch the list of fleets currently registered in
// the server
// The user can choose to filter parts of the fleets using an
// array of filters that will be applied to the SQL query.
// No controls is enforced on the filters so one should make
// sure that it's consistent with the underlying table.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// fleets available. Each one is appended `as-is` to the SQL
// query.
//
// Returns the list of fleets along with any errors. Note that
// in case the error is not `nil` the returned list is to be
// ignored.
func (p *FleetProxy) Fleets(filters []DBFilter) ([]Fleet, error) {
	// Create the query and execute it.
	props := []string{
		"id",
		"name",
		"uni",
		"objective",
		"arrival_time",
		"target_galaxy",
		"target_solar_system",
		"target_position",
		"arrival_time",
	}

	table := "fleets"

	query := fmt.Sprintf("select %s from %s", strings.Join(props, ", "), table)

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
		return nil, fmt.Errorf("Could not query DB to fetch fleets (err: %v)", err)
	}

	// Populate the return value: we should obtain a single
	// result, otherwise it's an issue.
	fleets := make([]Fleet, 0)
	var fleet Fleet

	for rows.Next() {
		err = rows.Scan(
			&fleet.ID,
			&fleet.Name,
			&fleet.UniverseID,
			&fleet.Objective,
			// TODO: This may fail, see when we have a real fleet.
			&fleet.ArrivalTime,
			&fleet.Galaxy,
			&fleet.System,
			&fleet.Position,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for fleet (err: %v)", err))
			continue
		}

		fleets = append(fleets, fleet)
	}

	return fleets, nil
}

// FleetComponents :
// Used to fetch data related to a fleet: this includes the
// individual components of the fleet (which is mostly used
// in the case of group attacks).
//
// The `filters` define some properties that are used when
// fetching the parent fleet for which components should be
// retrieved. It is assumed that these filteres allow to
// fetch a single fleet. If this is not the case an error
// is returned.
//
// Returns the components associated to the fleet along with
// any error.
func (p *FleetProxy) FleetComponents(filters []DBFilter) ([]FleetComponent, error) {
	// Fetch the fleet from the filters.
	fleets, err := p.Fleets(filters)
	if err != nil {
		return nil, fmt.Errorf("Could not fetch component for fleet (err: %v)", err)
	}

	// Check that we only found a single fleet matching the
	// input filters: this will be the fleet for which the
	// components should be retrieved.
	if len(fleets) != 1 {
		return nil, fmt.Errorf("Found %d fleet(s) matching filters, cannot fetch components", len(fleets))
	}

	fleet := fleets[0]

	// Create the query to fetch individual components of the
	// fleet and execute it.
	props := []string{
		"id",
		"player",
		"start_galaxy",
		"start_solar_system",
		"start_position",
		"speed",
		"joined_at",
	}

	table := "fleet_elements"
	where := fmt.Sprintf("fleet='%s'", fleet.ID)

	query := fmt.Sprintf("select %s from %s where %s", strings.Join(props, ", "), table, where)
	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch fleet \"%s\" details (err: %v)", fleet.ID, err)
	}

	// Populate the return content from the result of the
	// DB query.
	components := make([]FleetComponent, 0)
	comp := FleetComponent{
		FleetID: fleet.ID,
	}

	for rows.Next() {
		err = rows.Scan(
			&comp.ID,
			&comp.PlayerID,
			&comp.Galaxy,
			&comp.System,
			&comp.Position,
			&comp.Speed,
			&comp.JoinedAt,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for fleet \"%s\" component (err: %v)", fleet.ID, err))
			continue
		}

		err = p.fetchFleetComponentData(&comp)
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch data for fleet \"%s\" (component \"%s\" failed, err: %v)", fleet.ID, comp.PlayerID, err))
			continue
		}

		components = append(components, comp)
	}

	return components, nil
}

// fetchFleetComponentData :
// Used to fetch data related to a fleet component. It includes
// the actual ship involved in the component and their actual
// amount.
// individual components of the fleet (which is mostly used
// in the case of group attacks).
//
// The `com` references the fleet component for which the data
// should be fetched. We assume that the internal fields (and
// more specifically the identifier) are already populated.
//
// Returns any error.
func (p *FleetProxy) fetchFleetComponentData(comp *FleetComponent) error {
	// Check whether the fleet component has an identifier assigned.
	if comp.ID == "" {
		return fmt.Errorf("Unable to fetch data from fleet component with invalid identifier")
	}

	// Create the query to fetch individual components of the
	// fleet and execute it.
	props := []string{
		"ship",
		"amount",
	}

	table := "fleet_ships"
	where := fmt.Sprintf("fleet_element='%s'", comp.ID)

	query := fmt.Sprintf("select %s from %s where %s", strings.Join(props, ", "), table, where)
	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return fmt.Errorf("Could not query DB to fetch fleet component \"%s\" details (err: %v)", comp.ID, err)
	}

	// Populate the return content from the result of the
	// DB query.
	comp.Ships = make([]ShipInFleet, 0)
	var ship ShipInFleet

	for rows.Next() {
		err = rows.Scan(
			&ship.ShipID,
			&ship.Amount,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for fleet component \"%s\" (err: %v)", comp.ID, err))
			continue
		}

		comp.Ships = append(comp.Ships, ship)
	}

	return nil
}

// Create :
// Used to perform the creation of the fleet described
// by the input data to the DB. In case the creation can
// not be performed an error is returned.
//
// The `fleet` describes the element to create in DB. Its
// value may be modified by the function mainly to update
// the identifier of the fleet if none have been set.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`.
func (p *FleetProxy) Create(fleet *Fleet) error {
	// Assign a valid identifier if this is not already the case.
	if fleet.ID == "" {
		fleet.ID = uuid.New().String()
	}

	// Fetchthe universe related to the fleet to create.
	uni, err := p.fetchUniverse(fleet.UniverseID)
	if err != nil {
		return fmt.Errorf("Could not create fleet \"%s\", unable to fetch universe (err: %v)", fleet.ID, err)
	}

	// Validate that the input data describe a valid fleet.
	if !fleet.valid(uni) {
		return fmt.Errorf("Could not create fleet \"%s\", some properties are invalid", fleet.ID)
	}

	// Marshal the input fleet to pass it to the import script.
	data, err := json.Marshal(fleet)
	if err != nil {
		return fmt.Errorf("Could not import fleet \"%s\" (err: %v)", fleet.ID, err)
	}
	jsonToSend := string(data)

	query := fmt.Sprintf("select * from create_fleet('%s')", jsonToSend)
	_, err = p.dbase.DBExecute(query)

	// Check for errors.
	if err != nil {
		return fmt.Errorf("Could not import fleet \"%s\" (err: %s)", fleet.ID, err)
	}

	// All is well.
	return nil
}

// fetchUniverse :
// Used to fetch the universe from the DB. The input identifier
// is meant to represent a universe registered in the DB. We
// will make sure that it can be fetched and that a single item
// is matching in the DB.
// In case no universe can be found an error is returned.
//
// The `id` defines the index of the universe to fetch.
//
// Returns the universe corresponding to the input identifier
// along with any errors.
func (p *FleetProxy) fetchUniverse(id string) (Universe, error) {
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
	if len(unis) != 1 {
		return Universe{}, fmt.Errorf("Retrieved %d universes for id \"%s\" (expected 1)", len(unis), id)
	}

	return unis[0], nil
}

// CreateComponent :
// Used to perform the creation of a new component for a fleet.
// The component should describe the player willing to join the
// fleet along with some information about the starting position
// and the ships involved.
// We will make sure that the player belongs to the rigth uni,
// that the starting position is valid compared to the actual
// dimensions of the universe and that the fleet exists.
//
// The `comp` defines the fleet component to create.
//
// Returns any error in case the component cannot be added to
// the fleet for some reasons.
func (p *FleetProxy) CreateComponent(comp *FleetComponent) error {
	// TODO: Handle this.
	return fmt.Errorf("Not implemented")
}
