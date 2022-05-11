FROM ubuntu:focal-20220316 as builder

ADD docker /docker
COPY ./VCForPython27.msi /root/packages/VCForPython27.msi

RUN /docker/update-apt.sh
RUN /docker/install-awscli.sh
RUN /docker/install-packer.sh

WORKDIR /stackstate-agent/.ci-builders/windows/builder7
