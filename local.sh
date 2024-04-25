#!/bin/bash

# docker run --rm -it -v ${PWD}:${PWD} -e MAJOR_VERSION="3" -e CI_PROJECT_DIR=${PWD} --workdir=${PWD} artifactory.tooling.stackstate.io/docker-virtual/stackstate/datadog_build_deb_x64:8292f573 bash
SRC_PATH="/go/src/github.com/StackVista/stackstate-agent"
WHAT=$1

if [ -z "${WHAT}" ]; then
	echo "Usage: $0 [all | prep | deps_deb | build_binaries | build_cluster_agent | build_deb]"
	exit 1
fi

WHAT=$(echo "${WHAT}" | tr '[:lower:]' '[:upper:]')

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "PREP" ]; then
	mkdir -p /go/src/github.com/StackVista
    rm -rf /go/src/github.com/StackVista/stackstate-agent || true
    . /usr/local/rvm/scripts/rvm
    ln -s "${CI_PROJECT_DIR}" /go/src/github.com/StackVista/stackstate-agent
fi

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "DEPS_DEB" ]; then
    # shellcheck disable=SC2164
    go clean -modcache

    echo "          ---                      ---"
    echo "          --- Getting dependencies ---"
    echo "          ---                      ---"
    inv -e deps --verbose
    go mod vendor
    go mod tidy
    inv agent.version --major-version 3 -u > version.txt
    echo "          ---                      ---"
    echo "          --- Agent Version String ---"
    echo "          ---                      ---"
    cat version.txt
fi

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "BUILD_BINARIES" ]; then
    echo "          ---                      ---"
    echo "          --- Building agent       ---"
    echo "          ---                      ---"
    echo " ******** --- Building dogstatsd   ---"
    inv -e dogstatsd.build --static --major-version 3
    echo " ******** --- Building rtloader    ---"
    inv -e rtloader.make
    echo " ******** --- Installing rtloader  ---"
    inv -e rtloader.install
    echo " ******** --- Building agent       ---"
    # shellcheck disable=SC2164
    cd $SRC_PATH
    inv -e agent.build --major-version "3" --python-runtimes "3"
    # shellcheck disable=SC2164
    cd "$CI_PROJECT_DIR"
fi

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "BUILD_CLUSTER_AGENT" ]; then
    echo "          ---                        ---"
    echo "          --- Building cluster agent ---"
    echo "          ---                        ---"
    inv -e cluster-agent.build
fi

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "BUILD_DEB" ]; then
    echo "          ---                      ---"
    echo "          --- Building deb package  ---"
    echo "          ---                      ---"
    mv "${CI_PROJECT_DIR}"/.omnibus /omnibus || mkdir -p /omnibus
    inv agent.version --major-version 3
    cat version.txt || true
    source ./.gitlab-scripts/setup_artifactory.sh
    export OMNIBUS_BASE_DIR="/.omnibus"
    inv -e agent.omnibus-build --gem-path $CI_PROJECT_DIR/.gems --base-dir $OMNIBUS_BASE_DIR --go-mod-cache $CI_PROJECT_DIR/vendor --skip-deps --skip-sign --major-version 3 --python-runtimes 3
fi
