#!/usr/bin/env bash

set -e
NAMESPACE="mayastor-e2e"

# wait 100s for test-conductor to be running
for i in {0..100}; do
    res=$(kubectl get pod -n ${NAMESPACE} | grep test-conductor | grep 'Running\|Completed\|Error' || true)
    if [ "${res}" != "" ]; then
        break
    fi
    echo waiting
    if [ "$i" == "100" ]; then
        echo "timed out waiting for test-conductor"
        exit 1
    fi
    sleep 1
done

SINCEOPT=""
while [ true ]; do
    kubectl logs -n mayastor-e2e test-conductor --follow ${SINCEOPT}
    # If there is no output for 5 mins, logs --follow will timeout.
    # Found that the command may timeout after 4 hours even with output.
    # If the test-conductor is still running then re-stream the log by repeating the loop.
    # The "--since=1m" should mean we don't lose any output but we may get some duplicated lines.
    res=$(kubectl get pod -n ${NAMESPACE} | grep test-conductor | grep 'Running' || true)
    if [ "${res}" == "" ]; then
        break
    fi
    echo "============== output timeout - restreaming ==============="
    SINCEOPT='--since=1m'
done
res=$(kubectl get pod -n ${NAMESPACE} | grep test-conductor | grep 'Completed' || true)
if [ "${res}" == "" ]; then
    echo "test conductor has not completed successfully"
    exit 1
fi
echo "test conductor has completed"
