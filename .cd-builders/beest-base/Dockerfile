FROM golang:1.17.3-bullseye

RUN apt-get update && \
    apt-get install -y less direnv unzip pip bash-completion vim && \
    rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install -r requirements.txt && \
    rm requirements.txt

ARG TERRAGRUNT_V=0.35.18
RUN curl -L "https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_V}/terragrunt_linux_amd64" -o terragrunt && \
    chmod +x ./terragrunt && \
    mv ./terragrunt /usr/local/bin/

ARG TERRAFORM_V=1.0.11
RUN curl -O "https://releases.hashicorp.com/terraform/${TERRAFORM_V}/terraform_${TERRAFORM_V}_linux_amd64.zip" && \
    unzip terraform_${TERRAFORM_V}_linux_amd64.zip && \
    rm terraform_${TERRAFORM_V}_linux_amd64.zip && \
    mv ./terraform /usr/local/bin/

ARG PACKER_V=1.7.8
RUN curl -O "https://releases.hashicorp.com/packer/${PACKER_V}/packer_${PACKER_V}_linux_amd64.zip" && \
    unzip packer_${PACKER_V}_linux_amd64.zip && \
    rm packer_${PACKER_V}_linux_amd64.zip && \
    mv ./packer /usr/local/bin/

ARG KUBECTL_V=1.22.4
RUN curl -LO "https://dl.k8s.io/release/v${KUBECTL_V}/bin/linux/amd64/kubectl" && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/

ARG K9S_V=0.25.8
RUN curl -LO "https://github.com/derailed/k9s/releases/download/v${K9S_V}/k9s_Linux_x86_64.tar.gz" && \
    mkdir -p tmp_k9s && \
    tar xvzf k9s_Linux_x86_64.tar.gz -C tmp_k9s && \
    mv tmp_k9s/k9s /usr/local/bin/ && \
    rm -r k9s_Linux_x86_64.tar.gz tmp_k9s

#ARG KREW_V=0.4.2
#RUN curl -LO "https://github.com/kubernetes-sigs/krew/releases/download/v${KREW_V}/krew-linux_amd64.tar.gz" && \
#    tar zxvf krew-linux_amd64.tar.gz && \
#    ./krew-linux_amd64 install krew && \
#    PATH="/root/.krew/bin:$PATH" kubectl krew install ctx
#
#RUN kubectl krew install ctx && \
#    kubectl krew install ns

ARG HELM_V=3.7.1
RUN curl -O "https://get.helm.sh/helm-v${HELM_V}-linux-amd64.tar.gz" && \
    tar xvzf helm-v${HELM_V}-linux-amd64.tar.gz && \
    mv ./linux-amd64/helm /usr/local/bin/ && \
    rm -r helm-v${HELM_V}-linux-amd64.tar.gz linux-amd64/

ARG STSTOOLBOX_V=1.3.17
RUN curl -O "https://stseuw1-tooling-main-homebrew.s3.amazonaws.com/sts-toolbox/v${STSTOOLBOX_V}/sts-toolbox-${STSTOOLBOX_V}.linux-amd64.tar.gz" && \
    tar xvzf sts-toolbox-${STSTOOLBOX_V}.linux-amd64.tar.gz && \
    rm -r sts-toolbox-${STSTOOLBOX_V}.linux-amd64.tar.gz && \
    mv ./sts-toolbox /usr/local/bin/

ARG K3D_V=5.2.0
RUN curl -s "https://raw.githubusercontent.com/rancher/k3d/main/install.sh" | TAG=v${K3D_V} bash

ARG DOCKER_V=20.10.9
RUN curl -O "https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_V}.tgz" && \
    tar xvzf docker-${DOCKER_V}.tgz && \
    mv ./docker/docker /usr/local/bin/ && \
    rm -r docker-${DOCKER_V}.tgz docker
