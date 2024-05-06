#!/bin/bash

set -e

# We want to produce a final binary on a branded path, but it is convenient to run unit tests and the like on
# the original unbranded paths. Hence we allow for both.
if [[ "${BRANDED}" = "false" ]]; then
    SRC_PATH="/go/src/github.com/DataDog/datadog-agent"
    export AGENT_GITHUB_ORG=DataDog
    export AGENT_REPO_NAME=datadog-agent
else
    SRC_PATH="/go/src/github.com/StackVista/stackstate-agent"
    export AGENT_GITHUB_ORG=StackVista
    export AGENT_REPO_NAME=stackstate-agent
fi

WHAT=$1

if [ -z "${WHAT}" ]; then
	echo "Usage: $0 [shell | all | prep | deps_deb | build_binaries | build_cluster_agent | build_deb]"
	exit 1
fi

WHAT=$(echo "${WHAT}" | tr '[:lower:]' '[:upper:]')

if [ "${WHAT}" = "SHELL" ]; then
    docker run --rm -it -v ${PWD}:${PWD} -e MAJOR_VERSION="3" -e CI_PROJECT_DIR=${PWD} --workdir=${PWD} artifactory.tooling.stackstate.io/docker-virtual/stackstate/datadog_build_deb_x64:8292f573 bash
fi

# Prepare a copy of the agent in the SRC_DIR to make sure that in a containerized environment the source directory
# does not get tainted, and all files have the proper user for within the container.
function prepare() {
    . /usr/local/rvm/scripts/rvm

    if ! type "rsync" > /dev/null; then
      apt install rsync -y --no-install-recommends
    fi

    mkdir -p $SRC_PATH
    echo "Syncing files to $SRC_PATH"
    rsync -au "$CI_PROJECT_DIR"/. $SRC_PATH
    chown -R root:root $SRC_PATH
    cd "$SRC_PATH" || exit

    if [[ "${BRANDED}" != "false" ]]; then
        echo "Fixing import paths"
        ./fix_package_paths.sh "$SRC_PATH"

        echo "Running tidy after rewriting paths"
        # TODO: Ideally we'd not need to run this, but because we update the package paths, we need to update go.mod/sum/vendor
        # Alternative is to commit package renames, but that is also pretty messy
        go mod tidy
    fi
    cd "$CI_PROJECT_DIR" || exit
}

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "DEPS_DEB" ]; then
    prepare

    cd $SRC_PATH || exit

    echo "Running debian dependencies in $SRC_PATH"

    # shellcheck disable=SC2164
    go clean -modcache

    echo "          ---                      ---"
    echo "          --- Getting dependencies ---"
    echo "          ---                      ---"
    inv -e deps --verbose
    go mod tidy
    inv agent.version --major-version 3 -u > version.txt
    echo "          ---                      ---"
    echo "          --- Agent Version String ---"
    echo "          ---                      ---"
    cat version.txt

    cd "$CI_PROJECT_DIR" || exit
fi

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "BUILD_BINARIES" ]; then
    prepare

    cd $SRC_PATH || exit

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
    inv -e agent.build --major-version "3" --python-runtimes "3"

    cd "$CI_PROJECT_DIR" || exit
fi

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "BUILD_CLUSTER_AGENT" ]; then
    prepare

    cd $SRC_PATH || exit

    echo "          ---                        ---"
    echo "          --- Building cluster agent ---"
    echo "          ---                        ---"
    inv -e cluster-agent.build

    cd "$CI_PROJECT_DIR" || exit
fi

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "BUILD_DEB" ]; then
    prepare

    cd $SRC_PATH || exit

    echo "          ---                      ---"
    echo "          --- Building deb package  ---"
    echo "          ---                      ---"
    mv "$SRC_PATH"/.omnibus /omnibus || mkdir -p /omnibus
    inv agent.version --major-version 3
    cat version.txt || true
    source ./.gitlab-scripts/setup_artifactory.sh
    export OMNIBUS_BASE_DIR="/.omnibus"
    inv -e agent.omnibus-build --gem-path $SRC_PATH/.gems --base-dir $OMNIBUS_BASE_DIR --go-mod-cache $SRC_PATH/vendor --skip-deps --skip-sign --major-version 3 --python-runtimes 3

    cd "$CI_PROJECT_DIR" || exit
fi

if [ "${WHAT}" = "CMD" ]; then
    prepare

    cd $SRC_PATH || exit

    echo "          ---                         ---"
    echo "          --- Running command `$2`"
    echo "          ---                         ---"

    $2

    cd "$CI_PROJECT_DIR" || exit
fi
