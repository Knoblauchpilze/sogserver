# sogserver
The server part which should be used to create a persistent environment for oglike home edition.

# Installation

The installation process requires a working version of:
 * The [docker runtime](https://docs.docker.com/install/linux/docker-ce/ubuntu/).
 * The [migrate tool](https://github.com/golang-migrate/migrate).
 * The [go language](https://golang.org/doc/install).

First, clone the repo through:
```git clone git@github.com:Knoblauchpilze/sogserver.git```

Go to the project's repository `cd ~/path/to/the/repo`. From there, one needs to install the database within its own docker image and then build the server (and ultimately run it).

## Build the database docker image

Creating the database is useful so that the server can access tot he data model needed to perform the computations for the application. This includes all the hard-coded values used by the game during the simulation. This consists in two steps: building the db container and initializing it.

### Create the DB container

- Go to `og_db`.
- Create the db: `make docker_db`.
- Run the db: `make create_db`. Note that in case a previous operation already succeeded one should call `make remove_db` beforehand as a container with this name already exists.
- Initialize the database by calling the `make migrate` target: this will create the schema associated to the data model of the application and populate the needed fields.

### Iterate on the DB schema

In case some new information need to be added to the database one can use the migrations mechanism. By creating a new migration file in the relevant [directory](https://github.com/Knoblauchpilze/sogserver/tree/master/og_db/migrations) and naming accordingly (increment the number so that the `migrate` tool knows in which order migrations should be ran) it is possible to perform some modifications of the db by altering some properties. The migration should respect the existing constraints on the tables.
Once this is done one can rebuild the db by using the `make migrate` target which will only apply the migrations not yet persisted in the db schema.

### Managing the DB

If the db container has been stopped for some reasons one can relaunch it through the `make start_db` command. One can also directly connect to the db using the `make connect` command. The password to do so can be found in the configuration files.
In case a rebuild of the db is needed please proceed to launch the following commands:
 - `make remove_db` will stop the db container (if needed) and remove any existing images/container image referencing it.
 - `make docker_db` will rebuild the docker image of the db.
 - `make create_db` will run the docker image as a fully-fleshed container.
 - `make migrate` will initialize the db schema.

Note that these commands should be launched directly fro the `og_db` directory.

## Build the server

Once the directory is cloned, move to the project's repository with `cd ~/path/to/the/repo`. From there launch the following commands:
 * `go mod init oglike_server`
 * `make`

This should build the server and perform a launch of the executable in a controlled environment. To develop and integrate new features, several other targets are provided in the root `Makefile`:
 * `build`: build the server.
 * `clean`: clean any existing build results.
 * `info`: provide some information about the current git status of the project.
 * `install`: copy the latest result of the build to a sandbox environment.
 * `run`: perform a build of the server and run the latest result.

## Build the server docker image

The server can also be packaged independently in a docker image. This is also provided by the root `Makefile` with dedicated targets:
 * Remove any existing container (only if a `make create` has already been launched before): `make remove`.
 * Create the docker image: `make docker`.
 * Create the docker container from the image: `make create`.
 * Run the container: `make start`.

Note that the server for now is only reachable if it uses the port `3000` as defined in the `Dockerfile`. This is a limitation that it is not yet planned to correct.
To see the logs of the container one can use the `docker logs -f oglike_container` command. If the environment to use should be modified the environment variable `APP_ENVIRONMENT` can be overriden when creating the docker. So instead of the `make create` one can use the following option when launching the container: `--env APP_ENVIRONMENT=production`.

Also note that for now the container cannot reach the DB as the `5500` port (or the port used by the DB for that matter) cannot be accessed from the container.

# Usage

The server allows to query information from the DB through various endpoints. We distinguish between the `GET` semantic where the user wants to access some information and the `POST` requests typically used when some data should be created on the server. The `GET` syntax is similar for most of the resources. The user can query the collection of resources of a particular type through the `/resource-name` endpoint and individual elements of the collection through `/resource-name/resource-id` or using query parameters with something along the lines of `/resource-name?resource_id=id`.

This syntax is similar for most endpoints. It is consistent with what's expected of a [REST](https://en.wikipedia.org/wiki/Representational_state_transfer) API.

All the endpoints can be accessed through standard `HTTP` request at the port specified in the configuration files (see [configs](https://github.com/Knoblauchpilze/sogserver/tree/master/configs)).

## Universes

Provides information related to the universes under the `/universes` route. This will query a list of all the created universes. One can request the information of a specific universe using the `/universes/uni-id` syntax or through query parameters using `/universes?id=id` or `/universes?name=name` or both. This will fetch only universes matching the filters.

The user can also request the creation of a universe through the `/universe` endpoint. It requires to provide some data about the properties actually defining the universe under the `universe-data` key where the user needs to specify the following properties:
 * `id`: defines the identifier of the universe (will be auto-generated if left empty).
 * `name`: the display name for this universe.
 * `economic_speed`: an integer representing the multiplier against the canonical times related to economy (should be greater than `0`).
 * `fleet_speed`: similar to the `economic_speed` but defines the fleets travel speed.
 * `research_speed`: similar to `economic_speed` but defines a multiplier applied on research times.
 * `fleets_to_ruins_ratio`: a value in the range `[0; 1]` defining how much of the construction resources of a fleet end up in a ruins fields after a ship is destroyed.
 * `defenses_to_ruins_ratio`: a value similar to `fleets_to_ruins_ratio` but defining how much of the defenses resources go in the ruins.
 * `fleets_consumption_ratio`: a value in the range `[0; 1]` defining a multiplier on the canonical consumption of each ship.
 * `galaxies_count`: an integer larger than `0` defining how many galaxies are defined in the universe.
 * `galaxy_size`: an integer larger than `0` defining how many solar systems are defined in each galaxy.
 * `solar_system_size`: an integer larger than `0` defining how many planets exist in a single solar system.

## Accounts

Very similar to the `/universes` endpoint but allows to query information about the accounts. The routes are described below:
 * `/accounts`: queries information about all accounts.
 * `/accounts/account-id`: queries information about a specific account, retrieving its identifier, name and e-mail.
 * `/accounts?id=id`: queries information about a specific account (similarly to the above).
 * `/accounts?name=name`: queries information about accounts matching the specified name.
 * `/accounts?mail=mail`: queries information about accounts matching the specified e-mail.

The main purpose of an account is to actually represent the instance of a user in the game. One can create a new account through the `/account` endpoint and by providing data under the `account-data` key. The following properties should be defined to be able to successfully create the account:
 * `id`: defines the identifier of the account. Should be unique and will be auto-generated if left empty.
 * `name`: a display name for the account. Does not need to be unique but should not be empty.
 * `mail`: the e-mail address associated to the account. Should be unique among all accounts.
 * `password`: the password to use to log-in on this account.

## Buildings

Allows to fetch information about buildings that can be built on planets through the `/buildings` endpoint. Just like the other endpoints one can query a specific building through the query parameters `id` and `name`. Each entry in the return value contains the cost of the building, its effects in terms of storage and production, the potential consumption of resources and the tech tree. The tech tree represents the set of technologies that the player should have researched to gain access to this building and the set of buildings that should already exist on a planet before being allowed to build it.

## Technologies

Exactly the same as `buildings` but for technologies available in the game. The endpoint is `/technologies` and the query parameters to fetch specific information are `id` and `name`. Similar properties to buildings are returned like the costs and requirements of each technology.

## Ships

Works the same way as the others. The main endpoint is `/ships` and the properties are `id` and `name`. The returned values include the base armament capacities of each ship along with the definition of the rapid fires eacy ship has on other ships or defense systems. It also include information about the propulsion system used by the ships and their base speed.

## Defenses

Similar to the rest, the endpoint is `/defenses` and the properties are `id` and `name`. The costs along with the armament capacities are returned for each defense matching the filters.

## Planets

The planets endpoint allows to query planets on a variety of criteria. The main endpoint is accessible through the `/planets` route and serves all the planets matching the query filters. The user can query a particular planet by providing its identifier in the route (e.g. `/planets/planet_id`).

The user also has access to some query parameters:
 * `player`: defines a filter on the player owning the planet.
 * `name`: defines a filter on the name of the planet.
 * `galaxy` : defines a filter on the galaxy of the planet.
 * `solar_system` : defines a filter on the solar system of the planet.
 * `universe` : defines a filter on the position of the planet.
 * `id` : defines a filter on the identifier of the planet.

These filters can be combined between each other and it's always the case in a `AND` semantic (meaning that a planet must match all filters to be returned). The individual description of the planet regroups the ships existing on the planet, the buildings built on it, the defenses installed on it and also the resources that are currently present on it. It also defines the upgrade actions attached to the planet which are actions that aim at improving the infrastructure available on the planet.

## Players

The `/players` allows to access the individual instance of accounts in universes. A player is linked to a single account and each player can only be present once in a universe. Most of the functionalities are related to the universe the player belongs to, the account to which it is linked and also the technologies that are associated to the player. Note that all the technologies researched by this player are returned by this endpoint.

Here are the available query parameters:
 * `player` : defines a filter on the identifier of the player.
 * `account` : defines a filter on the account linked to the player: this allows to get all the universes into which an account is registered.
 * `universe` : defines a filter on the universe the player belongs to.
 * `name` : defines a filter on the name of the player.

It is possible to create a new player, which represents the instance of an account in a universe. This can be done through the `/player` endpoint and by specifying the data under the `player-data` key. The following properties can be specified:
 * `id`: an identifier representing this player. It will be auto-generated if left empty. It is automatically generated if not specified.
 * `account`: the identifier of the account linked to this player. Should be matching an existing account.
 * `universe`: the identifier of the universe into which the player should be created. No other `player` linked to the same account should exist in this universe.
 * `name`: the display name of the player in the universe. Should be unique but does not need to be (maybe we should modify that at some point).

## Construction actions

The `/planets` routes also serves the upgrade actions that are registered for a given planet. Upgrade actions are the core mechanism of the game allowing a player to improve a planet by building more levels of a building, research or more ships. It is always linked to a planet as we need the resources to perform the action.
Any upgrade action can be performed provided that the conditions are met for example regarding technologies dependencies or buildings dependencies in the case of ships for example. Various endpoints allow to create upgrade action for a planet through the `planets/planet_id/actions/XYZ` syntax where `XYZ` is one of `buildings`, `technologies`, `ships` or `defenses`.
In any case the data to provide to create a new upgrade action should be registered under the key `action-data`.

### Buildings

Buildings upgrade action should match either an upgrade of a building of one level or a destruction of the last level of a building. The `json` object to provide to create such an action should look like below:

```json:
{
	"element": "element_id",
  "current_level": 1,
  "desired_level": 2
}
```

Where the `"element"` key references an identifier describing the building to upgrade and the `"current_level"` and `"desired_level"` the current level of the building on the planet and its desired level. Internally this action also relies on the index of the planet and the completion time but these information are computed from the route on one part and computed internally on the second part (so that there's no possibility for someone to cheat).
Only one action of this type can be active at any moment for a given planet.

### Technologies

The technologies upgrade action allows to research a new technology for a player. While the technology will be applied to all the planets of the player at once when it's finished, only a single planet can be used to actually perform the research. The `json` object to provide to create such an action should look like below:

```json:
{
  "element": "element_id",
  "current_level": 0,
  "desired_level": 1
}
```

The `"element"` key defines the identifier of the technology which should be researched and both levels give indication on the current state for the player and the desired state. Note that unlike buildings, technologies cannot be un-researched. Internally the technology action is set up with the player owning the planet where it's started.
Only one action of this type can be active at any moment for a single player.

### Ships

The ships upgrade are a bit different from the technologies and buildings counterparts in the sense that they can be queued and there's no limit as to how many actions can be scheduled concurrently. The ships are built in the shipyard of the planet and they are queued after the last existing action. Typically if an action still requires 30 minutes to finish, requesting a new ship construction action will queue it after the 30 minutes of the current action have passed.

The `json` object to provide to the route looks like below:

```json:
{
  "element": "element_id",
  "amount": 25
}
```

The `"element"` describes the identifier of the ship to build and the `"amount"` defines how many ships should be built.

### Defenses

The defenses upgrade action is very similar to the `ships` case but it concerns actions. The actions are queued indifferently after the last construction action (either a defense one or a ship one). The associated `json` object is very similar to the one used for ships:

```json:
{
  "element": "element_id",
  "amount": 25
}
```

The `"element"` defines the identifier of the defense system to build while the `"amount"` defines the number of defense systems to build.

## Fleets

TODO: Should implement this.