#!/usr/bin/env bash

CLIENT:=kubectl
TIMEOUT=60
INTERVAL=5
end=$((SECONDS+TIMEOUT))

ds_name=$($(CLIENT) get ds --no-headers=true -A -l app=mixedcpus-plugin -o custom-columns=NAME:.metadata.name)
desired=$($(CLIENT) get ds --no-headers=true -A -l app=mixedcpus-plugin -o custom-columns=DESIRED:.status.currentNumberScheduled)

while [ $SECONDS -lt $end ]; do
    ready=$($(CLIENT) get ds --no-headers=true -A -l app=mixedcpus-plugin -o custom-columns=DESIRED:.status.numberReady)
    if [[ "${desired}" != "${ready}" ]]; then
        echo "DaemonSet: ${ds_name} not ready; DESIRED=${desired} READY=${ready}"
        sleep ${INTERVAL}
    else
        break
    fi
done

echo "DaemonSet: ${ds_name} is ready"
