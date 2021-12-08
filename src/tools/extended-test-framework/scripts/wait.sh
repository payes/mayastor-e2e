#!/usr/bin/env bash


# wait 100s for test-conductor to be running
for i in {0..100}; do
    kubectl get pod -n mayastor-e2e | grep test-conductor | grep "Running"
	if [ "$?" == "0" ]; then
        break
    fi
	if [ "$i" == "100" ]; then
		exit 1
	fi
	sleep 1
done

set -e

timeout 15d kubectl logs -n mayastor-e2e test-conductor --follow

echo "test completed"

