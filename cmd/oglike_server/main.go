package main

import (
	"flag"
	"fmt"

	// Note that this link: https://stackoverflow.com/questions/55442878/organize-local-code-in-packages-using-go-modules
	// proved helpful when trying to determine which syntax to adopt to use packages defined locally.
	"oglike_server/internal/routes"
	"oglike_server/pkg/arguments"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// TODO: Construction actions.
// TODO: Allow to delete a player.
// TODO: Allow to delete a planet.
// TODO: Use the token mechanism to make sure that a client has access
// to some information (typically to the data for a planet or a player).
// TODO: Add rapid fire in `ShipDesc`.
// TODO: Provide the script to actually perform the construction action.
// This script would be something like an update of the table by making
// sure that it only upgrade something consistent so typically:
// `update planets_buildings pb
//  set pb.level=s.desired_level
//  where
//    pb.planet = s.planet and
//    pb.building = s.building
//    pb.level = s.current_level
//  with source s as select * from json_populate_recordset(table::null, inputs)`
// We should also make sure to delete the upgrade action afterwards.
// We could also check the number of affected rows as returned by the
// Postgres driver to see whether we could update something.
// TODO: We should maybe find a way to transfer the mechanism for regular
// expressions used in the route in the `extractRouteVars` framework to allow
// to actually determine precisely which are the elements from the route and
// which are the extra pathes.
// TODO: Refine building, technology and ship and defense unmarshalling.

// usage :
// Displays the usage of the server. Typically requires a configuration
// file to be able to fetch the configuration variables to use during
// the execution of the server.
func usage() {
	fmt.Println("Usage:")
	fmt.Println("./oglike_server -config=[file] for configuration file to use (development/production)")
}

// main :
// Start the server and perform http listening.
func main() {
	// Define common flags.
	help := flag.Bool("h", false, "Print usage")
	conf := flag.String("config", "", "Configuration file to customize app behavior (development/production)")

	// Parse flags.
	flag.Parse()

	// Check for help flag.
	if *help {
		usage()
	}

	// Parse configuration if any.
	trueConf := ""
	if conf != nil {
		trueConf = *conf
	}
	metadata := arguments.Parse(trueConf)

	log := logger.NewStdLogger(metadata.InstanceID, metadata.PublicIPv4)

	// Handle last resort error handling to at least determine
	// what was the cause of the crash.
	defer func() {
		err := recover()
		if err != nil {
			log.Trace(logger.Fatal, fmt.Sprintf("App crashed after error: %v", err))
		}

		log.Release()
	}()

	// Create the server and set it up.
	DB := db.NewPool(log)
	server := routes.NewServer(metadata.Port, DB, log)

	err := server.Serve()
	if err != nil {
		panic(fmt.Errorf("Unexpected error while listening to port %d (err: %v)", metadata.Port, err))
	}
}
