package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	// Note that this link: https://stackoverflow.com/questions/55442878/organize-local-code-in-packages-using-go-modules
	// proved helpful when trying to determine which syntax to adopt to use packages define locally.
	"oglike_server/pkg/arguments"
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

// TODO: Remove this.
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(fmt.Sprintf("Handling request"))
	fmt.Fprintf(w, "Hi there, I love %s!\n", r.URL.Path[1:])
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

	l := logger.NewStdLogger(metadata.InstanceID, metadata.PublicIPv4)

	l.Trace(logger.Verbose, "Verbose message")
	l.Trace(logger.Debug, "Debug message")
	l.Trace(logger.Info, "Info message")
	l.Trace(logger.Notice, "Notice message")
	l.Trace(logger.Warning, "Warning message")
	l.Trace(logger.Error, "Error message")
	l.Trace(logger.Critical, "Critical message")
	l.Trace(logger.Fatal, "Fatal message")

	// TODO: Implement the server maybe using this design pattern:
	// https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":3007", nil))
}
