package game

import "fmt"

// attack :
// Used to perform the attack of the input planet
// by this fleet. Note that only the units that
// are deployed on the planet at this moment will
// be included in the fight.
//
// The `p` represents the element to be attacked.
// It can either be a planet or a moon.
//
// Return any error along with the name of the
// script to execute to finalize the execution of
// the fleet.
func (f *Fleet) attack(p *Planet) (string, error) {
	// TODO: Implement this.
	return "fleet_return_to_base", fmt.Errorf("Not implemented")
}
