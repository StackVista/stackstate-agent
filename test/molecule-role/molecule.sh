#!/bin/sh

set -e

if [ -z "$CI_COMMIT_REF_NAME" ]; then
  export AGENT_GITLAB_BRANCH=`git rev-parse --abbrev-ref HEAD`
else
  export AGENT_GITLAB_BRANCH=$CI_COMMIT_REF_NAME
fi

docker run --rm \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /tmp:/tmp \
    -v "$(pwd)":/tmp/$(basename "${PWD}") \
    -w /tmp/$(basename "${PWD}") \
    -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -e AWS_REGION -e USER -e AGENT_GITLAB_BRANCH -e MOLECULE_RUN_ID \
    quay.io/ansible/molecule:2.22rc3 \
    molecule "$@"
#    molecule --debug "$@"
