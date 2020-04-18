package data

import "fmt"

// buildingModule :
// Used to provide common functionailities to access the
// level of a building on a given planet by its name. It
// is sometimes more interesting than working only with
// identifiers.
//
// The `buildings` defines the table of association of
// any building name to the corresponding identifier in
// the DB.
//
// The `planet` defines the planet for which level of
// buildings should be fetched.
type buildingModule struct {
	buildings map[string]string
	planet    Planet
}

// getIDOf :
// Used to retrieve the identifier of the building with
// the specified name. If no such building can be found,
// an error is raised.
//
// The `building` defines the name of the building for
// which the identifier should be retrieved.
//
// Returns a string representing the identifier of the
// building with the input name along with any error in
// case the name does not correspond to any known item.
func (bm buildingModule) getIDOf(building string) (string, error) {
	id, ok := bm.buildings[building]

	if !ok {
		return "", fmt.Errorf("Cannot find building with name \"%s\"", building)
	}

	return id, nil
}

// getLevelOf :
// Used to retrieve the level of the building with the
// specified identifier on the planet associated to this
// building module.
//
// The `building` should correspond to a valid identifier
// of a building from the DB. If the identifier is not
// valid, an error will be raised.
//
// Returns the level of the building on the planet (which
// can be `0` if the planet does not define such a item)
// along with any error.
func (bm buildingModule) getLevelOf(building string) (int, error) {
	id, err := bm.getIDOf(building)

	if err != nil {
		return 0, fmt.Errorf("Unable to retrieve building \"%s\" on \"%s\" (err: %v)", building, bm.planet.ID, err)
	}

	// Try to find the building in the planet's elements.
	for _, b := range bm.planet.Buildings {
		if b.ID == id {
			return b.Level, nil
		}
	}

	// Could not find this building in the planet's items.
	return 0, nil
}
