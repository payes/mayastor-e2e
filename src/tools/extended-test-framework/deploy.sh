#!/usr/bin/env bash

set -e

TEST=""
OPERATION=""

help() {
  cat <<EOF
Usage: $0 [OPTIONS]

Options:
  --test <name> test_conductor test to run, steady_state or replica_perturbation
  --remove      remove instead of deploy
Examples:
  $0 --test steady_state
  $0 --remove
EOF
}

# Parse arguments
while [ "$#" -gt 0 ]; do
  case "$1" in

    -t|--test)
      shift
      case $1 in
            steady_state|non_steady_state)
                TEST=$1
                ;;
            *)
                echo "unrecognized test"
                help
                exit 1
        esac
      test=$1
      ;;
    -r|--remove)
      OPERATION="delete"
      set +e # we can ignore errors when undeploying
      ;;
    -h)
      help
      exit 0
      ;;
  esac
  shift
done

if [ -z ${TEST} ] && [ -z ${OPERATION} ] ; then
  echo "undefined test"
  help
  exit 1
fi

SCRIPTDIR=$(dirname "$(realpath "$0")")

if [ "${OPERATION}" == "delete" ]; then
  kubectl delete configmap etfw-config -n mayastor-e2e
  kubectl delete -f ${SCRIPTDIR}/deploy/workload_monitor.yaml
  kubectl delete -f ${SCRIPTDIR}/deploy/test_director.yaml
  kubectl delete pod -n mayastor-e2e test-conductor
  kubectl delete -f ${SCRIPTDIR}/deploy/test_conductor.yaml
  kubectl delete -f ${SCRIPTDIR}/deploy/test_namespace.yaml
else
  kubectl create -f ${SCRIPTDIR}/deploy/test_namespace.yaml
  kubectl create configmap etfw-config -n mayastor-e2e --from-file=${SCRIPTDIR}/deploy/test_conductor/${TEST}/config.yaml
  kubectl create -f ${SCRIPTDIR}/deploy/test_conductor.yaml
  kubectl create -f ${SCRIPTDIR}/deploy/test_conductor/${TEST}/test_conductor_pod.yaml
  kubectl create -f ${SCRIPTDIR}/deploy/test_director.yaml
  kubectl create -f ${SCRIPTDIR}/deploy/workload_monitor.yaml
fi

