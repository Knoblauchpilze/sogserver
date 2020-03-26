package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/pkg/logger"
)

// listAccounts :
// Used to retrieve a list of all the accounts created so far on
// the server along with some general information. Note that it
// is not directly an indication of the players registered in the
// universes.
//
// Returns the handler that can be executed to serve such requests.
func (s *server) listAccounts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars, err := s.extractRouteVars("/accounts", r)
		if err != nil {
			panic(fmt.Errorf("Error while serving accounts (err: %v)", err))
		}

		// We have to assume that no `extra route` is provided on this
		// endpoint.
		if vars.path != "" {
			s.log.Trace(logger.Warning, fmt.Sprintf("Detected ignored extra route \"%s\" when serving accounts", vars.path))
		}

		// Retrieve the accounts from the bridge.
		accs, err := s.accounts.Accounts()
		if err != nil {
			s.log.Trace(logger.Error, fmt.Sprintf("Unexpected error while fetching accounts (err: %v)", err))
			http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

			return
		}

		// Marshal the content of the accounts.
		out, err := json.Marshal(accs)
		if err != nil {
			s.log.Trace(logger.Error, fmt.Sprintf("Error while marshalling accounts (err: %v)", err))
			http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

			return
		}

		// Notify the client.
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(out)

		if err != nil {
			s.log.Trace(logger.Error, fmt.Sprintf("Error while sending accounts to client (err: %v)", err))
		}
	}
}

// listAccount :
// Analyze the route provided in input to retrieve the properties of
// all accounts matching the requested information. This is usually
// used in coordination with the `listAccounts` method where the user
// will first fetch a list of all accounts and then maybe use this
// list to query specific properties of a person. The return value
// includes the list of properties using a `json` format.
//
// Returns the handler that can be executed to serve such requests.
func (s *server) listAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars, err := s.extractRouteVars("/accounts", r)
		if err != nil {
			panic(fmt.Errorf("Error while serving account (err: %v)", err))
		}

		s.log.Trace(logger.Warning, fmt.Sprintf("Should serve account: vars are %v", vars))
	}
}
