#!/bin/bash

if [[ -z "${ECS_FARGATE}" ]]; then
    exit 0
fi

# Set a default config for ECS Fargate if found
# Don't override /etc/stackstate-agent/stackstate.yaml if it exists
if [[ ! -e /etc/stackstate-agent/stackstate.yaml ]]; then
    ln -s  /etc/stackstate-agent/stackstate-ecs.yaml \
           /etc/stackstate-agent/stackstate.yaml
fi

# Remove all default checks, AD will automatically enable fargate check
find /etc/stackstate-agent/conf.d/ -iname "*.yaml.default" | xargs grep -L 'ad_identifiers' | xargs rm -f
