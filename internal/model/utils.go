package model

import "github.com/google/uuid"

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
