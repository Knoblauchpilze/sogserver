package db

import (
	"fmt"
	"oglike_server/pkg/logger"
	"sync"
	"time"

	"github.com/jackc/pgx"
	"github.com/spf13/viper"
)

// configuration :
// Defines the possible options to define the way this DB
// object should try to connect to the underlying database.
// Common parameters allow to locate the database through
// a network address and provide some information about a
// various set of connection parameters (username, DB name
// and password).
//
// The `host` references the address at which the database
// is hosted and thus where we should try to connect to it.
// The default value is "localhost".
//
// The `port` describes the exposed port to connect to the
// database.
// The default value is 5432.
//
// The `name` defines the name of the database. This value
// should be set as we cannot assume anything regarding its
// value in general.
//
// The `user` defines the role that this object should use
// to connect to the DB. It should be specified from the
// configuration file.
//
// The `password` defines the password to use to access to
// the DB given the specified username. No default value is
// provided for this value.
//
// The `timeout` which separates two successive connection
// attemps to the DB. In case an attempt fails we will wait
// for this amount of time before trying again. This time
// is expressed in seconds.
// The default value is `5` seconds.
//
// The `connectionsPool` defines the number of concurrent
// connections that can be issued on the underlying DB. The
// larger this value the more stress will be put on the DB
// but the more clients will be able to concurrently access
// it.
// The default value is `5`.
type configuration struct {
	host            string
	port            int
	name            string
	user            string
	password        string
	timeout         int
	connectionsPool int
}

// DB :
// Describes a database object to provides a wrapper on the
// pgx handler. This is used as a convenience way to hide a
// part of the DB implementation to be used in other parts
// of an application.
// Compared to the base wrapper it handles a mechanism to
// try connecting to the DB until it comes online. It will
// also retrieve automatically the parameters to use to
// connect to the DB from the configuration file.
//
// The `pool` holds a reference on the database object. This
// value is not `nil` whenever a connection to the DB has
// been successfully established.
//
// The `lock` allows to protect the `pool` value from some
// concurrent accesses. This is typically useful when the
// connection to the DB is lost and we try to establish it
// again to prevent other clients to access the DB in the
// meantime.
//
// The `logger` allows to notify information and errors.
//
// The `config` describes the connection properties to use
// to perform the connection to the DB object. It is parsed
// upon building the object so that we don't attempt anything
// in case the configuration is not valid.
type DB struct {
	pool   *pgx.ConnPool
	lock   sync.Mutex
	logger logger.Logger
	config configuration
}

// parseConfiguration :
// Attempt to parse the configuration provided to this app
// to extract connection parameters to use for the DB. It
// relies on default value in case some values are not set
// and panics if some mandatory values cannot be found in
// the configuration.
//
// Returns the built-in configuration object.
func parseConfiguration() configuration {
	// Create a default configuration object.
	config := configuration{
		"localhost",
		5432,
		"",
		"",
		"",
		5,
		5,
	}

	// Fetch configuration values from the runtime.
	if viper.IsSet("Database.Host") {
		config.host = viper.GetString("Database.Host")
	}
	if viper.IsSet("Database.Port") {
		config.port = viper.GetInt("Database.Port")
	}
	if viper.IsSet("Database.Name") {
		config.name = viper.GetString("Database.Name")
	}
	if viper.IsSet("Database.User") {
		config.user = viper.GetString("Database.User")
	}
	if viper.IsSet("Database.Password") {
		config.password = viper.GetString("Database.Password")
	}
	if viper.IsSet("Database.Timeout") {
		config.timeout = viper.GetInt("Database.Timeout")
	}
	if viper.IsSet("Database.ConnectionsPool") {
		config.connectionsPool = viper.GetInt("Database.ConnectionsPool")
	}

	// Check whether we could find all the mandatory
	// configuration properties and that the rest of
	// the values are consistent.
	if len(config.name) == 0 {
		panic(fmt.Errorf("Invalid DB name fetched from configuration \"%s\"", config.name))
	}
	if len(config.user) == 0 {
		panic(fmt.Errorf("Invalid DB user fetched from configuration \"%s\"", config.user))
	}
	if len(config.password) == 0 {
		panic(fmt.Errorf("Invalid DB password fetched from configuration \"%s\"", config.password))
	}
	if config.port < 0 {
		panic(fmt.Errorf("Invalid DB port fetched from configuration %d", config.port))
	}
	if config.connectionsPool <= 0 {
		panic(fmt.Errorf("Invalid DB connections pool fetched from configuration %d", config.connectionsPool))
	}

	return config
}

// NewPool :
// Performs the creation of a new database object. The created
// object will try to connect to the database described in the
// configuration file until a connection is established.
// Until the connection is successfully established, calls to
// `DBExecute` or `DBQuery` will fail.
//
// The `logger` allows to specify the logging device to use.
//
// Returns the created database object.
func NewPool(logger logger.Logger) *DB {
	// Parse the configuration for the DB connection.
	config := parseConfiguration()

	// Verify the port.
	maxPort := 1 << 16
	if config.port >= maxPort {
		panic(fmt.Errorf("Cannot use port %d to connect to DB \"%s\"", config.port, config.name))
	}

	// Create the DB object.
	dbase := DB{
		nil,
		sync.Mutex{},
		logger,
		config,
	}

	// Try to connect to the DB.
	dbase.createPoolAttempt()

	// Create a ticker to maintain the connection with the
	// DB healthy in case of a disconnection later on.
	ticker := time.NewTicker(time.Second * time.Duration(config.timeout))
	go func() {
		for range ticker.C {
			dbase.Healthcheck()
		}
	}()

	// Return the created database.
	return &dbase
}

