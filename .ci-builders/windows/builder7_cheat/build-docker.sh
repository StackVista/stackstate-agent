#!/usr/bin/env bash

if [ -z "${AWS_PROFILE}" ]; then
    echo "No AWS_PROFILE set for env, quitting..."
    exit 1
fi

if [ ! -f "VCForPython27.msi" ]; then
    aws s3 cp s3://vcpython27/VCForPython27.msi .
fi

docker build -t wolverminion/packer-runner:0.0.6 .
