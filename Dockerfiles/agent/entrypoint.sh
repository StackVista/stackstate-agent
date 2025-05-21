#!/usr/bin/env bash
for file in $(ls etc/cont-init.d/ | grep -v "50-ci.sh"); do
    if [[ -f "etc/cont-init.d/${file}" ]]; then
        "etc/cont-init.d/${file}"
        RC=$?
        if [[ $RC -ne 0 ]]; then
            echo "Error in ${file} with return code ${RC}"
            exit $RC
        fi
    fi
done

/opt/stackstate-agent/bin/agent/agent run
