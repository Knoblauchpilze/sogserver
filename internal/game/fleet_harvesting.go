package game

import (
	"fmt"
	"oglike_server/pkg/db"
)

// harvest :
// Used to perform the harvesting of resources that
// are dispersed in the debris field specified by
// the target of the fleet.
//
// The `data` allows to access to the DB.
//
// Return any error along with the name of the
// script to execute to finalize the execution of
// the fleet.
func (f *Fleet) harvest(data Instance) (string, error) {
	// The harvesting operation requires to harvest a
	// part or all the resources that are dispersed in
	// a given debris field. The capacity available to
	// harvest is given by the cargo space of all the
	// recyclers belonging to the fleet.
	// We will assume that this function is only called
	// when the fleet is actually ready to process the
	// debris field and we won't check whether it is
	// the case.

	// If the fleet is not returning yet, process the
	// harvesting operation.
	if !f.returning {
		hp, err := newHarvestingProps(f, data.Ships)
		if err != nil {
			return "", ErrUnableToSimulateFleet
		}

		// Retrieve the description of the debris
		// field from the DB.
		df, err := NewDebrisFieldFromDB(f.TargetCoords, f.Universe, data)
		if err != nil {
			return "", ErrUnableToSimulateFleet
		}

		// Harvest this field.
		hp.harvest(df)

		// Now perform the update of resources to
		// the DB given what has been harvested.
		query := db.InsertReq{
			Script: "fleet_harvesting_success",
			Args: []interface{}{
				f.ID,
				df.ID,
				hp.collected,
			},
			Verbose: true,
		}

		err = data.Proxy.InsertToDB(query)
		if err != nil {

			fmt.Println(fmt.Sprintf("Error: %v", err))
			return "", err
		}
	}

	// The `fleet_return_to_base` script is actually
	// safe to call even in the case of a fleet that
	// should not yet return to its base. So we will
	// abuse this fact.
	return "fleet_return_to_base", nil
}
