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

The server can be contacted through standard `HTTP` request at the port specified in the configuration files (see [configs](https://github.com/Knoblauchpilze/sogserver/tree/master/configs)).
