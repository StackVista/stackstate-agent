RUNTIMETAG := $(shell date +%Y%m%d)
DOCKER_REPOSITORY ?= docker.io

version:
	echo $(RUNTIMETAG) | tail -1 > version.txt

build:
	docker build -t $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-circle:deb7_$(RUNTIMETAG) .
push:
	docker push $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-circle:deb7_$(RUNTIMETAG)

try:
	docker run -it $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-circle:deb7_$(RUNTIMETAG) sh

tag_latest:
	docker tag $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-circle:deb7_$(RUNTIMETAG) docker.io/stackstate/stackstate-agent-runner-circle:deb7_latest

push_latest:
	docker push $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-circle:deb7_latest
