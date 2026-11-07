FROM docker.io/python:3.11.4-slim-buster@sha256:c46b0ae5728c2247b99903098ade3176a58e274d9c7d2efeaaab3e0621a53935

RUN pip install boto3==1.8.0 sts-rpm-s3==0.3.3 invoke==1.7.0 distro==1.4.0 awscli==1.16
RUN apt update  && apt-get install -y libkrb5-dev unixodbc-dev rpm python-deltarpm createrepo yum-utils gnupg2

ADD rpmmacros /root/.rpmmacros

