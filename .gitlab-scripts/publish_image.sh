#!/bin/sh

TAG=$1
REPO=$2

cp ${@:3} Dockerfiles/agent

docker build -t stackstate/$REPO:$TAG -t stackstate/$REPO:latest Dockerfiles/agent
docker login -u $DOCKER_USER -p $DOCKER_PASS
docker push stackstate/stackstate-agent:$TAG
docker push stackstate/stackstate-agent:latest
