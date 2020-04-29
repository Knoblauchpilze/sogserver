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
// TODO: We could create for example a player module, which would fetch
// *ALL* the info of a player (including its construction actions, and
// its technologies and planets) and would detain a lock so that no one
// can access to the data as long as the player structure exists. This
// structure would contain both the players (with its name, universe,
// account, etc.), the associated list of planets and the list of techs
// researched by the player). Fecthing all this data would lock the
// player's name in the DB (so that other people accessing it would have
// to wait) and once all data is fetched (i.e. the structure is created)
// the lock would be released.
// The planet would also be linked to all the fleets that start from
// it or reach it (beware of dead lock for fleets moving from one
// planet of the player to another).
// Maybe the fleet model is no good because it does not have the
// intuitive notion of the planet it is attached to. But on the other
// hand it seems correct because a fleet may not be directed towards
// a planet.
// So we need:
//  - a `Player` struct more refined that what we have. It would include
//    the planets.
//  - a `Planet` struct more refined that what we have. It would include
//    the upgrade actions (and potentially the fleets).
// TODO: Some scripts might not work anymore due to changes in the way
// we have some Coordinates in the structure rather than some actual
// `Galaxy`, `System` and `Position` values.
// This can include the upgrade action scripts along with the actual code
// in the update functions.
// TODO: Should update the arrival time of the fleet from the server.
// TODO: The technology upgrade action does not include the player anymore.
// TODO: The planet should define a custom MarshalJSON method in order not
// to embed all the information of tech deps for ships, defenses etc.
// TODO: It seems that we don't check for `rows.Err` or `dbRes.Err` all
// the time.
// TODO: If we do:
// `curl http://localhost:3000/fleets/objectives?id=9a9c5df2-5d9e-4d0f-82dc-e633d123766d`
// we target the `fleets` endpoint while if we do:
// `curl http://localhost:3000/fleets/objectives/9a9c5df2-5d9e-4d0f-82dc-e633d123766d`
// we target the `fleets/objectives` endpoint.

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
