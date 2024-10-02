RUNTIMETAG := $(shell date +%Y%m%d)
DOCKER_REPOSITORY ?= docker.io
BASE_IMAGE ?= deb7_latest

version:
	echo $(RUNTIMETAG) | tail -1 > version.txt

build:
	docker build --build-arg BASE_IMAGE=$(BASE_IMAGE) -t $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-gitlab:deb7_$(RUNTIMETAG) .
try:
	docker run -it $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-gitlab:deb7_$(RUNTIMETAG) bash
push:
	docker push $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-gitlab:deb7_$(RUNTIMETAG)
