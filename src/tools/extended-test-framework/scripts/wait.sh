#!/usr/bin/env bash


# wait 100s for test-conductor to be created
for i in {0..100}; do
    kubectl get pod -n mayastor-e2e | grep test-conductor
	if [ "$?" == "0" ]; then
        break
    fi
	if [ "$i" == "100" ]; then
		exit 1
	fi
	sleep 1
done

# minutes in 15 days. We expect the test_conductor to run for no more that 14 days
LIMIT=$(( 60 * 24 * 15))
echo "waiting for test-conductor to complete"

for ((i=0; i<=${LIMIT}; i++)); do
	kubectl get pod -n mayastor-e2e test-conductor | grep 'Completed\|Error'
	if [ "$?" == "0" ]; then
		break
	fi
    if [ "$i" == "${LIMIT}" ]; then
		echo "timed out"
        exit 1
    fi
	sleep 60
done

echo "test completed"

