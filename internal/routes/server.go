package routes

import (
	"context"
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/internal/game"
	"oglike_server/internal/model"
	"oglike_server/pkg/background"
	"oglike_server/pkg/db"
	"oglike_server/pkg/dispatcher"
	"oglike_server/pkg/logger"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/handlers"
	"github.com/spf13/viper"
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
// The `accounts` is used in a similar way to `universes` but
// regroups information about the players' account registered
// so far in the server. Each account can have chidlren item
// in any number of universe. An account is really the base
// for a client to use any feature of the server.
//
// The `players` is used similarly to the `universes` but is
// handling the players data. A player is the instance of an
// account in a specific universe and is usually associated
// to planets and fleets.
//
// The `planets` defins a similar proxy to all the others and
// handles the planets data. Planets are linked to universes
// and players and most of the actions are related to at least
// a single planet.
//
// The `fleets` handles the fleets' data. Fleets are the way
// to make planets communicate and send info (resources and
// ships) in the game.
//
// The `actions` handle the upgrade actions that can be
// registered to improve the state of buildings on a planet.
// It also handles the technologies, ships and defenses that
// can be built on a planet. It makes sure that any request
// is consistent with the state of the planet where the
// action should be performed.
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
// The `log` allows to perform most of the logging on any
// action done by the server such as logging connections or
// generally any useful information that could be monitored
// by the execution system of the server.
//
// The `process` defines the background process that is used
// to make sure that certain operations are executed in a way
// that guarantee consistency of the data in the server's DB.
type Server struct {
	port      int
	router    *dispatcher.Router
	universes data.UniverseProxy
	accounts  data.AccountProxy
	players   data.PlayerProxy
	planets   data.PlanetProxy
	fleets    data.FleetProxy
	actions   data.ActionProxy

	og    game.Instance
	proxy db.Proxy
	log   logger.Logger

	process *background.Process
}

// ErrUnexpectedServeError : Indicates that an error occurred
// while serving the root endpoint.
var ErrUnexpectedServeError = fmt.Errorf("Unexpected error occurred while serving http requests")

// ErrServerShutdownError : Indicates that an error occurred
// while shutting down the server.
var ErrServerShutdownError = fmt.Errorf("Unexpected error occurred while shutting down the server")

// configuration :
// Defines the custom properties that can be defined for the
// server through the configuration file.
//
// The `BackgroundUpdate` defines the time interval between
// two consecutive update of the background processes of the
// serevr. This allows to keep the amount of pending tasks
// for the game actions to a reasonable level and prevent
// some weirdness that could appear with time management if
// the duration become too long.
// The duration is expressed in minutes and the default value
// is set to `60`.
type configuration struct {
	BackgroundUpdate time.Duration
}

// parseConfiguration :
// Used to parse the configuration file and environment
// variables provided when executing this server to get
// the values of the `Server` properties. These props
// allow to customize the behavior of the processes to
// be performed by the server.
//
// Returns the parsed configuration where all non-set
// properties have their default values.
func parseConfiguration() configuration {
	// Create the default configuration.
	config := configuration{
		BackgroundUpdate: 60 * time.Minute,
	}

	// Parse custom properties.
	if viper.IsSet("Server.BackgroundUpdate") {
		min := viper.GetInt("Server.BackgroundUpdate")
		config.BackgroundUpdate = time.Duration(min) * time.Minute
	}

	return config
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
	cm := model.NewCountriesModule(log)
	bm := model.NewBuildingsModule(log)
	tm := model.NewTechnologiesModule(log)
	sm := model.NewShipsModule(log)
	dm := model.NewDefensesModule(log)
	rm := model.NewResourcesModule(log)
	om := model.NewFleetObjectivesModule(log)
	mm := model.NewMessagesModule(log)

	err := cm.Init(proxy, false)
	if err != nil {
		panic(fmt.Errorf("Cannot create server (err: %v)", err))
	}

	err = bm.Init(proxy, false)
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

	err = om.Init(proxy, false)
	if err != nil {
		panic(fmt.Errorf("Cannot create server (err: %v)", err))
	}

	err = mm.Init(proxy, false)
	if err != nil {
		panic(fmt.Errorf("Cannot create server (err: %v)", err))
	}

	// Create the data model from it.
	ogDataModel := game.NewInstance(proxy, log)

	ogDataModel.Countries = cm
	ogDataModel.Buildings = bm
	ogDataModel.Technologies = tm
	ogDataModel.Ships = sm
	ogDataModel.Defenses = dm
	ogDataModel.Resources = rm
	ogDataModel.Objectives = om
	ogDataModel.Messages = mm

	// Create proxies on composite types.
	up := data.NewUniverseProxy(ogDataModel, log)
	ap := data.NewAccountProxy(ogDataModel, log)
	pp := data.NewPlayerProxy(ogDataModel, log)
	ppp := data.NewPlanetProxy(ogDataModel, log)
	fp := data.NewFleetProxy(ogDataModel, log)
	aap := data.NewActionProxy(ogDataModel, log)

	// Create the background process to ensure
	// data consistency in the game's DB.
	config := parseConfiguration()

	p := background.NewProcess(config.BackgroundUpdate, log)

	p.WithModule("cron").WithRetry().WithOperation(
		func() (bool, error) {
			defer ogDataModel.Unlock()
			ogDataModel.Lock()

			return true, nil
		},
	)

	return Server{
		port:   port,
		router: nil,

		universes: up,
		accounts:  ap,
		planets:   ppp,
		players:   pp,
		fleets:    fp,
		actions:   aap,

		og:    ogDataModel,
		proxy: proxy,
		log:   log,

		process: p,
	}
}

// Serve :
// Used to start listening to the port associated to this
// server and handle incoming requests. This will return
// an error in case something went wrong while listening
// to the port.
//
// Returns any error occurred during the serve operation.
func (s *Server) Serve() error {
	// Create a new router if one is not already started.
	if s.router != nil {
		panic(fmt.Errorf("Cannot start serving OG server, process already running"))
	}

	s.router = dispatcher.NewRouter(s.log)

	// Setup routes.
	s.routes()

	// Wrap the router in a server allowing all origins.
	aMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	aOrigins := handlers.AllowedOrigins([]string{"*"})
	aHeaders := handlers.AllowedHeaders([]string{"Origin", "X-Requested-With", "Content-Type", "Accept", "Authorization"})
	corsRouter := handlers.CORS(aHeaders, aOrigins, aMethods)(s.router)

	// Create the server which will serve requests. The
	// idiom used to serve requests is inspired from the
	// following link:
	// https://stackoverflow.com/questions/39320025/how-to-stop-http-listenandserve
	// which describes a way to gracefully shutdown a
	// HTTP server. We figure it's worth doing it.
	server := &http.Server{
		Addr:    ":" + strconv.FormatInt(int64(s.port), 10),
		Handler: corsRouter,
	}

	// Start the routine which will handle the automatic
	// update of some processes if it is not done often
	// enough.
	s.process.Start()

	// Serve the root path.
	var serveErr error
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				s.log.Trace(logger.Fatal, "server", fmt.Sprintf("Caught unexpected error while serving requests (err: %v)", err))

				serveErr = ErrUnexpectedServeError
			}

			wg.Done()

			s.log.Trace(logger.Notice, "server", "Server has stopped")
		}()

		s.log.Trace(logger.Notice, "server", "Server has started")

		// Serve the main endpoint and panic in case something
		// bad occurs in the process.
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// Setting up signal capturing.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (pkill -2).
	<-stop

	s.shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		s.log.Trace(logger.Error, "server", fmt.Sprintf("Caught unexpected error while shutting down server (err: %v)", err))

		return ErrServerShutdownError
	}

	// Wait for `ListenAndServe` to perform cleanup.
	wg.Wait()

	return serveErr
}

// shutdown :
// Requests the server to gracefully shutdown and
// terminate all the processes that are pending
// before doing so.
func (s *Server) shutdown() {
	s.process.Stop()
}
