#!/bin/bash

TARGET_BUCKET=$1

CODENAME=${2:-$CI_COMMIT_REF_NAME}
TARGET_CODENAME=${CODENAME:-dirty}


if [ -z ${TARGET_BUCKET+x} ]; then
	echo "Missing S3 bucket parameter"
	exit 1;
fi

if [ -z ${STACKSTATE_AGENT_VERSION+x} ]; then
	# Pick the latest tag by default for our version.
	STACKSTATE_AGENT_VERSION=$(cat $CI_PROJECT_DIR/version.txt)
fi
echo $STACKSTATE_AGENT_VERSION

rpm-s3 -b $TARGET_BUCKET -p "${TARGET_CODENAME}" $CI_PROJECT_DIR/outcomes/pkg/*.rpm
