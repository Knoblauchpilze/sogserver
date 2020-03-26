package data

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// UniverseProxy :
// Intended as a wrapper to access properties of universes
// and retrieve data from the database. This helps hiding
// the complexity of how the data is laid out in the `DB`
// and the precise name of tables from the exterior world.
//
// The `dbase` is the database that is wrapped by this
// object. It is checked for consistency upon building the
// wrapper.
//
// The `log` allows to perform display to the user so as
// to inform of potential issues and debug information to
// the outside world.
type UniverseProxy struct {
	dbase *db.DB
	log   logger.Logger
}

// NewUniverseProxy :
// Create a new proxy on the input `dbase` to access the
// properties of universes as registered in the DB. In
// case the provided DB is `nil` a panic is issued.
// Information in the following thread helped shape this
// component:
// https://www.reddit.com/r/golang/comments/9i5cpg/good_approach_to_interacting_with_databases/
//
// The `dbase` represents the database whose accesses are
// to be wrapped.
//
// The `log` will be used to notify information so that
// we can have an idea of the activity of this component.
// One possible example is for timing the requests.
//
// Returns the created proxy.
func NewUniverseProxy(dbase *db.DB, log logger.Logger) UniverseProxy {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create universes proxy from invalid DB"))
	}

	return UniverseProxy{dbase, log}
}

// Planets :
// Return a list of planets associated registered in the
// universe specified in input. It queries the DB to fetch
// the relevant data. Note that only the created planets
// will be returned.
//
// The `uni` describes the universe for which planetes
// should be returned. In case it does not represent a
// valid universe the returned planets list will most
// likely be empty.
//
// Returns the list of planets for this universe along
// with any error. In case the error is not `nil` the
// value of the array should be ignored.
func (p *UniverseProxy) Planets(uni Universe) ([]Planet, error) {
	// /universes/universe_id/planets
	return nil, fmt.Errorf("Not implemented")
}

// Buildings :
// Return a list of the buildings currently built on the
// planet specified as input. This will automatically use
// the universe associated to the planet to fetch needed
// data.
//
// The `planet` defines the planet for which the list of
// buildings should be fetched.
//
// Returns a list of buildings for the specified planet.
// If the planet's identifier is not valid the return
// list will most likely be invalid. In case no buildings
// are built on the planet the output list will also be
// empty. It should be ignored in case the error is not
// `nil`.
func (p *UniverseProxy) Buildings(planet Planet) ([]Building, error) {
	// /universes/universe_id/planet_id/buildings
	return nil, fmt.Errorf("Not implemented")
}

// Defenses :
// Return a list of the defenses currently built on the
// planet specified as input. This will automatically use
// the universe associated to the planet to fetch needed
// data.
//
// The `planet` defines the planet for which the list of
// defenses should be fetched.
//
// Returns a list of defenses for the specified planet.
// If the planet's identifier is not valid the return
// list will most likely be invalid. In case no defenses
// are built on the planet the output list will also be
// empty. It should be ignored in case the error is not
// `nil`.
func (p *UniverseProxy) Defenses(planet Planet) ([]Defense, error) {
	// /universes/universe_id/planet_id/defenses
	return nil, fmt.Errorf("Not implemented")
}

// Ships :
// Similar to `Defenses` but returns the number of ships
// currently available on the specified planet. Note that
// it does not include the fleets that are moving towards
// the planet or leaving from it.
//
// The `planet` defines the planet for which the list of
// ships should be fetched.
//
// Returns a list of the ships currently available on the
// specified planet. The list is empty if no ships are
// available and should be ignored if the associated error
// is not `nil`.
func (p *UniverseProxy) Ships(planet Planet) ([]Ship, error) {
	// /universes/universe_id/planet_id/ships
	return nil, fmt.Errorf("Not implemented")
}

// Fleets :
// Similar to the `Ships` method but returns the list of
// fleets that are directed towards or start from this
// planet. Note that it accounts both for friendly but
// also enemy fleets.
//
// The `planet` defines the planet for which the list of
// fleets should be fetched.
//
// Returns a list of the fleets currently directed towards
// the planet or leaving from it no matter their objectives.
// The list should be ignored if the error is not `nil`.
func (p *UniverseProxy) Fleets(planet Planet) ([]Fleet, error) {
	// /universes/universe_id/planet_id/fleets
	return nil, fmt.Errorf("Not implemented")
}
