package game

import (
	"fmt"

	"github.com/google/uuid"
)

// validUUID :
// Used to check whether the input string can be interpreted
// as a valid identifier.
//
// The `id` defines the element to check.
//
// Returns `true` if this identifier is valid and `false` if
// this is not the case.
func validUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// ErrInvalidElementID : Indicates that the element has no identifier.
var ErrInvalidElementID = fmt.Errorf("Empty or invalid identifier provided for element")

// ErrDuplicatedElement : Indicates that the element identifier is not unique.
var ErrDuplicatedElement = fmt.Errorf("Invalid not unique element")

// ErrElementNotFound : Indicates that no element with specified ID exists.
var ErrElementNotFound = fmt.Errorf("Identifier does not correspond to any known element")

// ErrInvalidName : Indicates that the name is invalid or already exists.
var ErrInvalidName = fmt.Errorf("Invalid or already existing name")

// ErrInvalidUpdateData : Indicates that the update data is not valid.
var ErrInvalidUpdateData = fmt.Errorf("Invalid update data")
