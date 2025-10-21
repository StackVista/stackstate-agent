FROM docker.io/stackstate/stackstate-agent-runner-gitlab:latest

RUN pip install awscli==1.16 boto3==1.8.0 sts-rpm-s3==0.3.3
#RUN apt-get install -y libkrb5-dev unixodbc-dev rpm python-deltarpm createrepo yum-utils

ADD rpmmacros /root/.rpmmacros
