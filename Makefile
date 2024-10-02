.SHELLFLAGS = -ec

SHELL         := /bin/bash
.DEFAULT_GOAL := dev

UID    ?= $(shell id -u)
GID    ?= $(shell id -g)
# Workaround for target completion, because Makefile does not like : in the target commands
colon  := :

PYTHON_RUNTIME    ?= 2

LOCAL_BUILD_IMAGE  = stackstate-agent-local-build
VOLUME_GO_PKG_NAME = ${LOCAL_BUILD_IMAGE}-go-volume
AGENT_SOURCE_MOUNT = /stackstate-agent-mount
PROJECT_DIR        = /go/src/github.com/StackVista/stackstate-agent
LOCAL_BUILDER_INIT = Dockerfiles/local_builder/local_init.sh

# Set this parameter to 'dit' to use the make command in your IDE, You can then execute shell commands allowing the container to run in detached mode
DOCKER_RUN_MODE   ?= it
DOCKER_ENV		   = --env PROJECT_DIR=${PROJECT_DIR} \
                     --env artifactory_user=${ARTIFACTORY_USER} \
                     --env artifactory_password=${ARTIFACTORY_PASSWORD} \
                     --env ARTIFACTORY_PYPI_URL="artifactory.tooling.stackstate.io/artifactory/api/pypi/pypi-local/simple" \
                     --env PYTHON_RUNTIME=${PYTHON_RUNTIME}


build:
	cd Dockerfiles/local_builder && \
	docker build -t ${LOCAL_BUILD_IMAGE} \
		--build-arg UID=${UID} \
		--build-arg GID=${GID} \
		.

# Volume sharing can be used for agent application development
dev: build
	docker run -${DOCKER_RUN_MODE} --rm \
        --name ${LOCAL_BUILD_IMAGE} \
        --mount source=${VOLUME_GO_PKG_NAME},target=/go/pkg \
        --volume ${PWD}${colon}${PROJECT_DIR} \
        ${DOCKER_ENV} ${LOCAL_BUILD_IMAGE}

# Source copy can be used for Omnibus package build
omnibus: build
	docker run -${DOCKER_RUN_MODE} --rm \
        --user root \
        --name ${LOCAL_BUILD_IMAGE} \
        --mount source=${VOLUME_GO_PKG_NAME},target=/go/pkg \
        --volume ${PWD}${colon}${AGENT_SOURCE_MOUNT}${colon}ro \
        --env AGENT_SOURCE_MOUNT=${AGENT_SOURCE_MOUNT} \
        ${DOCKER_ENV} ${LOCAL_BUILD_IMAGE} ${COPY_MOUNT}

stop:
	docker stop $(shell docker ps -a -q --filter="name=${LOCAL_BUILD_IMAGE}")

shell:
	docker exec -ti ${LOCAL_BUILD_IMAGE} bash --init-file /local_init.sh

# When starting the first time you always need to pull deps
shell_install_deps:
	docker exec -ti ${LOCAL_BUILD_IMAGE} \
		bash -c "source ${PROJECT_DIR}/${LOCAL_BUILDER_INIT} install-deps;"

# Build and test the rtloader
shell_rtloader_build_test:
	docker exec -ti ${LOCAL_BUILD_IMAGE} \
		bash -c "source ${PROJECT_DIR}/${LOCAL_BUILDER_INIT} rtloader-build-test;"

# Build the agent binary
shell_agent_build:
	docker exec -ti ${LOCAL_BUILD_IMAGE} \
		bash -c "source ${PROJECT_DIR}/${LOCAL_BUILDER_INIT} agent-build;"

# Build the agent omnibus package
shell_agent_omnibus_build:
	docker exec -ti ${LOCAL_BUILD_IMAGE} \
		bash -c "source ${PROJECT_DIR}/${LOCAL_BUILDER_INIT} agent-omnibus-build;"

# Run the go tests with race conditions
# Specify a env variable PACKAGE_TARGET to only run tests against a certain package for example: PACKAGE_TARGET="./pkg/collector/python"
shell_go_tests_with_race:
	docker exec -ti ${LOCAL_BUILD_IMAGE} \
		bash -c "source ${PROJECT_DIR}/${LOCAL_BUILDER_INIT} go-tests-with-race ${PACKAGE_TARGET};"

# Switch to python 2
shell_set_python_2:
	docker exec -ti ${LOCAL_BUILD_IMAGE} \
		bash -c "source ${PROJECT_DIR}/${LOCAL_BUILDER_INIT} set-python-2;"

# Switch to python 3
shell_set_python_3:
	docker exec -ti ${LOCAL_BUILD_IMAGE} \
		bash -c "source ${PROJECT_DIR}/${LOCAL_BUILDER_INIT} set-python-3;"

.PHONY: build dev omnibus stop shell shell_install_deps shell_rtloader_build_test shell_agent_build shell_agent_omnibus_build shell_go_tests_with_race shell_set_python_2 shell_set_python_3
