package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strings"

	"github.com/google/uuid"
)

// AccountProxy :
// Intended as a wrapper to access properties of accounts
// and retrieve data from the database. Internally uses the
// common proxy defined in this package.
type AccountProxy struct {
	commonProxy
}

// NewAccountProxy :
// Create a new proxy allowing to serve the requests
// related to accounts.
//
// The `dbase` represents the database to use to fetch
// data related to accounts.
//
// The `log` allows to notify errors and information.
//
// Returns the created proxy.
func NewAccountProxy(dbase *db.DB, log logger.Logger) AccountProxy {
	return AccountProxy{
		newCommonProxy(dbase, log),
	}
}

// Accounts :
// Allows to fetch the list of accounts currently registered
// in the DB. This defines how many unique players already
// have created at least an account in a universe.
// The user can choose to filter parts of the accounts using
// an array of filters that will be applied to the SQL query.
// No controls is enforced on the filters so one should make
// sure that it's consistent with the underlying table.
//
// The `filters` define some filtering property that can be
// applied to the SQL query to only select part of all the
// accounts available. Each one is appended `as-is` to the
// query.
//
// Returns the list of accounts along with any errors. Note
// that in case the error is not `nil` the returned list is
// to be ignored.
func (p *AccountProxy) Accounts(filters []DBFilter) ([]Account, error) {
	// Create the query and execute it.
	query := queryDesc{
		props: []string{
			"id",
			"mail",
			"name",
			"password",
		},
		table:   "accounts",
		filters: filters,
	}

	res, err := p.fetchDB(query)
	defer res.Close()

	// Check for errors.
	if err != nil {
		return nil, fmt.Errorf("Could not query DB to fetch accounts (err: %v)", err)
	}

	// Populate the return value.
	accounts := make([]Account, 0)
	var acc Account

	for res.next() {
		err = res.scan(
			&acc.ID,
			&acc.Mail,
			&acc.Name,
			&acc.Password,
		)

		if err != nil {
			p.log.Trace(logger.Error, fmt.Sprintf("Could not retrieve info for account (err: %v)", err))
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
func (p *AccountProxy) Create(acc *Account) error {
	// Assign a valid identifier if this is not already the case.
	if acc.ID == "" {
		acc.ID = uuid.New().String()
	}

	// Validate that the input data describe a valid account.
	if !acc.valid() {
		return fmt.Errorf("Could not create account \"%s\", some properties are invalid", acc.Name)
	}

	// Create the query and execute it.
	query := insertReq{
		script: "create_account",
		args:   []interface{}{*acc},
	}

	err := p.insertToDB(query)

	// Check for errors.
	if err != nil {
		// We will refine this process a bit to try to detect
		// cases where the user tries to insert an account
		// with an already existing e-mail.
		// In this case we should get an error indicating a
		// `23505` as return code. We will refine the error
		// in this case.
		msg := fmt.Sprintf("%v", err)

		if strings.Contains(msg, getDuplicatedElementErrorKey()) {
			return fmt.Errorf("Could not import account \"%s\", mail \"%s\" already exists (err: %s)", acc.Name, acc.Mail, msg)
		}

		return fmt.Errorf("Could not import account \"%s\" (err: %s)", acc.Name, msg)
	}

	return nil
}
