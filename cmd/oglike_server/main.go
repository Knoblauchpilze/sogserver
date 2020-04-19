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

// TODO: Allow to delete a player.
// TODO: Allow to delete a planet.
// TODO: Use the token mechanism to make sure that a client has access
// to some information (typically to the data for a planet or a player).
// TODO: We should maybe find a way to transfer the mechanism for regular
// expressions used in the route in the `extractRouteVars` framework to allow
// to actually determine precisely which are the elements from the route and
// which are the extra pathes.
// TODO: It seems like some assumptions we make when updating resources in
// DB (and more precisely resources count) do not play well when intervals
// reach lengths of more than a month/day/year. See here for details:
// https://stackoverflow.com/questions/952493/how-do-i-convert-an-interval-into-a-number-of-hours-with-postgres
// To avoid this maybe we could have some function that would run every
// night (or another duration which would be consistent with the maximum
// interval with no issues) to perform update for players that didn't
// connect for a long time.
// TODO: We should maybe load all the data model from the DB in a single
// object that would then be passed around to proxies to perform various
// validations.
// We could add methods like `isBuilding`, `isTechnology`, `getFromName`
// etc. and all relevant methods to access the data anywhere.
// We could go ahead and maybe fetch some information about the planet
// and player in a similar way: basically wrap it in some sort of object
// that would allow high-level functions to be mutualized when needed.
// This could also maybe play well with the fact that for now we didn't
// extend a lot the lock and thus we don't really guarantee anything as
// several clients can access the planets of a players independently etc.
// Typically to fetch the amount of metal on a planet, we can't do it
// easily now: we need to access the planet, but also the map between the
// resources identifiers and their name (which is not contained in the
// map). This could be solved by providing first a data model (where one
// could access the id of the metal resource from its identifier) and then
// some wrapper on the planet which would use the data model to interpret
// the quantities.

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
			log.Trace(logger.Fatal, "main", fmt.Sprintf("App crashed after error: %v", err))
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
