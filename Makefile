.SHELLFLAGS = -ec

SHELL         := /bin/bash

UID    ?= $(shell id -u)
GID    ?= $(shell id -g)

LOCAL_BUILD_IMAGE  = stackstate-agent-local-build
VOLUME_GO_PKG_NAME = ${LOCAL_BUILD_IMAGE}-go-volume
AGENT_SOURCE_MOUNT = /stackstate-agent-mount
PROJECT_DIR        = /go/src/github.com/StackVista/stackstate-agent

DOCKER_ENV		   = --env PROJECT_DIR=${PROJECT_DIR} \
                     --env artifactory_user=${ARTIFACTORY_USER} \
                     --env artifactory_password=${ARTIFACTORY_PASSWORD} \
                     --env ARTIFACTORY_PYPI_URL="artifactory.tooling.stackstate.io/artifactory/api/pypi/pypi-local/simple" \
                     --env PYTHON_RUNTIME=2


build:
	cd Dockerfiles/local_builder && \
	docker build -t ${LOCAL_BUILD_IMAGE} \
		--build-arg UID=${UID} \
		--build-arg GID=${GID} \
		.

source-shared: build
	docker run -it --rm \
        --name ${LOCAL_BUILD_IMAGE} \
        --mount source=${VOLUME_GO_PKG_NAME},target=/go/pkg \
        --volume ${PWD}:${PROJECT_DIR} \
        ${DOCKER_ENV} ${LOCAL_BUILD_IMAGE}

source-copy: build
	docker run -it --rm \
        --name ${LOCAL_BUILD_IMAGE} \
        --mount source=${VOLUME_GO_PKG_NAME},target=/go/pkg \
        --volume ${PWD}:${AGENT_SOURCE_MOUNT}:ro \
        --env AGENT_SOURCE_MOUNT=${AGENT_SOURCE_MOUNT} \
        ${DOCKER_ENV} ${LOCAL_BUILD_IMAGE} ${COPY_MOUNT}


shell:
	docker exec -ti ${LOCAL_BUILD_IMAGE} bash --init-file /local_init.sh

shell-root:
	docker exec --user root -ti ${LOCAL_BUILD_IMAGE} bash


.PHONY: build source-shared source-copy shell shell-root
