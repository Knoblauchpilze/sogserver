package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/KnoblauchPilze/sogserver/logger"
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

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(fmt.Sprintf("Handling request"))
	fmt.Fprintf(w, "Hi there, I love %s!\n", r.URL.Path[1:])
}

// main :
// Start the server and perform http listening.
func main() {
	// Handle help flag
	help := flag.Bool("h", false, "Print usage")
	if *help {
		usage()
	}

	logger := logger.NewStdLogger("", "127.0.0.1")

	logger.Trace(Verbose, "Verbose message")
	logger.Trace(Debug, "Debug message")
	logger.Trace(Info, "Info message")
	logger.Trace(Notice, "Notice message")
	logger.Trace(Warning, "Warning message")
	logger.Trace(Error, "Error message")
	logger.Trace(Critical, "Critical message")
	logger.Trace(Fatal, "Fatal message")

	// TODO: Implement the server maybe using this design pattern:
	// https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":3007", nil))
}
