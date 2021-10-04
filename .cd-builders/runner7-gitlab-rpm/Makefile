RUNTIMETAG := $(shell date +%Y%m%d)
TRY := 0
build:
	docker build -t docker.io/stackstate/stackstate-agent-runner-gitlab:centos6_latest_v7 .
push:
	docker tag docker.io/stackstate/stackstate-agent-runner-gitlab:centos6_latest_v7 docker.io/stackstate/stackstate-agent-runner-gitlab:centos6_$(RUNTIMETAG)_v7
	docker tag docker.io/stackstate/stackstate-agent-runner-gitlab:centos6_latest_v7 docker.io/stackstate/stackstate-agent-runner-gitlab:centos6_$(RUNTIMETAG)_v7_$(TRY)
	docker push docker.io/stackstate/stackstate-agent-runner-gitlab:centos6_latest_v7
	docker push docker.io/stackstate/stackstate-agent-runner-gitlab:centos6_$(RUNTIMETAG)_v7
	docker push docker.io/stackstate/stackstate-agent-runner-gitlab:centos6_$(RUNTIMETAG)_v7_$(TRY)
