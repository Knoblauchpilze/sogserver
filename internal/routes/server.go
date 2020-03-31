package routes

import (
	"fmt"
	"net/http"
	"oglike_server/internal/data"
	"oglike_server/pkg/db"
	"oglike_server/pkg/logger"
	"strconv"
)

// server :
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
// The `logger` allows to perform most of the logging on any action
// done by the server such as logging clients' connections, errors
// and generally some elements useful to track the activity of the
// server.
type server struct {
	port         int
	universes    data.UniverseProxy
	accounts     data.AccountProxy
	buildings    data.BuildingProxy
	technologies data.TechnologyProxy
	ships        data.ShipProxy
	defenses     data.DefenseProxy
	planets      data.PlanetProxy
	players      data.PlayersProxy
	log          logger.Logger
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
func NewServer(port int, dbase *db.DB, log logger.Logger) server {
	if dbase == nil {
		panic(fmt.Errorf("Cannot create server from empty database"))
	}

	return server{
		port,
		data.NewUniverseProxy(dbase, log),
		data.NewAccountProxy(dbase, log),
		data.NewBuildingProxy(dbase, log),
		data.NewTechnologyProxy(dbase, log),
		data.NewShipProxy(dbase, log),
		data.NewDefenseProxy(dbase, log),
		data.NewPlanetProxy(dbase, log),
		data.NewPlayersProxy(dbase, log),
		log,
	}
}

// Serve :
// Used to start listening to the port associated to this server
// and handle incoming requests. This will return an error in case
// something went wrong while listening to the port.
func (s *server) Serve() error {
	// Setup routes.
	s.routes()

	// Serve the root path.
	return http.ListenAndServe(":"+strconv.FormatInt(int64(s.port), 10), nil)
}
