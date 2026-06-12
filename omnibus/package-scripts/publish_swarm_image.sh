#!/bin/sh

set -xe

IMAGE_TAG="${1}"
IMAGE_REPO="${2}"
EXTRA_TAG="${3}"
BASE_REPO="${4}"
BASE_TAG="${5}"
DOCKERFILE_PATH="${CI_PROJECT_DIR}/Dockerfiles/swarm-agent"
REGISTRY="quay.io"
REGISTRY_DOCKERHUB="docker.io"
ORGANIZATION="stackstate"

echo "BASE_TAG=${BASE_TAG}"
echo "BASE_REPO=${BASE_REPO}"
echo "IMAGE_TAG=${IMAGE_TAG}"
echo "IMAGE_REPO=${IMAGE_REPO}"
echo "DOCKERFILE_PATH=${DOCKERFILE_PATH}"

BUILD_TAG="${IMAGE_REPO}:${IMAGE_TAG}"

docker login -u "${docker_user}" -p "${docker_password}" "${REGISTRY_DOCKERHUB}"
docker login -u "${quay_user}" -p "${quay_password}" "${REGISTRY}"

docker build --build-arg BASE_REPO="${REGISTRY}/${ORGANIZATION}/${BASE_REPO}" --build-arg BASE_TAG=${BASE_TAG} -t "${BUILD_TAG}" "${DOCKERFILE_PATH}"

DOCKER_TAG="${REGISTRY}/${ORGANIZATION}/${IMAGE_REPO}:${IMAGE_TAG}"

docker tag "${BUILD_TAG}" "${DOCKER_TAG}"
docker push "${DOCKER_TAG}"

if [ -n "$EXTRA_TAG" ]; then
    DOCKER_EXTRA_TAG="${REGISTRY}/${ORGANIZATION}/${IMAGE_REPO}:${EXTRA_TAG}"
    docker tag "${DOCKER_TAG}" "${DOCKER_EXTRA_TAG}"
    echo "Pushing release to ${EXTRA_TAG}"
    docker push "${DOCKER_EXTRA_TAG}"
fi

# Comment out the if and fi lines to test anchore scanning on any branch.
if [ ! -z "${CI_COMMIT_TAG}" ] || [ "${CI_COMMIT_BRANCH}" = "master" ]; then
    # for Anchore use publicly accessible image tag
    DOCKER_TAG="${REGISTRY}/${ORGANIZATION}/${IMAGE_REPO}:${EXTRA_TAG}"
    echo "Scanning image ${DOCKER_TAG} for vulnerabilities"
    omnibus/package-scripts/anchore_scan.sh -i "${DOCKER_TAG}" -n 0
fi
