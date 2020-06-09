package game

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

// ErrInvalidPassword : Indicates that the account has an invalid password.
var ErrInvalidPassword = fmt.Errorf("Empty password provided for account")

// ErrInvalidMail : Indicates that the mail does not seem to have a valid syntax.
var ErrInvalidMail = fmt.Errorf("Invalid syntax for mail")

// ErrDuplicatedMail : Indicates that the mail associated to an account already exists.
var ErrDuplicatedMail = fmt.Errorf("Already existing mail")

// valid :
// Determines whether the account is valid. By valid we only mean
// obvious syntax errors.
//
// Returns any error or `nil` if the account seems valid.
func (a *Account) valid() error {
	// Note that we *verified* the following regular expression
	// does compile so we don't check for errors.
	exp, _ := regexp.Compile("^[a-zA-Z0-9]*[a-zA-Z0-9_.+-][a-zA-Z0-9]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$")

	if !validUUID(a.ID) {
		return ErrInvalidElementID
	}
	if a.Name == "" {
		return ErrInvalidName
	}
	if a.Password == "" {
		return ErrInvalidPassword
	}
	if !exp.MatchString(a.Mail) {
		return ErrInvalidMail
	}

	return nil
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
	if !validUUID(a.ID) {
		return a, ErrInvalidElementID
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
				Values: []interface{}{a.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return a, err
	}
	if dbRes.Err != nil {
		return a, dbRes.Err
	}

	// Scan the account's data.
	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return a, ErrElementNotFound
	}

	err = dbRes.Scan(
		&a.ID,
		&a.Mail,
		&a.Name,
		&a.Password,
	)

	// Make sure that it's the only account.
	if dbRes.Next() {
		return a, ErrDuplicatedElement
	}

	return a, err
}

// SaveToDB :
// Used to save the content of this account to
// the DB. In case an error is raised during the
// operation a comprehensive error is returned.
//
// The `proxy` allows to access to the DB.
//
// Returns any error.
func (a *Account) SaveToDB(proxy db.Proxy) error {
	// Check consistency.
	if err := a.valid(); err != nil {
		return err
	}

	// Create the query and execute it: we will
	// use the dedicated handler to provide a
	// comprehensive error.
	query := db.InsertReq{
		Script: "create_account",
		Args:   []interface{}{a},
	}

	err := proxy.InsertToDB(query)

	return a.analyzeDBError(err)
}

// UpdateInDB :
// Used to update the content of the account in
// the DB. Only part of the account's data can
// be updated as specified by this function.
//
// The `proxy` allows to access to the DB.
//
// Returns any error.
func (a *Account) UpdateInDB(proxy db.Proxy) error {
	// Make sure that at least one of the name, mail
	// or password are valid. We want also to be sure
	// that in case the mail is provided it is valid.
	exp, _ := regexp.Compile("^[a-zA-Z0-9]*[a-zA-Z0-9_.+-][a-zA-Z0-9]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$")

	if a.Name == "" && a.Password == "" {
		// No mail nor password, check the case of the
		// mail: it should be provided and valid.
		if a.Mail == "" || !exp.MatchString(a.Mail) {
			return ErrInvalidUpdateData
		}
	}

	// Make sure that the email is valid if it is
	// provided: we tolerate empty mail though as
	// we know that either the name or the pwd
	// is provided.
	if a.Name != "" && !exp.MatchString(a.Mail) {
		return ErrInvalidUpdateData
	}

	// Create the query and execute it. In a
	// similar way we need to provide some
	// analysis of any error.
	query := db.InsertReq{
		Script: "update_account",
		Args: []interface{}{
			a.ID,
			struct {
				Name     string `json:"name"`
				Mail     string `json:"mail"`
				Password string `json:"password"`
			}{
				Name:     a.Name,
				Mail:     a.Mail,
				Password: a.Password,
			},
		},
	}

	err := proxy.InsertToDB(query)

	return a.analyzeDBError(err)
}

// analyzeDBError :
// used to perform the analysis of a DB error based on
// the structure of the accounts' table to produce a
// comprehensive error of what went wrong.
//
// The `err` defines the error to analyze.
//
// Returns a comprehensive error or the input error if
// nothing can be extracted from the input data.
func (a *Account) analyzeDBError(err error) error {
	// In case the error is not a `db.Error` we can't do
	// anything, so just return the input error.
	dbe, ok := err.(db.Error)
	if !ok {
		return err
	}

	// Otherwise we can try to make some sense of it.
	dee, ok := dbe.Err.(db.DuplicatedElementError)
	if ok {
		switch dee.Constraint {
		case "accounts_pkey":
			return ErrDuplicatedElement
		case "accounts_name_key":
			return ErrInvalidName
		case "accounts_mail_key":
			return ErrDuplicatedMail
		}
	}

	return dbe
}
