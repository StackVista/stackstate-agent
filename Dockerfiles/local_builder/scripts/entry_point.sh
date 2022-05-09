#!/bin/bash

# set -euxo pipefail

if [[ "$#" -ne "2" ]]; then
    echo "Expected argument: <agent-clone-source> <branch>"
fi

. /root/miniconda3/etc/profile.d/conda.sh
conda activate ddpy3

# mkdir -p /go/src/github.com/StackVista
# cd /go/src/github.com/StackVista
# git clone "$1" stackstate-agent
#
# cd stackstate-agent
#
# git checkout $2

cd /go/src/github.com/StackVista/stackstate-agent

source ./.gitlab-scripts/setup_artifactory.sh

inv -e deps --verbose --dep-vendor-only

exec bash --init-file Dockerfiles/local_builder/scripts/shell.sh
