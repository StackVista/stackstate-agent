#!/bin/sh

IMAGE_TAG=$1
IMAGE_REPO=$2
ARTIFACT_PATH=$3

echo $IMAGE_TAG
echo $IMAGE_REPO
echo $ARTIFACT_PATH

cp $ARTIFACT_PATH/*.deb Dockerfiles/agent

docker build -t stackstate/$IMAGE_REPO:$IMAGE_TAG -t stackstate/$IMAGE_REPO:latest Dockerfiles/agent
docker login -u $DOCKER_USER -p $DOCKER_PASS
docker push stackstate/stackstate-agent:$IMAGE_TAG
docker push stackstate/stackstate-agent:latest
