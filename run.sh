#!/bin/bash

docker run --rm -it \
	-e SRC_PATH:"/go/src/github.com/StackVista/stackstate-agent" \
	-e OMNIBUS_BASE_DIR:"/.omnibus" \
	-e OMNIBUS_BASE_DIR_WIN:"c:/omnibus-ruby" \
	-e BCC_VERSION:"v0.12.0" \
	-e SYSTEM_PROBE_GO_VERSION:"1.13.11" \
	-e DATADOG_AGENT_EMBEDDED_PATH:"/opt/datadog-agent/embedded" \
	-e ARCH:"amd64" \
	-e VCINSTALLDIR:"C:\\Program Files (x86)\\Microsoft Visual Studio\\2017\\Community" \
	-e MOLECULE_K8S_CLUSTER:"eks_test_cluster_1_21" \
	-e PRIMARY_MAJOR_VERSION:'2' \
	-e HELM_CHART_VERSION:'latest' \
	-e PROCESS_AGENT_TEST_REPO:"stackstate-process-agent-test" \
	-e CONDA_ENV:"ddpy2" \
  -e PYTHON_RUNTIMES:"2" \
  -e MAJOR_VERSION:"2" \
  -e STS_VER:"v2" \
  -e STS_AWS_RELEASE_BUCKET:"stackstate-agent-2" \
  -e STS_AWS_TEST_BUCKET:"stackstate-agent-2-test" \
  -e STS_AWS_RELEASE_BUCKET_YUM:"stackstate-agent-2-rpm" \
  -e STS_AWS_TEST_BUCKET_YUM:"stackstate-agent-2-rpm-test" \
  -e STS_AWS_RELEASE_BUCKET_WIN:"stackstate-agent-2" \
  -e STS_AWS_TEST_BUCKET_WIN:"stackstate-agent-2-test" \
  -e STS_DOCKER_RELEASE_REPO:"stackstate-agent-2" \
  -e STS_DOCKER_TEST_REPO:"stackstate-agent-2-test" \
  -e STS_DOCKER_RELEASE_REPO_TRACE:"stackstate-trace-agent" \
  -e STS_DOCKER_TEST_REPO_TRACE:"stackstate-trace-agent-test" \
  -e STS_DOCKER_RELEASE_REPO_CLUSTER:"stackstate-cluster-agent" \
  -e STS_DOCKER_TEST_REPO_CLUSTER:"stackstate-cluster-agent-test" \
	-e CI_PROJECT_DIR:"/go/src/github.com/StackVista/stackstate-agent" \
	-v ${PWD}:"/go/src/github.com/StackVista/stackstate-agent" \
	-v ~/.ansible:/root/.ansible \
	-v ~/.aws:/root/.aws \
	-e AWS_PROFILE=$AWS_PROFILE \
	wolverminion/packer-runner:0.0.6 bash

	# artifactory.tooling.stackstate.io/docker-virtual/stackstate/stackstate-agent-runner-gitlab:debian-20220505
