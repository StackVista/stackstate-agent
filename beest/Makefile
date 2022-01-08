.SHELLFLAGS = -ec

SHELL         := /bin/bash
.DEFAULT_GOAL := beest

UID    ?= $(shell id -u)
GID    ?= $(shell id -g)

OS := $(shell uname)
ifeq ($(OS),Linux)
    DOCKER_GID ?= $(shell stat -c '%g' /var/run/docker.sock)
endif
ifeq ($(OS),Darwin)
	DOCKER_GID ?= $(shell stat -f '%g' /var/run/docker.sock)
endif

BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD | cut -d'/' -f 2- | tr [:upper:] [:lower:])

CI_COMMIT_REF_NAME           ?= ${BRANCH}
STACKSTATE_BRANCH            ?= master
MAJOR_VERSION                ?= 2
STS_AWS_TEST_BUCKET          ?= stackstate-agent-2-test
STS_DOCKER_TEST_REPO         ?= stackstate-agent-2-test
STS_DOCKER_TEST_REPO_CLUSTER ?= stackstate-cluster-agent-test

TODAY			   = $(shell date +%Y%m%d)
BASE_KEEPER_IMAGE  = stackstate/stackstate-agent-2-test:beest-keeper-base-${TODAY}
IMAGE_NAME         = beest
CONTAINER_NAME     = beest
VOLUME_GO_PKG_NAME = beest-go-volume

DOCKER_RUN = docker run -it --rm \
             	--name ${CONTAINER_NAME} \
             	--volume ${PWD}:/go/src/app \
             	--volume ${HOME}/.aws:/home/keeper/.aws \
             	--volume /var/run/docker.sock:/var/run/docker.sock:ro \
             	--mount source=${VOLUME_GO_PKG_NAME},target=/go/pkg \
             	--network host \
             	--env AGENT_CURRENT_BRANCH=${CI_COMMIT_REF_NAME} \
             	--env STACKSTATE_BRANCH=${STACKSTATE_BRANCH} \
             	--env MAJOR_VERSION=${MAJOR_VERSION} \
             	--env STS_AWS_TEST_BUCKET=${STS_AWS_TEST_BUCKET} \
             	--env STS_DOCKER_TEST_REPO=${STS_DOCKER_TEST_REPO} \
             	--env STS_DOCKER_TEST_REPO_CLUSTER=${STS_DOCKER_TEST_REPO_CLUSTER} \
             	${IMAGE_NAME}:${BRANCH}

DOCKER_RUN_M = docker run --rm \
             	--volume ${PWD}:/go/src/app \
             	--env AGENT_CURRENT_BRANCH=${CI_COMMIT_REF_NAME} \
             	--env STACKSTATE_BRANCH=${STACKSTATE_BRANCH} \
             	--env MAJOR_VERSION=${MAJOR_VERSION} \
             	--env STS_AWS_TEST_BUCKET=${STS_AWS_TEST_BUCKET} \
             	--env STS_DOCKER_TEST_REPO=${STS_DOCKER_TEST_REPO} \
             	--env STS_DOCKER_TEST_REPO_CLUSTER=${STS_DOCKER_TEST_REPO_CLUSTER} \
             	--workdir /go/src/app \
             	${BASE_KEEPER_IMAGE}


build-base:
	docker build -t ${BASE_KEEPER_IMAGE} -f ./keeper.base.dockerfile .

push-base: build-base
	docker push ${BASE_KEEPER_IMAGE}

beest:
	${DOCKER_RUN_M} ./beest.sh $(action)


build:
	docker build -t ${IMAGE_NAME}:${BRANCH} \
		--build-arg UID=${UID} \
		--build-arg GID=${GID} \
		--build-arg DOCKER_GID=${DOCKER_GID} \
		-f ./keeper.dockerfile .

start: build
	${DOCKER_RUN}

shell:
	docker exec -ti ${CONTAINER_NAME} bash --init-file ./bootstrap.sh

shell-root:
	docker exec --user root -ti ${CONTAINER_NAME} bash


raw-debug: build
	docker run -it --rm ${IMAGE_NAME}:${BRANCH}

.PHONY: beest
