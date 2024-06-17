# StackState Agent

Contains the code for the StackState agent V2. Agent integrations are not included in this project and can be found [here](https://github.com/StackVista/stackstate-agent-integrations).

## Installation

Installation instructions are available on the [StackState docs site](https://docs.stackstate.com/stackpacks/integrations/agent).

## Getting development started

To build the Agent you need:
 * [Go](https://golang.org/doc/install) 1.16 or later. You'll also need to set your `$GOPATH` and have `$GOPATH/bin` in your path.
 * Python 3.7+ along with development libraries for tooling. You will also need Python 2.7 if you are building the Agent with Python 2 support.
 * Python dependencies. You may install these with `pip install -r requirements.txt`
   This will also pull in [Invoke](http://www.pyinvoke.org) if not yet installed.
 * CMake version 3.12 or later and a C++ compiler

**Note:** you may want to use a python virtual environment to avoid polluting your
      system-wide python environment with the agent build/dev dependencies. You can
      create a virtual environment using `virtualenv` and then use the `invoke agent.build`
      parameters `--python-home-2=<venv_path>` and/or `--python-home-3=<venv_path>`
      (depending on the python versions you are using) to use the virtual environment's
      interpreter and libraries. By default, this environment is only used for dev dependencies
      listed in `requirements.txt`.

**Note:** You may have previously installed `invoke` via brew on MacOS, or `pip` in
      any other platform. We recommend you use the version pinned in the requirements
      file for a smooth development/build experience.

Builds and tests are orchestrated with `invoke`, type `invoke --list` on a shell
to see the available tasks.

To start working on the Agent, you can build the `master` branch:

1. Checkout the repo: `git clone https://github.com/StackVista/stackstate-agent.git $GOPATH/src/github.com/DataDog/datadog-agent`.
2. cd into the project folder: `cd $GOPATH/src/github.com/StackVista/stackstate-agent`.
3. Install go tools: `invoke install-tools`.
4. Install go dependencies: `invoke deps`.
   Make sure that `$GOPATH/bin` is in your `$PATH` otherwise this step might fail.
5. Create a development `datadog.yaml` configuration file in `dev/dist/datadog.yaml`, containing a valid API key: `api_key: <API_KEY>`
6. Build the agent with `invoke agent.build --build-exclude=systemd`.
7. When editing code in VS Code or in Intellij configure it to use the same tags as are used by the previous invoke command (they are visisble in the output, current set is `consul ec2 process python gce cri zk containerd zlib jmx secrets kubelet kubeapiserver jetson docker etcd apm netcgo orchestrator`). In VS Code this is configurable for the workspace on the Go plugin settings.

    By default, the Agent will be built to use Python 3 but you can select which Python version you want to use:

      - `invoke agent.build --python-runtimes 2` for Python2 only
      - `invoke agent.build --python-runtimes 3` for Python3 only
      - `invoke agent.build --python-runtimes 2,3` for both Python2 and Python3

     You can specify a custom Python location for the agent (useful when using
     virtualenvs):

       invoke agent.build \
         --python-runtimes 2,3 \
         --python-home-2=$GOPATH/src/github.com/StackVista/stackstate-agent/venv2 \
         --python-home-3=$GOPATH/src/github.com/StackVista/stackstate-agent/venv3 .

    Running `invoke agent.build`:

     * Discards any changes done in `bin/agent/dist`.
     * Builds the Agent and writes the binary to `bin/agent/agent`.
     * Copies files from `dev/dist` to `bin/agent/dist`. See `https://github.com/DataDog/datadog-agent/blob/main/dev/dist/README.md` for more information.

     If you built an older version of the agent, you may have the error `make: *** No targets specified and no makefile found.  Stop.`. To solve the issue, you should remove `CMakeCache.txt` from `rtloader` folder with `rm rtloader/CMakeCache.txt`.

## Testing

Run tests using `invoke test`. During development, add the `--skip-linters` option to skip straight to the tests.
```
invoke test --targets=./pkg/aggregator/... --skip-linters
```

When testing code that depends on [rtloader](/rtloader), build and install it first.
```
invoke rtloader.make && invoke rtloader.install
invoke test --targets=./pkg/collector/python --skip-linters
```

## Run

You can run the agent with:
```
./bin/agent/agent run -c bin/agent/dist/stackstate.yaml
```

The file `bin/agent/dist/datadog.yaml` is copied from `dev/dist/datadog.yaml` by `invoke agent.build` and must contain a valid api key.

## Install

### Linux

##### Official

To install the official release:

    $ curl -o- https://stackstate-agent-3.s3.amazonaws.com/install.sh | STS_API_KEY="xxx" STS_URL="yyy" bash
     or
    $ wget -qO- https://stackstate-agent-3.s3.amazonaws.com/install.sh | STS_API_KEY="xxx" STS_URL="yyy" bash

##### Test

If you want to install a branch version use the test repository:

    $ curl -o- https://stackstate-agent-3-test.s3.amazonaws.com/install.sh | STS_API_KEY="xxx" STS_URL="yyy" CODE_NAME="PR_NAME" bash
     or
    $ wget -qO- https://stackstate-agent-3-test.s3.amazonaws.com/install.sh | STS_API_KEY="xxx" STS_URL="yyy" CODE_NAME="PR_NAME" bash

and replace `PR_NAME` with the branch name (e.g. `master`, `STAC-xxxx`).

### Docker

##### Official

    $ docker pull artifactory.tooling.stackstate.io/docker-virtual/stackstate/stackstate-agent-2:latest

##### Test

    $ docker pull artifactory.tooling.stackstate.io/docker-virtual/stackstate/stackstate-agent-2-test:latest

### Windows

##### Official

To install the official release:

    $ . { iwr -useb https://stackstate-agent-3.s3.amazonaws.com/install.ps1 } | iex; install -stsApiKey "xxx" -stsUrl "yyy"

##### Test

If you want to install a branch version use the test repository:

    $ . { iwr -useb https://stackstate-agent-3-test.s3.amazonaws.com/install.ps1 } | iex; install -stsApiKey "xxx" -stsUrl "yyy" -codeName "PR_NAME"

and replace `PR_NAME` with the branch name (e.g. `master`, `STAC-xxxx`).

#### Arguments

Other arguments can be passed to the installation command.

Linux arguments:

- `STS_HOSTNAME` = Instance hostname
- `$HOST_TAGS` = Agent host tags to use for all topology component (by default `os:linux` will be added)
- `SKIP_SSL_VALIDATION` = Skip ssl certificates validation when talking to the backend (defaults to `false`)
- `STS_INSTALL_ONLY` = Agent won't be automatically started after installation

Windows arguments:

- `hostname` = Instance hostname
- `tags` = Agent host tags to use for all topology component (by default `os:windows` will be added)
- `skipSSLValidation` = Skip ssl certificates validation when talking to the backend (defaults to `false`)
- `agentVersion` = Version of the Agent to be installed (defaults to `latest`)

##### Omnibus notes for windows build process

We ended up checking in a patched gem file under omnibus/vendor/cache/libyajl2-1.2.1.gem, to make windows builds work with newer msys toolchain.
The source of this can be found here https://github.com/StackVista/libyajl2-gem/tree/1.2.0-fixed-lssp. Ideally we'd be able to drop this hack once we bump the ruby version > 2.6.5 because libyajl2 compiles proper on those ruby versions.

## GitLab cluster agent pipeline

If you want to speed up the GitLab pipeline and run only the steps related to the cluster agent, include the string `[cluster-agent]` in your commit message.

## Testing cluster-agent helm chart

The acceptance tests in our pipeline use the stackstate-agent helm chart to install the agent in a test cluster. If you make changes to the stackstate-agent helm chart, you probably want to test if our acceptance tests will work after your changes.

When you open a merge request on the helm-chart repository, a test version of that chart will be published to a test helm repository [stackstate-test](https://helm-test.stackstate.io). You can add that test repo in your machine by running the following commands:

```shell
helm repo add stackstate-test https://helm-test.stackstate.io && helm repo update
```

You can then install this version of the cluster-agent helm chart by running:

```shell
helm upgrade --install \
  --create-namespace \
  --namespace <namespace> \
  --set-string 'stackstate.apiKey'='<api-key>' \
  --set-string 'stackstate.cluster.name'='<cluster-name' \
  --set-string 'stackstate.url'='<stackstate-url>' \
  stackstate-agent stackstate-test/stackstate-agent --version <version>
```

`<version>` is the new version you've set on `helm-charts/stable/stackstate-agent/Chart.yaml` on your feature branch.

To use this version in the `stackstate-agent` pipeline, create a branch and update the `AGENT_HELM_CHART_VERSION` variable on `.gitlab-ci-agent.yml`, with that the pipeline will use the test helm repository that was updated by the helm-charts pipeline.

