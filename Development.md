# Build and distribute StackState Agent in linux using Docker

Using our builder image clone and checkout the public repo and th <<branch>> you are interested of:
```bash
$ docker run --rm -ti registry.tooling.stackstate.io/quay/stackstate/stackstate-agent-runner-gitlab:deb7_20211210 bash

$ export CI_PROJECT_DIR=/go/src/github.com/StackVista/stackstate-agent && \
  mkdir -p /go/src/github.com/StackVista && \
  cd src/github.com/StackVista && \
  git clone https://github.com/StackVista/stackstate-agent && \
  cd stackstate-agent && \
  git checkout <<branch>>
```

Remember to `git pull` every time you push a change.

### Configure Artifactory

We use some private python libraries for our integrations therefore you need to configure artifactory as pypi repository:
```bash
$ export GITLAB_PACKAGE_REGISTRY_PYPI_SIMPLE_URL=gitlab.com/api/v4/projects/71271774/packages/pypi/simple && \
  export artifactory_user=... && \
  export artifactory_password=... && \
  source ./.gitlab-scripts/setup_artifactory.sh
```

### Build using Python3 interpreter
```bash
$ conda activate ddpy3 && \
  inv deps && \
  inv agent.clean && \
  inv -e agent.omnibus-build --base-dir /omnibus --skip-deps --skip-sign --major-version 3 --python-runtimes 3
```

### Build using Python2 interpreter
```bash
$ conda activate ddpy2 && \
  inv deps && \
  inv agent.clean && \
  inv -e agent.omnibus-build --base-dir /omnibus --skip-deps --skip-sign --major-version 2 --python-runtimes 2
```

## Use local copy

Instead of cloning the repo you could use directly your local one:
```bash
$ docker run --rm -it --name stackstate-agent-builder --mount type=bind,source="${PWD}",target=/root/stackstate-agent,readonly registry.tooling.stackstate.io/quay/stackstate/stackstate-agent-runner-gitlab:deb7_20211210 bash

$ export CI_PROJECT_DIR=/go/src/github.com/StackVista/stackstate-agent && \
  mkdir -p /go/src/github.com/StackVista && \
  cd src/github.com/StackVista

$ cp -r /root/stackstate-agent /go/src/github.com/StackVista
```

Now [configure Artifatory](#configure-artifactory) then build using either [Python2](#build-using-python2-interpreter) or [Python3](#build-using-python3-interpreter).
Remember to copy every time you make a change on your local copy.
