#!/bin/bash

source /root/.bashrc
conda activate ddpy3
source ./.gitlab-scripts/setup_artifactory.sh

# inv -e deps --verbose --dep-vendor-only
#
# inv rtloader.clean && inv rtloader.make --python-runtimes 3 && inv rtloader.test
