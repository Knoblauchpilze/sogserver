package model

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"regexp"
)

// MessagesModule :
// This structure allows to manipulate the data related to
// the messages that can be received by a player as part of
// the game. Each message has a specific structure and can
// be linked to some other messages. It can also have some
// arguments to fill the placeholders defined in the body
// of the message.
//
// The `descs` defines the list of structure for each msg
// defined in the game.
type MessagesModule struct {
	associationTable
	baseModule

	descs map[string]messageDesc
}

// Message :
// Defines a message in the game. A message can be
// received by the player by doing some actions.
// Each message has a list of placeholders to allow
// some form of parameterization.
//
// The `ID` defines its identifier in the DB. It is
// used in most other table to actually reference a
// message.
//
// The `Type` defines a general type for this msg
// which usually refer to its content.
//
// The `Name` defines the display name for the msg.
//
// The `Content` represents the string for this
// message which will be displayed in the player's
// mailbox.
//
// The `Arguments` defines the list of arguments
// that can be used to customize the message.
type Message struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"`
	Name      string   `json:"name"`
	Content   string   `json:"content"`
	Arguments []string `json:"arguments"`
}

// messageDesc :
// Defines the structure of a message as fetched
// from the DB.
//
// The `kind` defines a general string defining
// the category attached to a message.
//
// The `content` defines the string representing
// the content for this message.
//
// The `arguments` defines the list of arguments
// to customize this message.
type messageDesc struct {
	kind      string
	content   string
	arguments []string
}

// newMessageDesc :
// Used to build the message description from the
// input message. It will mainly analyze the content
// of the message to extract arguments.
//
// The `msg` defines the input message.
//
// Returns the created message description.
func newMessageDesc(msg Message) messageDesc {
	desc := messageDesc{
		kind:      msg.Type,
		content:   msg.Content,
		arguments: []string{},
	}

	// Find arguments.
	re := regexp.MustCompile("\\$[A-Z_]+")

	args := re.FindAllString(desc.content, -1)
	if args != nil {
		desc.arguments = args
	}

	return desc
}

// NewMessagesModule :
// Used to create a new messages module which is
// initialized with no content. The module will
// stay invalid until the `init` method is called
// with a valid DB.
//
// The `log` defines the logging layer to forward
// to the base `baseModule` element.
func NewMessagesModule(log logger.Logger) *MessagesModule {
	return &MessagesModule{
		associationTable: newAssociationTable(),
		baseModule:       newBaseModule(log, "messages"),

		descs: nil,
	}
}

// valid :
// Refinement of the base `associationTable` valid method
// in order to perform some checks on additional tables
// defined for this module.
//
// Returns `true` if the association table is valid and
// the internal resources as well.
func (mm *MessagesModule) valid() bool {
	return mm.associationTable.valid() && len(mm.descs) > 0
}

// Init :
// Implementation of the `DBModule` interface to allow
// fetching information from the input DB and load to
// local memory.
//
// The `proxy` represents the main data source to use
// to initialize the messages data.
//
// The `force` allows to erase any existing information
// and reload everything from the DB.
//
// Returns any error.
func (mm *MessagesModule) Init(proxy db.Proxy, force bool) error {
	// Prevent reload if not needed.
	if mm.valid() && !force {
		return nil
	}

	// Initialize internal values.
	mm.descs = make(map[string]messageDesc)

	// Initialize general information for messages.
	err := mm.fetchMessages(proxy)
	if err != nil {
		mm.trace(logger.Error, fmt.Sprintf("Could not initialize module (err: %v)", err))
		return err
	}

	return nil
}

