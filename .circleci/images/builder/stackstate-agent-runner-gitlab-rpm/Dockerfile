FROM stackstate/stackstate-agent-runner-circle:rpm_base.latest

RUN pip install awscli boto

RUN yum -y install net-snmp-devel python-deltarpm yum-utils libtool rpm-sign

# todo: if works, work rpm-s3
RUN cd /usr/local && \
    git clone https://github.com/crohr/rpm-s3 --recursive && \
    echo '[s3]\ncalling_format = boto.s3.connection.OrdinaryCallingFormat' > /etc/boto.cfg
    # Use HTTP for debugging traffic to S3.
    #echo '[Boto]\nis_secure = False' >> /etc/boto.cfg
ENV PATH /usr/local/rpm-s3/bin:$PATH
ADD rpmmacros /root/.rpmmacros
