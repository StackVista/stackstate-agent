##########################################
# Preparation stage: extract and cleanup #
##########################################
FROM registry.tooling.stackstate.io/library/docker/ubuntu:20.04 AS extract
ARG ARCH="amd64"
ENV ARCH=$ARCH
COPY stackstate-agent*_${ARCH}.deb /
WORKDIR /output

RUN dpkg -x /stackstate-agent*_${ARCH}.deb . \
 && rm -rf usr etc/init lib \
    opt/stackstate-agent/sources \
    opt/stackstate-agent/embedded/share/doc \
    opt/stackstate-agent/embedded/share/man \
    # remove libcurl 4.4.0 - it gets removed later anyway but the docker store scanner doesn't see that
    opt/stackstate-agent/embedded/lib/libcurl.so.4.4.0 \
    # ditto for this older libsystemd
    usr/lib/x86_64-linux-gnu/libsystemd.so.0.21.0 \
 && find opt/stackstate-agent/ -iname "*.a" -delete \
 && rm -rf opt/stackstate-agent/bin/agent/dist/jmx \
 && ln -s /opt/stackstate-agent/embedded/ssl etc/ssl \
 && mkdir conf.d checks.d

# Configuration:
#   - copy default config files
COPY stackstate*.yaml etc/stackstate-agent/

FROM registry.suse.com/bci/bci-base:15.6 AS software
# as per comment here: https://github.com/rancher/rancher/blob/main/package/Dockerfile#L12
# No-op command to create an explicit layer - this fixes a weird buildkit/buildx bug on macos arm
RUN :
####################################
# Actual docker image construction #
####################################

FROM registry.suse.com/bci/bci-micro:15.6
LABEL maintainer="StackState <info@stackstate.com>"
# Install system packages using builder image that has zypper

ARG user="stackstate-agent"
ENV DOCKER_STS_AGENT=true \
    PATH=/opt/stackstate-agent/bin/agent/:/opt/stackstate-agent/embedded/bin/:$PATH \
    CURL_CA_BUNDLE=/opt/stackstate-agent/embedded/ssl/certs/cacert.pem

# Copy agent from extract stage
COPY --from=extract /output/ /
COPY --from=software /usr/bin/find /usr/bin/find
RUN rm -rf /etc/services.d/network

COPY cont-init.d /etc/cont-init.d/
COPY probe.sh initlog.sh secrets-helper/readsecret.py /

RUN echo "$user:x:1000:1000::/home/$user:/bin/bash" >> /etc/passwd && \
    echo "$user:x:1000:" >> /etc/group && \
    mkdir /home/$user && \
    chown -R $user:$user /home/$user && \
    chown -R $user:root /etc/$user/ /var/log/$user/ && \
    chmod g+r,g+w,g+X -R /etc/stackstate-agent/ /var/log/stackstate-agent/ && \
    chmod 755 /probe.sh /initlog.sh && \
    chown root:root /readsecret.py && \
    chmod 500 /readsecret.py && \
    rm /opt/stackstate-agent/embedded/bin/process-agent

USER 1000:1000

# Expose DogStatsD and trace-agent ports
EXPOSE 8125/udp 8126/tcp

HEALTHCHECK --interval=2m --timeout=5s --retries=2 \
  CMD ["/probe.sh"]

# Leave following directories RW to allow use of kubernetes readonlyrootfs flag
VOLUME ["/etc/stackstate-agent", "/var/log/stackstate", "/tmp"]

COPY entrypoint.sh /
CMD ["/entrypoint.sh"]
