FROM artifactory.tooling.stackstate.io/docker-virtual/stackstate/stackstate-agent-runner-gitlab:debian-20220505

ARG UID
ARG GID
ARG MYUSER=dev

# Add local user to match id/gid of host user
RUN (groupadd -g $GID $MYUSER || true) && \
    useradd -m -u $UID -g $GID $MYUSER -s /bin/sh

# Ensure local user can run all root commands
RUN usermod -aG root $MYUSER && \
    chmod g+rx /root && \
    chown -R $MYUSER /go/src

# Ensure local user can run omnibus build
RUN mkdir /opt/stackstate-agent /etc/stackstate-agent /var/log/stackstate-agent \
          /var/log/sa /etc/init && \
    chown -R $MYUSER /opt/stackstate-agent /etc/stackstate-agent /var/log/stackstate-agent \
                     /var/log/sa /etc/init /etc/init.d /lib/systemd/system /usr/bin

USER $MYUSER

COPY ./local_init.sh /local_init.sh

# external docker volume
# if we don't precreate it we do not get the permission we want
RUN mkdir /go/pkg && \
    chmod a+w /go/pkg

ENTRYPOINT []
CMD ["bash", "--init-file", "/local_init.sh"]
