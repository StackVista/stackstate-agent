.SHELLFLAGS = -ec

SHELL         := /bin/bash
.DEFAULT_GOAL := push

GIT_BRANCH = $(shell git rev-parse --abbrev-ref HEAD)

STACKSTATE_BRANCH            ?= master
MAJOR_VERSION                ?= 2
STS_AWS_TEST_BUCKET          ?= stackstate-agent-2-test
STS_DOCKER_TEST_REPO         ?= stackstate-agent-2-test
STS_DOCKER_TEST_REPO_CLUSTER ?= stackstate-cluster-agent-test

RUNTIMETAG         := $(shell date +%Y%m%d)
DOCKER_REPOSITORY  ?= docker.io/stackstate/beest
BASE_KEEPER_IMAGE   = ${DOCKER_REPOSITORY}:${RUNTIMETAG}
DOCKER_ENV		    = --env AGENT_CURRENT_BRANCH=${GIT_BRANCH} \
			    	  --env STACKSTATE_BRANCH=${STACKSTATE_BRANCH} \
			    	  --env MAJOR_VERSION=${MAJOR_VERSION} \
			    	  --env STS_AWS_TEST_BUCKET=${STS_AWS_TEST_BUCKET} \
			    	  --env STS_DOCKER_TEST_REPO=${STS_DOCKER_TEST_REPO} \
			    	  --env STS_DOCKER_TEST_REPO_CLUSTER=${STS_DOCKER_TEST_REPO_CLUSTER}

build:
	docker build -t ${BASE_KEEPER_IMAGE} .

push: build
	docker push ${BASE_KEEPER_IMAGE}


try:
	docker run -it --rm \
		--volume ${PWD}/../..:/root/beest \
		--env CI_PROJECT_DIR=/root \
		${DOCKER_ENV} \
		${BASE_KEEPER_IMAGE}
