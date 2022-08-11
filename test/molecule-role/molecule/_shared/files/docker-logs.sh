#!/bin/bash
set -eou pipefail

id

FOLDER="docker_logs"
CONTAINER_NAMES=$(docker ps -a --format "{{.Names}}")
if [ ! -d "${FOLDER}" ]; then
    mkdir "${FOLDER}"
fi

for CONTAINER_NAME in ${CONTAINER_NAMES}; do
	CONTAINER_LOG_PATH=$(docker inspect --format='{{.LogPath}}' "${CONTAINER_NAME}")
        echo "${CONTAINER_LOG_PATH}"
        if [ -n "${CONTAINER_LOG_PATH}" ]; then
            cp "${CONTAINER_LOG_PATH}" "${FOLDER}/${CONTAINER_NAME}.log"
        fi
done

journalctl -u docker -S today --no-tail > "${FOLDER}/dockerd.log"

tar -zcvf "${FOLDER}.tar.gz" "${FOLDER}"
