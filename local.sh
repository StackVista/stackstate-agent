#!/bin/bash

set -e

# We want to produce a final binary on a branded path, but it is convenient to run unit tests and the like on
# the original unbranded paths. Hence we allow for both.
if [[ "${RELOCATED}" != "true" ]]; then
    SRC_PATH="/go/src/github.com/DataDog/datadog-agent"
    export AGENT_GITHUB_ORG=DataDog
    export AGENT_REPO_NAME=datadog-agent
else
    SRC_PATH="/go/src/github.com/StackVista/stackstate-agent"
    export AGENT_GITHUB_ORG=StackVista
    export AGENT_REPO_NAME=stackstate-agent
fi

export PYTHON_RUNTIMES="3"
export BRANDED=${BRANDED:-true}
export OMNIBUS_FORCE_PACKAGES=true

WHAT=$1

if [ -z "${WHAT}" ]; then
	echo "Usage: $0 [shell | all | prep | deps_deb | build_binaries | build_cluster_agent | build_agent | build_deb | build_agent_image | build_cluster_agent_image | build_images | unit_tests | copy_rtloader | lint | cmd | up_to_cluster_agent | up_to_agent | up_to_build_deb | up_to_unit_tests]"
	exit 1
fi

WHAT=$(echo "${WHAT}" | tr '[:lower:]' '[:upper:]')

if [ "${WHAT}" = "SHELL" ]; then
    if [ ! -f "${PWD}/.bash_history" ]; then
        touch ${PWD}/.bash_history
    fi

    # Detect Docker socket and CLI for building images from inside the container
    DOCKER_MOUNTS=""
    if [ -S /var/run/docker.sock ]; then
        DOCKER_MOUNTS="-v /var/run/docker.sock:/var/run/docker.sock:z"
    fi
    DOCKER_CLI="$(which docker 2>/dev/null || true)"
    if [ -n "$DOCKER_CLI" ]; then
        DOCKER_MOUNTS="$DOCKER_MOUNTS -v ${DOCKER_CLI}:/usr/local/bin/docker:z"
    fi

    docker run --rm -it \
        --network host \
        -v ${PWD}:${PWD}:z \
        -v ${PWD}/.bash_history:/root/.bash_history:z \
        $DOCKER_MOUNTS \
        -e MAJOR_VERSION="3" \
        -e USER_ID="$(id -u)" -e GROUP_ID="$(id -g)" \
        -e CI_PROJECT_DIR=${PWD} \
        -e GITLAB_PACKAGE_REGISTRY_PYPI_SIMPLE_URL="${GITLAB_PACKAGE_REGISTRY_PYPI_SIMPLE_URL}" \
        -e GITLAB_PACKAGE_REGISTRY_USER="${GITLAB_PACKAGE_REGISTRY_USER}" \
        -e GITLAB_PACKAGE_REGISTRY_TOKEN="${GITLAB_PACKAGE_REGISTRY_TOKEN}" \
        -e BRANDED=${BRANDED} \
        -e REGISTRY="${REGISTRY}" \
        -e ORG="${ORG:-stackstate}" \
        --workdir=${PWD} \
        registry.tooling.stackstate.io/quay/stackstate/datadog_build_linux_x64:4ed2400d bash
fi

# Prepare a copy of the agent in the SRC_DIR to make sure that in a containerized environment the source directory
# does not get tainted, and all files have the proper user for within the container.
function prepare() {
    . /usr/local/rvm/scripts/rvm
    # Ensure gofmt is available for fix_branding.sh
    export PATH="/usr/local/go/bin:$PATH"

    if ! type "rsync" > /dev/null; then
      apt install rsync -y --no-install-recommends
    fi

    mkdir -p $SRC_PATH
    echo "Syncing files to $SRC_PATH"
    rsync -au "$CI_PROJECT_DIR"/. $SRC_PATH
    chown -R root:root $SRC_PATH
    rm -rf "$SRC_PATH/rtloader/build" || true
    cd "$SRC_PATH" || exit

    if [[ "${RELOCATED}" = "true" ]]; then
        git config --global --add safe.directory "$SRC_PATH"

        echo "Fixing import paths"
        ./fix_package_paths.sh "$SRC_PATH"

        echo "Cleaning module cache to remove stale entries"
        go clean -modcache

        echo "Removing go.sum and vendor to force clean regeneration"
        rm -f go.sum
        rm -rf vendor

        echo "Syncing workspace after path transformation"
        go work sync

        echo "Vendoring workspace modules"
        go work vendor
    fi
    if [[ "${BRANDED}" = "true" ]]; then
        echo "Applying Branding to $SRC_PATH..."
        ./fix_branding.sh "$SRC_PATH"
    fi
    cd "$CI_PROJECT_DIR" || exit
}

