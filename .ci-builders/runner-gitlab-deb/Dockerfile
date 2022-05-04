FROM docker.io/stackstate/stackstate-agent-runner-gitlab:latest

RUN pip install awscli boto3 sts-rpm-s3
#RUN apt-get install -y libkrb5-dev unixodbc-dev rpm python-deltarpm createrepo yum-utils

ADD rpmmacros /root/.rpmmacros
