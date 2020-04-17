package data

import (
	"fmt"
)

// universesDependentProxy :
// Used to describe a proxy which uses the universe proxy
// as a composition of its internal attribute. It means
// that during its operations, the proxy can be using some
// info about universes from the DB. Rather than owning
// duplicate piece of code to fetch the universes it uses
// the base `UniverseProxy` to handle the requests.
//
// The `uniProxy` provides a way to access to the universes
// from the DB.
type universesDependentProxy struct {
	uniProxy UniverseProxy
}

// newUniversesDependentProxy :
// Performs the creation of a new proxy from the input
// argument, giving access to the universes features
// of the input argument.
//
// The `unis` defines a way to access to universes in
// the main DB.
//
// Returns the created object and panics if something is
// not right when creating the proxy.
func newUniversesDependentProxy(unis UniverseProxy) universesDependentProxy {
	return universesDependentProxy{
		unis,
	}
}

// fetchUniverse :
// Used to fetch the universe from the DB with an identifier
// matching the input one. If no such universe can be fetched
// an error is returned. If more that one universe seems to
// match the input identifier an error is returned as well.
//
// The `id` defines the index of the universe to fetch.
//
// Returns the universe corresponding to the input identifier
// along with any errors.
func (p *universesDependentProxy) fetchUniverse(id string) (Universe, error) {
	// Create the db filters from the input identifier.
	filters := make([]DBFilter, 1)

	filters[0] = DBFilter{
		"id",
		[]string{id},
	}

	unis, err := p.uniProxy.Universes(filters)

	// Check for errors and cases where we retrieve several
	// universes.
	if err != nil {
		return Universe{}, err
	}
	if len(unis) > 1 {
		err = fmt.Errorf("Retrieved %d universes for id \"%s\"", len(unis), id)
	}

	return unis[0], err
}

// playersDependentProxy :
// Similar to the `universeDependentProxy` but for cases
// where a proxy needs to have access to players during
// its life cycle.
//
// The `plyProxy` provides a way to access to the players
// from the DB.
type playersDependentProxy struct {
	plyProxy PlayerProxy
}

// newPlayersDependentProxy :
// Creates a new players dependent proxy from the input
// arguments.
//
// The `dbase` defines the main DB that should be wrapped
// by this object.
//
// Returns the created object and panics if something is
// not right when creating the proxy.
func newPlayersDependentProxy(players PlayerProxy) playersDependentProxy {
	return playersDependentProxy{
		players,
	}
}

// fetchPlayer :
// Allows to query the player with the specified identifier
// from the main DB. In case the player is not found or more
// than one player are found an error is returned.
//
// The `id` defines the identifier of the player to fetch.
//
// Returns the player corresponding to the identifier along with
// any error.
func (p *playersDependentProxy) fetchPlayer(id string) (Player, error) {
	// Create the db filters from the input identifier.
	filters := make([]DBFilter, 1)

	filters[0] = DBFilter{
		"id",
		[]string{id},
	}

	players, err := p.plyProxy.Players(filters)

	// Check for errors and cases where we retrieve several
	// players.
	if err != nil {
		return Player{}, err
	}
	if len(players) != 1 {
		return Player{}, fmt.Errorf("Retrieved %d players for id \"%s\" (expected 1)", len(players), id)
	}

	return players[0], nil
}

// planetsDependentProxy :
// Similar to the `universeDependentProxy` but for cases
// where a proxy needs to have access to planets during
// its life cycle.
//
// The `plaProxy` provides a way to access to the planets
// from the DB.
type planetsDependentProxy struct {
	plaProxy PlanetProxy
}

// newPlanetsDependentProxy :
// Creates a new planets dependent proxy from the input
// arguments.
//
// The `dbase` defines the main DB that should be wrapped
// by this object.
//
// The `log` defines the logger allowing to notify errors
// or info to the user.
//
// The `planets` defines a way to access to planets from
// the main DB.
//
// Returns the created object and panics if something is
// not right when creating the proxy.
func newPlanetsDependentProxy(planets PlanetProxy) planetsDependentProxy {
	return planetsDependentProxy{
		planets,
	}
}

// fetchPlanet :
// Allows to query the planet with the specified identifier
// from the main DB. In case the planet is not found or more
// than one planet are found an error is returned.
//
// The `id` defines the identifier of the planet to fetch.
//
// Returns the planet corresponding to the identifier along
// with any error.
func (p *planetsDependentProxy) fetchPlanet(id string) (Planet, error) {
	// Create the db filters from the input identifier.
	filters := make([]DBFilter, 1)

	filters[0] = DBFilter{
		"p.id",
		[]string{id},
	}

	planets, err := p.plaProxy.Planets(filters)

	// Check for errors and cases where we retrieve several
	// players.
	if err != nil {
		return Planet{}, err
	}
	if len(planets) != 1 {
		return Planet{}, fmt.Errorf("Retrieved %d planets for id \"%s\" (expected 1)", len(planets), id)
	}

	return planets[0], nil
}