if [ "${WHAT}" = "PREP" ]; then
    prepare
fi

if [ "${WHAT}" = "ALL" ]; then
    prepare
fi

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "DEPS_DEB" ] || [ "${WHAT}" = "UP_TO_CLUSTER_AGENT" ] || [ "${WHAT}" = "UP_TO_AGENT" ] || [ "${WHAT}" = "UP_TO_BUILD_DEB" ] || [ "${WHAT}" = "UP_TO_UNIT_TESTS" ]; then
    if [ "${WHAT}" = "DEPS_DEB" ] || [ "${WHAT}" = "UP_TO_CLUSTER_AGENT" ] || [ "${WHAT}" = "UP_TO_AGENT" ] || [ "${WHAT}" = "UP_TO_BUILD_DEB" ] || [ "${WHAT}" = "UP_TO_UNIT_TESTS" ]; then
        prepare
    fi

    cd $SRC_PATH || exit

    echo "Running debian dependencies in $SRC_PATH"

    # shellcheck disable=SC2164
    go clean -modcache

    echo "          ---                      ---"
    echo "          --- Getting dependencies ---"
    echo "          ---                      ---"
    inv -e deps --verbose
    go mod tidy

    if [[ "${RELOCATED}" != "true" ]]; then
        echo "          ---                      ---"
        echo "          --- Regenerating vendor  ---"
        echo "          ---                      ---"
        GOWORK=off go mod vendor
        go work sync
        go work vendor
    fi

    inv agent.version -u > version.txt
    echo "          ---                      ---"
    echo "          --- Agent Version String ---"
    echo "          ---                      ---"
    cat version.txt

    cd "$CI_PROJECT_DIR" || exit
fi

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "BUILD_CLUSTER_AGENT" ] || [ "${WHAT}" = "UP_TO_CLUSTER_AGENT" ] || [ "${WHAT}" = "UP_TO_AGENT" ] || [ "${WHAT}" = "UP_TO_BUILD_DEB" ] || [ "${WHAT}" = "UP_TO_UNIT_TESTS" ]; then
    if [ "${WHAT}" = "BUILD_CLUSTER_AGENT" ]; then
        prepare
    fi

    cd $SRC_PATH || exit

    echo "          ---                        ---"
    echo "          --- Building cluster agent ---"
    echo "          ---                        ---"
    inv -e cluster-agent.build

    cd "$CI_PROJECT_DIR" || exit
fi

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "BUILD_AGENT" ] || [ "${WHAT}" = "UP_TO_AGENT" ] || [ "${WHAT}" = "UP_TO_BUILD_DEB" ] || [ "${WHAT}" = "UP_TO_UNIT_TESTS" ]; then
    if [ "${WHAT}" = "BUILD_AGENT" ] ; then
        prepare
    fi

    cd $SRC_PATH || exit

    echo "          ---                      ---"
    echo "          --- Building agent       ---"
    echo "          ---                      ---"
    echo " ******** --- Building dogstatsd   ---"
    # inv -e dogstatsd.build --static --major-version 3
    echo " ******** --- Building rtloader    ---"
    inv -e rtloader.make
    echo " ******** --- Installing rtloader  ---"
    inv -e rtloader.install
    echo " ******** --- Building agent       ---"
    # shellcheck disable=SC2164
    inv -e agent.build

    cd "$CI_PROJECT_DIR" || exit
