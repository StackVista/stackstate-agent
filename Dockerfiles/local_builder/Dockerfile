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
    chown -R dev /go/src

USER $MYUSER

COPY ./shell_init.sh /shell_init.sh

# external docker volume
# if we don't precreate it we do not get the permission we want
RUN mkdir /go/pkg && \
    chmod a+w /go/pkg

WORKDIR /go/src/github.com/StackVista/stackstate-agent

ENTRYPOINT []
CMD ["bash", "--init-file", "/shell_init.sh"]
