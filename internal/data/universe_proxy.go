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
// properties of universes as registered in the DB.
// In case the provided DB is `nil` a panic is issued.
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
