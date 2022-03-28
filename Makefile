.SHELLFLAGS = -ec

SHELL         := /bin/bash
.DEFAULT_GOAL := start

UID ?= $(shell id -u)
GID ?= $(shell id -g)
OS := $(shell uname)

ifeq ($(OS),Linux)
    DOCKER_GID ?= $(shell stat -c '%g' /var/run/docker.sock)
    UNAME_S = Linux
endif
ifeq ($(OS),Darwin)
    DOCKER_GID ?= $(shell stat -f '%g' /var/run/docker.sock)
    UNAME_S = Darwin
endif

GIT_BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
BRANCH = $(shell echo ${GIT_BRANCH} | tr [:upper:] [:lower:])

MAJOR_VERSION ?= 3

AGENT_DEVELOP_IMAGE     = agent-build:${BRANCH}
AGENT_DEVELOP_CONTAINER = agent-local-build
VOLUME_GO_PKG_NAME      = agent-local-build-volume
DOCKER_ENV		        = --env AGENT_CURRENT_BRANCH=${GIT_BRANCH} \
					      --env MAJOR_VERSION=${MAJOR_VERSION} \
					      --env artifactory_user=$artifactory_user \
					      --env artifactory_password=$artifactory_password \
                          --env ARTIFACTORY_PYPI_URL="artifactory.tooling.stackstate.io/artifactory/api/pypi/pypi-local/simple"

build:
	docker build -t ${AGENT_DEVELOP_IMAGE} \
		--build-arg UID=${UID} \
		--build-arg GID=${GID} \
		--build-arg DOCKER_GID=${DOCKER_GID} \
		--build-arg UNAME_S=${UNAME_S} \
		-f ./Dockerfiles/local_builder/Dockerfile ./Dockerfiles/local_builder

start: build
#    $(info This docker file will setup a docker container with a clone of the current agent directory.)
#    $(info The current directory will be mounted as a volume and can be pulled from, but the build is fully separated form the host system.)
	docker run -it --rm \
        --name ${AGENT_DEVELOP_CONTAINER} \
        --volume ${PWD}:/go/src/app \
        --volume /var/run/docker.sock:/var/run/docker.sock:ro \
        --mount source=${VOLUME_GO_PKG_NAME},target=/go/pkg \
        --network host \
        ${DOCKER_ENV} \
        ${AGENT_DEVELOP_IMAGE}

shell:
	docker exec -ti ${AGENT_DEVELOP_CONTAINER} bash --init-file ./shell.sh

shell-root:
	docker exec --user root -ti ${AGENT_DEVELOP_CONTAINER} bash

.PHONY: build start shell shell-root
