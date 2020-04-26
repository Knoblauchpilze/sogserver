package model

import "fmt"

// associationTable :
// Defines a common base for most of the data model modules
// where ids should be mapped to names and vice-versa. It
// provides a simple association table and methods allowing
// to query this info.
//
// The `idsToNames` defines a map allowing to convert an
// identifier to a human readable name.
//
// The `namesToIDs` defines a way to convert a  name to its
// DB identifier.
type associationTable struct {
	idsToNames map[string]string
	namesToIDs map[string]string
}

// ErrInvalidTable :
// Indicates that the association table that is provided
// to perform an operation is not valid.
var ErrInvalidTable = fmt.Errorf("Invalid association table")

// ErrDuplicatedID :
// Indicates that a registration operation could not be
// performed because the id already exists.
var ErrDuplicatedID = fmt.Errorf("ID is duplicated")

// ErrDuplicatedName :
// Indicates that a registration operation could not be
// performed because the name already exists.
var ErrDuplicatedName = fmt.Errorf("Name is duplicated")

// ErrInvalidAssociation :
// Indicates that the association could not be performed
// because at least one of the values is invalid.
var ErrInvalidAssociation = fmt.Errorf("Association has one invalid member")

// ErrNotFound :
// Used to indicate that the searched element was not
// found.
var ErrNotFound = fmt.Errorf("Element does not exist")

// valid :
// Used to determine whether this association table is
// valid or not. We consider that the validity status
// is linked to whether there is some data in the tables.
// If at least one element is registered, we consider it
// valid. This helps determining when there was an obvious
// error when performing the initialization from the DB.
//
// Returns `true` if the module is valid.
func (at *associationTable) valid() bool {
	litn := len(at.idsToNames)
	lnti := len(at.namesToIDs)

	return litn != 0 && lnti != 0 && litn == lnti
}

// registerAssociation :
// Used to perform the insertion of the input couple key
// value in both association tables.
// Both values should not be empty and none of them should
// exist in the association tables already, otherwise an
// error is returned.
//
// The `id` defines the identifier of the association.
//
// The `name` defines the name of the association.
//
// Returns any error.
func (at *associationTable) registerAssociation(id string, name string) error {
	if len(id) == 0 || len(name) == 0 {
		return ErrInvalidAssociation
	}

	_, iok := at.idsToNames[id]
	_, nok := at.namesToIDs[id]

	if iok {
		return ErrDuplicatedID
	}
	if nok {
		return ErrDuplicatedName
	}

	at.idsToNames[id] = name
	at.namesToIDs[name] = id

	return nil
}

// existsID :
// Used to determine whether the identifier exists in
// the association table. Not that this method does
// succeed even if the association table is not valid.
//
// The `id` defines the identifier to search for.
//
// Returns `true` if the identifier exists.
func (at *associationTable) existsID(id string) bool {
	_, ok := at.idsToNames[id]
	return ok
}

// existsName :
// Used to determine whether the name exists in this
// association table. Not that this method succeeds
// even if the association table is not valid.
//
// The `name` defines the name to search for.
//
// Returns `true` if the name exists.
func (at *associationTable) existsName(name string) bool {
	_, ok := at.namesToIDs[name]
	return ok
}

// getNameFromID :
// Used to retrieve the name of an element from the
// input identifier. In case the name of the element
// is unknown or the module is not valid an error is
// returned.
//
// The `ID` defines the identifier for which the name
// should be retrieved.
//
// Returns the name of the element corresponding to
// the input identifier.
func (at *associationTable) getNameFromID(ID string) (string, error) {
	if !at.valid() {
		return "", ErrInvalidTable
	}

	name, ok := at.idsToNames[ID]

	if !ok {
		return "", ErrNotFound
	}

	return name, nil
}

// getIDFromName :
// Similar to `getNameFromID` but performs the reverse
// query. Failure scenarios are similar.
//
// The `Name` defines the element name for which the
// identifier should be retrieved.
//
// Returns the identifier of the element corresponding
// to the input name.
func (at *associationTable) getIDFromName(name string) (string, error) {
	if !at.valid() {
		return "", ErrInvalidTable
	}

	id, ok := at.namesToIDs[name]

	if !ok {
		return "", ErrNotFound
	}

	return id, nil
}
