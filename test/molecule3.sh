#!/usr/bin/env bash

# Determine if you are running the script from a CI that contains a commit sha or from localhost
# These variables will be used within the build process to determine if encryption should be applied and where things are copied to
export DEV_MODE="false"
export DEV_PATH=""
if [ -z "$CI_COMMIT_SHA" ]; then
    echo ""
    echo "------------ DEV MODE ENABLED --------------"
    export DEV_MODE="true"
    export DEV_PATH="${PWD}/.."
    echo "DEV_MODE: ${DEV_MODE}"
    echo "DEV_PATH: ${DEV_PATH}"
    echo "---------------------------------------------------------"
    echo ""
fi



export CONDA_BASE="${HOME}/miniconda3"

# see if conda is available -- when running locally and use the conda base path
if [ -x "$(command -v conda)" ]; then
    CONDA_BASE=$(conda info --base)
fi

source $CONDA_BASE/etc/profile.d/conda.sh
conda env list | grep 'molecule' &> /dev/null
if [ $? != 0 ]; then
   conda create -n molecule python=3.6.12 -y || true
fi

set -e

export STACKSTATE_BRANCH=${STACKSTATE_BRANCH:-master}

export MAJOR_VERSION=${MAJOR_VERSION:-3}
export STS_AWS_TEST_BUCKET=${STS_AWS_TEST_BUCKET:-stackstate-agent-3-test}
export STS_DOCKER_TEST_REPO=${STS_DOCKER_TEST_REPO:-stackstate-agent-test}
export STS_DOCKER_TEST_REPO_CLUSTER=${STS_DOCKER_TEST_REPO_CLUSTER:-stackstate-cluster-agent-test}

if [[ -z $CI_COMMIT_REF_NAME ]]; then
  export AGENT_CURRENT_BRANCH=`git rev-parse --abbrev-ref HEAD`
else
  export AGENT_CURRENT_BRANCH=$CI_COMMIT_REF_NAME
fi

conda activate molecule

pip3 install -r molecule-role/requirements-molecule3.txt

# reads env file to file variables for molecule jobs locally
ENV_FILE=./.env
if test -f "$ENV_FILE"; then
    echo ""
    echo "------------ Sourcing env file with contents ------------"
    echo "$(cat $ENV_FILE)"
    echo "---------------------------------------------------------"
    echo ""
    source $ENV_FILE
fi

cd molecule-role

# Allows the yaml to be tested before spinning up and instance
yamllint -c .yamllint .

echo "MOLECULE_RUN_ID=${CI_JOB_ID:-unknown}"
echo "AGENT_CURRENT_BRANCH=${AGENT_CURRENT_BRANCH}"

if [[ $1 == "--bypass" ]]; then
    echo ""
    echo "------------ Bypass supplied running molecule command directly with parameters --------------"
    echo ""
    echo "Running Molecule Command: molecule ${*:2}"
    molecule "${*:2}"
    exit 0
fi

AVAILABLE_MOLECULE_SCENARIOS=("compose" "integrations" "kubernetes" "localinstall" "secrets" "swarm" "vms")
if [[ ! " ${AVAILABLE_MOLECULE_SCENARIOS[*]} " =~ $1 ]]; then
    echo ""
    echo "------------ Invalid Molecule Scenario Supplied ('$1') --------------"
    echo ""
    echo "Available Molecule Scenarios:"
    for i in "${AVAILABLE_MOLECULE_SCENARIOS[@]}"
    do
        echo "  - $i"
    done
    echo "---------------------------------------------------------"
    echo ""
    exit 1
fi

CONFIG_TARGET="setup"
EXTRA_PARAMETERS=""
AVAILABLE_MOLECULE_PROCESS=("create" "prepare" "test" "destroy" "login")
if [[ ! " ${AVAILABLE_MOLECULE_PROCESS[*]} " =~ $2 ]]; then
    echo ""
    echo "------------ Invalid Molecule Process Supplied ('$2') --------------"
    echo ""
    echo "Available Molecule Processes:"
    echo "  - create"
    echo "      Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur auctor"
    echo "  - prepare"
    echo "      Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur auctor"
    echo "  - test"
    echo "      Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur auctor"
    echo "  - destroy"
    echo "      Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur auctor"
    echo "---------------------------------------------------------"
    echo ""
    exit 1

elif [[ $2 == "prepare" ]]; then
    CONFIG_TARGET="run"
    EXTRA_PARAMETERS="--force"

elif [[ $2 == "test" ]] || [[ $2 == "login" ]]; then
    CONFIG_TARGET="run"
fi

if [[ $DEV_MODE ]]; then
    echo ""
    echo "------------------------ DEV MODE WARNING --------------------------------"
    echo "------------ INSTANCES CREATED WITH DEV MODE WILL NOT BE DESTROYED ----------"
    echo "------ MEANING IF YOU LEAVE THIS ZOMBIE INSTANCE UP IT WILL COST PER HOUR ----"
    echo "-----------------------------------------------------------------------------"
    sleep 10
fi

echo "Running Molecule Command: molecule --base-config ./molecule/$1/provisioner.$CONFIG_TARGET.yml $2 --scenario-name $1 $EXTRA_PARAMETERS"
molecule --base-config "./molecule/$1/provisioner.$CONFIG_TARGET.yml" "$2" --scenario-name "$1" $EXTRA_PARAMETERS
