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

## Resources

Allows to fetch information about resources on the game through the `/resources` endpoint. Filtering can be done through the `id` and `name` query parameter to retrieve information on a particular resource. Each entry returns the information for a resource as follow:
 * `id`: the identifier of the resource.
 * `name`: the in game display name of the resource.
 * `base_production`: the base production for this resource on a planet without any infrastructure.
 * `base_storage`: the base storage for this resource on a planet with no infrastructure.
 * `base_amount`: the starting amount of this resource on a newly created planet.
 * `movable`: whether or not this resource can be moved (i.e. carried to another location by a fleet).
 * `storable`: whether or not this resource can be stored. This indicates whether the production of a resource will accumulate on the planet over time.

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
  "amount": 3
}
```

The `"element"` defines the identifier of the defense system to build while the `"amount"` defines the number of defense systems to build.

## Fleets

Fleets are a backbone of the game as they allow interactions between worlds and players. A fleet can be hostile or friendly and is composed of several ships carrying some resources. Each fleet can have a variety of objectives which define the purpose of the fleet.

### Objectives

A comprehensive description of the fleets objectives can be accessed through the `/fleets/objectives` route. It contains a description of the objective with the authrozied ships for this objective and whether or not this objective is hostile to the target destination.
Filters can be applied as defined below:
 * `id`: defines a filter on a specific fleet objective.
 * `name` : defines a filter on the name of the player.

### Fleet types

We distinguish two main fleet types: regular fleet and ACS fleets. Regular fleet are composed of a single batch of ships sent by a unique player from a unique location. An ACS operation on the other hand joins several fleets that can be sent by several players from distinct location. A mechanism ensures that the fleet is still considered as a single element and will be processed as so when it reaches its destination.
Both types of fleet have a similar set of operation and are accessible through `/fleets` or `/fleets/acs`.

### Regular fleets

Regular fleets can be fetched from the `/fleets` route and can be filtered using the following properties:
 * `id`: defines a filter on the identifier of a fleet.
 * `universe`: defines a filter on the universe to which the fleet belongs.
 * `objective`: defines a filter on the objective of the fleet.
 * `source`: defines a filter on the source element of the fleet: can refer to either a planet or a moon.
 * `target`: defines a filter on the target element of the fleet: can refer to either a planet or a moon. Note that in case the fleet is not directed (in case of a harvesting or colonization operation for example) filtering on this field will yield no results.
 * `galaxy`: defines a filter on the galaxy of the target of the fleet.
 * `solar_system`: defines a filter on the solar system of the target of the fleet.
 * `position`: defines a filter on the position of the target of the fleet.
 * `acs`: defines a filter on the ACS to which this fleet should belong.

The information returned by the fleet contains the general information such as the universe to which it belongs, the source and destination elements, the arrival and return time, etc. It also contains a list of the ships defined for this fleet and the cargo that is carried by the fleet.

A new fleet can be created through the `/fleets` endpoint. A fleet should define at least the following elements to be successfully created under the `fleet-data` key:

```json:
{
  "id": "fleet_id",
  "universe": "universe_id",
  "objective": "objective_id",
  "player": "player_id",
  "source": "id_of_source_celestial_body",
  "source_type": "planet|moon"
  "target_coordinates": {
    "galaxy": 1,
    "system": 260,
    "position": 10,
    "location": "planet|moon|debris"
  },
  "target": "id_of_target_celestial_body_if_any",
  "acs": "id_of_parent_acs_attack_if_any",
  "speed": 0.2,
  "deployment_time": 3600, /* 1 hour */
  "ships": [
    {
      "ship": "ship_id_1",
      "count": 1
    },
    {
      "ship": "ship_id_2",
      "count": 2
    }
  ]
  "cargo": [
    {
      "resource": "res_id_1",
      "amount": 26.2
    },
    {
      "resource": "res_id_2",
      "amount": 14
    }
  ]
}
```

The properties are listed below:
* `id`: the identifier of the fleet (will be automatically generated if empty).
* `universe`: the universe to which this fleet belongs.
* `objective`: the identifier of the objective for this fleet.
* `player`: the identifier of the player owning this fleet. The player should be registered in the `universe` specified.
* `source`: the identifier of the source location of the fleet (either a planet or a moon). The source should belong to the `player` and **must** be valid as a fleet alwyas starts from an existing location.
* `source_type`: defines the type of the `source` element: can be either `"planet"` or `"moon"`. Any other value will be interpreted as an error.
* `target_coordinates`: the location of the destination of the fleet. These coordinates should be consistent with the provided `universe` and with the actual objective of the fleet. The location of the target withtin the specified coordinate should be specified here (so either `planet` or `moon` or `debris`).
* `target`: the identifier of the existing celestial body to which this fleet is directed. Might be empty in case the `target_coordinates` indicate a `debris` location.
* `acs`: the identifier of the `ACS` operation to which this fleet belongs. When using the `/fleets` endpoint the `ACS` operation **must** exist already.
* `speed`: a floating point value in the range `]0; 1]` indicating the percentage of the maximal speed this fleet will be travelling at. The maximum speed is computed in the server from the speed of the ships belonging to the fleet.
* `DeploymentTime`: specifies the amount of time a fleet should be deployed at its destination before returning to its source location. This value is expressed in seconds and will be forcibly set to `0` in case the fleet objective does not allow any sort of deployment.
* `Ships`: defines an array for the ships belonging to the fleet. Each ship is referenced by its identifier (see the [Ships](https://github.com/Knoblauchpilze/sogserver#ships) section) and a count. The ships provided should be consistent with what's deployed on the source location. Each ship count should be stricly positive.
* `Cargo`: defines an array for the resources carried by the fleet. This amount should be consistent with both the amount stored on the planet and by the cargo capacity of the ships. Each amount should be stricly positive to be valid.

### ACS fleets

Creating an ACS fleet (for Alliance Combat System) or fetching the related data is very similar to creating a regular fleet. In order to fetch a particular ACS operation's data one should use the `/fleets/acs` endpoint. The filtering properties are defined below:
* `id`: defines a filter on the identifier of the ACS operation.
* `universe`: defines a filter on the identifier into whith the ACS takes place.
* `objective`: defines a filter on the objective of the ACS operation.
* `target`: defines a filter on the destination of the ACS. This value is guaranteed to be not null as we don't allow ACS operation on debris fiedls or anything similar.
* `target_type`: defines a filter to further refine the destination of the ACS fleet so either `"planet"` or `"moon"`.

The ACS fleet is composed several individual fleets all having the same objective, destination and arrival time. In order to create an ACS operation one must create a regular fleet and mark it for ACS operation. The server will valdiate the data as a regular fleet and perform the necessary adjustment so that the fleet is included in a ACS. An ID will be assigned to the ACS operation automatically and returned as an indication that the process was successful. The route to do so is `/fleets/acs` and the fleet's component (the first and only member at the time of creation) should be provided under the `fleet-data` key. The properties to do so are exactly the same as described in the [regular fleets](https://github.com/Knoblauchpilze/sogserver#regular-fleets) section as we want to create a valid component for this ACS.
