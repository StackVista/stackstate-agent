#!/bin/sh

set -xe

IMAGE_TAG="${1}"
IMAGE_REPO="${2}"
DOCKERFILE_PATH="${3}"
PUSH_LATEST="${4:-false}"
REGISTRY="${5:-docker.io}"

echo "${IMAGE_TAG}"
echo "${IMAGE_REPO}"
echo "${DOCKERFILE_PATH}"

docker build -t "${REGISTRY}/stackstate/${IMAGE_REPO}:${IMAGE_TAG}" "${DOCKERFILE_PATH}"
docker login -u "${DOCKER_USER}" -p "${DOCKER_PASS}"
docker push "${REGISTRY}/stackstate/${IMAGE_REPO}:${IMAGE_TAG}"

if [ "$PUSH_LATEST" = "true" ]; then
    docker tag "${REGISTRY}/stackstate/${IMAGE_REPO}:${IMAGE_TAG}" "${REGISTRY}/stackstate/${IMAGE_REPO}:latest"
    echo 'Pushing release to latest'
    docker push "${REGISTRY}/stackstate/${IMAGE_REPO}:latest"
fi
