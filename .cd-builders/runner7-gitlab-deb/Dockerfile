FROM docker.io/stackstate/stackstate-agent-runner-circle:deb7_latest

RUN apt-get install -y nano
RUN echo 'export PATH=$HOME/bin:$PATH' >> /root/.bashrc

#RUN pip install awscli boto3 sa-rpm-s3
#RUN apt-get install -y libkrb5-dev unixodbc-dev rpm python-deltarpm createrepo yum-utils

#ADD rpmmacros /root/.rpmmacros
SHELL ["/bin/bash", "-c"]