// createPoolAttempt :
// Used to try to connect to the database described in the configuration
// file. The connection is assigned to the internal attribute only if it
// has succeeded.
// Note that this method does not check whether the connection to the DB
// is already healthy: we assume that calling this method is either the
// result of checking this situation beforehand or follows a change in
// the connection parameters.
//
// Returns `true` if the attempot succeeeded (i.e. if we are successfully
// connected to the DB) and `false` otherwise.
func (dbase *DB) createPoolAttempt() bool {
	config := dbase.config
	dbase.logger.Trace(logger.Info, fmt.Sprintf("Attempting to connect to \"%s\" (user: \"%s\", host: \"%s:%d\")", config.name, config.user, config.host, config.port))

	port := uint16(config.port)

	// Try to connect to the database.
	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     config.host,
			Database: config.name,
			Port:     port,
			User:     config.user,
			Password: config.password,
		},
		MaxConnections: config.connectionsPool,
		AcquireTimeout: 0,
	})

	// Check whether the connection was successful.
	if err != nil {
		dbase.logger.Trace(logger.Warning, fmt.Sprintf("Failed to connect to DB \"%s\" (err : %v)", config.name, err))
		return false
	}

	dbase.logger.Trace(logger.Info, fmt.Sprintf("Connection to DB \"%s\" with username \"%s\" succeeded", config.name, config.user))

	// Assign the database connection to the internal
	// attribute while maintaining thread safety.
	dbase.lock.Lock()
	func() {
		defer dbase.lock.Unlock()
		dbase.pool = pool
	}()

	return true
}

// Healthcheck :
// Used to check the health of the connection to the DB. In case
// the connection is found not to be healthy, a new attempt is
// scheduled immediately.
// Note that if the connection with the database has been lost
// for some readon, the test performed by this method will not
// allow to detect it and will thus return as if the connection
// was still healthy.
// This could be resolved by actually querying the DB (because
// it seems that a query does flag the connection as invalid or
// rather indicates that the connection is not valid anymore so
// we effectively reach `0` in the current connection count) but
// for now we will consider that it's enough. In any case when
// a user actually performs a request we detect that and the
// healthcheck fails so a new connection attempt is scheduled.
func (dbase *DB) Healthcheck() {
	// Retrieve the current connection status.
	dbIsNil := false
	var stat pgx.ConnPoolStat

	dbase.lock.Lock()
	func() {
		defer dbase.lock.Unlock()

		dbIsNil = (dbase.pool == nil)
		if !dbIsNil {
			stat = dbase.pool.Stat()
		}
	}()

	// Check whether the connection is healthy.
	if dbIsNil || stat.CurrentConnections == 0 {
		dbase.createPoolAttempt()
	}
}

// DbExecute :
// Attempts to perform the input query with the specified arguments on
// the internal database connection.
// Note that if the connection has not yet been established with the DB
// an error is returned.
//
// The `query` represents the request to execute. This is usually built
// by knowing in advance the structure of the DB.
//
// The `args` are arguments to pass to the query. Depending on the query
// it can be anything relevant to make the query succeed (such as data
// to be inserted in the DB, etc.).
//
// Returns the result of the query along with any errors.
func (dbase *DB) DBExecute(query string, args ...interface{}) (*pgx.CommandTag, error) {
	dbase.lock.Lock()
	if dbase.pool == nil {
		dbase.lock.Unlock()
		return nil, fmt.Errorf("Cannot execute query on DB \"%s\" (err: connection is invalid)", dbase.config.name)
	}

	var tag pgx.CommandTag
	var err error

	func() {
		defer dbase.lock.Unlock()
		tag, err = dbase.pool.Exec(query, args...)
	}()

	return &tag, err
}

// DbQuery :
// Attempts to execute the input query with the specified arguments on
// the internal database connection. This method is very similar to the
// `DBExecute` but fetch information from the DB rather than inserting
// some data in it (or performing any kind of modifications for that
// matter).
// Note that if the connection has not yet been established with the DB
// an error is returned.
//
// The `query` represents the request to execute.
//
// The `args` are arguments to pass to the query.
//
// Returns the result of the query along with any errors.
func (dbase *DB) DBQuery(query string, args ...interface{}) (*pgx.Rows, error) {
	dbase.lock.Lock()
	if dbase.pool == nil {
		dbase.lock.Unlock()
		return nil, fmt.Errorf("Cannot execute query on DB \"%s\" (err: connection is invalid)", dbase.config.name)
	}

	var r *pgx.Rows
	var err error

	func() {
		defer dbase.lock.Unlock()
		r, err = dbase.pool.Query(query, args...)
	}()

	return r, err
}
