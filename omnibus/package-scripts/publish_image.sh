#!/bin/sh

set -xe

IMAGE_TAG="${1}-${ARCH}"
IMAGE_REPO="${2}"
DOCKERFILE_PATH="${3}"
EXTRA_TAG="${4}-${ARCH}"
REGISTRY="quay.io"
ORGANIZATION="stackstate"
GITLAB_PACKAGE_REGISTRY_PYPI_SIMPLE_URL="https://gitlab.com/api/v4/projects/71271774/packages/pypi/simple"
S6_ARCH="${5}"

echo "IMAGE_TAG=${IMAGE_TAG}"
echo "IMAGE_REPO=${IMAGE_REPO}"
echo "DOCKERFILE_PATH=${DOCKERFILE_PATH}"

BUILD_TAG="${IMAGE_REPO}:${IMAGE_TAG}"

# shellcheck disable=SC2154
docker login -u "${quay_user}" -p "${quay_password}" "${REGISTRY}"
docker login -u "${REGISTRY_USER}" -p "${REGISTRY_PASSWORD}" "${REGISTRY_HOST}"

docker build --build-arg ARCH="${ARCH}" --build-arg S6_ARCH="${S6_ARCH}" -t "${BUILD_TAG}" "${DOCKERFILE_PATH}"


DOCKER_TAG="${REGISTRY}/${ORGANIZATION}/${IMAGE_REPO}:${IMAGE_TAG}"

docker tag "${BUILD_TAG}" "${DOCKER_TAG}"
docker push "${DOCKER_TAG}"

if [ -n "$EXTRA_TAG" ]; then
    DOCKER_EXTRA_TAG="${REGISTRY}/${ORGANIZATION}/${IMAGE_REPO}:${EXTRA_TAG}"
    docker tag "${DOCKER_TAG}" "${DOCKER_EXTRA_TAG}"
    echo "Pushing release to ${EXTRA_TAG}"
    docker push "${DOCKER_EXTRA_TAG}"
fi
