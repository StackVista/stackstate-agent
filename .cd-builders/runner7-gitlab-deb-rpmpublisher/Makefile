RUNTIMETAG := $(shell date +%Y%m%d)
DOCKER_REPOSITORY ?= docker.io
BASE_RUNNER_IMAGE_VERSION ?= publisher_latest

build:
	docker build --build-arg BASE_RUNNER_IMAGE_VERSION=$(BASE_RUNNER_IMAGE_VERSION) -t $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-gitlab:deb-rpmpublisher_$(RUNTIMETAG) .
push:
	docker push $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-gitlab:deb-rpmpublisher_$(RUNTIMETAG)
