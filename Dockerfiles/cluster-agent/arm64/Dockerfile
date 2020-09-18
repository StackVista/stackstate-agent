########################################
# Preparation stage: layout and chmods #
########################################

FROM debian:bullseye-slim as builder

WORKDIR /output

COPY datadog-cluster-agent.arm64 opt/datadog-agent/bin/datadog-cluster-agent
COPY ./conf.d etc/datadog-agent/conf.d
COPY ./datadog-cluster.yaml etc/datadog-agent/datadog-cluster.yaml
COPY ./security-agent-policies/compliance/containers/ etc/datadog-agent/compliance.d
COPY entrypoint.sh .
COPY readsecret.sh .

RUN chmod 755 entrypoint.sh \
    && chown root:root readsecret.sh \
    && chmod 550 readsecret.sh \
    && chmod g+r,g+w,g+X -R etc/datadog-agent/ \
    && chmod +x opt/datadog-agent/bin/datadog-cluster-agent \
    && ln -s /opt/datadog-agent/bin/datadog-cluster-agent opt/datadog-agent/bin/agent

####################################
# Actual docker image construction #
####################################

FROM debian:bullseye-slim

LABEL maintainer "StackState <info@stackstate.com>"

ENV PATH="/opt/stackstate-agent/bin/:$PATH"

RUN apt-get update \
    && apt full-upgrade -y \
    && apt-get install --no-install-recommends -y ca-certificates curl \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=builder /output /

# Allow running as an unprivileged user:
# - General case is the stackstate-agent user
# - OpenShift uses a random UID in the root group
#
# Containerd does not preserve permissions when mounting a volume on top
# of an empty folder. Creating .placeholder files as a workaround.
#
RUN adduser --system --no-create-home --disabled-password --ingroup root stackstate-agent \
    && mkdir -p /var/log/stackstate/ /conf.d \
    && touch /var/log/stackstate/.placeholder \
    && touch /tmp/.placeholder \
    && chown -R stackstate-agent:root /etc/stackstate-agent/ /var/log/stackstate/ /conf.d /tmp/ \
    && chmod g+r,g+w,g+X -R /etc/stackstate-agent/ /var/log/stackstate/ /conf.d /tmp/

# Incompatible with the custom metrics API on port 443
# Set STS_EXTERNAL_METRICS_PROVIDER_PORT to a higher value to run as non-root
# USER stackstate-agent

# Leave following directories RW to allow use of readonly rootfs
VOLUME ["/etc/stackstate-agent", "/var/log/stackstate", "/tmp"]

ENTRYPOINT ["/entrypoint.sh"]

# No docker healthcheck, use a HTTP check
# on port 5005 and/or 443 on Kubernetes

CMD ["stackstate-cluster-agent", "start"]
