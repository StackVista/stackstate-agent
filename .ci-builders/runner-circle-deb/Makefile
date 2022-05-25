RUNTIMETAG := $(shell date +%Y%m%d)
DOCKER_REPOSITORY ?= docker.io

build:
	docker build -t $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-gitlab:deb_$(RUNTIMETAG) .
push:
	docker push $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-gitlab:deb_$(RUNTIMETAG)

tag_latest:
	docker tag $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-gitlab:deb_$(RUNTIMETAG) $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-gitlab:deb_latest

push_latest:
	docker push $(DOCKER_REPOSITORY)/stackstate/stackstate-agent-runner-gitlab:deb_latest
