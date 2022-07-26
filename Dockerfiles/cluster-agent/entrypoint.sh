#!/bin/bash

# Copyright 2019 StackState
# Unless explicitly stated otherwise all files in this repository are licensed
# under the Apache License Version 2.0.
# This product includes software developed at Datadog (https://www.datadoghq.com/).
# Copyright 2016-present Datadog, Inc.


##### Core config #####

if [[ -z "$STS_API_KEY" ]]; then
    echo "You must set an STS_API_KEY environment variable to run the StackState Cluster Agent container"
    exit 1
fi

##### Copy the custom confs removing any ".." folder in the paths #####
find /conf.d -name '*.yaml' | sed -E "s#/\.\.[^/]+##" | xargs -I{} cp --parents -fv {} /etc/datadog-agent/

##### Starting up #####
export PATH="/opt/stackstate-agent/bin/stackstate-cluster-agent/:/opt/stackstate-agent/embedded/bin/":$PATH

exec "$@"
