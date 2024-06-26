#!/usr/bin/env bash

if [ -z "$1" ] || [ "$1" == "help" ]; then
    echo "Example Charts of how this is used on the Gitlab CI: https://miro.com/app/board/o9J_lzUC0FM=/"
    echo ""
    echo "WARNING: If you create any instance from you local machine please delete it seeing that Lambda does not clean dev instances thus the EC2 costs will increase the longer that instances stays up"
    echo ""
    echo "    First step is to create the EC2 machine"
    echo "    -  ./molecule3.sh <scenario> create"
    echo ""
    echo "    After that we copy over all the required files, install updates and deps, cache images etc."
    echo "    - ./molecule3.sh <scenario> prepare"
    echo ""
    echo "    Now you can either login into your machine with SSH or"
    echo "    -  ./molecule3.sh <scenario> login"
    echo ""
    echo "    Run the docker-compose and the unit tests (Note that everytime you run this a docker-compose cleanup is also ran to cleanup your prev run)"
    echo "    -  ./molecule3.sh <scenario> test"
    echo ""
    echo "    Destroy the EC2 machine and Keypair you created"
    echo "    -  ./molecule3.sh <scenario> destroy"
    echo ""
    echo "Available scenarios"
    echo "- compose"
    echo "- integrations"
    echo "- kubernetes"
    echo "- localinstall"
    echo "- secrets"
    echo "- swarm"
    echo "- vms"
    exit 1
fi

export CONDA_BASE="${HOME}/miniconda3"

# see if conda is available -- when running locally and use the conda base path
if [ -x "$(command -v conda)" ]; then
    CONDA_BASE=$(conda info --base)
fi

source $CONDA_BASE/etc/profile.d/conda.sh
conda env list | grep 'molecule' &> /dev/null
if [ $? != 0 ]; then
   conda create -n molecule python=3.9.7 -y || true
fi

set -e

export STACKSTATE_BRANCH=${STACKSTATE_BRANCH:-master}
export MAJOR_VERSION=${MAJOR_VERSION:-3}
export STS_AWS_TEST_BUCKET=${STS_AWS_TEST_BUCKET:-stackstate-agent-2-test}
export STS_DOCKER_TEST_REPO=${STS_DOCKER_TEST_REPO:-stackstate-agent-2-test}
export STS_DOCKER_TEST_REPO_CLUSTER=${STS_DOCKER_TEST_REPO_CLUSTER:-stackstate-cluster-agent-test}
export LC_ALL=en_US.utf-8
export LANG=en_US.utf-8

if [[ -z $CI_COMMIT_REF_NAME ]]; then
  export AGENT_CURRENT_BRANCH=`git rev-parse --abbrev-ref HEAD`
else
  export AGENT_CURRENT_BRANCH=$CI_COMMIT_REF_NAME
fi

conda activate molecule

pip3 install -r molecule-role/requirements-molecule3.txt

# reads env file to file variables for molecule jobs locally
if test -f "./.envrc"; then
    source "./.envrc"
elif test -f "./.env"; then
    source "./.env"
else
    echo "No env variables file found, Create either a .envrc or .env file to auto load env variables if required"
fi

cd molecule-role

# Allows the yaml to be tested before spinning up and instance
yamllint -c .yamllint .

echo "MOLECULE_RUN_ID=${CI_JOB_ID:-unknown}"
echo "AGENT_CURRENT_BRANCH=${AGENT_CURRENT_BRANCH}"

# TODO: Remove if kubernetes works
if [[ $1 == "--bypass" ]]; then
    echo ""
    echo "------------ Bypass supplied running molecule command directly with parameters --------------"
    echo ""
    echo "Running Molecule Command: molecule ${*:2}"
    molecule "${*:2}"
    exit 0
fi

# Helper for the first parameter defined
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

# Helper for the second parameter defined
AVAILABLE_MOLECULE_PROCESS=("lint" "create" "prepare" "test" "destroy" "login" "unit-test")
if [[ ! " ${AVAILABLE_MOLECULE_PROCESS[*]} " =~ $2 ]]; then
    echo ""
    echo "------------ Invalid Molecule Process Supplied ('$2') --------------"
    echo ""
    echo "Available Molecule Processes:"
    echo "  - lint"
    echo "  - create"
    echo "  - prepare"
    echo "  - test"
    echo "  - unit-test"
    echo "  - destroy"
    echo "---------------------------------------------------------"
    echo ""
    exit 1
fi

# Determine if you are running the script from a CI that contains a commit sha or from localhost
# These variables will be used within the build process to determine if encryption should be applied and where things are copied to
export DEV_MODE="false"

if [ -z "$CI_COMMIT_SHA" ]; then
    export DEV_MODE="true" && echo "DEV_MODE: $DEV_MODE"

    echo "------------------------ DEV MODE WARNING --------------------------------"
    echo "------------ INSTANCES CREATED WITH DEV MODE WILL NOT BE DESTROYED ----------"
    echo "------ MEANING IF YOU LEAVE THIS ZOMBIE INSTANCE UP IT WILL COST PER HOUR ----"
    echo "-----------------------------------------------------------------------------"

    sleep 5
fi

remove_molecule_cache_folder()
{
    MOLECULE_CACHE_PATH="$HOME/.cache/molecule/molecule-role/$1"
    if [ -d "$MOLECULE_CACHE_PATH" ]; then
        rm -rf "$MOLECULE_CACHE_PATH";
    fi
}

execute_molecule()
{
    all_args=("$@")
    molecule --base-config "./molecule/$1/provisioner.$2.yml" "$3" --scenario-name "$1" "${all_args[@]:3}"
}

if [[ $2 == "lint" ]]; then
    execute_molecule "$1" setup lint

elif [[ $2 == "create" ]]; then
    execute_molecule "$1" setup create

elif [[ $2 == "prepare" ]]; then
    remove_molecule_cache_folder "$1"
    execute_molecule "$1" run create

elif [[ $2 == "test" ]]; then
    remove_molecule_cache_folder "$1"
    execute_molecule "$1" run test

elif [[ $2 == "unit-test" ]]; then
    if [ $DEV_MODE == 'true' ]; then
        execute_molecule "$1" run test
    else
        echo "Molecule 'unit-test' can only be ran in dev mode, it's recommended to rather run the full 'test' command instead of 'unit-test'"
        exit 1
    fi

elif [[ $2 == "login" ]]; then
    # Login is used on dev only, thus we restore the .cache file that contains the ssh key
    # This allows the dev to connect back into the server
    # For some reason the ssh key is deleted after the prepare phase is done thus why we have a backup and
    # we can not setup a custom scenario for login thus we have to restore the key inside the sh script
    cp ".cache/molecule/molecule-role/$1/ssh_key" "$HOME/.cache/molecule/molecule-role/$1/ssh_key"
    chmod 600 "$HOME/.cache/molecule/molecule-role/$1/ssh_key"

    if [ -z "$3" ]; then
        execute_molecule "$1" run login
    else
        execute_molecule "$1" run login -h "$3"
    fi

elif [[ $2 == "destroy" ]]; then
    execute_molecule "$1" setup destroy
fi

