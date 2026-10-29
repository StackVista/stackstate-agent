FROM golang:1.27.3-trixie@sha256:c65ecfe613b539ac74c62a4c3b2c27e8f8a6076384c8f27e8a3c0f3be17cc385

RUN cross_debian_arch=$(uname -m | sed -e 's/aarch64/amd64/'  -e 's/x86_64/arm64/'); \
    cross_pkg_arch=$(uname -m | sed -e 's/aarch64/x86-64/' -e 's/x86_64/aarch64/'); \
    dpkg --add-architecture amd64 && dpkg --add-architecture arm64 && \
    apt-get update -y && apt-get dist-upgrade -y && \
    apt-get install -y sudo less curl wget make git cmake unzip \
        libc6-dev g++ gcc pkgconf clang-17 clang-format-17 \
        gcc-${cross_pkg_arch}-linux-gnu libc6-${cross_debian_arch}-cross \
        musl-dev:amd64 musl-dev:arm64 && \
    apt-get clean autoclean && \
    apt-get autoremove --yes

ARG UID
ARG GID

RUN addgroup --gid ${GID} build
RUN useradd -ms /bin/bash build -u ${UID} -g ${GID}
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
RUN adduser build sudo
RUN chown -R build:build /go
ENV GOTOOLCHAIN=auto

USER build

WORKDIR /app

# Add host-profiler tools to PATH for convenience
ENV PATH="/app/tools/host-profiler:${PATH}"

VOLUME /go

CMD ["build_and_launch.sh"]
