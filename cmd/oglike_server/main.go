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

// TODO: Allow to delete a player.
// TODO: Allow to delete a planet.
// TODO: Use the token mechanism to make sure that a client has access
// to some information (typically to the data for a planet or a player).
// TODO: It seems like some assumptions we make when updating resources in
// DB (and more precisely resources count) do not play well when intervals
// reach lengths of more than a month/day/year. See here for details:
// https://stackoverflow.com/questions/952493/how-do-i-convert-an-interval-into-a-number-of-hours-with-postgres
// To avoid this maybe we could have some function that would run every
// night (or another duration which would be consistent with the maximum
// interval with no issues) to perform update for players that didn't
// connect for a long time.
// TODO: Multiple ships or defenses actions overlaps while they should be
// added at the end of each other.
// TODO: Regarding the lock system. What we do for accounts and universes
// is very interesting because it means that we can fetch info in an atomic
// way where the data is fetched once and for all so that we don't have to
// handle cases where it can be modified while fetching it.
// A possible option to account for that would be to modify the fecting of
// planets for example where we would get all the info fetched by `fetchBuildings`
// and the like in a single big query which would return the list of tables
// values. It is not the preferred solution because it's kind of ugly.
// What we could do though, is make sure that the way we fetch the data is
// consistent with providing consistent info.
// Typically, in the case where there are no fleet, the data of a planet
// is quite self consistent: the upgrade actions will be updated, and the
// resources and all is well. The only part that requires some sort of a
// lock is the technologies but even then: actually we're fecthing the
// technologies in a single query and the case where we use them is to
// create an action So even if we fetch inconsistent data (typically a
// research that is too high or too low compared to what's really in the
// DB) it would only mean that the action could either be forbidden while
// in fact it should be authorized, or the other way around. This can be
// handled by the client which would display an adequate message provided
// the payload of the answer is also relevant. This would correspond to
// cases where the user clicked on a research on two separate tabs or
// something similar: very unlikely.
// Regarding the fleet, it might have potential impacts. Typically if we
// want to restore the fleets endpoint, we might want to lock a fleet.
// We also might want to lock the planet related to the fleet (if any)
// so that we can perform the fight or any operation while nobody is
// using the planet. This would give the lock order `Fleet > Planet`.
// On the other hand, when fetching the data for a planet, we want the
// fleets targetting it to be updated which gives the reverse lock order
// of `Planet > Fleet`. But why do we want to lock ?
// We want to lock to be sure that the information fetched by the fleet
// and processed by it cannot be made inconsistent by for exemple some
// research that is being started while the fleet is processed. So it
// seems important to lock the resources.
// Moreover it does not seem possible to kind of fetch information and
// then be done by assuming that they stay consistent because part of
// the fleet's processing will be off DB. So we have an issue.
// Finally (final nail in the coffin) there is an issue with the current
// system because the planet's code reads something like this:
//      p.fetchShipUpgrades(data)
//      p.fetchDefenseUpgrades(data)
//      p.updateResources(data)
//      p.fetchFleets(data)
// Where the `fetchShipUpgrades` will update the ship upgrade to the
// current time: but what if a fleet was scheduled in the mean time ?
// We have no easy way to handle this right now.
// So what can we do ?
// Maybe a solution where we have some kind of `process upgrade actions`
// function or routine which is called before actually fetching the
// data from the DB. This function would cycle through the list of
// upgrade actions for *all* players (maybe of a universe) and perform
// the ones that are past due. This would make sure that we can handle
// any actions independently of how we access the data (because we can
// have a locker on this process which can be triggered only once for
// example) and be sure that no one else would try to access the res
// of the DB in the meantime. We might have to refactor the actual
// actions for a ship/defense upgrade to create as many actions as
// needed so that we can put a fleet or any action in between two
// consecutive actions.
// Once the actions have been performed we can fetch the data safely
// knowing that nothing can be accessing it anymore.
// Which would then bring the fact that a fleet is a single element
// and if we want to fetch the fleet we only lock it (and not the
// target planet or anything).
// On the other hand, the planet will try to lock all the fleets
// that targets it and also the fleets that started from it (maybe).
// We could also revamp the way we handle fleets by not having by
// default a fleet and several fleet components but rather move
// part of the logic of the fleet component to the fleet itself
// and rather see the ACS as an extension of a regular fleet.
// For example we could have the fleet table look similar to what
// it is now. But the fleet_ships would reference not the fleet
// components but the fleets themselves.
// We could create a table like `fleets_acs_attack` and a similar
// one for the defense which would define some simple id and also
// probably the arrival_time.
// The `fleets_acs_attack_components` would reference both the
// `fleets` table and the `fleets_acs_attack` to define players
// that participate in the fleet. This would also be used in case
// we need to update the arrival time.
// So the `fleet` would define a `ACS Data` struct or something
// that would be used to determine whether this fleet is part of
// a larger attack.
// Regarding the actions queue, we could remove the call to the
// upgrade scripts from actual data (like in planet's method like
// `fetchDefenseUpgrades` for example) and deport everything in a
// common element in the `Instance` data object (or maybe even a
// method from the Instance itself). This queue would be lockable
// through some sort of mutex (maybe a concurrent lock ?) through
// a common (and always identical) resource string. Or maybe a
// simple mutex could do the trick.
// Maybe we shouldn't even have a concurrent lock but rather be
// locking the action queue as long as needed to perform the
// process that we want, and then release it. It would be very
// close to actually locking completely the server for every
// request but on the other hand it might be the simplest option.
// TODO: Restore the possibility to create ACS.

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
