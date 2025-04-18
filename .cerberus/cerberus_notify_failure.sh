#!/usr/bin/env bash

# This script does a call ot to the Cerberus Lambda running in AWS
# The Cerberus code can be found at: https://gitlab.com/stackvista/devops/cerberus

set -exuo pipefail

PAYLOAD=$(cat << END_OF_PAYLOAD
{
    "action": "notify",
    "context": {
        "project.id": "$CI_PROJECT_ID",
        "project.name": "$CI_PROJECT_TITLE",
        "project.slug": "$CI_PROJECT_PATH",
        "commit.sha": "$CI_COMMIT_SHA",
        "commit.title": "$CI_COMMIT_TITLE",
        "branch": "$CI_COMMIT_REF_NAME",
        "pipeline": "$CI_PIPELINE_ID",
        "suite": "$SUITE",
        "channel": "$SLACK_CI_REPORT_CHANNEL"
    }
}
END_OF_PAYLOAD
)

curl --verbose --fail --data "${PAYLOAD?Payload is empty}" "${CERBERUS_LAMBDA_URL?No URL Provided}"
