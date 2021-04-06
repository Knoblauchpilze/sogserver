package model

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// CountriesModule :
// Defines the list of countries available for the game
// usually referencing the available traductions for an
// in-game element.
type CountriesModule struct {
	associationTable
	baseModule
}

// CountryDesc :
// Defines the abstract representation of a country as
// registered in the DB. Basically just a name and the
// corresponding identifier.
type CountryDesc struct {
	// ID defines the identifier of the country as it
	// appears in the DB.
	ID string `json:"id"`

	// Name defines the name of the country in a human
	// readable way.
	Name string `json:"name"`
}

// NewCountriesModule :
// Used to create a new countries module initialized with
// no content (as no DB is provided yet). The module will
// stay invalid until the `init` method is called with a
// valid DB.
//
// The `log` defines the logging layer to forward to the
// base `baseModule` element.
func NewCountriesModule(log logger.Logger) *CountriesModule {
	return &CountriesModule{
		associationTable: newAssociationTable(),
		baseModule:       newBaseModule(log, "countries"),
	}
}

// Init :
// Implementation of the `DBModule` interface to allow
// fetching information from the input DB and load to
// local memory.
//
// The `proxy` represents the main data source to use
// to initialize the resources data.
//
// The `force` allows to erase any existing information
// and reload everything from the DB.
//
// Returns any error.
func (cm *CountriesModule) Init(proxy db.Proxy, force bool) error {
	// Prevent reload if not needed.
	if cm.valid() && !force {
		return nil
	}

	// Perform the DB query through a dedicated DB proxy.
	query := db.QueryDesc{
		Props: []string{
			"id",
			"name",
		},
		Table:   "countries",
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	// Check for errors.
	if err != nil {
		cm.trace(logger.Error, fmt.Sprintf("Unable to initialize countries module (err: %v)", err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		cm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize countries module (err: %v)", rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var ID, name string

	override := false
	inconsistent := false

	for rows.Next() {
		err = rows.Scan(
			&ID,
			&name,
		)

		if err != nil {
			cm.trace(logger.Error, fmt.Sprintf("Failed to initialize country from row (err: %v)", err))
			inconsistent = true

			continue
		}

		// Check whether a country with this identifier exists.
		if cm.existsID(ID) {
			cm.trace(logger.Error, fmt.Sprintf("Prevented override of country \"%s\"", ID))
			override = true

			continue
		}

		// Register this country in the association table.
		err = cm.registerAssociation(ID, name)
		if err != nil {
			cm.trace(logger.Error, fmt.Sprintf("Cannot register country \"%s\" (id: \"%s\") (err: %v)", name, ID, err))
			inconsistent = true

			continue
		}
	}

	if override || inconsistent {
		return ErrInconsistentDB
	}

	return nil
}

// GetCountryFromID :
// Used to retrieve information on the country that
// corresponds to the input identifier. If no country
// with this ID exists an error is returned.
//
// The `id` defines the identifier of the country to
// fetch.
//
// Returns the description for this country along with
// any errors.
func (cm *CountriesModule) GetCountryFromID(id string) (CountryDesc, error) {
	// Find this element in the association table.
	if !cm.existsID(id) {
		cm.trace(logger.Error, fmt.Sprintf("Cannot retrieve desc for country \"%s\"", id))
		return CountryDesc{}, ErrNotFound
	}

	// We assume at this point that the identifier (and
	// thus the name) both exists so we discard errors.
	name, _ := cm.getNameFromID(id)

	res := CountryDesc{
		ID:   id,
		Name: name,
	}

	return res, nil
}

// GetCountryFromName :
// Calls internally the `GetCountryFromID` in order
// to forward the call to the above method. Failures
// happen in similar cases.
//
// The `name` defines the name of the country for
// which a description should be provided.
//
// Returns the description for this country along
// with any errors.
func (cm *CountriesModule) GetCountryFromName(name string) (CountryDesc, error) {
	// Find this element in the association table.
	id, err := cm.GetIDFromName(name)
	if err != nil {
		cm.trace(logger.Error, fmt.Sprintf("Cannot retrieve desc for country \"%s\" (err: %v)", name, err))
		return CountryDesc{}, ErrNotFound
	}

	return cm.GetCountryFromID(id)
}

// Countries :
// Used to retrieve the countries matching the input
// filters from the data model. Note that if the DB
// has not yet been polled to retrieve data, we will
// return an error.
// The process will consist in first fetching all the
// IDs of the countries matching the filters, and then
// build the rest of the data from the already fetched
// values.
//
// The `proxy` defines the DB to use to fetch the list
// of countries.
//
// The `filters` represent the list of filters to apply
// to the data fecthing. This will select only part of
// all the available countries.
//
// Returns the list of countries matching the filters
// along with any error.
func (cm *CountriesModule) Countries(proxy db.Proxy, filters []db.Filter) ([]CountryDesc, error) {
	// Initialize the module if for some reasons it is still
	// not valid.
	if !cm.valid() {
		err := cm.Init(proxy, true)
		if err != nil {
			return []CountryDesc{}, err
		}
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table:   "countries",
		Filters: filters,
	}

	IDs, err := cm.fetchIDs(query, proxy)
	if err != nil {
		cm.trace(logger.Error, fmt.Sprintf("Unable to fetch countries (err: %v)", err))
		return []CountryDesc{}, err
	}

	// Now build the data from the fetched identifiers.
	descs := make([]CountryDesc, 0)
	for _, ID := range IDs {
		name, err := cm.getNameFromID(ID)
		if err != nil {
			cm.trace(logger.Error, fmt.Sprintf("Unable to fetch country \"%s\" (err: %v)", ID, err))
			continue
		}

		desc := CountryDesc{
			ID:   ID,
			Name: name,
		}

		descs = append(descs, desc)
	}

	return descs, nil
}
