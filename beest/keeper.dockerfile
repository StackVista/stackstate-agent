ARG BASE_IMAGE
FROM ${BASE_IMAGE}

ARG UID
ARG GID
ARG DOCKER_GID
ARG MYUSER=keeper
ARG UNAME_S

RUN (groupadd -g $GID $MYUSER || true) && \
    useradd -m -u $UID -g $GID $MYUSER -s /bin/sh
RUN if [ $UNAME_S = "Linux" ]; then \
      groupadd -g $DOCKER_GID docker && \
      usermod -a -G docker $MYUSER; \
    fi

USER $MYUSER

COPY direnv.toml /home/$MYUSER/.config/direnv/
RUN chown -R $MYUSER:$MYUSER /home/$MYUSER/.config/

# external docker volume
# if we don't precreate it we do not get the permission we want
RUN mkdir /go/pkg && \
    chmod a+w /go/pkg

WORKDIR /go/src/app

CMD ["bash", "--init-file", "./bootstrap.sh"]
