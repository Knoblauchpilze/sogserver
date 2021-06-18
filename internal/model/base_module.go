package model

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// DBModule :
// As most of the modules defined in this package rely
// on accessing information from the DB, this handler
// defines a common interface to standardize the way
// modules will use the DB. This will be used in the
// server to keep collections of module rather than a
// single instance of each element.
//
// This interface is composed of a single method that
// can be called to launch the refresh of the data of
// the module from the provided DB. Note that the user
// can force the refresh if needed.
type DBModule interface {
	Init(proxy db.Proxy, force bool) error
}

// ErrNotInitialized :
// Used to indicate the a DB module failed to correctly
// initialize its components from the provided DB.
var ErrNotInitialized = fmt.Errorf("unable to initialize DB module")

// ErrInconsistentDB :
// Used to indicate that some consistency errors were
// detected when loading data from the DB.
var ErrInconsistentDB = fmt.Errorf("detected inconsistencies in DB model")

// ErrInvalidID :
// Used to indicate that the provided identifier is not
// valid according to the internal data.
var ErrInvalidID = fmt.Errorf("invalid identifier")

// baseModule :
// This module allows to group the logging behavior of
// the DB modules by providing a concept of prefix that
// can be applied to log coming from a specific module.
// This helps grouping the failure more precisely.
//
// The `log` defines a convenient way to notify errors
// and information to the user.
//
// The `module` defines a string that will be used as
// a prefix to all messages passed to the logger.
//
// The `level` allows to precisely configure how much
// logs will be produced by this module by preventing
// some severity messages to be passed to the logging
// layer.
type baseModule struct {
	log    logger.Logger
	module string
	level  logger.Severity
}

// newBaseModule :
// Used to create a new base module from the logger and
// the module name (which will be prefixing any log req
// to this module).
//
// The `log` defines that a way to notify information
// and errors.
//
// The `module` defines a string that will be used as a
// prefix to all log messages triggered by this module.
//
// Returns the created module.
func newBaseModule(log logger.Logger, module string) baseModule {
	return baseModule{
		log:    log,
		module: module,
		level:  logger.Debug,
	}
}

// trace :
// Used as a wrapper around the internal logger object to
// benefit from the module defined for this element along
// with the log level.
// Calls that suceed in the log level verification are
// forwarded to the underlying logging layer.
//
// The `level` defines the severity of the message.
//
// The `msg` defines the content of the log to display.
func (bm *baseModule) trace(level logger.Severity, msg string) {
	// Log the message only if its severity is greater than
	// the current authorized log level.
	if level >= bm.level {
		bm.log.Trace(level, bm.module, msg)
	}
}

// fetchIDs :
// Used to perform the query and retrieve all the results
// in a single slice of strings. This can only be applied
// to the case where the query returns a single element
// of type string.
//
// The `query` defines the query to perform.
//
// The `proxy` defines the proxy to use to perform the
// query.
//
// Returns the list of identifiers fetched from the DB
// along with any errors.
func (bm *baseModule) fetchIDs(query db.QueryDesc, proxy db.Proxy) ([]string, error) {
	// Perform the query.
	dbRes, err := proxy.FetchFromDB(query)

	if err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Unable to fetch IDs (err: %v)", err))
		return []string{}, err
	}
	defer dbRes.Close()

	if dbRes.Err != nil {
		bm.trace(logger.Error, fmt.Sprintf("Invalid query to initialize IDs (err: %v)", dbRes.Err))
		return []string{}, dbRes.Err
	}

	// Fetch identifiers.
	var ID string
	IDs := make([]string, 0)

	for dbRes.Next() {
		err := dbRes.Scan(&ID)

		if err != nil {
			bm.trace(logger.Error, fmt.Sprintf("Failed to initialize ID from row (err: %v)", err))
			continue
		}

		IDs = append(IDs, ID)
	}

	return IDs, nil
}
