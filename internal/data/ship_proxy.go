package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// ShipProxy :
// Intended as a wrapper to access properties of ships and
// retrieve data from the database. Internally uses the
// common proxy defined in this package.
type ShipProxy struct {
	commonProxy
}

// NewShipProxy :
// Create a new proxy allowing to serve the requests
// related to ships.
//
// The `dbase` represents the database to use to fetch
// data related to ships.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewShipProxy(dbase *db.DB, log logger.Logger) ShipProxy {
	return ShipProxy{
		newCommonProxy(dbase, log),
	}
}

// Ships :
// Allows to fetch the list of ships currently available for a
// player to build. This list normally never changes as it's
// very unlikely to create a new ship.
// The user can choose to filter parts of the ships using an
// array of filters that will be applied to the SQL query.
// No controls is enforced on the filters so one should make
// sure that it's consistent with the underlying table.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// ships available. Each one is appended `as-is` to the query.
//
// Returns the list of ships along with any errors. Note that
// in case the error is not `nil` the returned list is to be
// ignored.
func (p *ShipProxy) Ships(filters []DBFilter) ([]ShipDesc, error) {
	// Create the query and execute it.
	query := queryDesc{
		props: []string{
			"id",
			"name",
		},
		table:   "ships",
		filters: filters,
	}

	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch ships (err: %v)", err)
	}

	// Populate the return value.
	ships := make([]ShipDesc, 0)
	var desc ShipDesc

	for res.next() {
		err = res.scan(
			&desc.ID,
			&desc.Name,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for ship (err: %v)", err))
			continue
		}

		desc.Cost, err = fetchElementCost(p.dbase, desc.ID, "ship", "ships_costs")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch cost for ship \"%s\" (err: %v)", desc.ID, err))
		}

		desc.BuildingsDeps, err = fetchElementDependency(p.dbase, desc.ID, "ship", "tech_tree_ships_vs_buildings")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch building dependencies for ship \"%s\" (err: %v)", desc.ID, err))
		}

		desc.TechnologiesDeps, err = fetchElementDependency(p.dbase, desc.ID, "ship", "tech_tree_ships_vs_technologies")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch technologies dependencies for ship \"%s\" (err: %v)", desc.ID, err))
		}

		desc.RFVSShips, err = p.fetchRapidFires(desc.ID, "ships_rapid_fire")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch rapid fires for ship \"%s\" against other ships (err: %v)", desc.ID, err))
		}

		desc.RFVSDefenses, err = p.fetchRapidFires(desc.ID, "ships_rapid_fire_defenses")
		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not fetch rapid fires for ship \"%s\" against defenses (err: %v)", desc.ID, err))
		}

		ships = append(ships, desc)
	}

	return ships, nil
}

// fetchRapidFires :
// Used to retrieve the rapid fire defined for the input ship
// (described by its identifier) in the input table. The RFs
// are returned in a corresponding slice along with any error.
//
// The `ship` describes the identifier of the ship for which
// the rapid fires should be retrieved.
//
// The `table` defines the table into which the rapid fires
// should be fetched. We define two main kind of rapid fire:
// from a ship to another ship and from a ship to a defense
// system.
//
// Returns the list of rapid fires this ship possess against
// the other elements along with any error.
func (p *ShipProxy) fetchRapidFires(ship string, table string) ([]RapidFire, error) {
	// Check consistency.
	if ship == "" {
		return []RapidFire{}, fmt.Errorf("Cannot fetch rapid fire for invalid ships")
	}

	// Create the query and execute it.
	query := queryDesc{
		props: []string{
			"target",
			"rapid_fire",
		},
		table: table,
		filters: []DBFilter{
			{
				"ship",
				[]string{ship},
			},
		},
	}

	res, err := p.fetchDB(query)
	defer res.Close()

	if err != nil {
		return []RapidFire{}, fmt.Errorf("Could not retrieve rapid fires for \"%s\" (err: %v)", ship, err)
	}

	// Populate the rapid fires.
	var gError error

	rfs := make([]RapidFire, 0)
	var rf RapidFire

	for res.next() {
		err = res.scan(
			&rf.Receiver,
			&rf.RF,
		)

		if err != nil {
			gError = fmt.Errorf("Could not retrieve info for rapid fires of \"%s\" (err: %v)", ship, err)
		}

		rfs = append(rfs, rf)
	}

	return rfs, gError
}
