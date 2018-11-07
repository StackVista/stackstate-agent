#!/bin/sh

CODENAME=${1-:$CI_COMMIT_REF_NAME}
TARGET_CODENAME=${CODENAME:-dirty}

TARGET_BUCKET=${2:-STS_AWS_BUCKET}

if [ -z ${STS_AWS_BUCKET+x} ]; then
	echo "Missing AGENT_S3_BUCKET in environment"
	exit 1;
fi

if [ -z ${STACKSTATE_AGENT_VERSION+x} ]; then
	# Pick the latest tag by default for our version.
	STACKSTATE_AGENT_VERSION=$(./version.sh)
	# But we will be building from the master branch in this case.
fi
echo $STACKSTATE_AGENT_VERSION

ls $CI_PROJECT_DIR/outcomes/pkg/*.*

deb-s3 upload --sign=${SIGNING_KEY_ID} --codename ${TARGET_CODENAME} --bucket ${TARGET_BUCKET:-stackstate-agent-test} $CI_PROJECT_DIR/outcomes/pkg/*.deb

