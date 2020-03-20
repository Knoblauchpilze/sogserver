package webutils

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx"
	"github.com/mycsHQ/webutils/logger"
	"github.com/spf13/viper"
)

// DB :
// Describes a database object. Provides a wrapper on the
// pgx types.
//
// The `pool` holds a reference on the database object. This
// value is not nil whenever a connection to the db has been
// successfully established.
//
// The `logger` allows to notify information and errors.
type DB struct {
	pool   *pgx.ConnPool
	logger logger.Logger
}

// NewPool :
// Performs the creation of a new database object. The created
// object will try to connect to the database described in the
// configuration file until a connection is established.
// Until the connection is successfully established, calls to
// `DbExecute` or `DbQuery` will fail.
//
// The `logger` allows to specify the logging device to use.
//
// Returns the created database object.
func NewPool(logger logger.Logger) *DB {
	// Create the db.
	db := DB{nil, logger}

	// Try to connect.
	db.createPoolAttempt()

	// Create a ticker to maintain the connection with the
	// db healthy.
	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for range ticker.C {
			db.Healthcheck()
		}
	}()

	// Return the created database.
	return &db
}

// createPoolAttempt :
// Used to try to connect to the database described in the
// configuration file. The connection is assigned to the internal
// attribute only if it has succeeded.
func (db *DB) createPoolAttempt() {
	// Retrieve connection parameters.
	host := viper.GetString("database.host")
	port := viper.GetString("database.port")
	user := viper.GetString("database.user")
	password := viper.GetString("database.password")
	dbName := viper.GetString("database.dbname")

	portUInt64, _ := strconv.ParseUint(port, 10, 16)
	portInt := uint16(portUInt64)

	db.logger.Trace("info", fmt.Sprintf("Trying to connect to db \"%s\" (user: \"%s\", host: \"%s:%s\")", dbName, user, host, port))

	// Try to connect to the database.
	pool, poolErr := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     host,
			Database: dbName,
			Port:     portInt,
			User:     user,
			Password: password,
		},
		MaxConnections: 40,
		AcquireTimeout: 0,
	})

	// Check whether the connection was successful.
	if poolErr != nil {
		db.logger.Trace("warning", fmt.Sprintf("DB connection pool setup failed (waiting 5s before reattempting)"))
		return
	}

	db.logger.Trace("info", fmt.Sprintf("DB connection to \"%s\" setup succeed", dbName))

	// Assign the database connection.)
	db.pool = pool
}

// Healthcheck :
// Used to check the health of the connection to the db.
// If the connection is not healthy a new attempt is scheduled.
func (db *DB) Healthcheck() {
	// Retrieve the current connection status.
	var stat pgx.ConnPoolStat
	if db.pool != nil {
		stat = db.pool.Stat()
	}

	// Check whether the connection is healthy.
	if db.pool == nil || stat.CurrentConnections == 0 {
		db.logger.Trace("info", "DB connection pool is not ready, reattempting to connect")
		db.createPoolAttempt()
	}
}

// DbExecute :
// Attempts to perform the input query with the specified arguments on the internal
// database connection.
// Note that if the connection has not yet been established with the db an error is
// returned.
// The `query` represents the request to execute.
// The `args` are arguments to pass to the query.
// Returns the result of the query along with any errors.
func (db *DB) DbExecute(query string, args ...interface{}) (*pgx.CommandTag, error) {
	if db.pool == nil {
		return nil, fmt.Errorf("Cannot executre query on db, invalid connection")
	}

	// Execute the query.
	tag, err := db.pool.Exec(query, args...)

	// All is well.
	return &tag, err
}

// DbQuery :
// Attempts to execute the input query with the specified arguments on the internal
// database connection.
// Note that if the connection has not yet been established with the db an error is
// returned.
// The `query` represents the request to execute.
// The `args` are arguments to pass to the query.
// Returns the result of the query along with any errors.
func (db *DB) DbQuery(query string, args ...interface{}) (*pgx.Rows, error) {
	if db.pool == nil {
		return nil, fmt.Errorf("Cannot executre query on db, invalid connection")
	}

	// Query the db.
	r, err := db.pool.Query(query, args...)

	// All is well.
	return r, err
}
