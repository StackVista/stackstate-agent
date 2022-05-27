ARG BASE_IMAGE
FROM ${BASE_IMAGE}

ARG UID
ARG GID
ARG MYUSER=dev

# Add local user to match id/gid of host OS user
RUN (groupadd -g $GID $MYUSER || true) && \
    useradd -m -u $UID -g $GID $MYUSER -s /bin/sh

# Ensure local user can run all root commands
RUN usermod -aG root $MYUSER && \
    chmod g+rx /root && \
    chown -R $MYUSER /go/src

# Ensure local user can run omnibus build
RUN mkdir /opt/stackstate-agent && \
    chown -R $MYUSER /opt/stackstate-agent

USER $MYUSER

COPY ./local_init.sh /local_init.sh

# external docker volume
# if we don't precreate it we do not get the permission we want
RUN mkdir /go/pkg && \
    chmod a+w /go/pkg

ENTRYPOINT []
CMD ["bash", "--init-file", "/local_init.sh"]
