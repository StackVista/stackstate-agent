#!/bin/bash

# docker run --rm -it -v ${PWD}:${PWD} -e MAJOR_VERSION="3" -e CI_PROJECT_DIR=${PWD} --workdir=${PWD} artifactory.tooling.stackstate.io/docker-virtual/stackstate/datadog_build_deb_x64:8292f573 bash

WHAT=$1

if [ -z "${WHAT}" ]; then
	echo "Usage: $0 [all | prep | deps | build]"
	exit 1
fi

WHAT=$(echo "${WHAT}" | tr '[:lower:]' '[:upper:]')

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "PREP" ]; then
	mkdir -p /go/src/github.com/StackVista
    rm -rf /go/src/github.com/StackVista/stackstate-agent || true
    . /usr/local/rvm/scripts/rvm
    ln -s "${CI_PROJECT_DIR}" /go/src/github.com/StackVista/stackstate-agent
fi

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "DEPS" ]; then
    # shellcheck disable=SC2164
    go clean -modcache

    echo "---                      ---"
    echo "--- Getting dependencies ---"
    echo "---                      ---"
    inv -e deps --verbose
    inv agent.version --major-version 3 -u > version.txt
    echo "---                      ---"
    echo "--- Agent Version String ---"
    echo "---                      ---"
    cat version.txt
fi

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "BUILD" ]; then
    echo "---                      ---"
    echo "--- Building agent       ---"
    echo "---                      ---"
    inv -e dogstatsd.build --static --major-version 3
fi
