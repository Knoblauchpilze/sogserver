package model

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// TechnologiesModule :
// Refines the concept of `progressCostsModule` for the
// particular case of technologies. A technology is some
// useful tool to enable developping more powerful ships
// and defense systems in the game.
// Each technology has some dependencies (meaning that
// it can't be researched without these prerequisites)
// and a progression costs system which indicates that
// the higher the researched level, the more expensive
// it is.
type TechnologiesModule struct {
	progressCostsModule
}

// TechnologyDesc :
// Describes the information associated to a technology.
// It defines its identifier (which is used in various
// places in the DB to reference it) and its name that
// is usually used for display purposes.
//
// The `Cost` allows to compute the cost of this item
// at any level.
type TechnologyDesc struct {
	UpgradableDesc

	Cost ProgressCost `json:"cost"`
}

// NewTechnologiesModule :
// Creates a new module allowing to handle technologies
// defined in the game. Progression rules and initial
// costs are fetched in a way consistent with what is
// defined in the `upgradadablesModule`. This module is
// meant to fetch the names of the technologies as an
// additional process.
//
// The `log` defines the logging layer to forward to the
// base `progressCostsModule` element.
func NewTechnologiesModule(log logger.Logger) *TechnologiesModule {
	return &TechnologiesModule{
		progressCostsModule: *newProgressCostsModule(log, Technology, "technology"),
	}
}

// Init :
// Provide some more implementation to retrieve data from
// the DB by fetching the technologies' identifiers and
// display names. This will constitute the base from which
// the upgradable module can attach the progression rules.
//
// The `dbase` represents the main data source to use
// to initialize the technologies data.
//
// The `force` allows to erase any existing information
// and reload everything from the DB.
//
// Returns any error.
func (tm *TechnologiesModule) Init(dbase *db.DB, force bool) error {
	if dbase == nil {
		tm.trace(logger.Error, fmt.Sprintf("Unable to initialize module from nil DB"))
		return db.ErrInvalidDB
	}

	// Prevent reload if not needed.
	if tm.valid() && !force {
		return nil
	}

	proxy := db.NewProxy(dbase)

	// Load the names and base information for each technology.
	// This operation is performed first so that the rest of
	// the data can be checked against the actual list of techs
	// that are defined in the game.
	err := tm.initNames(proxy)
	if err != nil {
		tm.trace(logger.Error, fmt.Sprintf("Could not initialize module (err: %v)", err))
		return err
	}

	// Perform the initialization of the progression rules,
	// and various data from the base handlers.
	err = tm.progressCostsModule.Init(dbase, force)
	if err != nil {
		tm.trace(logger.Error, fmt.Sprintf("Failed to initialize base module (err: %v)", err))
		return err
	}

	return nil
}

// initNames :
// Used to perform the fetching of the definition of techs
// from the input proxy. It will only get some basic info
// about each one such as their names and identifiers in
// order to populate the minimum information to fact-check
// the rest of the data (like the progression costs rules,
// etc.).
//
// The `proxy` defines a convenient way to access to the DB.
//
// Returns any error.
func (tm *TechnologiesModule) initNames(proxy db.Proxy) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
			"name",
		},
		Table:   "technologies",
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		tm.trace(logger.Error, fmt.Sprintf("Unable to initialize technologies (err: %v)", err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		tm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize technologies (err: %v)", rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var ID, name string

	override := false
	inconsistent := false

	for rows.Next() {
		err := rows.Scan(
			&ID,
			&name,
		)

		if err != nil {
			tm.trace(logger.Error, fmt.Sprintf("Failed to initialize technology from row (err: %v)", err))
			continue
		}

		// Check whether a technology with this identifier exists.
		if tm.existsID(ID) {
			tm.trace(logger.Error, fmt.Sprintf("Prevented override of technology \"%s\"", ID))
			override = true

			continue
		}

		// Register this technology in the association table.
		err = tm.registerAssociation(ID, name)
		if err != nil {
			tm.trace(logger.Error, fmt.Sprintf("Cannot register technology \"%s\" (id: \"%s\") (err: %v)", name, ID, err))
			inconsistent = true
		}
	}

	if override || inconsistent {
		return ErrInconsistentDB
	}

	return nil
}

// Technologies :
// Used to retrieve the technologies matching the input
// filters from the data model. Note that if the DB has
// not yet been polled to retrieve data, we will return
// an error.
// The process will consist in first fetching all the IDs
// of the technologies matching the filters, and then
// build the rest of the data from the already fetched
// values.
//
// The `filters` represent the list of filters to apply
// to the data fecthing. This will select only part of
// all the available technologies.
//
// Returns the list of technologies matching the filters
// along with any error.
func (tm *TechnologiesModule) Technologies(filters []db.Filter) ([]TechnologyDesc, error) {
	// TODO: Handle this.
	return nil, fmt.Errorf("Not implemented")
}
