package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/logger"
	"strings"
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
		err = marshalAndSend(accs, w)
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

		// Extract parts of the route: we first need to remove the first
		// '/' character.
		purged := vars.path[1:]
		parts := strings.Split(purged, "/")

		// Depending on the number of parts in the route, we will call
		// the suited handler.
		switch len(parts) {
		case 1:
			// Assume that a request on the account itself is equivalent
			// to requesting the universes into which it is present.
			fallthrough
		case 2:
			if len(parts) > 1 && parts[1] != "players" {
				s.log.Trace(logger.Warning, fmt.Sprintf("Detected ignored extra route \"%s\" when serving accounts for player \"%s\"", parts[1], parts[0]))
			}
			s.listPlayersForAccount(w, parts[0])
			return
		case 3:
			s.listPlayerProps(w, parts, vars.params)
			return
		default:
			// Can't do anything.
		}

		s.log.Trace(logger.Error, fmt.Sprintf("Unhandled request for account \"%s\"", purged))
		http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)
	}
}

// createAccount :
// Produce a handler that can be used to perform the creation of the
// account of players when they register in the system. This is the
// first action that the user must do before being able to register
// in a universe.
//
// Returns the handler to execute to handle accounts' creation.
func (s *server) createAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars, err := s.extractRouteVars("/account", r)
		if err != nil {
			panic(fmt.Errorf("Error while creating account (err: %v)", err))
		}

		// The route should not contain anymore data.
		if vars.path != "" {
			s.log.Trace(logger.Warning, fmt.Sprintf("Detected ignored extra route \"%s\" when creating account", vars.path))
		}

		// Extract the data to use to create the account from the input
		// request: this can be done conveniently through the server's
		// base method.
		var uniData routeData
		uniData, err = s.extractRouteData(r)
		if err != nil {
			panic(fmt.Errorf("Error while fetching data to create account (err: %v)", err))
		}

		// Unmarshal the content into a valid account.
		var acc data.Account
		err = json.Unmarshal([]byte(uniData.value), &acc)
		if err != nil {
			panic(fmt.Errorf("Error while parsing data to create account (err: %v)", err))
		}

		// Note that we don't actually check that the account is unique
		// so far in the database. We count on the fact that the mail
		// must be unique for all the accounts. Typically the client to
		// allow the user to create an account should fetch the list of
		// existing mails and prevent the request to be issued with an
		// invalid mail (or an already existing one).
		// At this step we consider that preventing the insertion of the
		// duplicated account is enough.
		err = s.accounts.Create(&acc)
		if err != nil {
			s.log.Trace(logger.Error, fmt.Sprintf("Could not create account from name \"%s\" (err: %v)", acc.Name, err))
			http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

			return
		}

		// We need to return a valid status code and the address of
		// the created resource, as described in the following post:
		// https://stackoverflow.com/questions/1829875/is-it-ok-by-rest-to-return-content-after-post
		resource := fmt.Sprintf("/accounts/%s", acc.ID)
		notifyCreation(resource, w)
	}
}

// listPlayersForAccount :
// Used to forrmat the list of universes into which the input account
// is registered into a json structure and send it back to the client
// through the provided response writer.
//
// The `w` argument defines the response writer to send back data to
// the client.
//
// The `account` represents the player's identifier which should be
// used to fetch the associated accounts.
func (s *server) listPlayersForAccount(w http.ResponseWriter, account string) {
	// Fetch accounts.
	accs, err := s.accounts.Characters(account)
	if err != nil {
		s.log.Trace(logger.Error, fmt.Sprintf("Unexpected error while fetching account \"%s\" (err: %v)", account, err))
		http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

		return
	}

	// Marshal and send the content.
	err = marshalAndSend(accs, w)
	if err != nil {
		s.log.Trace(logger.Error, fmt.Sprintf("Unexpected error while sending data for \"%s\" (err: %v)", account, err))
	}
}

// listPlayerProps :
// Used to retrieve information about a certain player and provide
// the corresponding info to the client through the response writer
// given as input.
//
// The `w` is the response writer to use to send the response back
// to the client.
//
// The `params` represents the aprameters provided to filter the
// data to retrieve for this account. The first element of this
// array is guaranteed to correspond to the identifier of the
// account for which the data should be retrieved.
//
// The `filters` correspond to the query filter that are set as an
// additional filtering layer to query only specific properties of
// the account.
func (s *server) listPlayerProps(w http.ResponseWriter, params []string, filters map[string]string) {
	//  We know that the first elements of the `params` array should
	// correspond to the account's identifier (i.e. the root value
	// which owns all the individual player's in the universes, and
	// the rest of the values correspond to filtering properties to
	// query only specific information.
	// We have the following routes possible:
	//  - `account_id/player_id` -> query params will be ignored.
	//  - `account_id/player_id/item` -> only retrieve the info of
	//    the `item` for the specified player id.
	user := params[0]

	if len(params) == 1 {
		s.listPlayersForAccount(w, user)
		return
	}

	// We need to feetch the player's data from its identifier.
	accountID := params[1]

	players, err := s.accounts.Characters(user)
	if err != nil {
		s.log.Trace(logger.Error, fmt.Sprintf("Unexpected error while fetching account \"%s\" (err: %v)", user, err))
		http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

		return
	}

	// Scan to find the account corresponding to the specified id.
	var player data.Player
	id := 0
	found := false

	for ; id < len(players) && !found; id++ {
		player = players[id]

		if player.ID == accountID {
			found = true
		}
	}

	if !found {
		s.log.Trace(logger.Error, fmt.Sprintf("Unable to find account \"%s\" associated to user \"%s\" (err: %v)", accountID, user, err))
		http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

		return
	}

	// Retrieve specific information of the player.
	var errSend error
	var planets []data.Planet
	var researches []data.Research
	var fleets []data.Fleet

	switch params[2] {
	case "planets":
		planets, err = s.accounts.Planets(player.ID)
		if err == nil {
			errSend = marshalAndSend(planets, w)
		}
	case "researches":
		researches, err = s.accounts.Researches(player)
		if err == nil {
			errSend = marshalAndSend(researches, w)
		}
	case "fleets":
		fleets, err = s.accounts.Fleets(player)
		if err == nil {
			errSend = marshalAndSend(fleets, w)
		}
	}

	// Notify errors.
	if err != nil {
		s.log.Trace(logger.Error, fmt.Sprintf("Unable to fetch data for account \"%s\" (err: %v)", accountID, err))
		http.Error(w, InternalServerErrorString(), http.StatusInternalServerError)

		return
	}

	if errSend != nil {
		s.log.Trace(logger.Error, fmt.Sprintf("Unexpected error while sending data for account \"%s\" (err: %v)", accountID, err))
	}
}
