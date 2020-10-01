#!/usr/bin/env bash

set -e

export STACKSTATE_BRANCH=${STACKSTATE_BRANCH:-master}

if [[ -z $CI_COMMIT_REF_NAME ]]; then
  export AGENT_CURRENT_BRANCH=`git rev-parse --abbrev-ref HEAD`
else
  export AGENT_CURRENT_BRANCH=$CI_COMMIT_REF_NAME
fi


conda create -n molecule python=3.6.12 -y || true
conda activate molecule

pip install -r molecule-role/requirements-molecule3.txt

cd molecule-role

echo =====MOLECULE_RUN_ID=${CI_JOB_ID}======AGENT_CURRENT_BRANCH=${CI_COMMIT_REF_NAME}=======

molecule "$@"
