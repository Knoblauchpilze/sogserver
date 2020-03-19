package main

import (
	"flag"
	"fmt"
)

// usage :
// Displays the usage of the server. Typically requires a configuration file
// to be able to fetch the configuration variables to use during the execution
// of the server.
func usage() {
	// Server usage: nyce_renderer -config=[file]
	fmt.Println("Usage:")
	fmt.Println("-config=[file] for configuration file to use (local/master/staging/production)")
}

// main :
// Start the server and perform http listening.
func main() {
	// Handle help flag
	help := flag.Bool("h", false, "Print usage")
	if *help {
		usage()
	}

	// TODO: Implement the server.
	fmt.Println("Hello from oglike_server !")
}
