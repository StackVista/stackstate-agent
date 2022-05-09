#!/bin/bash

source ~/.bashrc
conda activate ddpy3
cd /go/src/github.com/StackVista/stackstate-agent

inv -e deps --verbose --dep-vendor-only

inv rtloader.clean && inv rtloader.make --python-runtimes 3 && inv rtloader.test


source ./.gitlab-scripts/setup_artifactory.sh
