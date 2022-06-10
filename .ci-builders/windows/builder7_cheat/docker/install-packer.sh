#!/usr/bin/env sh

#curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -
#apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
#apt-get install -y packer=1.8.0
export HASHICORP_RELEASES=https://releases.hashicorp.com
export VERSION="1.8.0"

gpg --keyserver keyserver.ubuntu.com --recv-keys C874011F0AB405110D02105534365D9472D7468F && \
    mkdir -p /tmp/build && \
    cd /tmp/build && \

    packerArch='amd64' && \

#    apkArch="$(apk --print-arch)" && \
#    case "${apkArch}" in \
#        aarch64) packerArch='arm64' ;; \
#        armhf) packerArch='arm' ;; \
#        x86) packerArch='386' ;; \
#        x86_64) packerArch='amd64' ;; \
#        *) echo >&2 "error: unsupported architecture: ${apkArch} (see ${HASHICORP_RELEASES}/packer/${VERSION}/)" && exit 1 ;; \
#    esac && \

    wget ${HASHICORP_RELEASES}/packer/${VERSION}/packer_${VERSION}_linux_${packerArch}.zip && \
    wget ${HASHICORP_RELEASES}/packer/${VERSION}/packer_${VERSION}_SHA256SUMS && \
    wget ${HASHICORP_RELEASES}/packer/${VERSION}/packer_${VERSION}_SHA256SUMS.sig && \
    gpg --batch --verify packer_${VERSION}_SHA256SUMS.sig packer_${VERSION}_SHA256SUMS && \
    grep packer_${VERSION}_linux_${packerArch}.zip packer_${VERSION}_SHA256SUMS | sha256sum -c && \
    unzip -d /tmp/build packer_${VERSION}_linux_${packerArch}.zip && \
    cp /tmp/build/packer /bin/packer && \
    cd /tmp && \
    rm -rf /tmp/build && \
    gpgconf --kill all && \
#    apk del gnupg openssl && \
    rm -rf /root/.gnupg && \
    # Tiny smoke test to ensure the binary we downloaded runs
    packer version
