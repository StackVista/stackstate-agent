.SHELLFLAGS = -ec

SHELL := /bin/bash

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


IMAGE_NAME         = test-shards
CONTAINER_NAME     = sts-test
VOLUME_GO_PKG_NAME = sts-test-go-volume


build:
	docker build -t ${IMAGE_NAME}:${BRANCH} --build-arg UID=${UID} --build-arg GID=${GID} --build-arg DOCKER_GID=${DOCKER_GID} .

up: build
	docker run -it --rm \
		--name ${CONTAINER_NAME} \
		--volume ${PWD}:/go/src/app \
		--volume ${HOME}/.aws:/home/tester/.aws:ro \
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

exec:
	docker exec -ti ${CONTAINER_NAME} bash --init-file ./bootstrap.sh

exec-root:
	docker exec --user root -ti ${CONTAINER_NAME} bash


raw-debug:
	docker run -it --rm ${IMAGE_NAME}:${BRANCH}
