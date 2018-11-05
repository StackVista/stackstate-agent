#!/bin/sh

CURRENT_BRANCH=${CI_COMMIT_REF_NAME:-dirty}

if [ -z ${STS_AWS_BUCKET+x} ]; then
	echo "Missing AGENT_S3_BUCKET in environment"
	exit 1;
fi

if [ -z ${STACKSTATE_AGENT_VERSION+x} ]; then
	# Pick the latest tag by default for our version.
	STACKSTATE_AGENT_VERSION=$(./version.sh)
	# But we will be building from the master branch in this case.
fi
echo $PROCESS_AGENT_VERSION
FILENAME="process-agent-amd64-$STACKSTATE_AGENT_VERSION"
WORKSPACE=${WORKSPACE:-$PWD/../}
agent_path="$WORKSPACE"

deb-s3 upload --codename ${CURRENT_BRANCH:-dirty} --bucket ${STS_AWS_BUCKET:-stackstate-agent-test} $CI_PROJECT_DIR/outcomes/pkg/*.deb

