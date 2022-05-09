#!/usr/bin/env sh

export DEBIAN_FRONTEND=noninteractive
export TZ=Etc/UTC

apt-get update
apt-get upgrade -y
apt-get install -y curl \
    unzip \
    software-properties-common \
    git bash wget openssl gnupg

add-apt-repository --yes --update ppa:ansible/ansible

apt-get install -y ansible
