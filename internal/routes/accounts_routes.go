package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/logger"
)

// listAccounts :
// Used to perform the creation of a handler allowing to serve
// the requests on accounts.
//
// Returns the handler that can be executed to serve said reqs.
func (s *server) listAccounts() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("accounts")

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

// createAccount :
// Used to perform the creation of a handler allowing to server
// the requests to create accounts.
//
// Returns the handler to execute to perform said requests.
func (s *server) createAccount() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewCreateResourceEndpoint("accounts")

	// Configure the endpoint.
	ed.WithDataKey("account-data")
	ed.WithCreationFunc(
		func(input RouteData) ([]string, error) {
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
					s.log.Trace(logger.Error, fmt.Sprintf("Could not create account from data \"%s\" (err: %v)", rawData, err))
					continue
				}

				// Create the account.
				err = s.accounts.Create(&acc)
				if err != nil {
					s.log.Trace(logger.Error, fmt.Sprintf("Could not register account from data \"%s\" (err: %v)", rawData, err))
					continue
				}

				// Successfully created an account.
				s.log.Trace(logger.Notice, fmt.Sprintf("Created new account \"%s\" with id \"%s\"", acc.Name, acc.ID))
				resources = append(resources, acc.ID)
			}

			// Return the path to the resources created during the process.
			return resources, nil
		},
	)

	return ed.ServeRoute(s.log)
}
