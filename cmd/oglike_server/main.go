package main

import (
	"flag"
	"fmt"
	"runtime/debug"

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
// TODO: Multiple ships or defenses actions overlaps while they should be
// added at the end of each other.
// TODO: Maybe to fetch the result of the fleets that are not directed towards
// on of their planet, the players could issue a request to the actual target
// planets of the fleets (which is available in the fleet component which is
// retrieved (or could be) with the planet and get the result of the fleet
// this way.

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
			stack := string(debug.Stack())
			log.Trace(logger.Fatal, "main", fmt.Sprintf("App crashed after error: %v (stack: %s)", err, stack))
		}

		log.Release()
	}()

	// Create the server and set it up.
	DB := db.NewPool(log)
	proxy := db.NewProxy(DB)

	server := routes.NewServer(metadata.Port, proxy, log)

	err := server.Serve()
	if err != nil {
		panic(fmt.Errorf("Unexpected error while listening to port %d (err: %v)", metadata.Port, err))
	}
}
