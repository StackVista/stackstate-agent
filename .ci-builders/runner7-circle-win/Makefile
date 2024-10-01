RUNTIMETAG := $(shell date +%Y%m%d)

build:
	docker build -t docker.io/stackstate/stackstate-agent-runner-gitlab:win_$(RUNTIMETAG) .
push:
	docker push docker.io/stackstate/stackstate-agent-runner-gitlab:win_$(RUNTIMETAG)

tag_latest:
	docker tag docker.io/stackstate/stackstate-agent-runner-gitlab:win_$(RUNTIMETAG) docker.io/stackstate/stackstate-agent-runner-gitlab:win_latest

push_latest:
	docker push docker.io/stackstate/stackstate-agent-runner-gitlab:win_latest

