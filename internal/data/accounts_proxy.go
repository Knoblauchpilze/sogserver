package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// AccountProxy :
// Intended as a wrapper to access properties of accounts
// and retrieve data from the database. This helps hiding
// the complexity of how the data is laid out in the `DB`
// and the precise name of tables from the exterior world.
//
// The `dbase` is the database that is wrapped by this
// object. It is checked for consistency upon building the
// wrapper.
//
// The `log` allows to perform display to the user so as
// to inform of potential issues and debug information to
// the outside world.
type AccountProxy struct {
	dbase *db.DB
	log   logger.Logger
}

// NewAccountProxy :
// Create a new proxy on the input `dbase` to access the
// properties of accounts as registered in the DB.
// In case the provided DB is `nil` a panic is issued.
//
// The `dbase` represents the database whose accesses are
//
// The `log` will be used to notify information so that
// we can have an idea of the activity of this component.
// One possible example is for timing the requests.
//
// Returns the created proxy.
func NewAccountProxy(dbase *db.DB, log logger.Logger) AccountProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create accounts proxy from invalid DB"))
	}

	return AccountProxy{dbase, log}
}

// Accounts :
// Allows to fetch the list of accounts currently registered
// in the DB. This defines how many unique players already
// have created at least an account in a universe.
//
// Returns the list of accounts along with any errors. Note
// that in case the error is not `nil` the returned list is
// to be ignored.
func (p *AccountProxy) Accounts() ([]Account, error) {
	return nil, fmt.Errorf("Not implemented")
}

// Characters :
// Return the list of universes into which the input user
// is registered. The input stirng is interpreted as the
// identifier of a player's account and use to query the
// corresponding information in the database.
//
// The `user` is a string representing the identifier of
// the account for this user.
//
// Returns the list of players' data registered for this
// account along with any error. In case the error is not
// `nil` the value of the array should be ignored.
func (p *AccountProxy) Characters(user string) ([]Player, error) {
	return nil, fmt.Errorf("Not implemented")
}

// Planets :
// Return a list of planets associated to the player for
// the relevant universe and account. It queries the DB
// to fetch the relevant data and return it through an
// array of planets.
//
// The `player` describes the account for which planets
// should be fetched. We assume that it contains valid
// data. If this is not the case no planets will likely
// be retrieved.
//
// Returns the list of planets for this account along
// with any error. In case the error is not `nil` the
// value of the array should be ignored.
func (p *AccountProxy) Planets(player Player) ([]Planet, error) {
	// /accounts/account_id/player_id/planets
	return nil, fmt.Errorf("Not implemented")
}

// Researches :
// Return a list of the current researches developed on
// the account of the specified player. It queries the
// DB to fetch the relevant data and return it through
// an array.
//
// The `player` describes the account for which researches
// should be fetched. We assume that it contains valid
// data. If this is not the case no researches will likely
// be retrieved.
//
// Returns the list of researches for this account
// along with any error. In case the error is not
// `nil` the value of the array should be ignored.
func (p *AccountProxy) Researches(player Player) ([]Research, error) {
	// /accounts/account_id/player_id/researches
	return nil, fmt.Errorf("Not implemented")
}

// Fleets :
// Return a list of the current fleets deployed on the
// account of the player. It is a convenience method
// compared to fetching the fleet of a single planet
// and looping on all the planets of the account.
// The internal DB is queried to fetch the relevant
// information.
//
// The `player` describes the account for which fleets
// should be fetched. We assume that it contains valid
// data. If this is not the case no fleets will likely
// be retrieved.
//
// Returns the list of fleets for this account along
// with any error. In case the error is not `nil` the
// value of the array should be ignored.
func (p *AccountProxy) Fleets(player Player) ([]Fleet, error) {
	// /accounts/account_id/player_id/fleets
	return nil, fmt.Errorf("Not implemented")
}
