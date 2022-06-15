ARG BASE_REPO=quay.io/stackstate/stackstate-agent-2-test
ARG BASE_TAG=master
FROM ${BASE_REPO}:${BASE_TAG}

# Disable all checks
RUN rm -rf /etc/stackstate-agent/conf.d/*

# Enable docker_swarm check
COPY ./conf.d /etc/stackstate-agent/conf.d/

# Disable process agent
RUN rm -rf /etc/services.d/process

# Disable trace agent
RUN rm -rf /etc/services.d/trace
