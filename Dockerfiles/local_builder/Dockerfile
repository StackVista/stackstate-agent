FROM artifactory.tooling.stackstate.io/docker-virtual/stackstate/stackstate-agent-runner-gitlab:debian-20220826-patched-20230511

ARG UID
ARG GID
ARG MYUSER=dev

# Used for copy agent source from mount point
RUN apt-get install -y --no-install-recommends inotify-tools rsync

# Add local user to match id/gid of host user
RUN (groupadd -g $GID $MYUSER || true) && \
    useradd -m -u $UID -g $GID $MYUSER -s /bin/sh

# Ensure local user can run all root commands
RUN usermod -aG root $MYUSER && \
    chmod g+rx /root && \
    chown -R $MYUSER /go/src

USER $MYUSER

COPY ./local_init.sh /local_init.sh

# external docker volume
# if we don't precreate it we do not get the permission we want
RUN mkdir /go/pkg && \
    chmod a+w /go/pkg

ENTRYPOINT []
CMD ["bash", "--init-file", "/local_init.sh"]
