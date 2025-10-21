#!/usr/bin/env bash

file="01-check-apikey.sh"
"etc/cont-init.d/${file}"
RC=$?
if [[ $RC -ne 0 ]]; then
    echo ""
    echo "=================================================================================="
    echo "You must set an STS_API_KEY environment variable to run the StackState Agent container"
    echo "=================================================================================="
    echo ""
    sleep 20
    exit $RC
fi

# These files would have been executed by systemd in the order of the implied priority of the number in the filename.
FILES="50-kubernetes.sh 51-docker.sh 59-defaults.sh 60-network-check.sh 60-sysprobe-check.sh 89-copy-customfiles.sh"

# But since we are not using systemd, we will execute them in the order they are listed here.
for file in ${FILES}; do
    if [[ -f "etc/cont-init.d/${file}" ]]; then
        "etc/cont-init.d/${file}"
        RC=$?
        if [[ $RC -ne 0 ]]; then
            echo "Error in ${file} with return code ${RC}"
#            exit $RC
            # We log any errors for debugging purposes, but do not exit.
            # The next if-block will handle the case where no config file is found.
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
