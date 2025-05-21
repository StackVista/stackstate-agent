##########################################
# Preparation stage: extract and cleanup #
##########################################
FROM artifactory.tooling.stackstate.io/docker-virtual/ubuntu:20.04 as extract
ARG ARCH="amd64"
ARG S6_ARCH="amd64"
ARG WITH_JMX
ENV ARCH=$ARCH
ENV S6_ARCH=$S6_ARCH
COPY stackstate-agent*_${ARCH}.deb /
WORKDIR /output

# Get s6-overlay and check gpg signature
# Hack: PGP key store locally as it's currently not available from https://keybase.io/justcontainers/key.asc
COPY hack/s6-pgp-key /tmp/s6-pgp-key
ENV S6_VERSION v1.21.2.2
ADD https://github.com/just-containers/s6-overlay/releases/download/${S6_VERSION}/s6-overlay-${S6_ARCH}.tar.gz /output/s6.tgz
ADD https://github.com/just-containers/s6-overlay/releases/download/${S6_VERSION}/s6-overlay-${S6_ARCH}.tar.gz.sig /tmp/s6.tgz.sig
RUN apt-get update \
 && apt-get install --no-install-recommends -y gpg gpg-agent curl ca-certificates \
 && cat /tmp/s6-pgp-key | gpg --import \
 && gpg --verify /tmp/s6.tgz.sig /output/s6.tgz

# Extract and cleanup:
#   - unused systemd unit
#   - GPL sources for embedded software  # FIXME: move upstream
#   - docs and manpages                  # FIXME: move upstream
#   - static libraries                   # FIXME: move upstream
#   - jmxfetch on nojmx build
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
 && if [ -z "$WITH_JMX" ]; then rm -rf opt/stackstate-agent/bin/agent/dist/jmx; fi \
 && ln -s /opt/stackstate-agent/embedded/ssl etc/ssl \
 && mkdir conf.d checks.d

# Configuration:
#   - copy default config files
COPY stackstate*.yaml etc/stackstate-agent/


####################################
# Actual docker image construction #
####################################

FROM registry.suse.com/bci/bci-micro:15.6 as final
RUN : # No-op command to create an explicit layer - this fixes a weird buildkit/buildx bug on macos arm

# Temporary build stage image
FROM opensuse/leap:latest AS chroot-builder
# Install system packages using builder image that has zypper
COPY --from=final / /chroot/

LABEL maintainer "StackState <info@stackstate.com>"
ARG WITH_JMX
ENV DOCKER_STS_AGENT=true \
    PATH=/opt/stackstate-agent/bin/agent/:/opt/stackstate-agent/embedded/bin/:$PATH \
    CURL_CA_BUNDLE=/opt/stackstate-agent/embedded/ssl/certs/cacert.pem

COPY installs.sh /
COPY jmx-installs.sh /

# Install openjdk-8-jre-headless on jmx flavor
#RUN if [ -n "$WITH_JMX" ]; then zypper refresh \
# && mkdir /usr/share/man/man1 \
# && zypper --installroot /chroot -n in --no-recommends java-1_8_0-openjdk-headless \
# && zypper --non-interactive clean --all \
# && rm -rf /var/cache/zypp/* /tmp/* /var/tmp/*; fi
RUN if [ -n "$WITH_JMX" ]; then \
      zypper refresh; \
      mkdir /usr/share/man/man1; \
      zypper --installroot /chroot -n in --no-recommends java-1_8_0-openjdk-headless; \
      zypper --non-interactive clean --all; \
      rm -rf /var/cache/zypp/* /tmp/* /var/tmp/*; fi

# make sure we have recent dependencies
#RUN zypper --non-interactive refresh && \
#    zypper --non-interactive update && \
#    zypper --installroot /chroot -n in --no-recommends \
#      systemd libncurses6 libgcrypt20 && \
#    rm -f /usr/sbin/runuser /usr/lib64/libdb-5.3.so && \
#    zypper --non-interactive clean --all && \
#    rm -rf /var/cache/zypp/* /tmp/* /var/tmp/*
RUN zypper --non-interactive refresh; \
    zypper --non-interactive update; \
    zypper --installroot /chroot -n in --no-recommends systemd libncurses6 libgcrypt20; \
    rm -f /usr/sbin/runuser /usr/lib64/libdb-5.3.so; \
    zypper --non-interactive clean --all; \
    rm -rf /var/cache/zypp/* /tmp/* /var/tmp/*;

# Copy agent from extract stage
COPY --from=extract /output/ /chroot/
RUN rm -rf /etc/services.d/network

COPY cont-init.d /chroot/etc/cont-init.d/
COPY probe.sh initlog.sh secrets-helper/readsecret.py /chroot/

FROM final
COPY --from=chroot-builder /chroot/ /

LABEL maintainer "StackState <info@stackstate.com>"
ARG WITH_JMX
ENV DOCKER_STS_AGENT=true \
    PATH=/opt/stackstate-agent/bin/agent/:/opt/stackstate-agent/embedded/bin/:$PATH \
    CURL_CA_BUNDLE=/opt/stackstate-agent/embedded/ssl/certs/cacert.pem

RUN adduser --system --no-create-home --disabled-password --ingroup root stackstate-agent \
 && rm /var/run \
 && chown -R stackstate-agent:root /etc/stackstate-agent/ /var/log/stackstate-agent/ \
 && chmod g+r,g+w,g+X -R /etc/stackstate-agent/ /var/log/stackstate-agent/ \
 && chmod 755 /probe.sh /initlog.sh \
 && chown root:root /readsecret.py \
 && chmod 500 /readsecret.py \
# Remove the process-agent from the agent container, as it is not used and a source of vulnerabilities
 && rm opt/stackstate-agent/embedded/bin/process-agent

# Expose DogStatsD and trace-agent ports
EXPOSE 8125/udp 8126/tcp

HEALTHCHECK --interval=2m --timeout=5s --retries=2 \
  CMD ["/probe.sh"]

# Leave following directories RW to allow use of kubernetes readonlyrootfs flag
VOLUME ["/etc/stackstate-agent", "/var/log/stackstate", "/tmp"]

CMD ["/opt/stackstate-agent/bin/agent/agent", "run"]