fi

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "BUILD_DEB" ] || [ "${WHAT}" = "UP_TO_BUILD_DEB" ] || [ "${WHAT}" = "UP_TO_UNIT_TESTS" ]; then
    if [ "${WHAT}" = "BUILD_DEB" ]; then
        prepare
    fi

    cd $SRC_PATH || exit

    echo "          ---                      ---"
    echo "          --- Building deb package  ---"
    echo "          ---                      ---"
    mv "$SRC_PATH"/.omnibus /omnibus || mkdir -p /omnibus
    inv agent.version
    cat version.txt || true
    source setup_artifact_registry.sh
    export OMNIBUS_BASE_DIR="/.omnibus"
    inv -e omnibus.build --gem-path $SRC_PATH/.gems --base-dir $OMNIBUS_BASE_DIR --go-mod-cache $SRC_PATH/vendor --skip-deps --skip-sign

        # Prepare outputs
    mkdir -p $SRC_PATH/outcomes/pkg && mkdir -p $SRC_PATH/outcomes/dockerfiles && mkdir -p $SRC_PATH/outcomes/binary
    cp -r $OMNIBUS_BASE_DIR/pkg $SRC_PATH/outcomes
    cp -r $SRC_PATH/Dockerfiles $SRC_PATH/outcomes
#    - cp -r /opt/stackstate-agent/embedded/bin/trace-agent  $SRC_PATH/outcomes/binary/

    ls -la $SRC_PATH/outcomes/Dockerfiles

        # Prepare cache
        # Drop packages for cache
    rm -rf /omnibus/pkg
        # Drop agent for cache (will be resynced anyway)
    rm -rf /omnibus/src/datadog-agent
        # Drop symlink because it will fail the build when coming from a cache
    rm /omnibus/src/datadog-agent/src/github.com/StackVista/stackstate-agent/vendor/github.com/coreos/etcd/cmd/etcd || echo "Not found"
    mv /omnibus $SRC_PATH/.omnibus

    cd "$CI_PROJECT_DIR" || exit
fi

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "UNIT_TESTS" ] || [ "${WHAT}" = "UP_TO_UNIT_TESTS" ]; then
    if [ "${WHAT}" = "UNIT_TESTS" ]; then
        prepare
    fi

    cd $SRC_PATH || exit

    echo "          ---                      ---"
    echo "          --- Running Unit Tests   ---"
    echo "          ---                      ---"

    inv -e agent.build --race
    # TODO: check why formatting rules differ from previous step
    # - gofmt -l -w -s ./pkg ./cmd
    inv -e rtloader.test
    invoke install-tools
    echo "inv -e test --coverage --race --profile --cpus 4"
    inv -e test --coverage --race --profile --cpus 4

    cd "$CI_PROJECT_DIR" || exit
fi

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "COPY_RTLOADER" ]; then
    if [ "${WHAT}" = "COPY_RTLOADER" ]; then
        prepare
    fi

    rm -rf "$CI_PROJECT_DIR/rtloader/build" || true
    rm -rf "$SRC_PATH/rtloader/build" || true

    cd $SRC_PATH || exit

    echo "          ---                              ---"
    echo "          --- Copying Rtloader to source   ---"
    echo "          ---                              ---"

    inv -e rtloader.make

    cp -a "$SRC_PATH"/rtloader/build "$CI_PROJECT_DIR"/rtloader/
    chown "$USER_ID":"$GROUP_ID" -R "$CI_PROJECT_DIR"/rtloader/build

    cd "$CI_PROJECT_DIR" || exit
fi

if [ "${WHAT}" = "ALL" ] || [ "${WHAT}" = "LINT" ]; then
    if [ "${WHAT}" = "LINT" ]; then
        prepare
    fi

    cd $SRC_PATH || exit

    echo "          ---                      ---"
    echo "          --- Running Lint         ---"
    echo "          ---                      ---"

    inv -e lint-go

    cd "$CI_PROJECT_DIR" || exit
fi

