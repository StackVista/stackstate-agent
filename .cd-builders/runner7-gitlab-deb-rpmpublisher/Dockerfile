ARG BASE_RUNNER_IMAGE_VERSION=deb7_latest
FROM docker.io/stackstate/stackstate-agent-runner-circle:${BASE_RUNNER_IMAGE_VERSION}

RUN pip install boto3 sa-rpm-s3 invoke distro==1.4.0 awscli==1.16
RUN apt update  && apt-get install -y libkrb5-dev unixodbc-dev rpm python-deltarpm createrepo yum-utils

ADD rpmmacros /root/.rpmmacros
