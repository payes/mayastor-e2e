#!/usr/bin/env bash

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

LIMIT=10080
for ((i=0; i<=${LIMIT}; i++)); do
	kubectl get pod -n mayastor-e2e test-conductor | grep 'Completed\|Error'
	if [ "$?" == "0" ]; then
		exit 0
	fi
	        if [ "$i" == "${LIMIT}" ]; then
		echo "timed out"
                exit 1
        fi
	sleep 60
	echo "waiting"
done


