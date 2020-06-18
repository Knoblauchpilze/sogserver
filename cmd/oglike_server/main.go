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

// TODO: Use the token mechanism to make sure that a client has access
// to some information (typically to the data for a planet or a player).
// TODO: We don't really have a proper mechanism for messages in the case
// of fleet fight report. How could we do that ? We have a similar issue
// in the case of an esionage report where all the info is available but
// we don't really know how to persist the info.
// Maybe we could do something similar to the `espionage_report` message
// where only a single value is persisted and the data can be computed
// client-side. Typically for the espionage maybe we could persist the
// info-level along with a token and the client would then issue a request
// on the `planet` endpoint to fetch the info accessible given the info
// level provided by the spying. This would play nicely with the sort of
// authentication through tokens to access endpoints.
// In the case of a fight report we could maybe save only the rng seed
// used during the fight along some modifications like the techno of the
// attackers (or maybe just fecthing them when the message is actually
// interpreted is enough) and the amount of ships.
// https://lng.xooit.com/t1488-Mettre-en-page-un-RC-avec-ogame-winner.htm
// TODO: Should probably extract the generation of the report in its own
// function because in the case of an ACS we want a single report with
// all the participants.

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
