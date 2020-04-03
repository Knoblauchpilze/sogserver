package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"
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
type FleetProxy struct {
	dbase *db.DB
	log   logger.Logger
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
// Returns the created proxy.
func NewFleetProxy(dbase *db.DB, log logger.Logger) FleetProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create fleets proxy from invalid DB"))
	}

	return FleetProxy{dbase, log}
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
		"objective",
		"arrival_time",
		"galaxy",
		"solar_system",
		"position",
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

	fmt.Println(fmt.Sprintf("Query: \"%s\"", query))

	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch fleets (err: %v)", err)
	}

	// Populate the return value: we should obtain a single
	// result, otherwise it's an issue.
	fleets := make([]Fleet, 0)
	var fleet Fleet

	galaxy := 0
	system := 0
	position := 0

	for rows.Next() {
		err = rows.Scan(
			&fleet.ID,
			&fleet.Name,
			&fleet.Objective,
			// TODO: This may fail, see when we have a real fleet.
			&fleet.ArrivalTime,
			&galaxy,
			&system,
			&position,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for fleet (err: %v)", err))
			continue
		}

		fleet.Coords = Coordinate{
			galaxy,
			system,
			position,
		}

		// Fetch individual components of the fleet.
		err = p.fetchFleetData(&fleet)
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch data for fleet \"%s\" (err: %v)", fleet.ID, err))
			continue
		}

		fleets = append(fleets, fleet)
	}

	return fleets, nil
}

// fetchFleetData :
// Used to fetch data related to a fleet: this includes the
// individual components of the fleet (which is mostly used
// in the case of group attacks).
//
// The `fleet` references the fleet for which the data should
// be fetched. We assume that the internal fields (and more
// specifically the identifier) are already populated.
//
// Returns any error.
func (p *FleetProxy) fetchFleetData(fleet *Fleet) error {
	// Check whether the fleet has an identifier assigned.
	if fleet.ID == "" {
		return fmt.Errorf("Unable to fetch data from fleet with invalid identifier")
	}

	// Create the query to fetch individual components of the
	// fleet and execute it.
	props := []string{
		"player",
		"ship",
		"amount",
		"start_galaxy",
		"start_solar_system",
		"start_position",
	}

	table := "fleet_ships"
	where := fmt.Sprintf("fleet='%s'", fleet.ID)

	query := fmt.Sprintf("select %s from %s where %s", strings.Join(props, ", "), table, where)
	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return fmt.Errorf("Could not query DB to fetch fleet \"%s\" details (err: %v)", fleet.ID, err)
	}

	// Populate the return content from the result of the
	// DB query.
	fleet.Components = make([]FleetComponent, 0)
	var comp FleetComponent

	galaxy := 0
	system := 0
	position := 0

	for rows.Next() {
		err = rows.Scan(
			&comp.PlayerID,
			&comp.ShipID,
			&comp.Amount,
			&galaxy,
			&system,
			&position,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for fleet \"%s\" component (err: %v)", fleet.ID, err))
			continue
		}

		comp.Coords = Coordinate{
			galaxy,
			system,
			position,
		}

		fleet.Components = append(fleet.Components, comp)
	}

	return nil
}
