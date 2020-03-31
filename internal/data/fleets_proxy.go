package data

import (
	"fmt"
	"oglike_server/pkg/logger"
	"strings"
)

// Fleet :
// Used to retrieve the fleet described by the identifier
// in input. Only general information about the fleet is
// returned (i.e. no details composition).
//
// The `fleet` describes the identifier of the fleet for
// which data should be fetched.
//
// Returns the corresponding fleet or an error in case a
// fleet with said identifier cannot be found.
func (p *UniverseProxy) Fleet(fleet string) (Fleet, error) {
	// Create the query and execute it.
	props := []string{
		"f.id",
		"f.name",
		"fo.name",
		"f.arrival_time",
		"f.galaxy",
		"f.solar_system",
		"f.position",
	}

	table := "fleets f inner join fleet_objectives fo"
	joinCond := "f.objective=fo.id"
	where := fmt.Sprintf("f.id='%s'", fleet)

	query := fmt.Sprintf("select %s from %s on %s where %s", strings.Join(props, ", "), table, joinCond, where)
	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return Fleet{}, fmt.Errorf("Could not query DB to fetch fleet \"%s\" (err: %v)", fleet, err)
	}

	// Populate the return value: we should obtain a single
	// result, otherwise it's an issue.
	var fl Fleet

	galaxy := 0
	system := 0
	position := 0

	rows.Next()

	// Scan the first row.
	func() {
		defer func() {
			errScan := recover()
			if errScan != nil {
				err = fmt.Errorf("Could not perform scanning (err: %v)", errScan)
			}
		}()

		err = rows.Scan(
			&fl.ID,
			&fl.Name,
			&fl.Objective,
			// TODO: This may fail, see when we have a real fleet.
			&fl.ArrivalTime,
			&galaxy,
			&system,
			&position,
		)
	}()

	// Check for errors.
	if err != nil {
		return fl, fmt.Errorf("Could not retrieve info for fleet \"%s\" (err: %v)", fleet, err)
	}

	fl.Coords = Coordinate{
		galaxy,
		system,
		position,
	}

	// Skip remaining values (if any), and indicate the problem
	// if there are more.
	count := 0
	for rows.Next() {
		count++
	}
	if count > 0 {
		err = fmt.Errorf("Found %d values for fleet \"%s\"", count, fleet)
	}

	return fl, err
}

// FleetDetails :
// Used to retrieve the components of the fleet described
// as input parameter. This will list all the ships and
// potential players participating in the fleet.
// If the fleet is not valid the most liekely result is
// that no components will be found. If this is the case
// an error will be raised.
//
// The `fleet` describe the fleet for which components are
// to be retrieved.
//
// Returns the list of individual components for the fleet
// along with any errors.
func (p *UniverseProxy) FleetDetails(fleet Fleet) ([]FleetComponent, error) {
	// Create the query and execute it.
	props := []string{
		"s.player",
		"fs.ship",
		"fs.amount",
		"fs.start_galaxy",
		"fs.start_solar_system",
		"fs.start_position",
	}

	table := "fleet_ships fs inner join fleets f"
	joinCond := "fs.fleet=f.id"
	where := fmt.Sprintf("f.id='%s'", fleet.ID)

	query := fmt.Sprintf("select %s from %s on %s where %s", strings.Join(props, ", "), table, joinCond, where)
	rows, err := p.dbase.DBQuery(query)

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch fleet \"%s\" details (err: %v)", fleet.ID, err)
	}

	// Populate the return content from the result of the
	// DB query.
	components := make([]FleetComponent, 0)
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

		components = append(components, comp)
	}

	return components, nil
}
