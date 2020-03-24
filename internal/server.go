package internal

import (
	"fmt"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// server :
// Defines a server that can be used to handle the interaction with
// the OG database. This server handles can be built from the input
// database and logger and will perform the listening to handle the
// clients' requests.
// This article helped a bit to set up and describe the data model
// and structures used to describe the server.
// https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years
//
// The `port` allows to determine which port should be used by the
// server to accept incoming requests. This is usually specified in
// the configuration so as not to conflict with any other API.
//
// The `dbase` represents a connection to the database containing
// the data to be served to clients. This is assumed to be a valid
// item and will be used to issue internal requests when data needs
// to be fetched.
//
// The `logger` allows to perform most of the logging on any action
// done by the server such as logging clients' connections, errors
// and generally some elements useful to track the activity of the
// server.
type server struct {
	port  int
	dbase *db.DB
	log   logger.Logger
}

// NewServer :
// Create a new server with the input elements to use internally to
// access data and perform logging.
// In case any of the arguments are not valid a panic is issued to
// indicate the failure.
//
// The `port` defines the port to listen to by the server.
//
// The `dbase` represents a pointer to the database to use to fetch
// data when needed to answer clients' requests.
//
// The `log` is used to notify from various processes in the server
// and keep track of the activity.
func NewServer(port int, dbase *db.DB, log logger.Logger) server {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create server from empty database"))
	}

	return server{port, dbase, log}
}
