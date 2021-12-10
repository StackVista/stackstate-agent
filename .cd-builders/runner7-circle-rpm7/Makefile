RUNTIMETAG := $(shell date +%Y%m%d)
DOCKER_REPOSITORY ?= docker.io

version:
	echo $(RUNTIMETAG) | tail -1 > version.txt

build:
	docker build -t $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-circle:centos7_$(RUNTIMETAG) .

push:
	docker push $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-circle:centos7_$(RUNTIMETAG)

tag_latest:
	docker tag $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-circle:centos7_$(RUNTIMETAG) $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-circle:centos7_latest

push_latest:
	docker push $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-circle:centos7_latest

