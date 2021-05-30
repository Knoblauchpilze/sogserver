package game

import (
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"time"
)

// DebrisField :
// Defines a debris field in the game. Such a field is
// linked to a position and is created each time a fleet
// interacts with a planet or moon. It can receive some
// resources in case ships get destroyed during a battle.
// The debris object describe the list of resources that
// are dispersed in the field.
//
// The `ID` defines the identifier of the debris fields.
//
// The `Universe` defines the parent universe to which
// this debris field belong..
//
// The `Coordinates` defines the position of the debris
// field within its parent universe.
//
// The `Resources` defines the list of resources that
// are dispersed in the field. Might be empty in case
// the field has been harvested recently for example.
//
// The `CreatedAt` defines the creation time for this
// field.
type DebrisField struct {
	ID          string                 `json:"id"`
	Universe    string                 `json:"universe"`
	Coordinates Coordinate             `json:"coordinates"`
	Resources   []model.ResourceAmount `json:"resources"`
	CreatedAt   time.Time              `json:"created_at"`
}

// NewDebrisFieldFromDB :
// Used to fetch the content of the debris field in
// the DB from its identifier.
//
// The `ID` defines the ID of the debris field to get.
// It is fetched from the DB and should refer to an
// existing debris field.
//
// The `data` allows to actually perform the DB
// requests to fetch the debris field's data.
//
// Returns the debris field as fetched from the DB along
// with any errors.
func NewDebrisFieldFromDB(ID string, data Instance) (DebrisField, error) {
	// Create the fleet.
	df := DebrisField{
		ID: ID,
	}

	// Consistency.
	if !validUUID(df.ID) {
		return df, ErrInvalidElementID
	}

	// Fetch the debris field's content.
	err := df.fetchGeneralInfo(data)
	if err != nil {
		return df, err
	}

	err = df.fetchResources(data)
	if err != nil {
		return df, err
	}

	return df, nil
}

// NewDebrisFieldFromCoords :
// Used to fetch the content of the debris field in
// the DB from the input coordinates. No checks are
// performed on the coordinates relatively to the
// parent universe but it is considered an error if
// no debris field is returned.
//
// The `coordinates` defines the position of the
// field to fetch.
//
// The `universe` defines the identifier of the
// parent universe
//
// The `data` allows to actually perform the DB
// requests to fetch the field's data.
//
// Returns the field as fetched from the DB along
// with any errors.
func NewDebrisFieldFromCoords(coordinates Coordinate, universe string, data Instance) (DebrisField, error) {
	// Create the debris field.
	df := DebrisField{
		Universe:    universe,
		Coordinates: coordinates,
	}

	// Consistency.
	if !validUUID(universe) {
		return df, ErrInvalidElementID
	}

	// Fetch the debris field's content.
	err := df.fetchDescription(data)
	if err != nil {
		return df, err
	}

	err = df.fetchResources(data)
	if err != nil {
		return df, err
	}

	return df, nil
}

// fetchGeneralInfo :
// Used internally when building a debris field from
// the DB to retrieve general information such as the
// coordinates.
//
// The `data` defines the object to access the DB.
//
// Returns any error.
func (df *DebrisField) fetchGeneralInfo(data Instance) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"universe",
			"galaxy",
			"solar_system",
			"position",
			"created_at",
		},
		Table: "debris_fields",
		Filters: []db.Filter{
			{
				Key:    "id",
				Values: []interface{}{df.ID},
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

	// Scan the fleet's data.
	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return ErrElementNotFound
	}

	var g, s, p int

	err = dbRes.Scan(
		&df.Universe,
		&g,
		&s,
		&p,
		&df.CreatedAt,
	)

	var errC error
	df.Coordinates, errC = newCoordinate(g, s, p, Debris)
	if errC != nil {
		return errC
	}

	// Make sure that it's the only fleet.
	if dbRes.Next() {
		return ErrDuplicatedElement
	}

	return err
}

// fetchDescription :
// Used to fetch the relevant data for this debris
// field. We assume that the universe and coords
// are already assigned.
//
// The `data` allows to access to the DB.
//
// Returns any error.
func (df *DebrisField) fetchDescription(data Instance) error {
	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"id",
			"created_at",
		},
		Table: "debris_fields",
		Filters: []db.Filter{
			{
				Key:    "universe",
				Values: []interface{}{df.Universe},
			},
			{
				Key:    "galaxy",
				Values: []interface{}{df.Coordinates.Galaxy},
			},
			{
				Key:    "solar_system",
				Values: []interface{}{df.Coordinates.System},
			},
			{
				Key:    "position",
				Values: []interface{}{df.Coordinates.Position},
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

	// Scan the fleet's data.
	atLeastOne := dbRes.Next()
	if !atLeastOne {
		return ErrElementNotFound
	}

	// Populate the debris field' properties.
	err = dbRes.Scan(
		&df.ID,
		&df.CreatedAt,
	)

	// Make sure that it's the only debris field.
	if dbRes.Next() {
		return ErrDuplicatedElement
	}

	return err
}

// fetchResources :
// Used to fetch the resources dispersed in the
// debris field.
//
// The `data` allows to access to the DB.
//
// Returns any error.
func (df *DebrisField) fetchResources(data Instance) error {
	df.Resources = make([]model.ResourceAmount, 0)

	// Create the query and execute it.
	query := db.QueryDesc{
		Props: []string{
			"res",
			"amount",
		},
		Table: "debris_fields_resources",
		Filters: []db.Filter{
			{
				Key:    "field",
				Values: []interface{}{df.ID},
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

	// Populate the return value.
	var ra model.ResourceAmount

	for dbRes.Next() {
		err = dbRes.Scan(
			&ra.Resource,
			&ra.Amount,
		)

		if err != nil {
			return err
		}

		df.Resources = append(df.Resources, ra)
	}

	return nil
}

// amountDispersed :
// Allows to fetch the total amount of resources
// dispersed in this field.
//
// Returns the total amount of resources that are
// dispersed in the field.
func (df *DebrisField) amountDispersed() float32 {
	tot := float32(0.0)

	for _, r := range df.Resources {
		tot += r.Amount
	}

	return tot
}
