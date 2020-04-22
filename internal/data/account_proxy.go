package data

import (
	"fmt"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"

	"github.com/google/uuid"
)

// AccountProxy :
// Intended as a wrapper to access properties of all the
// accounts and retrieve data from the database. In most
// cases we need to access some properties of a single
// account through a provided identifier.
type AccountProxy struct {
	commonProxy
}

// NewAccountProxy :
// Create a new proxy allowing to serve the requests
// related to accounts.
//
// The `data` defines the data model to use to fetch
// information and verify actions.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewAccountProxy(data model.Instance, log logger.Logger) AccountProxy {
	return AccountProxy{
		commonProxy: newCommonProxy(data, log, "accounts"),
	}
}

// Accounts :
// Return a list of accounts registered so far in all the
// values defined in the DB. The input filters might help
// to narrow the search a bit by providing some properties
// the accounts to look for should have.
//
// The `filters` define some filtering property that can
// be applied to the SQL query to only select part of all
// the accounts available. Each one is appended `as-is`
// to the query.
//
// Returns the list of accounts registered in the DB and
// matching the input list of filters. In case the error
// is not `nil` the value of the array should be ignored.
func (p *AccountProxy) Accounts(filters []db.Filter) ([]model.Account, error) {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
		},
		Table:   "accounts",
		Filters: filters,
	}

	res, err := p.proxy.FetchFromDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not query DB to fetch accounts (err: %v)", err))
		return []model.Account{}, err
	}

	// We now need to retrieve all the identifiers that matched
	// the input filters and then build the corresponding unis
	// object for each one of them.
	var ID string
	IDs := make([]string, 0)

	for res.Next() {
		err = res.Scan(&ID)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Error while fetching account ID (err: %v)", err))
			continue
		}

		IDs = append(IDs, ID)
	}

	accounts := make([]model.Account, 0)

	for _, ID = range IDs {
		acc, err := model.NewAccountFromDB(ID, p.data)

		if err != nil {
			p.trace(logger.Error, fmt.Sprintf("Unable to fetch account \"%s\" data from DB (err: %v)", ID, err))
			continue
		}

		accounts = append(accounts, acc)
	}

	return accounts, nil
}

// Create :
// Used to perform the creation of the account described
// by the input data to the DB. In case the creation can
// not be performed an error is returned.
//
// The `acc` describes the element to create in DB. This
// value may be modified by the function mainly to update
// the identifier of the account if none have been set.
//
// The return status indicates whether the creation could
// be performed: if this is not the case the error is not
// `nil`.
func (p *AccountProxy) Create(acc model.Account) error {
	// Assign a valid identifier if this is not already the case.
	if acc.ID == "" {
		acc.ID = uuid.New().String()
	}

	// Check consistency.
	if !acc.Valid() {
		return model.ErrInvalidAccount
	}

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_account",
		Args:   []interface{}{acc},
	}

	err := p.proxy.InsertToDB(query)

	// TODO: Restore checks to indicate which type of error occurred.
	// Check for errors.
	if err != nil {
		p.trace(logger.Error, fmt.Sprintf("Could not create account \"%s\" (err: %v)", acc.Name, err))
		return err
	}

	p.trace(logger.Notice, fmt.Sprintf("Created new account \"%s\" with id \"%s\"", acc.Name, acc.ID))

	return nil
}
