.SHELLFLAGS = -ec

SHELL         := /bin/bash
.DEFAULT_GOAL := start

UID    ?= $(shell id -u)
GID    ?= $(shell id -g)

BASE_AGENT_IMAGE   = "artifactory.tooling.stackstate.io/docker-virtual/stackstate/stackstate-agent-runner-gitlab:debian-20220505"
LOCAL_BUILD_IMAGE  = stackstate-agent-local-build
VOLUME_GO_PKG_NAME = ${LOCAL_BUILD_IMAGE}-go-volume
DOCKER_ENV		   = --env artifactory_user=${ARTIFACTORY_USER} \
					 --env artifactory_password=${ARTIFACTORY_PASSWORD} \
					 --env ARTIFACTORY_PYPI_URL="artifactory.tooling.stackstate.io/artifactory/api/pypi/pypi-local/simple"

build:
	docker build -t ${LOCAL_BUILD_IMAGE} \
		--build-arg BASE_IMAGE=${BASE_AGENT_IMAGE} \
		--build-arg UID=${UID} \
		--build-arg GID=${GID} \
		-f ./rtloader/Dockerfile .

start: build
	docker run -it --rm \
        --name ${LOCAL_BUILD_IMAGE} \
        --volume ${PWD}:/go/src/github.com/StackVista/stackstate-agent \
        --mount source=${VOLUME_GO_PKG_NAME},target=/go/pkg \
        ${DOCKER_ENV} ${LOCAL_BUILD_IMAGE}

orig:
	docker run -it --rm \
        --volume ${PWD}:/go/src/github.com/StackVista/stackstate-agent \
        --mount source=${VOLUME_GO_PKG_NAME},target=/go/pkg \
        --entrypoint /go/src/github.com/StackVista/stackstate-agent/Dockerfiles/local_builder/scripts/entry_point.sh \
        ${DOCKER_ENV} ${BASE_AGENT_IMAGE}


shell:
	docker exec -ti ${LOCAL_BUILD_IMAGE} bash --init-file ./bootstrap.sh

shell-root:
	docker exec --user root -ti ${LOCAL_BUILD_IMAGE} bash

.PHONY: build start shell shell-root
