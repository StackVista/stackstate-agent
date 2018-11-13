#!/bin/sh

IMAGE_REPO=$1
ARTIFACT_PATH=$2
IMAGE_TAG=$(inv version)

echo $IMAGE_TAG
echo $IMAGE_REPO
echo $ARTIFACT_PATH

cp $ARTIFACT_PATH/*.deb Dockerfiles/agent

docker build -t stackstate/$IMAGE_REPO:$IMAGE_TAG -t stackstate/$IMAGE_REPO:latest Dockerfiles/agent
docker login -u $DOCKER_USER -p $DOCKER_PASS
docker push stackstate/stackstate-agent:$IMAGE_TAG
docker push stackstate/stackstate-agent:latest
