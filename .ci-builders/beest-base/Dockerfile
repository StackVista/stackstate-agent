FROM golang:1.17.8-bullseye

RUN apt-get update && \
    apt-get install -y --no-install-recommends less direnv unzip python3-pip bash-completion vim jq && \
    rm -rf /var/lib/apt/lists/*

COPY requirements* ./
RUN pip install -r requirements-pip-full.txt

COPY ansible.cfg /etc/ansible/
RUN ansible-galaxy collection install -r requirements-galaxy.yml

ARG TERRAFORM_V=1.0.11
RUN curl -O "https://releases.hashicorp.com/terraform/${TERRAFORM_V}/terraform_${TERRAFORM_V}_linux_amd64.zip" && \
    unzip terraform_${TERRAFORM_V}_linux_amd64.zip && \
    rm terraform_${TERRAFORM_V}_linux_amd64.zip && \
    mv ./terraform /usr/local/bin/

# ensure https://terragrunt.gruntwork.io/docs/getting-started/supported-terraform-versions/
ARG TERRAGRUNT_V=0.35.18
RUN curl -L "https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_V}/terragrunt_linux_amd64" -o terragrunt && \
    chmod +x ./terragrunt && \
    mv ./terragrunt /usr/local/bin/

ARG PACKER_V=1.7.8
RUN curl -O "https://releases.hashicorp.com/packer/${PACKER_V}/packer_${PACKER_V}_linux_amd64.zip" && \
    unzip packer_${PACKER_V}_linux_amd64.zip && \
    rm packer_${PACKER_V}_linux_amd64.zip && \
    mv ./packer /usr/local/bin/

ARG AWSCLI_V=2.7.4
RUN curl -L "https://awscli.amazonaws.com/awscli-exe-linux-x86_64-${AWSCLI_V}.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && \
    ./aws/install && \
    rm awscliv2.zip

ARG KUBECTL_V=1.22.4
RUN curl -LO "https://dl.k8s.io/release/v${KUBECTL_V}/bin/linux/amd64/kubectl" && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/

ARG K9S_V=0.25.18
RUN curl -LO "https://github.com/derailed/k9s/releases/download/v${K9S_V}/k9s_Linux_x86_64.tar.gz" && \
    mkdir -p tmp_k9s && \
    tar xvzf k9s_Linux_x86_64.tar.gz -C tmp_k9s && \
    mv tmp_k9s/k9s /usr/local/bin/ && \
    rm -r k9s_Linux_x86_64.tar.gz tmp_k9s

ARG STERN_V=1.11.0
RUN curl -L "https://github.com/wercker/stern/releases/download/${STERN_V}/stern_linux_amd64" -o stern && \
    chmod +x ./stern && \
    mv ./stern /usr/local/bin/

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

ARG K3D_V=5.2.0
RUN curl -s "https://raw.githubusercontent.com/rancher/k3d/main/install.sh" | TAG=v${K3D_V} bash

ARG DOCKER_V=20.10.9
RUN curl -O "https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_V}.tgz" && \
    tar xvzf docker-${DOCKER_V}.tgz && \
    mv ./docker/docker /usr/local/bin/ && \
    rm -r docker-${DOCKER_V}.tgz docker

ARG STSTOOLBOX_V=1.3.29
RUN curl -O "https://stseuw1-tooling-main-homebrew.s3.amazonaws.com/sts-toolbox/v${STSTOOLBOX_V}/sts-toolbox-${STSTOOLBOX_V}.linux-amd64.tar.gz" && \
    tar xvzf sts-toolbox-${STSTOOLBOX_V}.linux-amd64.tar.gz && \
    rm -r sts-toolbox-${STSTOOLBOX_V}.linux-amd64.tar.gz && \
    mv ./sts-toolbox /usr/local/bin/

ARG STSCLI_V=0.3.7
RUN curl --insecure -O "https://cli-dl.stackstate.com.s3.amazonaws.com/stackstate-cli/v${STSCLI_V}/stackstate-cli-full-${STSCLI_V}.linux-x86_64.tar.gz" && \
    tar xvzf stackstate-cli-full-${STSCLI_V}.linux-x86_64.tar.gz && \
    rm -r stackstate-cli-full-${STSCLI_V}.linux-x86_64.tar.gz && \
    mv ./sts /usr/local/bin/
