ARG AGENT_VERSION=7
FROM registry.datadoghq.com/agent:${AGENT_VERSION}

RUN apt-get update
RUN rm -vf /etc/ssl/openssl.cnf
RUN DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC apt-get install -y gdb build-essential strace less vim

# Install go
ARG GO_SHA256_AMD64=708effb774be8237570d0add163225abbdfaf4fca28b2611df167beba4feef89
ARG GO_SHA256_ARM64=d0507e9e9d7fe012aae570108cbd76c15de879e17130ab8cb90d4d7445cb1f2e
RUN arch="$(dpkg --print-architecture)" \
 && case "${arch}" in \
      amd64) go_sha256="${GO_SHA256_AMD64}" ;; \
      arm64) go_sha256="${GO_SHA256_ARM64}" ;; \
      *) echo "Unsupported architecture: ${arch}" >&2; exit 1 ;; \
    esac \
 && curl -fSL -o golang.tgz "https://go.dev/dl/go1.26.6.linux-${arch}.tar.gz" \
 && echo "${go_sha256}  golang.tgz" | sha256sum -c -
RUN tar xzvf golang.tgz
RUN ln -s /go /goroot

# Install go-delve
RUN /go/bin/go install github.com/go-delve/delve/cmd/dlv@latest

# Install datadog-agent-dbg at the right version
RUN apt-get install -o Dpkg::Options::="--force-confold" --no-install-recommends -y gnupg dirmngr ca-certificates \
 && apt-key adv --recv-keys --keyserver hkp://keyserver.ubuntu.com:80 E6266D4AC0962C7D
ARG AGENT_VERSION
ARG APT_REPO=apt.datadoghq.com
ARG APT_DISTRIBUTION=stable
RUN echo "deb https://${APT_REPO}/ ${APT_DISTRIBUTION} 7 6" > /etc/apt/sources.list.d/datadog.list \
 && apt-get update -o Dir::Etc::sourcelist="sources.list.d/datadog.list" \
 && cd /root/ \
 && DEB_AGENT_VERSION=$(echo "${AGENT_VERSION}" | sed "s/-/~/") \
 && apt-get download datadog-agent-dbg=1:${DEB_AGENT_VERSION}-1 \
 && dpkg -X ./datadog-agent-dbg*.deb / \
 && rm -f ./datadog-agent-dbg*.deb

# Add GDB debug script for libpython
RUN TARGET_FILE=$(find /opt/datadog-agent/.debug/opt/datadog-agent/embedded/lib/ -name 'libpython3.*0.dbg' | sed 's/.dbg/-gdb.py/') \
 && curl https://raw.githubusercontent.com/python/cpython/master/Tools/gdb/libpython.py -o ${TARGET_FILE}

RUN mkdir -p /usr/lib/debug/.build-id/$(readelf -n /opt/datadog-agent/bin/agent/agent | perl -nE 'say $1 if /Build ID: (..)(.*)/')

RUN cp /opt/datadog-agent/.debug/opt/datadog-agent/bin/agent/agent.dbg \
       /usr/lib/debug/.build-id/$(readelf -n /opt/datadog-agent/bin/agent/agent | perl -nE 'say "$1/$2.debug" if /Build ID: (..)(.*)/')

# Configure  gdb
COPY .gdbinit /root/

ENTRYPOINT ["gdb", "/opt/datadog-agent/bin/agent/agent"]
