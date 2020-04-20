package routes

import (
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/internal/model"
	"oglike_server/pkg/db"
	"oglike_server/pkg/dispatcher"
	"oglike_server/pkg/logger"
	"strconv"
)

// Server :
// Defines a server that can be used to handle the interaction with
// the OG database. This server handles can be built from the input
// database and logger and will perform the listening to handle the
// clients' requests.
// This article helped a bit to set up and describe the data model
// and structures used to describe the server:
// https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years
//
// The `port` allows to determine which port should be used by the
// server to accept incoming requests. This is usually specified in
// the configuration so as not to conflict with any other API.
//
// The `router` defines the element to use to perform the routing
// and receive clients requests. This object will be populated to
// reflect the routes available on this server and started upon
// calling the `Serve` method.
//
// The `universes` represents a proxy object allowing to interact
// and retrieve properties of universes from the main DB. It is used
// as a way to hide the complexity of the DB and only use high-level
// functions that do not rely on the internal schema of the DB to
// work.
//
// The `accounts` fills a similar role to `universes` but is related
// to accounts information.
//
// The `buildings` represents the proxy to use to perform requests
// concerning buildings and to access information about this topic.
//
// The `technologies` fills a similar purpose as `buildings` but
// for technologies related requests.
//
// The `ships` fills a similar purpose as `buildings` but for ships
// related requests.
//
// The `defenses` fills a similar purpose as `buildings` but for
// defenses related requests.
//
// The `planets` fills a similar purpose to `universes` but for the
// planets registered in the game.
//
// The `players` fills a similar purpose to `accounts` but for the
// players registered in each universe.
//
// The `fleets` filles a similar purpose to `planets` but for the
// fleets registered in the game.
//
// The `upgradeAction` defines a proxy that can be used to serve
// the various upgrade actions handled by the game. It handles
// both the creation of the actions and their fetching.
//
//
// The `dbase` defines the DB to use to access to the data.
//
// The `logger` allows to perform most of the logging on any action
// done by the server such as logging clients' connections, errors
// and generally some elements useful to track the activity of the
// server.
type Server struct {
	port          int
	router        *dispatcher.Router
	universes     data.UniverseProxy
	accounts      data.AccountProxy
	buildings     *model.BuildingsModule
	technologies  *model.TechnologiesModule
	ships         *model.ShipsModule
	defenses      *model.DefensesModule
	planets       data.PlanetProxy
	players       data.PlayerProxy
	fleets        data.FleetProxy
	upgradeAction data.ActionProxy

	dbase *db.DB
	log   logger.Logger
}

// NewServer :
// Create a new server with the input elements to use internally to
// access data and perform logging.
// In case any of the arguments are not valid a panic is issued to
// indicate the failure.
//
// The `port` defines the port to listen to by the server.
//
// The `dbase` represents a pointer to the database to use to fetch
// data when needed to answer clients' requests.
//
// The `log` is used to notify from various processes in the server
// and keep track of the activity.
func NewServer(port int, dbase *db.DB, log logger.Logger) Server {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create server from empty database"))
	}

	uniProxy := data.NewUniverseProxy(dbase, log)
	playerProxy := data.NewPlayerProxy(dbase, log)
	planetProxy := data.NewPlanetProxy(dbase, log, uniProxy)

	// Create modules to handle data model and initialize each one
	// of them.
	bm := model.NewBuildingsModule(log)
	tm := model.NewTechnologiesModule(log)
	sm := model.NewShipsModule(log)
	dm := model.NewDefensesModule(log)

	err := bm.Init(dbase, false)
	if err != nil {
		panic(fmt.Errorf("Cannot create server (err: %v)", err))
	}

	err = tm.Init(dbase, false)
	if err != nil {
		panic(fmt.Errorf("Cannot create server (err: %v)", err))
	}

	err = sm.Init(dbase, false)
	if err != nil {
		panic(fmt.Errorf("Cannot create server (err: %v)", err))
	}

	err = dm.Init(dbase, false)
	if err != nil {
		panic(fmt.Errorf("Cannot create server (err: %v)", err))
	}

	return Server{
		port,
		nil,
		uniProxy,
		data.NewAccountProxy(dbase, log),
		bm,
		tm,
		sm,
		dm,
		planetProxy,
		playerProxy,
		data.NewFleetProxy(dbase, log, uniProxy, playerProxy),
		data.NewActionProxy(dbase, log, planetProxy, playerProxy),
		dbase,
		log,
	}
}

// Serve :
// Used to start listening to the port associated to this server
// and handle incoming requests. This will return an error in case
// something went wrong while listening to the port.
func (s *Server) Serve() error {
	// Create a new router if one is not already started.
	if s.router != nil {
		panic(fmt.Errorf("Cannot start serving OG server, process already running"))
	}

	s.router = dispatcher.NewRouter(s.log)

	// Setup routes.
	s.routes()

	// Register the router as the main listener.
	http.Handle("/", s.router)

	// Serve the root path.
	return http.ListenAndServe(":"+strconv.FormatInt(int64(s.port), 10), nil)
}
