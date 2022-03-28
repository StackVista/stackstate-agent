FROM artifactory.tooling.stackstate.io/docker-virtual/stackstate/stackstate-agent-runner-gitlab:latest7

ARG UID
ARG GID
ARG DOCKER_GID
ARG MYUSER=developer
ARG MYUSER_HOME=/home/$MYUSER
ARG UNAME_S
ENV DD_PIP_VERSION $DD_PIP_VERSION
ENV DD_SETUPTOOLS_VERSION $DD_SETUPTOOLS_VERSION
ARG DD_CONDA_VERSION=4.7.10

COPY scripts/entry_point.sh /entry_point.sh
COPY scripts/shell.sh /shell.sh

RUN chmod +x /entry_point.sh
RUN chmod +x /shell.sh

RUN (groupadd -g $GID $MYUSER || true) && \
    useradd -m -u $UID -g $GID -G sudo $MYUSER -s /bin/sh
RUN if [ $UNAME_S = "Linux" ]; then \
      groupadd -g $DOCKER_GID docker && \
      usermod -a -G docker $MYUSER; \
    fi

RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

RUN chmod a+w /go/src
RUN chmod a+w /go/src/github.com

USER $MYUSER

# Conda
RUN curl -sL -o $MYUSER_HOME/miniconda.sh https://repo.anaconda.com/miniconda/Miniconda3-${DD_CONDA_VERSION}-Linux-x86_64.sh
RUN bash $MYUSER_HOME/miniconda.sh -b
#RUN source $HOME/miniconda3/etc/profile.d/conda.sh
ENV PKG_CONFIG_LIBDIR "${PKG_CONFIG_LIBDIR}:$MYUSER_HOME/miniconda3/lib/pkgconfig"
ENV PATH "$MYUSER_HOME/miniconda3/bin:${PATH}"
RUN conda init bash

# Setup pythons
RUN conda create -n ddpy2 python python=2
RUN conda create -n ddpy3 python python=3.8

# Update pip, setuptools and misc deps
RUN /bin/bash -c "source $MYUSER_HOME/miniconda3/etc/profile.d/conda.sh && conda activate ddpy2 \
    && pip install -i https://pypi.python.org/simple pip==${DD_PIP_VERSION} \
    && pip install --ignore-installed setuptools==${DD_SETUPTOOLS_VERSION} \
    && pip install invoke distro==1.4.0 awscli==1.16.240"

# Update pip, setuptools and misc deps
RUN /bin/bash -c  "source $MYUSER_HOME/miniconda3/etc/profile.d/conda.sh && conda activate ddpy3 \
    && pip install -i https://pypi.python.org/simple pip==${DD_PIP_VERSION} \
    && pip install --ignore-installed setuptools==${DD_SETUPTOOLS_VERSION} \
    && pip install invoke distro==1.4.0 awscli==1.16.240"

# external docker volume
# if we don't precreate it we do not get the permission we want
RUN mkdir /go/pkg && \
    chmod a+w /go/pkg

WORKDIR /go/src/app

ENTRYPOINT ["/entry_point.sh"]
