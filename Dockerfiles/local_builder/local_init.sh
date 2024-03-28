#!/bin/bash

## This script will activate conda and rvm
## if the source mount env var is valued then we work on a copy of the Agent source code

source /root/.bashrc
conda activate ddpy${PYTHON_RUNTIME}

if [[ ! -z "${AGENT_SOURCE_MOUNT}" ]]; then
    echo "Agent source mount provided: ${AGENT_SOURCE_MOUNT}"

    if [[ -d ${AGENT_SOURCE_MOUNT} ]]; then

        pidof inotifywait
        if [ $? -ne 0 ]; then
            echo "inotifywait not running"

            echo -e "\nCopying ..."
            mkdir -p ${PROJECT_DIR}
            rsync -a ${AGENT_SOURCE_MOUNT}/ ${PROJECT_DIR}

            echo -e "\n--> Open a new shell with 'make shell', so that this terminal will keep syncing changes\n"

            while inotifywait -r -e modify,create,delete,move ${AGENT_SOURCE_MOUNT}; do
                rsync -av ${AGENT_SOURCE_MOUNT}/ ${PROJECT_DIR}
            done
        else
            echo "inotifywait already running"
        fi

    else
        echo "Mount directory does not exist!"
        exit 1
    fi
fi

cd ${PROJECT_DIR}
source .gitlab-scripts/setup_artifactory.sh

# List of rtloader commands
COMMAND__RTLOADER_CLEAN="inv rtloader.clean"
COMMAND__RTLOADER_MAKE="inv rtloader.make --python-runtimes $PYTHON_RUNTIME"
COMMAND__RTLOADER_TEST="inv rtloader.test"

# List of agent commands
COMMAND__AGENT_BUILD="inv -e agent.build --major-version 2 --python-runtimes $PYTHON_RUNTIME"
COMMAND__AGENT_CLEAN="inv agent.clean"
COMMAND__AGENT_OMNIBUS_BUILD="inv -e agent.omnibus-build --base-dir ~/.omnibus --skip-deps --skip-sign --major-version 2 --python-runtimes $PYTHON_RUNTIME"

# List of go commands
COMMAND__INSTALL_DEPS="inv -e deps --verbose --dep-vendor-only"
COMMAND__GO_TESTS_WITH_RACE="inv -e test --targets=${2:-.} --coverage --race --profile --fail-on-fmt --cpus 4 --major-version 2 --python-runtimes $PYTHON_RUNTIME --skip-linters"

# List of python commands
COMMAND__SET_PYTHON_2="export PYTHON_RUNTIME=2"
COMMAND__SET_PYTHON_3="export PYTHON_RUNTIME=3"
COMMAND__ACTIVATE_DDPY_ENV="conda activate ddpy\$PYTHON_RUNTIME"

# Determine if a command should be execute or if it is general shell access
case $1 in
  install-deps)
    $COMMAND__INSTALL_DEPS
    ;;

  rtloader-build-test)
    $COMMAND__RTLOADER_CLEAN
    $COMMAND__RTLOADER_MAKE
    $COMMAND__RTLOADER_TEST
    ;;

  agent-build)
    $COMMAND__AGENT_BUILD
    ;;

  agent-clean)
    $COMMAND__AGENT_CLEAN
    ;;

  agent-omnibus-build)
    $COMMAND__AGENT_OMNIBUS_BUILD
    ;;

  go-tests-with-race)
    $COMMAND__GO_TESTS_WITH_RACE
    ;;

  set-python-2)
    $COMMAND__SET_PYTHON_2
    conda activate ddpy2
    ;;

  set-python-3)
    $COMMAND__SET_PYTHON_3
    conda activate ddpy3
    ;;

  *)
    cat << EOF

    ---------------------------------------------------------------------------------------
    Here few helpful commands to get you started (check .gitlab-ci-agent.yml for more):
      # When starting the first time you always need to pull deps
      $COMMAND__INSTALL_DEPS

      # Build and test the rtloader
      $COMMAND__RTLOADER_CLEAN && $COMMAND__RTLOADER_MAKE && $COMMAND__RTLOADER_TEST

      # Build the agent binary
      $COMMAND__AGENT_BUILD

      # Build the agent omnibus package
      $COMMAND__AGENT_OMNIBUS_BUILD

      # Run the go tests with race conditions
      $COMMAND__GO_TESTS_WITH_RACE

      # Clean temporary objects and binary artifacts
      $COMMAND__AGENT_CLEAN

      # Switch to python 2
      $COMMAND__SET_PYTHON_2 && $COMMAND__ACTIVATE_DDPY_ENV

      # Switch to python 3
      $COMMAND__SET_PYTHON_3 && $COMMAND__ACTIVATE_DDPY_ENV

    ---------------------------------------------------------------------------------------

EOF
    ;;
esac



