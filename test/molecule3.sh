#!/usr/bin/env bash

set -e

source $HOME/miniconda3/etc/profile.d/conda.sh

export STACKSTATE_BRANCH=${STACKSTATE_BRANCH:-master}

if [[ -z $CI_COMMIT_REF_NAME ]]; then
  export AGENT_CURRENT_BRANCH=`git rev-parse --abbrev-ref HEAD`
else
  export AGENT_CURRENT_BRANCH=$CI_COMMIT_REF_NAME
fi

conda env list | grep 'molecule' &> /dev/null
if [ $? != 0 ]; then
   conda create -n molecule python=3.6.12 -y || true
fi

conda activate molecule

pip install -r molecule-role/requirements-molecule3.txt

cd molecule-role

echo =====MOLECULE_RUN_ID=${CI_JOB_ID}======AGENT_CURRENT_BRANCH=${CI_COMMIT_REF_NAME}=======

molecule "$@"
