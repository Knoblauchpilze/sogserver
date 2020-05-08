package data

import (
	"fmt"
	"oglike_server/internal/game"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"

	"github.com/google/uuid"
)

// UniverseProxy :
// Intended as a wrapper to access properties of all the
// universes and retrieve data from the database. In most
// cases we need to access some properties of a single
// universe through a provided identifier.
type UniverseProxy struct {
	commonProxy
}

// NewUniverseProxy :
// Create a new proxy allowing to serve the requests
// related to universes.
//
// The `data` defines the data model to use to fetch
// information and verify requests.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewUniverseProxy(data model.Instance, log logger.Logger) UniverseProxy {
	return UniverseProxy{
		commonProxy: newCommonProxy(data, log, "universes"),
	}
}

// Universes :
// Return a list of universes registered so far in all the
// values defined in the DB. The input filters might help
// to narrow the search a bit by providing some properties
// the universe to look for should have.
//
// The `filters` define some filtering property that can
// be applied to the SQL query to only select part of all
// the universes available. Each one is appended `as-is`
// to the query.
//
// Returns the list of universes registered in the DB and
// matching the input list of filters. In case the error
// is not `nil` the value of the array should be ignored.
func (p *UniverseProxy) Universes(filters []db.Filter) ([]game.Universe, error) {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table:   "universes",
		Filters: filters,
	}

	dbRes, err := p.data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not query DB to fetch universes (err: %v)", err))
		return []game.Universe{}, err
	}
	if dbRes.Err != nil {
		p.trace(logger.Error, fmt.Sprintf("Invalid query to fetch universes (err: %v)", dbRes.Err))
		return []game.Universe{}, dbRes.Err
	}

	// We now need to retrieve all the identifiers that matched
	// the input filters and then build the corresponding item
	// object for each one of them.
	var ID string
	IDs := make([]string, 0)

	for dbRes.Next() {
		err = dbRes.Scan(&ID)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Error while fetching universe ID (err: %v)", err))
			continue
		}

		IDs = append(IDs, ID)
	}

	universes := make([]game.Universe, 0)

	for _, ID = range IDs {
		uni, err := game.NewUniverseFromDB(ID, p.data)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Unable to fetch universe \"%s\" data from DB (err: %v)", ID, err))
			continue
		}

		universes = append(universes, uni)
	}

	return universes, nil
}

// Create :
// Used to perform the creation of the universe described
// by the input data to the DB. In case the creation cannot
// be performed an error is returned.
//
// The `uni` describes the element to create in DB. This
// value may be modified by the function mainly to update
// the identifier of the universe if none have been set.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`. It also returns the identifier of the universe
// that was created: this is helpful in case there is no
// input identifier provided.
func (p *UniverseProxy) Create(uni game.Universe) (string, error) {
	// Assign a valid identifier if this is not already the case.
	if uni.ID == "" {
		uni.ID = uuid.New().String()
	}

	err := uni.SaveToDB(p.data.Proxy)
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not create universe \"%s\" (err: %v)", uni.Name, err))
		return uni.ID, err
	}

	p.trace(logger.Notice, fmt.Sprintf("Created new universe \"%s\" with id \"%s\"", uni.Name, uni.ID))

	return uni.ID, nil
}
