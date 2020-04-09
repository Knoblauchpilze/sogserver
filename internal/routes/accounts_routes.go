package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/handlers"
	"oglike_server/pkg/logger"
)

// listAccounts :
// Used to perform the creation of a handler allowing to serve
// the requests on accounts.
//
// Returns the handler that can be executed to serve said reqs.
func (s *server) listAccounts() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewEndpointDesc("accounts")

	allowed := map[string]string{
		"account_id":   "id",
		"account_name": "name",
		"account_mail": "mail",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithIDFilter("id")
	ed.WithDataFunc(
		func(filters []data.DBFilter) (interface{}, error) {
			return s.accounts.Accounts(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// accountCreator :
// Implements the interface requested by the creation handler in
// the `handlers` package. The main functions are describing the
// interface to create a new account from the data fetched from
// the input request.
//
// The `proxy` defines the proxy to use to interact with the DB
// when creating the data.
//
// The `log` allows to notify problems and information during an
// account's creation.
type accountCreator struct {
	proxy data.AccountProxy
	log   logger.Logger
}

// Route :
// Implementation of the method to get the route name to create some
// new accounts.
// Returns the name of the route.
func (ac *accountCreator) Route() string {
	return "accounts"
}

// DataKey :
// Implementation of the method to get the name of the key used to
// pass data to the server.
// Returns the name of the key.
func (ac *accountCreator) DataKey() string {
	return "account-data"
}

// Create :
// Implementation of the method to perform the creation of the data
// related to the new accounts. We will use the internal proxy to
// request the DB to create a new account.
//
// The `input` represent the data fetched from the input request and
// should contain the properties of the accounts to create.
//
// Return the targets of the created resources along with any error.
func (ac *accountCreator) Create(input handlers.RouteData) ([]string, error) {
	// We need to iterate over the data retrieved from the route and
	// create accounts from it.
	var acc data.Account
	resources := make([]string, 0)

	// Prevent request with no data.
	if len(input.Data) == 0 {
		return resources, fmt.Errorf("Could not perform creation of account with no data")
	}

	for _, rawData := range input.Data {
		// Try to unmarshal the data into a valid `Account` struct.
		err := json.Unmarshal([]byte(rawData), &acc)
		if err != nil {
			ac.log.Trace(logger.Error, fmt.Sprintf("Could not create account from data \"%s\" (err: %v)", rawData, err))
			continue
		}

		// Create the account.
		err = ac.proxy.Create(&acc)
		if err != nil {
			ac.log.Trace(logger.Error, fmt.Sprintf("Could not register account from data \"%s\" (err: %v)", rawData, err))
			continue
		}

		// Successfully created an account.
		ac.log.Trace(logger.Notice, fmt.Sprintf("Created new account \"%s\" with id \"%s\"", acc.Name, acc.ID))
		resources = append(resources, acc.ID)
	}

	// Return the path to the resources created during the process.
	return resources, nil
}

// createAccount :
// Creates a handler allowing to server requests to create new
// accounts in the main DB. This rely on the handler structure
// provided by the `handlers` package which allows to mutualize
// the extraction of the data from the input request and the
// general flow to perform the creation.
//
// Returns the handler which can be executed to perform such
// requests.
func (s *server) createAccount() http.HandlerFunc {
	return handlers.ServeCreationRoute(
		&accountCreator{
			s.accounts,
			s.log,
		},
		s.log,
	)
}
