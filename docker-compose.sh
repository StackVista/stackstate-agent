#!/usr/bin/env bash

# Run the agent of the current branch in a docker compose
# We then export the logs to a file

AGENT_VERSION=$(git rev-parse --abbrev-ref HEAD)

export STACKSTATE_BRANCH="master"
export AGENT_VERSION="$AGENT_VERSION"

# Source the molecule env files
source ./test/.env
source ./test/.envrc

if [[ -n $quay_user ]] && [[ -n $quay_password ]]; then
    echo "Quay login with user: $quay_user"
    docker login -u "${quay_user}" -p "${quay_password}" "quay.io"
fi

# Logging
echo "AGENT_VERSION: $AGENT_VERSION"

rm docker-compose-*.log
docker-compose kill
docker rmi "stackstate/${STS_DOCKER_TEST_REPO:-stackstate-agent-test}:$AGENT_VERSION" --force

# Startup a new docker instance outputting the logs
docker-compose up --detach

# Compose container id
CONTAINER_ID=$(docker ps | grep 'quay.io/stackstate/stackstate-receiver' | awk '{ print $1 }')

sleep 15

# Record all the logs to files
docker-compose logs --no-color --follow --tail 1 zookeeper >& "docker-compose-zookeeper-$AGENT_VERSION".log &
docker-compose logs --no-color --follow --tail 1 kafka >& "docker-compose-kafka-$AGENT_VERSION".log &
docker-compose logs --no-color --follow --tail 1 stackstate-agent >& "docker-compose-agent-$AGENT_VERSION".log &
docker exec "$CONTAINER_ID" tail -f /opt/docker/logs/stackstate-receiver.log >& "docker-compose-receiver-$AGENT_VERSION".log &

