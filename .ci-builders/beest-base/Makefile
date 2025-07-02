.SHELLFLAGS = -ec

SHELL         := /bin/sh
.DEFAULT_GOAL := push

RUNTIMETAG         := $(shell date +%Y%m%d)
DOCKER_REGISTRY    ?= docker.io

# For local development, you have to change this to something you have access to push to. For Louis Parkin, it is wolverminion
BASE_KEEPER_IMAGE_LOCAL ?= wolverminion/beest:local
BASE_KEEPER_IMAGE   = ${DOCKER_REGISTRY}/stackstate/beest:${RUNTIMETAG}

GITLAB_PACKAGE_REGISTRY_PYPI_URL         ?= gitlab.com/api/v4/projects/71271774/packages/pypi/simple
GITLAB_PACKAGE_REGISTRY_USER             ?= NA
GITLAB_PACKAGE_REGISTRY_TOKEN         ?= NA
DOCKER_BUILD_ARG              = --build-arg GITLAB_PACKAGE_REGISTRY_PYPI_URL=${GITLAB_PACKAGE_REGISTRY_PYPI_URL} \
                                --build-arg GITLAB_PACKAGE_REGISTRY_USER=${GITLAB_PACKAGE_REGISTRY_USER} \
                                --build-arg GITLAB_PACKAGE_REGISTRY_TOKEN=${GITLAB_PACKAGE_REGISTRY_TOKEN}

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

build-local:
	docker build ${DOCKER_BUILD_ARG} -t ${BASE_KEEPER_IMAGE_LOCAL} -f ./local.dockerfile .

push: build
	docker push ${BASE_KEEPER_IMAGE}

push-local: build-local
	docker push ${BASE_KEEPER_IMAGE_LOCAL}

kaniko:
	/kaniko/executor ${DOCKER_BUILD_ARG} --destination ${BASE_KEEPER_IMAGE}

try:
	docker run -it --rm \
		--volume ${PWD}/../..:/root/beest \
		--env CI_PROJECT_DIR=/root \
		${DOCKER_ENV} \
		${BASE_KEEPER_IMAGE}


.PHONY: build push try
