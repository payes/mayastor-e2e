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

for i in {0..1000}; do
	kubectl get pod -n mayastor-e2e test-conductor | grep Completed
	if [ "$?" == "0" ]; then
		exit 0
	fi
	        if [ "$i" == "1000" ]; then
		echo "timed out"
                exit 1
        fi
	sleep 10
	echo "waiting"
done


