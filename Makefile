# Main `Makefile` to build the server. The layout of this file has
# been inspired by the following links:
# https://github.com/golang-standards/project-layout
# https://github.com/helm/helm/blob/master/Makefile

# Common build variables.
BINDIR      := $(CURDIR)/bin
DIST_DIRS   := find * -type d -exec
TARGETS     := linux/amd64
TARGET_OBJS ?= linux-amd64.tar.gz linux-amd64.tar.gz.sha256 linux-amd64.tar.gz.sha256sum
BINNAME     ?= oglike_server

GOPATH        = $(shell go env GOPATH)
ARCH          = $(shell uname -p)

# Docker setup.
SERVER_IMAGE_NAME     = oglike_image
SERVER_CONTAINER_NAME = oglike_container
SERVER_PORT           = 3007

# Go options.
PKG        := ./...
TAGS       :=
TESTS      := .
TESTFLAGS  :=
LDFLAGS    := -w -s
GOFLAGS    :=
SRC        := $(shell find . -type f -name '*.go' -print)

# Git information.
GIT_COMMIT = $(shell git rev-parse HEAD)
GIT_SHA    = $(shell git rev-parse --short HEAD)
GIT_TAG    = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY  = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")

.PHONY: all
all: build

# Target defining the build operation for the server.
build: $(BINDIR)/$(BINNAME)

$(BINDIR)/$(BINNAME): $(SRC)
	GO111MODULE=on go build $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(BINNAME) ./cmd/oglike_server

# Target to clean any existing build results.
clean:
	@rm -rf $(BINDIR)
	@rm -rf sandbox

# Target providing information about the current version and
# git status of the project.
info:
	 @echo "Version:           ${VERSION}"
	 @echo "Git Tag:           ${GIT_TAG}"
	 @echo "Git Commit:        ${GIT_COMMIT}"
	 @echo "Git Tree State:    ${GIT_DIRTY}"

# Target performing an install of the server into a sandbox
# environment resembling the production setup.
install: build
	@mkdir -p sandbox
	@cp scripts/*.sh sandbox
	@cp -r $(BINDIR) sandbox
	@cp configs/*.yml sandbox

# Target providing a way to compile and run the server.
run: install
	@cd sandbox && ./run.sh local

# Target allowing to build the docker image for the server.
docker:
	docker build -t ${SERVER_IMAGE_NAME} .

# Target allowing to remove any existing docker image of the server.
remove: stop
	docker rm ${SERVER_CONTAINER_NAME}

# Target allowing to create the docker image for the server.
create:
	docker run -d --name ${SERVER_CONTAINER_NAME} -p ${SERVER_PORT}:${SERVER_PORT} ${SERVER_IMAGE_NAME}

# Target allowing to start the docker image for the server.
start:
	docker start ${SERVER_CONTAINER_NAME}

# Target allowing to stop the docker image fot the server.
stop:
	docker stop ${SERVER_CONTAINER_NAME}
