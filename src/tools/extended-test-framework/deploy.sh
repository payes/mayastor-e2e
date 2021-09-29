#!/usr/bin/env bash

set -e

TEST=""

help() {
  cat <<EOF
Usage: $0 [OPTIONS]

Options:
  --test <name>             test_conductor test to run, steady_state or replica_perturbation
Examples:
  $0 --test steady_state
EOF
}

# Parse arguments
while [ "$#" -gt 0 ]; do
  case "$1" in

    -t|--test)
      shift
      case $1 in
            steady_state|replica_perturbation)
                TEST=$1
                ;;
            *)
                echo "unrecognized test"
                help
                exit 1
        esac
      test=$1
      ;;
  esac
  shift
done

if [ -z ${TEST} ]; then
  echo "undefined test"
  help
  exit 1
fi

SCRIPTDIR=$(dirname "$(realpath "$0")")
kubectl create -f ${SCRIPTDIR}/test_namespace.yaml
kubectl create configmap etfw-config -n mayastor-e2e --from-file=${SCRIPTDIR}/test_conductor/cmd/test_conductor_${TEST}/config.yaml
kubectl create -f ${SCRIPTDIR}/test_conductor/cmd/test_conductor_${TEST}/test_conductor.yaml
kubectl create -f ${SCRIPTDIR}/test_director.yaml
kubectl create -f ${SCRIPTDIR}/workload_monitor.yaml

