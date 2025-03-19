#!/bin/sh

set -xe

IMAGE_TAG="${1}-${ARCH}"
IMAGE_REPO="${2}"
DOCKERFILE_PATH="${3}"
EXTRA_TAG="${4}-${ARCH}"
K8S_REPO="${5}"
REGISTRY="quay.io"
ORGANIZATION="stackstate"
ARTIFACTORY_URL="artifactory.tooling.stackstate.io/artifactory/api/pypi/pypi-local/simple"
S6_ARCH="${6}"

echo "IMAGE_TAG=${IMAGE_TAG}"
echo "IMAGE_REPO=${IMAGE_REPO}"
echo "DOCKERFILE_PATH=${DOCKERFILE_PATH}"

BUILD_TAG="${IMAGE_REPO}:${IMAGE_TAG}"

# shellcheck disable=SC2154
docker login -u "${quay_user}" -p "${quay_password}" "${REGISTRY}"
docker login -u "${artifactory_user}" -p "${artifactory_password}" "${ARTIFACTORY_URL}"

docker build --build-arg ARCH="${ARCH}" --build-arg S6_ARCH="${S6_ARCH}" -t "${BUILD_TAG}" "${DOCKERFILE_PATH}"


DOCKER_TAG="${REGISTRY}/${ORGANIZATION}/${IMAGE_REPO}:${IMAGE_TAG}"

docker tag "${BUILD_TAG}" "${DOCKER_TAG}"
docker push "${DOCKER_TAG}"

if [ -n "$EXTRA_TAG" ]; then
    DOCKER_EXTRA_TAG="${REGISTRY}/${ORGANIZATION}/${IMAGE_REPO}:${EXTRA_TAG}"
    docker tag "${DOCKER_TAG}" "${DOCKER_EXTRA_TAG}"
    echo "Pushing release to ${EXTRA_TAG}"
    docker push "${DOCKER_EXTRA_TAG}"

    # If K8S_REPO is not equal to "NOP" and is set then push the image to the k8s repo
    if [ -n "${K8S_REPO}" ] && [ "${K8S_REPO}" != "NOP" ]; then
        DOCKER_K8S_TAG="${REGISTRY}/${ORGANIZATION}/${K8S_REPO}:${EXTRA_TAG}"
        docker tag "${DOCKER_TAG}" "${DOCKER_K8S_TAG}"
        echo "Pushing release to ${K8S_REPO}"
        docker push "${DOCKER_K8S_TAG}"
    fi

fi
