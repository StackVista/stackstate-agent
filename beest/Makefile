include sut/yards/splunk/Makefile

.SHELLFLAGS = -ec

SHELL         := /bin/bash
.DEFAULT_GOAL := start

UID    ?= $(shell id -u)
GID    ?= $(shell id -g)

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
GIT_SHA    = $(shell git rev-parse --short=8 HEAD)
BRANCH = $(shell echo ${GIT_BRANCH} | tr [:upper:] [:lower:])

MAJOR_VERSION                ?= 2
STS_AWS_TEST_BUCKET          ?= stackstate-agent-3-test
STS_DOCKER_TEST_REPO         ?= stackstate-agent-test
STS_DOCKER_TEST_REPO_CLUSTER ?= stackstate-cluster-agent-test
PROCESS_AGENT_TEST_REPO      ?= stackstate-process-agent-test

BASE_KEEPER_IMAGE  = artifactory.tooling.stackstate.io/docker-virtual/stackstate/beest:20230515
KEEPER_IMAGE       = beest:${BRANCH}
KEEPER_CONTAINER   = beest
VOLUME_GO_PKG_NAME = beest-go-volume
DOCKER_ENV		   = --env AGENT_CURRENT_BRANCH=${GIT_BRANCH} \
					 --env AGENT_CURRENT_BRANCH_REF=${GIT_SHA} \
					 --env MAJOR_VERSION=${MAJOR_VERSION} \
					 --env STS_AWS_TEST_BUCKET=${STS_AWS_TEST_BUCKET} \
					 --env STS_DOCKER_TEST_REPO=${STS_DOCKER_TEST_REPO} \
					 --env STS_DOCKER_TEST_REPO_CLUSTER=${STS_DOCKER_TEST_REPO_CLUSTER} \
					 --env PROCESS_AGENT_TEST_REPO=${PROCESS_AGENT_TEST_REPO}

# Core commands
login:
	docker login artifactory.tooling.stackstate.io -u ${artifactory_user} -p ${artifactory_password}

build: login
	docker build -t ${KEEPER_IMAGE} \
		--build-arg BASE_IMAGE=${BASE_KEEPER_IMAGE} \
		--build-arg UID=${UID} \
		--build-arg GID=${GID} \
		--build-arg DOCKER_GID=${DOCKER_GID} \
		--build-arg UNAME_S=${UNAME_S} \
		-f ./keeper.dockerfile .

start: build
	docker run -it --rm \
        --name ${KEEPER_CONTAINER} \
        --volume ${PWD}:/go/src/app \
        --volume ${PWD}/../dev/agent-v1-to-v2-pickle-conversion:/go/src/app/tools/agent-v1-to-v2-pickle-conversion \
        --volume /var/run/docker.sock:/var/run/docker.sock:ro \
        --mount source=${VOLUME_GO_PKG_NAME},target=/go/pkg \
        --network host \
        ${DOCKER_ENV} \
        ${KEEPER_IMAGE}

stop:
	-docker kill ${KEEPER_CONTAINER}

shell:
	docker exec -ti ${KEEPER_CONTAINER} bash --init-file ./bootstrap.sh

shell-root:
	docker exec --user root -ti ${KEEPER_CONTAINER} bash

# Beest go unit test
go-test:
	go test -v ./...

# Beest verification fixture tests
matcher-test:
	PYTHONPATH=./testframework/stscliv1/ pytest testframework

.PHONY: build start shell shell-root go-test matcher-test
