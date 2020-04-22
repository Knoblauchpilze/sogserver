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
// Defines a server that can be used to handle the interaction
// with the OG database. This server handles can be built from
// the input database and logger and will perform the listening
// to handle the clients' requests.
// This article helped to set up and describe some aspects of
// the data model used by the server:
// https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years
//
// The `port` allows to determine which port should be used by
// the server to accept incoming requests. This is usually set
// in the configuration so as not to conflict with any other API.
//
// The `router` defines the element to use to perform the routing
// and receive clients requests. This object will be populated to
// reflect the routes available on this server and started upon
// calling the `Serve` method.
//
// The `universes` represents a proxy object allowing to access
// and modify the properties of universes from the main DB. It
// is used as a way to hide the complexity of the DB and only
// use high-level functions that do not rely on the internal
// schema of the DB to work.
//
//
// The `accounts` fills a similar role to `universes` but is related
// to accounts information.
//
// The `players` fills a similar purpose to `accounts` but for the
// players registered in each universe.
//
// The `planets` fills a similar purpose to `universes` but for the
// planets registered in the game.
//
// The `fleets` filles a similar purpose to `planets` but for the
// fleets registered in the game.
//
// The `actions` defines a proxy that can be used to serve
// the various upgrade actions handled by the game. It handles
// both the creation of the actions and their fetching.
//
//
// The `og` defines the data model associated to this server.
// It helps to serve information and is used by composite
// types to access base properties of the data model such as
// the prod level for a building, the cost of a ship, etc.
// in order to build the more complex processings required
// by the game.
//
// The `proxy` defines the DB to use to access to the data.
//
// The `logger` allows to perform most of the logging on any
// action done by the server such as logging connections or
// generally any useful information that could be monitored
// by the execution system of the server.
type Server struct {
	port      int
	router    *dispatcher.Router
	universes data.UniverseProxy
	accounts  data.AccountProxy
	players   data.PlayerProxy
	planets   data.PlanetProxy
	fleets    data.FleetProxy
	actions   data.ActionProxy

	og    model.Instance
	proxy db.Proxy
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
func NewServer(port int, proxy db.Proxy, log logger.Logger) Server {
	// Create modules to handle data model and initialize each one
	// of them.
	bm := model.NewBuildingsModule(log)
	tm := model.NewTechnologiesModule(log)
	sm := model.NewShipsModule(log)
	dm := model.NewDefensesModule(log)
	rm := model.NewResourcesModule(log)

	err := bm.Init(proxy, false)
	if err != nil {
		panic(fmt.Errorf("Cannot create server (err: %v)", err))
	}

	err = tm.Init(proxy, false)
	if err != nil {
		panic(fmt.Errorf("Cannot create server (err: %v)", err))
	}

	err = sm.Init(proxy, false)
	if err != nil {
		panic(fmt.Errorf("Cannot create server (err: %v)", err))
	}

	err = dm.Init(proxy, false)
	if err != nil {
		panic(fmt.Errorf("Cannot create server (err: %v)", err))
	}

	err = rm.Init(proxy, false)
	if err != nil {
		panic(fmt.Errorf("Cannot create server (err: %v)", err))
	}

	// Create the data model from it.
	ogDataModel := model.Instance{
		Proxy:        proxy,
		Buildings:    bm,
		Technologies: tm,
		Ships:        sm,
		Defenses:     dm,
		Resources:    rm,
	}

	// Create proxies on composite types.
	up := data.NewUniverseProxy(ogDataModel, log)
	pp := data.NewPlayerProxy(ogDataModel, log)
	ppp := data.NewPlanetProxy(ogDataModel, log)

	return Server{
		port:   port,
		router: nil,

		universes: up,
		accounts:  data.NewAccountProxy(dbase, log),
		planets:   ppp,
		players:   pp,
		fleets:    data.NewFleetProxy(dbase, log, up, pp),
		actions:   data.NewActionProxy(dbase, log, ppp, pp),

		og:    ogDataModel,
		proxy: proxy,
		log:   log,
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