// fetchMessages :
// Used internally to fetch the messages from the DB.
// The input proxy will be used to access the info
// and populate internal tables.
//
// The `proxy` defines a convenient way to access to
// the DB.
//
// Returns any error.
func (mm *MessagesModule) fetchMessages(proxy db.Proxy) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"mi.id",
			"mt.type",
			"mi.name",
			"mi.content",
		},
		Table:   "messages_ids mi inner join messages_types mt on mi.type = mt.id",
		Filters: []db.Filter{},
	}

	rows, err := proxy.FetchFromDB(query)
	defer rows.Close()

	if err != nil {
		mm.trace(logger.Error, fmt.Sprintf("Unable to initialize messages (err: %v)", err))
		return ErrNotInitialized
	}
	if rows.Err != nil {
		mm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize messages (err: %v)", rows.Err))
		return ErrNotInitialized
	}

	// Analyze the query and populate internal values.
	var msg Message

	override := false
	inconsistent := false

	for rows.Next() {
		err := rows.Scan(
			&msg.ID,
			&msg.Type,
			&msg.Name,
			&msg.Content,
		)

		if err != nil {
			mm.trace(logger.Error, fmt.Sprintf("Failed to initialize message from row (err: %v)", err))
			inconsistent = true

			continue
		}

		// Check whether a message with this identifier exists.
		if mm.existsID(msg.ID) {
			mm.trace(logger.Error, fmt.Sprintf("Prevented override of message \"%s\"", msg.ID))
			override = true

			continue
		}

		// Register this message in the association table.
		err = mm.registerAssociation(msg.ID, msg.Name)
		if err != nil {
			mm.trace(logger.Error, fmt.Sprintf("Cannot register message \"%s\" (id: \"%s\") (err: %v)", msg.Name, msg.ID, err))
			inconsistent = true

			continue
		}

		// Register the directed and hostile status.
		_, ok := mm.descs[msg.ID]

		if ok {
			mm.trace(logger.Error, fmt.Sprintf("Overriding message structure for \"%s\"", msg.ID))
			override = true
		}

		mm.descs[msg.ID] = newMessageDesc(msg)
	}

	if override || inconsistent {
		return ErrInconsistentDB
	}

	return nil
}

// GetMessageFromID :
// Used to retrieve information on the message
// that corresponds to the input identifier. If
// no such message exists an error is returned.
//
// The `id` defines the identifier of the msg
// to fetch.
//
// Returns the description for this message and
// any errors.
func (mm *MessagesModule) GetMessageFromID(id string) (Message, error) {
	// Find this element in the association table.
	if !mm.existsID(id) {
		mm.trace(logger.Error, fmt.Sprintf("Cannot retrieve desc for message \"%s\"", id))
		return Message{}, ErrNotFound
	}

	// We assume at this point that the identifier (and
	// thus the name) both exists so we discard errors.
	name, _ := mm.getNameFromID(id)

	desc := mm.descs[id]

	res := Message{
		ID:        id,
		Name:      name,
		Type:      desc.kind,
		Content:   desc.content,
		Arguments: desc.arguments,
	}

	return res, nil
}

// GetMessageFromName :
// Calls internally the `GetMessageFromID` in order
// to forward the call to the above method. Failures
// happen in similar cases.
//
// The `name` defines the name of the message for
// which a description should be provided.
//
// Returns the description for this message along
// with any errors.
func (mm *MessagesModule) GetMessageFromName(name string) (Message, error) {
	// Find this element in the association table.
	id, err := mm.GetIDFromName(name)
	if err != nil {
		mm.trace(logger.Error, fmt.Sprintf("Cannot retrieve desc for message \"%s\" (err: %v)", name, err))
		return Message{}, ErrNotFound
	}

	return mm.GetMessageFromID(id)
}

// Messages :
// Used to retrieve the messages matching the input
// filters from the data model. Note that if the DB
// has not yet been polled to retrieve data, we will
// return an error.
//
// The `proxy` defines the DB to use to fetch the
// messages description.
//
// The `filters` represent the list of filters to apply
// to the data fecthing. This will select only part of
// all the available objectives.
//
// Returns the list of objectives matching the filters
// along with any error.
func (mm *MessagesModule) Messages(proxy db.Proxy, filters []db.Filter) ([]Message, error) {
	// Initialize the module if for some reasons it is still
	// not valid.
	if !mm.valid() {
		err := mm.Init(proxy, true)
		if err != nil {
			return []Message{}, err
		}
	}

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"mi.id",
		},
		Table:   "messages_ids mi inner join messages_types mt on mi.type = mt.id",
		Filters: filters,
	}

	IDs, err := mm.fetchIDs(query, proxy)
	if err != nil {
		mm.trace(logger.Error, fmt.Sprintf("Unable to fetch messages (err: %v)", err))
		return []Message{}, err
	}

	// Now build the data from the fetched identifiers.
	descs := make([]Message, 0)
	for _, ID := range IDs {
		name, err := mm.getNameFromID(ID)
		if err != nil {
			mm.trace(logger.Error, fmt.Sprintf("Unable to fetch message \"%s\" (err: %v)", ID, err))
			continue
		}

		desc := Message{
			ID:   ID,
			Name: name,
		}

		d, ok := mm.descs[ID]
		if !ok {
			mm.trace(logger.Error, fmt.Sprintf("Unable to fetch structure for message \"%s\"", ID))
			continue
		} else {
			desc.Type = d.kind
			desc.Content = d.content
			desc.Arguments = d.arguments
		}

		descs = append(descs, desc)
	}

	return descs, nil
}
