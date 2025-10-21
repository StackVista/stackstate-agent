FROM docker.io/stackstate/stackstate-agent-runner-gitlab:centos6

RUN pip install awscli==1.16 boto3==1.8.0 sts-rpm-s3==0.3.3

RUN yum -y install net-snmp-devel python-deltarpm yum-utils libtool rpm-sign createrepo

ADD rpmmacros /root/.rpmmacros
