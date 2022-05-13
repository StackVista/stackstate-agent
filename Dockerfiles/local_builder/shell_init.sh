#!/bin/bash

source /root/.bashrc
conda activate ddpy2
source /go/src/github.com/StackVista/stackstate-agent/.gitlab-scripts/setup_artifactory.sh

# inv -e deps --verbose --dep-vendor-only
#
# inv rtloader.clean && inv rtloader.make --python-runtimes 2 && inv rtloader.test
#
# inv -e agent.omnibus-build --gem-path $CI_PROJECT_DIR/.gems --base-dir $OMNIBUS_BASE_DIR --skip-deps --skip-sign --major-version 2 --python-runtimes 2
# inv -e agent.omnibus-build --skip-deps --skip-sign --major-version 2 --python-runtimes 2
