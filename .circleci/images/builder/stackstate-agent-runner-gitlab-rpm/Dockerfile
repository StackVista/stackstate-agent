FROM stackstate/stackstate-agent-runner-circle:rpm_base.latest

RUN yum install -y python-devel.x86_64

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

ENV PATH=/go/src/github.com/bin:$PATH

ENV GOPATH /go
