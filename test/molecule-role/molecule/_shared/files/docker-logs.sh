#!/bin/env bash
FOLDER="docker_logs"
CONTAINER_NAMES=$(docker ps -a --format "{{.Names}}")
mkdir "${FOLDER}"
for CONTAINER_NAME in ${CONTAINER_NAMES}; do
	CONTAINER_LOG_PATH=$(docker inspect --format='{{.LogPath}}' "${CONTAINER_NAME}")
        echo "${CONTAINER_LOG_PATH}"
        if [ -n "${CONTAINER_LOG_PATH}" ]; then
            cp "${CONTAINER_LOG_PATH}" "${FOLDER}"
        fi
done

tar -zcvf "${FOLDER}.tar.gz" "${FOLDER}"
