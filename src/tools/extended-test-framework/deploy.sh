#!/usr/bin/env bash

set -e

SCRIPTDIR=$(dirname "$(realpath "$0")")
kubectl create -f ${SCRIPTDIR}/test_namespace.yaml
kubectl create configmap etfw-config -n mayastor-e2e --from-file=${SCRIPTDIR}/test_conductor/config.yaml
kubectl create -f ${SCRIPTDIR}/test_conductor.yaml
kubectl create -f ${SCRIPTDIR}/test_director.yaml
kubectl create -f ${SCRIPTDIR}/workload_monitor.yaml

