RUNTIMETAG := $(shell date +%Y%m%d)
DOCKER_REPOSITORY ?= docker.io
BASE_IMAGE ?= deb7_latest

build:
	docker build --build-arg BASE_IMAGE=$(BASE_IMAGE) -t $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-gitlab:deb-rpmpublisher_$(RUNTIMETAG) .
push:
	docker push $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-gitlab:deb-rpmpublisher_$(RUNTIMETAG)