## Docker image build steps
## These require Docker socket + CLI to be mounted (done automatically by the SHELL command).
## Run after ./local.sh all (or the individual build steps) has produced artifacts in outcomes/.

ARCH="${ARCH:-amd64}"
IMAGE_TAG="${IMAGE_TAG:-$(git rev-parse --short HEAD 2>/dev/null || echo 'local')}"
ORG="${ORG:-stackstate}"
AGENT_IMAGE_REPO="stackstate-k8s-agent"
CLUSTER_IMAGE_REPO="stackstate-k8s-cluster-agent"

function check_docker() {
    if ! command -v docker &>/dev/null; then
        echo "ERROR: docker CLI not available inside this container."
        echo "Run './local.sh shell' instead of './local.sh cmd' to get Docker socket + CLI mounted."
        exit 1
    fi
    if ! docker info &>/dev/null; then
        echo "ERROR: Cannot connect to Docker daemon. Is the socket mounted?"
        echo "Check that /var/run/docker.sock exists and is accessible."
        exit 1
    fi
}

function build_agent_image() {
    check_docker

    local dockerfile_dir="$CI_PROJECT_DIR/outcomes/Dockerfiles/agent"
    if [ ! -d "$dockerfile_dir" ]; then
        echo "ERROR: $dockerfile_dir not found. Run build_deb first."
        exit 1
    fi

    local deb_file
    deb_file=$(ls "$CI_PROJECT_DIR"/outcomes/pkg/stackstate-agent_*_${ARCH}.deb 2>/dev/null | head -1)
    if [ -z "$deb_file" ]; then
        echo "ERROR: No .deb found at outcomes/pkg/stackstate-agent_*_${ARCH}.deb"
        exit 1
    fi

    echo "Copying $deb_file -> $dockerfile_dir/"
    cp "$deb_file" "$dockerfile_dir/"

    local local_tag="${AGENT_IMAGE_REPO}:${IMAGE_TAG}"
    echo "Building agent image: $local_tag"
    docker build --build-arg ARCH="${ARCH}" -t "$local_tag" "$dockerfile_dir"

    if [ -n "$REGISTRY" ]; then
        local remote_tag="${REGISTRY}/${ORG}/${local_tag}"
        docker tag "$local_tag" "$remote_tag"
        echo "Tagged: $remote_tag"
        echo "Push with: docker push $remote_tag"
    fi
}

function build_cluster_agent_image() {
    check_docker

    local dockerfile_dir="$CI_PROJECT_DIR/outcomes/Dockerfiles/cluster-agent"
    if [ ! -d "$dockerfile_dir" ]; then
        echo "ERROR: $dockerfile_dir not found. Run build_cluster_agent first."
        exit 1
    fi

    # CI copies from bin/, local.sh build also puts it there
    local binary="$CI_PROJECT_DIR/bin/stackstate-cluster-agent"
    if [ ! -f "$binary" ]; then
        binary="$SRC_PATH/bin/stackstate-cluster-agent"
    fi
    if [ ! -f "$binary" ]; then
        echo "ERROR: stackstate-cluster-agent binary not found in bin/"
        exit 1
    fi

    echo "Copying $binary -> $dockerfile_dir/"
    cp "$binary" "$dockerfile_dir/"

    local local_tag="${CLUSTER_IMAGE_REPO}:${IMAGE_TAG}"
    echo "Building cluster-agent image: $local_tag"
    docker build -t "$local_tag" "$dockerfile_dir"

    if [ -n "$REGISTRY" ]; then
        local remote_tag="${REGISTRY}/${ORG}/${local_tag}"
        docker tag "$local_tag" "$remote_tag"
        echo "Tagged: $remote_tag"
        echo "Push with: docker push $remote_tag"
    fi
}

if [ "${WHAT}" = "BUILD_AGENT_IMAGE" ]; then
    build_agent_image
fi

if [ "${WHAT}" = "BUILD_CLUSTER_AGENT_IMAGE" ]; then
    build_cluster_agent_image
fi

if [ "${WHAT}" = "BUILD_IMAGES" ]; then
    build_agent_image
    build_cluster_agent_image
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
