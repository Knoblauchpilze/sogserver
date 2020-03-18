# sogserver
The server part which should be used to create a persistent environment for oglike home edition.

# Installation

The installation process requires a working version of:
 * The [docker runtime](https://docs.docker.com/install/linux/docker-ce/ubuntu/)
 * A working version of the [migrate tool](https://github.com/golang-migrate/migrate)
 * A working version of the [go language](https://golang.org/doc/install)

First, clone the repo through:
```git clone git@github.com:Knoblauchpilze/sogserver.git```

Go to the project's repository `cd ~/path/to/the/repo`. From there, one need to install the database within its own docker image and then build the server (and ultimately run it).

## Build the database docker image

- Go to `og_db`.
- Create the db: `make docker_db`.
- Run the db: `make create_db`.
- Connect to the db if needed through the `make connect`.

## Build the server

- Compile: `make r`
- Install: `make install`

# Usage

The server can be contacted through standard `HTTP` request at the port specified in the configuration.
