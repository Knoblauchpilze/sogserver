package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/handlers"
	"oglike_server/pkg/logger"
)

// accountAdapter :
// Implements the interface requested by the general handler in
// the `handlers` package. The main functions are describing the
// interface to retrieve information about the accounts from a
// database.
//
// The `proxy` defines the proxy to use to interact with the DB
// when fetching the data.
type accountAdapter struct {
	proxy data.AccountProxy
}

// Route :
// Implementation of the method to get the route name to access
// the accounts' data.
// Returns the name of the route.
func (aa *accountAdapter) Route() string {
	return "accounts"
}

// ParseFilters :
// Implementation of the method to get the filters from the route
// variables. This will allow to fetch accounts through a precise
// identifier.
//
// The `vars` are the variables (both route element and the query
// parameters) retrieved from the input request.
//
// Returns the list of filters extracted from the input info.
func (aa *accountAdapter) ParseFilters(vars handlers.RouteVars) []handlers.Filter {
	filters := make([]handlers.Filter, 0)

	// Traverse the input parameters and select only the ones relevant
	// to accounts querying.
	allowed := map[string]string{
		"account_id":   aa.proxy.GetIdentifierDBColumnName(),
		"account_name": "name",
		"account_mail": "mail",
	}

	for key, values := range vars.Params {
		// Check whether this key corresponds to an account filter. If
		// this is the case we can register it.
		filterName, ok := allowed[key]

		if ok && len(values) > 0 {
			// None of the filters associated to the account are numeric
			// for now. If it's the case later on it would have to be
			// modified.
			filter := handlers.Filter{
				Key:     filterName,
				Options: values,
			}

			filters = append(filters, filter)
		}
	}

	// We also need to fetch parts of the route that can be used to
	// provide a filter on the account's identifier. More precisely
	// the route can define something like:
	// `/accounts/account-id` which we will interpret as a filter on
	// the account's identifier.
	// Note that we assume that if the route contains more than `1`
	// element it *always* contains an identifier as second token.
	if len(vars.RouteElems) > 0 {
		uni := vars.RouteElems[0]

		// Append the identifier filter to the existing list.
		found := false
		for id := range filters {
			if filters[id].Key == aa.proxy.GetIdentifierDBColumnName() {
				found = true
				filters[id].Options = append(filters[id].Options, uni)
			}
		}

		if !found {
			filters = append(
				filters,
				handlers.Filter{
					Key:     aa.proxy.GetIdentifierDBColumnName(),
					Options: []string{uni},
				},
			)
		}
	}

	return filters
}

// Data :
// Implementation of the method to get the data related to accounts
// from the internal DB. We will use the internal DB proxy to get the
// info while still applying the filters.
//
// The `filters` represent the filters extracted from the route and
// as provided by the `ParseFilters` method. We need to convert it
// into a semantic that can be interpreted by the DB.
//
// Returns the data related to the accounts along with any errors.
func (aa *accountAdapter) Data(filters []handlers.Filter) (interface{}, error) {
	// Convert the input request filters into DB filters.
	dbFilters := make([]data.DBFilter, 0)
	for _, filter := range filters {
		dbFilters = append(
			dbFilters,
			data.DBFilter{
				Key:    filter.Key,
				Values: filter.Options,
			},
		)
	}

	// Use the DB proxy to fetch the data.
	return aa.proxy.Accounts(dbFilters)
}

// accountCreator :
// Implements the interface requested by the creation handler in
// the `handlers` package. The main functions are describing the
// interface to retrieve information about the accounts from a
// database.
//
// The `proxy` defines the proxy to use to interact with the DB
// when fetching the data.
//
// The `log` allows to notify problems and information during a
// universe's creation.
type accountCreator struct {
	proxy data.AccountProxy
	log   logger.Logger
}

// Route :
// Implementation of the method to get the route name to create some
// new accounts.
// Returns the name of the route.
func (ac *accountCreator) Route() string {
	return "account"
}

// AccessRoute :
// Implementation of the method to get the route name to access to
// the data created by this handler. This is basically the `accounts`
// route.
func (ac *accountCreator) AccessRoute() string {
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

// listAccounts :
// Creates a handler allowing to serve requests on the accounts
// by interrogating the main DB. We uses the handler structure in
// the `handlers` package and provide the needed endpoint desc as
// requested.
//
// Returns the handler that can be executed to serve such requests.
func (s *server) listAccounts() http.HandlerFunc {
	return handlers.ServeRoute(
		&accountAdapter{
			s.accounts,
		},
		s.log,
	)
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
