.SHELLFLAGS = -ec

SHELL         := /bin/sh
.DEFAULT_GOAL := push

RUNTIMETAG         := $(shell date +%Y%m%d)
DOCKER_REGISTRY    ?= docker.io
BASE_KEEPER_IMAGE   = ${DOCKER_REGISTRY}/stackstate/beest:${RUNTIMETAG}

ARTIFACTORY_PYPI_URL         ?= artifactory.tooling.stackstate.io/artifactory/api/pypi/pypi-local/simple
ARTIFACTORY_USER             ?= NA
ARTIFACTORY_PASSWORD         ?= NA
DOCKER_BUILD_ARG              = --build-arg ARTIFACTORY_PYPI_URL=${ARTIFACTORY_PYPI_URL} \
                                --build-arg ARTIFACTORY_USER=${ARTIFACTORY_USER} \
                                --build-arg ARTIFACTORY_PASSWORD=${ARTIFACTORY_PASSWORD}

GIT_BRANCH                    = $(shell git rev-parse --abbrev-ref HEAD)
MAJOR_VERSION                ?= 2
STS_AWS_TEST_BUCKET          ?= stackstate-agent-3-test
STS_DOCKER_TEST_REPO         ?= stackstate-agent-test
PROCESS_AGENT_TEST_REPO      ?= stackstate-process-agent-test
STS_DOCKER_TEST_REPO_CLUSTER ?= stackstate-cluster-agent-test
DOCKER_ENV		              = --env AGENT_CURRENT_BRANCH=${GIT_BRANCH} \
			    	            --env MAJOR_VERSION=${MAJOR_VERSION} \
			    	            --env STS_AWS_TEST_BUCKET=${STS_AWS_TEST_BUCKET} \
			    	            --env STS_DOCKER_TEST_REPO=${STS_DOCKER_TEST_REPO} \
			    	            --env STS_DOCKER_TEST_REPO_CLUSTER=${STS_DOCKER_TEST_REPO_CLUSTER} \
                                --env PROCESS_AGENT_TEST_REPO=${PROCESS_AGENT_TEST_REPO}

build:
	docker build ${DOCKER_BUILD_ARG} -t ${BASE_KEEPER_IMAGE} .

push: build
	docker push ${BASE_KEEPER_IMAGE}

kaniko:
	/kaniko/executor ${DOCKER_BUILD_ARG} --destination ${BASE_KEEPER_IMAGE}

try:
	docker run -it --rm \
		--volume ${PWD}/../..:/root/beest \
		--env CI_PROJECT_DIR=/root \
		${DOCKER_ENV} \
		${BASE_KEEPER_IMAGE}


.PHONY: build push try
