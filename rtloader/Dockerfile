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

# external docker volume
# if we don't precreate it we do not get the permission we want
RUN mkdir /go/pkg && \
    chmod a+w /go/pkg

WORKDIR /go/src/github.com/StackVista/stackstate-agent

#ENTRYPOINT ["rtloader/entrypoint.sh"]
CMD ["bash", "--init-file", "rtloader/entrypoint.sh"]
