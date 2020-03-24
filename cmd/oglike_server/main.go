package main

import (
	"flag"
	"fmt"

	// Note that this link: https://stackoverflow.com/questions/55442878/organize-local-code-in-packages-using-go-modules
	// proved helpful when trying to determine which syntax to adopt to use packages define locally.
	"oglike_server/internal"
	"oglike_server/pkg/arguments"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
)

// usage :
// Displays the usage of the server. Typically requires a configuration file
// to be able to fetch the configuration variables to use during the execution
// of the server.
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
	server := internal.NewServer(metadata.Port, DB, log)

	err := server.Serve()
	if err != nil {
		panic(fmt.Errorf("Unexpected error while listening to port %d (err: %v)", metadata.Port, err))
	}
}
