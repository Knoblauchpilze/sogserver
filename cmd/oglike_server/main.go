package main

import (
	"flag"
	"fmt"
	"net/http"

	// Note that this link: https://stackoverflow.com/questions/55442878/organize-local-code-in-packages-using-go-modules
	// proved helpful when trying to determine which syntax to adopt to use packages define locally.
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

type Han struct {
	base *db.DB
}

// TODO: Remove this.
func (h *Han) handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(fmt.Sprintf("Handling request"))

	rows, err := h.base.DBQuery("select * from universes")
	if err != nil {
		fmt.Println(fmt.Sprintf("Query failed with err %v", err))
	} else {
		fmt.Println(fmt.Sprintf("Query succeeded with result %v", *rows))

		var name string
		var id string

		row := 1
		for rows.Next() {
			rows.Scan(
				&name,
				&id,
			)

			fmt.Println(fmt.Sprintf("Row %d has name \"%s\" and id \"%s\"", row, name, id))

			row++
		}
	}

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

	DB := db.NewPool(log)

	log.Trace(logger.Verbose, "Verbose message")
	log.Trace(logger.Debug, "Debug message")
	log.Trace(logger.Info, "Info message")
	log.Trace(logger.Notice, "Notice message")
	log.Trace(logger.Warning, "Warning message")
	log.Trace(logger.Error, "Error message")
	log.Trace(logger.Critical, "Critical message")
	log.Trace(logger.Fatal, "Fatal message")

	h := Han{DB}

	// TODO: Implement the server maybe using this design pattern:
	// https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years
	http.HandleFunc("/", h.handler)
	http.ListenAndServe(":3007", nil)
}
