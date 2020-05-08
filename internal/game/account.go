package game

import (
	"fmt"
	"oglike_server/internal/model"
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
func NewAccountFromDB(ID string, data model.Instance) (Account, error) {
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

	// Create the query and execute it.
	query := db.InsertReq{
		Script: "create_account",
		Args:   []interface{}{a},
	}

	err := proxy.InsertToDB(query)

	// Analyze the error in order to provide some
	// comprehensive message.
	dbe, ok := err.(db.Error)
	if !ok {
		return err
	}

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
