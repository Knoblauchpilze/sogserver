package game

import "fmt"

// destroy :
// Used to attempt to destroy the moon in argument
// by this fleet. This may fail and result in the
// total destruction of the fleet.
//
// The `m` represents the moon to destroy.
//
// Return any error along with the name of the
// script to execute to finalize the execution of
// the fleet.
func (f *Fleet) destroy(m *Planet) (string, error) {
	// TODO: Implement this.
	return "fleet_return_to_base", fmt.Errorf("Not implemented")
}
