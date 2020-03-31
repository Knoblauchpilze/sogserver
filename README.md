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

Provides information related to the universes under the `/universes` route. This will query a list of all the created universes. One can request the information of a specific universe using the `/universes/uni-id` syntax or through query parameters using `/universes?universe_id=id` or `/universes?universe_name=name` or both. This will fetch only universes matching the filters.

## Accounts

Very similar to the `/universes` endpoint but allows to query information about the accounts. The routes are described below:
 * `/accounts`: queries information about all accounts.
 * `/accounts/account-id`: queries information about a specific account, retrieving its identifier, name and e-mail.
 * `/accounts?account_id=id`: queries information about a specific account (similarly to the above).
 * `/accounts?account_name=name`: queries information about accounts matching the specified name.
 * `/accounts?account_mail=mail`: queries information about accounts matching the specified e-mail.

## Buildings

Allows to fetch information about buildings that can be built on planets through the `/buildings` endpoint. Just like the other endpoints one can query a specific building through the query parameters `building_id` and `building_name`.

## Technologies

Exactly the same as `buildings` but for technologies available in the game. The endpoint is `/technologies` and the query parameters to fetch specific information are `technology_id` and `technology_name`.

## Ships

Works the same way as the others. The main endpoint is `/ships` and the properties are `ship_id` and `ship_name`.

## Defenses

Similar to the rest, the endpoint is `/defenses` and the properties are `defense_id` and `defense_name`.

## Planets

The planets endpoint allows to query planets on a variety of criteria. The main endpoint is accessible through the `/planets` route and serves all the planets registered in the server no matter the universe (thus it's not very helpful). The user can query a particular planet by providing its identifier in the route (e.g. `/planets/planet_id`).

The user also has access to some query parameters:
 * planet_id":    pa.proxy.GetIdentifierDBColumnName(),
 * `planet_name`: defines a filter on the name of the planet.
 * `galaxy` : defines a filter on the galaxy of the planet.
 * `solar_system` : defines a filter on the solar system of the planet.
 * `universe` : defines a filter on the position of the planet.
 * `player_id` : defines a filter on the identifier of the player owning the planet.
 * `account_id` : defines a filter on the identifier of the account owning the planet.

These filters can be combined between each other and it's always the case in a `AND` semantic (meaning that a planet must match all filters to be returned). The individual description of the planet regroups the ships existing on the planet, the buildings built on it, the defenses installed on it and also the resources that are currently present on it.

## Players

The `/players` allows to access the individual instance of accounts in universes. A player is linked to a single account and each player can only be present once in a universe. Most of the functionalities are related to the universe the player belongs to, the account to which it is linked and also the technologies that are associated to the player.

Here are the available query parameters:
 * `player_id` : defines a filter on the identifier of the player.
 * `account_id` : defines a filter on the account linked to the player: this allows to get all the universes into which an account is registered.
 * `universe_id` : defines a filter on the universe the player belongs to.
 * `player_name` : defines a filter on the name of the player.

## Fleets

The `/fleets` endpoint allows to fetch information on the fleet currently moving through the server. There is no real way to distinguish between fleets of different universes but we instead rely on providing the coordinates of the destination of the fleet and thus access the fleet through its target planet. The available query parameters are:
 * `fleet_id` : defines a filter on the identifier of the fleet (can be accessed directly through the route).
 * `fleet_name` : defines a filter on the name of the fleet (if any).
 * `galaxy` : defines a filter on the target galaxy of the fleet.
 * `solar_system` : defines a filter on the target solar system of the fleet.
 * `position` : defines a filter on the target position of the fleet.

For each query the detailed information of the fleet are retrieved which describe the number of ships and the composition of the successive waves of the fleet (players involved, starting position of each one, etc.).

For now there are some missing possibilities but it will be added on the go.
