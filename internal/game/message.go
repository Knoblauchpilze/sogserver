package game

import (
	"oglike_server/pkg/db"
	"time"
)

// Message :
// Defines a message that is directed towards a player.
// Messages allow to notify the player of certain info
// such as a fleet's arrival to its destination or the
// potential spying operation of another player.
// The message contains a description that allows to
// actually display its content in the client UI.
//
// The `ID` defines the identifier of the message, as
// registered in the DB. It uniquely identifies the
// message.
//
// The `Player` defines the identifier of the player
// that is receiving this message.
//
// The `Content` actually defines the template of the
// message. It contains placeholders where the data
// from the `arguments` will be injected.
//
// The `CreatedAt` defines the creation time of the
// message.
//
// The `Arguments` define an ordered list of args to
// be injected into the content of the message. This
// list might be empty if the message does not use
// any argument.
type Message struct {
	ID        string    `json:"id"`
	Player    string    `json:"player"`
	Message   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Arguments []string  `json:"arguments"`
}

// NewMessageFromDB :
// Used to fetch the content of the message from
// the input DB and populate all internal fields
// from it. In case the DB cannot be fetched or
// some errors are encoutered, the return value
// will include a description of the error.
//
// The `ID` defines the ID of the message to get.
// It is fetched from the DB and is assumed to
// refer to an existing message.
//
// The `data` allows to actually perform the DB
// requests to fetch the message's data.
//
// Returns the message as fetched from the DB
// along with any errors.
func NewMessageFromDB(ID string, data Instance) (Message, error) {
	// Create the message.
	m := Message{
		ID: ID,
	}

	// Consistency.
	if !validUUID(m.ID) {
		return m, ErrInvalidElementID
	}

	// Fetch the message's content.
	err := m.fetchGeneralInfo(data)
	if err != nil {
		return m, err
	}

	err = m.fetchArguments(data)
	if err != nil {
		return m, err
	}

	return m, err
}

// fetchGeneralInfo :
// Used internally when building a message from
// the DB to retrieve general information such
// as the type of the message and its creation
// date.
//
// The `data` defines the object to access the
// DB.
//
// Returns any error.
func (m *Message) fetchGeneralInfo(data Instance) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
			"player",
			"message",
			"created_at",
		},
		Table: "messages_players",
		Filters: []db.Filter{
			{
				Key:    "id",
				Values: []interface{}{m.ID},
			},
		},
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}
	if dbRes.Err != nil {
		return dbRes.Err
	}

	// Scan the account's data.
	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return ErrElementNotFound
	}

	err = dbRes.Scan(
		&m.ID,
		&m.Player,
		&m.Message,
		&m.CreatedAt,
	)

	// Make sure that it's the only message.
	if dbRes.Next() {
		return ErrDuplicatedElement
	}

	return err
}

// fetchArguments :
// Similar to `fetchGeneralInfo` but used to
// fetch the arguments to inject in the message
// text.
//
// The `data` defines the object to access the
// DB.
//
// Returns any error.
func (m *Message) fetchArguments(data Instance) error {
	m.Arguments = make([]string, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"argument",
		},
		Table: "messages_arguments",
		Filters: []db.Filter{
			{
				Key:    "message",
				Values: []interface{}{m.ID},
			},
		},
		Ordering: "order by position",
	}

	dbRes, err := data.Proxy.FetchFromDB(query)
	defer dbRes.Close()

	// Check for errors.
	if err != nil {
		return err
	}
	if dbRes.Err != nil {
		return dbRes.Err
	}

	// Populate the return value.
	var arg string

	for dbRes.Next() {
		err = dbRes.Scan(
			&arg,
		)

		if err != nil {
			return err
		}

		m.Arguments = append(m.Arguments, arg)
	}

	return nil
}
