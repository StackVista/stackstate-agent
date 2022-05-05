FROM docker.io/python:3.10.4-slim-buster

RUN pip install boto3 sts-rpm-s3==0.3.3 invoke distro==1.4.0 awscli==1.16
RUN apt update  && apt-get install -y libkrb5-dev unixodbc-dev rpm python-deltarpm createrepo yum-utils gnupg2

ADD rpmmacros /root/.rpmmacros

