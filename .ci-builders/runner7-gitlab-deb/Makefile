RUNTIMETAG := $(shell date +%Y%m%d)
DOCKER_REPOSITORY ?= docker.io
BASE_RUNNER_IMAGE_VERSION ?= deb7_latest

version:
	echo $(RUNTIMETAG) | tail -1 > version.txt

build:
	docker build --build-arg BASE_RUNNER_IMAGE_VERSION=$(BASE_RUNNER_IMAGE_VERSION) -t $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-gitlab:deb7_$(RUNTIMETAG) .
try:
	docker run -it $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-gitlab:deb7_$(RUNTIMETAG) bash
push:
	docker push $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-gitlab:deb7_$(RUNTIMETAG)
