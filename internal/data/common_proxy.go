package data

import (
	"fmt"
	"oglike_server/internal/locker"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// commonProxy :
// Intended as a common wrapper to access the main DB
// through a convenience way. It holds most of the
// common resources needed to acces the DB and notify
// errors/information to the user about processes that
// may occur while fetching data. This helps hiding
// the complexity of how the data is laid out in the
// `DB` and the precise name of tables from the rest
// of the application.
// The following link contains useful information on
// the paradigm we're following with this object:
// https://www.reddit.com/r/golang/comments/9i5cpg/good_approach_to_interacting_with_databases/
//
// The `proxy` is a reference to an object allowing
// to manipulate the main DB. It provides convenience
// methods to execute a query and insert some data
// into it.
//
// The `log` allows to perform display to the user so as
// to inform of potential issues and debug information to
// the outside world.
//
// The `lock` allows to lock specific resources when some
// data should be retrieved. For example in the case of
// a planet, one might first want to update the upgrade
// actions that are built on this planet in order to be
// sure to get up-to-date content.
// This typically include checking whether some buildings
// have reached completion and upgrading the resources
// that are present on the planet.
// Each of these actions are executed in a lazy update
// fashion where the action is created and then performed
// only when needed: we store the completion time and if
// the data needs to be accessed we upgrade it.
// This mechanism requires that when the data needs to be
// fetched for a planet an update operation should first
// be performed to ensure that the data is up-to-date.
// As no data is usually shared across elements of a same
// collection we don't see the need to lock all of them
// when a single one should be updated. Using the structure
// defined in the `ConcurrentLock` we have a way to lock
// only some elements which is exactly what we need.
//
// The `module` defines a string that will be used to
// perform some logs with a qualified service.
//
// The `data` defines a convenience object which regroups
// all the data fetched from the main DB in a easy-to-use
// object: it can be used to fetch information of various
// kind without needing to worry about how the data is
// actually fetched. It can be used to verify specific
// criteria when performing an action for example.
type commonProxy struct {
	proxy  db.Proxy
	log    logger.Logger
	lock   *locker.ConcurrentLocker
	module string
	data   model.Instance
}

// ErrInvalidResource :
// Used in case the resource requested to lock is not
// valid.
var ErrInvalidResource = fmt.Errorf("Invalid resource provided as lock")

// ErrLock :
// Used in case an error occurs when interacting with
// a lock.
var ErrLock = fmt.Errorf("Error while using resource lock")

// ErrInvalidOperation :
// Used in case the operation requested to be performed
// while a lock is held fails.
var ErrInvalidOperation = fmt.Errorf("Invalid query performed for resource")

// newCommonProxy :
// Performs the creation of a new common proxy from the
// input database and logger.
//
// The `dbase` defines the main DB that should be wrapped
// by this object.
//
// The `data` defines the data model that should be used
// by this proxy to query information.
//
// The `log` defines the logger allowing to notify errors
// or info to the user.
//
// The `module` defines a string identtofying the module
// to associate to this proxy.
//
// Returns the created object and panics if something is
// not right when creating the proxy.
func newCommonProxy(data model.Instance, log logger.Logger, module string) commonProxy {
	return commonProxy{
		proxy:  data.Proxy,
		log:    log,
		lock:   locker.NewConcurrentLocker(log),
		module: module,
		data:   data,
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
func (cp *commonProxy) trace(level logger.Severity, msg string) {
	cp.log.Trace(level, cp.module, msg)
}

// performWithLock :
// Used to exectue the specified query on the internal DB
// while making sure that the lock on the specified ID is
// acquired and released when needed.
//
// The `resource` represents an identifier of the resoure
// to access with the query: this method makes sure that
// a lock on this resource is created and handled so as
// to ensure that a single process is accessing it at any
// time.
//
// The `query` represents the operation to perform on the
// DB which should be protected with a lock. It should
// consist in a valid SQL query.
//
// Returns any error occurring during the process.
func (cp commonProxy) performWithLock(resource string, req db.InsertReq) error {
	// Prevent invalid resource identifier.
	if resource == "" {
		return ErrInvalidResource
	}

	// Acquire a lock on this resource.
	resLock := cp.lock.Acquire(resource)
	defer cp.lock.Release(resLock)

	// Perform the update: we will wrap the function inside
	// a dedicated handler to make sure that we don't lock
	// the resource more than necessary.
	var err error
	var errRelease error
	var errExec error

	func() {
		resLock.Lock()
		defer func() {
			if rawErr := recover(); rawErr != nil {
				err = fmt.Errorf("Error occured while executing query (err: %v)", rawErr)
			}
			errRelease = resLock.Release()
		}()

		// Perform the update.
		errExec = cp.proxy.InsertToDB(req)
	}()

	// Return any error.
	if errExec != nil {
		cp.trace(logger.Error, fmt.Sprintf("Invalid query executed for resource \"%s\" (err: %v)", resource, err))
		return db.ErrInvalidQuery
	}
	if errRelease != nil {
		cp.trace(logger.Error, fmt.Sprintf("Unable to release locker on \"%s\" propertly (err: %v)", resource, err))
		return ErrLock
	}
	if err != nil {
		cp.trace(logger.Error, fmt.Sprintf("Detected error while performing operation for \"%s\" (err: %v)", resource, err))
		return ErrInvalidOperation
	}

	return nil
}
