#!/bin/sh

set -xe

IMAGE_TAG="${1}"
IMAGE_REPO="${2}"
DOCKERFILE_PATH="${3}"
EXTRA_TAG="${4}"
REGISTRY="quay.io"
ORGANIZATION="stackstate"
ARTIFACTORY_URL="artifactory.tooling.stackstate.io/artifactory/api/pypi/pypi-local/simple"

echo "IMAGE_TAG=${IMAGE_TAG}"
echo "IMAGE_REPO=${IMAGE_REPO}"
echo "DOCKERFILE_PATH=${DOCKERFILE_PATH}"

BUILD_TAG="${IMAGE_REPO}:${IMAGE_TAG}"

# shellcheck disable=SC2154
docker login -u "${quay_user}" -p "${quay_password}" "${REGISTRY}"
# shellcheck disable=SC2154
docker login -u "${artifactory_user}" -p "${artifactory_password}" "${ARTIFACTORY_URL}"

docker build -t "${BUILD_TAG}" "${DOCKERFILE_PATH}"


DOCKER_TAG="${REGISTRY}/${ORGANIZATION}/${IMAGE_REPO}:${IMAGE_TAG}"

docker tag "${BUILD_TAG}" "${DOCKER_TAG}"
docker push "${DOCKER_TAG}"

if [ -n "$EXTRA_TAG" ]; then
    DOCKER_EXTRA_TAG="${REGISTRY}/${ORGANIZATION}/${IMAGE_REPO}:${EXTRA_TAG}"
    docker tag "${DOCKER_TAG}" "${DOCKER_EXTRA_TAG}"
    echo "Pushing release to ${EXTRA_TAG}"
    docker push "${DOCKER_EXTRA_TAG}"

    # Comment out the if and fi lines to test anchore scanning on any branch.
    if [ ! -z "${CI_COMMIT_TAG}" ] || [ "${CI_COMMIT_BRANCH}" = "master" ]; then
        # for Anchore use publicly accessible image tag
        DOCKER_TAG="${REGISTRY}/${ORGANIZATION}/${IMAGE_REPO}:${EXTRA_TAG}"
        echo "Scanning image ${DOCKER_TAG} for vulnerabilities"
        omnibus/package-scripts/anchore_scan.sh -i "${DOCKER_TAG}" -n 0
    fi
fi
