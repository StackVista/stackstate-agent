#!/usr/bin/env bash

FILES="50-kubernetes.sh 51-docker.sh 59-defaults.sh 60-network-check.sh 60-sysprobe-check.sh 89-copy-customfiles.sh"

for file in ${FILES}; do
    if [[ -f "etc/cont-init.d/${file}" ]]; then
        "etc/cont-init.d/${file}"
        RC=$?
        if [[ $RC -ne 0 ]]; then
            echo "Error in ${file} with return code ${RC}"
#            exit $RC
        fi
    fi
done

if [[ ! -e /etc/stackstate-agent/stackstate.yaml ]]; then
    echo "Despite our best efforts, we could not find a config file to use."
    echo "Please mount a config file to /etc/stackstate-agent/stackstate.yaml"
    echo "or set the STS_CONFIG_FILE environment variable to a config file."
    echo "Exiting..."
    sleep 20
    exit 1
fi

/opt/stackstate-agent/bin/agent/agent run
