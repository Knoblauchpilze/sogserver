package routes

import (
	"encoding/json"
	"net/http"
	"oglike_server/internal/game"
	"oglike_server/pkg/db"
)

// listAccounts :
// Used to perform the creation of a handler allowing to serve
// the requests on accounts.
//
// Returns the handler that can be executed to serve said reqs.
func (s *Server) listAccounts() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewGetResourceEndpoint("accounts")

	allowed := map[string]string{
		"id":   "id",
		"name": "name",
		"mail": "mail",
	}

	// Configure the endpoint.
	ed.WithFilters(allowed).WithResourceFilter("id").WithModule("accounts").WithLocker(s.og)
	ed.WithDataFunc(
		func(filters []db.Filter) (interface{}, error) {
			return s.accounts.Accounts(filters)
		},
	)

	return ed.ServeRoute(s.log)
}

// createAccount :
// Used to perform the creation of a handler allowing to serve
// the requests to create accounts.
//
// Returns the handler to execute to perform said requests.
func (s *Server) createAccount() http.HandlerFunc {
	// Create the endpoint with the suited route.
	ed := NewCreateResourceEndpoint("accounts")

	// Configure the endpoint.
	ed.WithDataKey("account-data").WithModule("accounts").WithLocker(s.og)
	ed.WithCreationFunc(
		func(input RouteData) ([]string, error) {
			// We need to iterate over the data retrieved from the route and
			// create accounts from it.
			var acc game.Account
			resources := make([]string, 0)

			// Prevent request with no data.
			if len(input.Data) == 0 {
				return resources, ErrNoData
			}

			for _, rawData := range input.Data {
				// Try to unmarshal the data into a valid `Account` struct.
				err := json.Unmarshal([]byte(rawData), &acc)
				if err != nil {
					return resources, ErrInvalidData
				}

				// Create the account.
				res, err := s.accounts.Create(acc)
				if err != nil {
					return resources, err
				}

				// Successfully created an account.
				resources = append(resources, res)
			}

			// Return the path to the resources created during the process.
			return resources, nil
		},
	)

	return ed.ServeRoute(s.log)
}
