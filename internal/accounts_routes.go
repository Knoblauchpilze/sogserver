package internal

import (
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
		route, err := s.extractRoute(r, "/accounts")
		if err != nil {
			panic(fmt.Errorf("Error while serving accounts (err: %v)", err))
		}

		s.log.Trace(logger.Warning, fmt.Sprintf("Should serve list of all accounts (route: \"%s\")", route))
	}
}

// listAccount :
// Provide detailed information about a single account referenced
// by the input route. If no such account exist this will be sent
// back to the client.
//
// Returns the handler to be executed to serve these requests.
func (s *server) listAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, err := s.extractRoute(r, "/accounts/")
		if err != nil {
			panic(fmt.Errorf("Error while serving account (err: %v)", err))
		}

		s.log.Trace(logger.Warning, fmt.Sprintf("Should serve info for account \"%s\"", account))
	}
}
