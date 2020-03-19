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

- Go to `og_db`.
- Create the db: `make docker_db`.
- Run the db: `make create_db`.
- Connect to the db if needed through the `make connect` target.

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

# Usage

The server can be contacted through standard `HTTP` request at the port specified in the configuration files (see [configs](https://github.com/Knoblauchpilze/sogserver/tree/master/configs)).
