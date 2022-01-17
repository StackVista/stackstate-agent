.SHELLFLAGS = -ec

SHELL         := /bin/bash
.DEFAULT_GOAL := start

UID    ?= $(shell id -u)
GID    ?= $(shell id -g)

OS := $(shell uname)
ifeq ($(OS),Linux)
    DOCKER_GID ?= $(shell stat -c '%g' /var/run/docker.sock)
endif
ifeq ($(OS),Darwin)
	DOCKER_GID ?= $(shell stat -f '%g' /var/run/docker.sock)
endif

GIT_BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
BRANCH = $(shell echo ${GIT_BRANCH} | cut -d'/' -f 2- | tr [:upper:] [:lower:])

STACKSTATE_BRANCH            ?= master
MAJOR_VERSION                ?= 2
STS_AWS_TEST_BUCKET          ?= stackstate-agent-2-test
STS_DOCKER_TEST_REPO         ?= stackstate-agent-2-test
STS_DOCKER_TEST_REPO_CLUSTER ?= stackstate-cluster-agent-test

BASE_KEEPER_IMAGE  = stackstate/stackstate-agent-2-test:beest-keeper-base-20220113
KEEPER_IMAGE       = beest:${BRANCH}
KEEPER_CONTAINER   = beest
VOLUME_GO_PKG_NAME = beest-go-volume
DOCKER_ENV		   = --env AGENT_CURRENT_BRANCH=${GIT_BRANCH} \
					 --env STACKSTATE_BRANCH=${STACKSTATE_BRANCH} \
					 --env MAJOR_VERSION=${MAJOR_VERSION} \
					 --env STS_AWS_TEST_BUCKET=${STS_AWS_TEST_BUCKET} \
					 --env STS_DOCKER_TEST_REPO=${STS_DOCKER_TEST_REPO} \
					 --env STS_DOCKER_TEST_REPO_CLUSTER=${STS_DOCKER_TEST_REPO_CLUSTER}


build-base:
	docker build -t ${BASE_KEEPER_IMAGE} -f ./keeper.base.dockerfile .

push-base: build-base
	docker push ${BASE_KEEPER_IMAGE}

build:
	docker build -t ${KEEPER_IMAGE} \
		--build-arg BASE_IMAGE=${BASE_KEEPER_IMAGE} \
		--build-arg UID=${UID} \
		--build-arg GID=${GID} \
		--build-arg DOCKER_GID=${DOCKER_GID} \
		-f ./keeper.dockerfile .

start: build
	docker run -it --rm \
        --name ${KEEPER_CONTAINER} \
        --volume ${PWD}:/go/src/app \
        --volume ${HOME}/.aws:/home/keeper/.aws \
        --volume /var/run/docker.sock:/var/run/docker.sock:ro \
        --mount source=${VOLUME_GO_PKG_NAME},target=/go/pkg \
        --network host \
        ${DOCKER_ENV} \
        ${KEEPER_IMAGE}

shell:
	docker exec -ti ${KEEPER_CONTAINER} bash --init-file ./bootstrap.sh

shell-root:
	docker exec --user root -ti ${KEEPER_CONTAINER} bash


gitlab-debug:
	docker run -it --rm \
		--volume ${PWD}:/root/beest \
		--env CI_PROJECT_DIR=/root \
		${DOCKER_ENV} \
		${BASE_KEEPER_IMAGE}
