package model

import (
	"fmt"
	"oglike_server/pkg/db"
	"regexp"
)

// Account :
// Defines a player's account within the OG context. It is
// not related to any universe and defines what could be
// called the root account for each player. It is then used
// each time the user wants to join a new universe so as to
// merge all these accounts in a single entity.
//
// The `ID` defines the identifier of the player, which is
// used to uniquely distinguish between two accounts.
//
// The `Name` describes the user provided name for this
// account. It can be duplicated among several accounts
// as we're using the identifier to guarantee uniqueness.
//
// The `Mail` defines the email address associated to the
// account. It can be used to make sure that no two accounts
// share the same address.
//
// The `Password` defines the password that the user should
// enter to grant access to the account.
type Account struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Mail     string `json:"mail"`
	Password string `json:"password"`
}

// ErrInvalidAccount :
// Used to indicate that the account provided in input is
// not valid.
var ErrInvalidAccount = fmt.Errorf("Invalid account with no identifier")

// ErrDuplicatedAccount :
// Used to indicate that the account's identifier provided
// input is not unique in the DB.
var ErrDuplicatedAccount = fmt.Errorf("Invalid not unique account")

// Valid :
// Used to determine whether the parameters defined for this
// account are consistent with what is expected. It is mostly
// used to check that the name is valid and that the e-mail
// address makes sense.
func (a *Account) Valid() bool {
	// Note that we *verified* the following regular expression
	// does compile so we don't check for errors.
	exp, _ := regexp.Compile("^[a-zA-Z0-9]*[a-zA-Z0-9_.+-][a-zA-Z0-9]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$")

	// Check common properties.
	return validUUID(a.ID) &&
		a.Name != "" &&
		exp.MatchString(a.Mail) &&
		a.Password != ""
}

// NewAccountFromDB :
// Used to fetch the content of the account from
// the input DB and populate all internal fields
// from it. In case the DB cannot be fetched or
// some errors are encoutered, the return value
// will include a description of the error.
//
// The `ID` defines the ID of the account to get.
// It should be fetched from the DB and is assumed
// to refer to an existing account.
//
// The `data` allows to actually perform the DB
// requests to fetch the account's data.
//
// Returns the account as fetched from the DB
// along with any errors.
func NewAccountFromDB(ID string, data Instance) (Account, error) {
	// Create the account.
	a := Account{
		ID: ID,
	}

	// Consistency.
	if a.ID == "" {
		return a, ErrInvalidAccount
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
			"mail",
			"name",
			"password",
		},
		Table: "accounts",
		Filters: []db.Filter{
			{
				Key:    "id",
				Values: []string{a.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return a, err
	}

	// Scan the account's data.
	err = dbRes.Scan(
		&a.ID,
		&a.Mail,
		&a.Name,
		&a.Password,
	)

	// Make sure that it's the only account.
	if dbRes.Next() {
		return a, ErrDuplicatedAccount
	}

	return a, nil
}
