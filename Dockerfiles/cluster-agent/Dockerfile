########################################
# Preparation stage: layout and chmods #
########################################

FROM registry.tooling.stackstate.io/library/docker/ubuntu:20.04 as builder

WORKDIR /output

COPY stackstate-cluster-agent opt/stackstate-agent/bin/stackstate-cluster-agent
COPY ./conf.d etc/stackstate-agent/conf.d
COPY ./stackstate-cluster.yaml etc/stackstate-agent/stackstate-cluster.yaml
# TODO: are templates no longer present?
#COPY ./dist/templates opt/stackstate-agent/bin/dist/templates
COPY entrypoint.sh .

RUN chmod 755 entrypoint.sh \
    && chmod g+r,g+w,g+X -R etc/stackstate-agent/ \
    && chmod +x opt/stackstate-agent/bin/stackstate-cluster-agent \
    && ln -s /opt/stackstate-agent/bin/stackstate-cluster-agent opt/stackstate-agent/bin/agent

####################################
# Actual docker image construction #
####################################

FROM registry.suse.com/bci/bci-base:15.6 AS software
# as per comment here: https://github.com/rancher/rancher/blob/main/package/Dockerfile#L12
# No-op command to create an explicit layer - this fixes a weird buildkit/buildx bug on macos arm
RUN :

FROM registry.suse.com/bci/bci-micro:15.6 AS final
LABEL maintainer="StackState <info@stackstate.com>"

ARG user="stackstate-agent"
ENV PATH="/opt/stackstate-agent/bin/:$PATH"

COPY --from=builder /output /
COPY --from=software /usr/bin/find /usr/bin/find

# Allow running as an unprivileged user:
# - General case is the stackstate-agent user
# - OpenShift uses a random UID in the root group
#
# Containerd does not preserve permissions when mounting a volume on top
# of an empty folder. Creating .placeholder files as a workaround.
#
RUN echo "$user:x:1000:1000::/home/$user:/bin/bash" >> /etc/passwd && \
    echo "$user:x:1000:" >> /etc/group && \
    mkdir -p /var/log/stackstate/ /conf.d && \
    touch /var/log/stackstate/.placeholder && \
    touch /tmp/.placeholder && \
    mkdir /home/$user && \
    chown -R $user:$user /home/$user && \
    chown -R stackstate-agent:root /etc/stackstate-agent/ /var/log/stackstate/ /conf.d /tmp/ && \
    chmod g+r,g+w,g+X -R /etc/stackstate-agent/ /var/log/stackstate/ /conf.d /tmp/

USER 1000:1000
# Incompatible with the custom metrics API on port 443
# Set STS_EXTERNAL_METRICS_PROVIDER_PORT to a higher value to run as non-root
# USER stackstate-agent

# Leave following directories RW to allow use of readonly rootfs
VOLUME ["/etc/stackstate-agent", "/var/log/stackstate", "/tmp"]

ENTRYPOINT ["/entrypoint.sh"]

# No docker healthcheck, use a HTTP check
# on port 5005 and/or 443 on Kubernetes

CMD ["stackstate-cluster-agent", "start"]
