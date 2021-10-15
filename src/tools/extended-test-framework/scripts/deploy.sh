#!/usr/bin/env bash

set -e

TESTARG=""
OPERATION=""
PLANARG=""
DURATIONARG=""

help() {
  cat <<EOF
Usage: $0 [OPTIONS]
or:    $0 --remove

Options:
  --test <name>          test_conductor test to run, steady_state, non_steady_state or non_steady_state_multi_vols
  --plan <test plan ID>  specify the test plan to receive the test runs
  --duration <duration>  set the overal run time for the test with units, e.g. 12d, 34h, 56m27s etc
or
  --remove               remove instead of deploy
Examples:
  $0 --test steady_state --plan AB-123 --duration 6d12h
  $0 --remove
EOF
}

# Parse arguments
while [ "$#" -gt 0 ]; do
  case "$1" in

    -t|--test)
      shift
      case $1 in
            steady_state|non_steady_state|non_steady_state_multi_vols)
                TESTARG=$1
                ;;
            *)
                echo "unrecognized test"
                help
                exit 1
		;;
      esac
      ;;
    -d|--duration)
      shift
      DURATIONARG=$1
      ;;
    -p|--plan)
      shift
      PLANARG=$1
      ;;
    -r|--remove)
      OPERATION="delete"
      set +e # we can ignore errors when undeploying
      ;;
    -h)
      help
      exit 0
      ;;
    *)
      echo "unrecognized parameter"
      help
      exit 1
      ;;
  esac
  shift
done

if [ -z ${OPERATION} ]; then
  if [ -z ${TESTARG} ]; then
    echo "undefined test"
    help
    exit 1
  fi
  if [ -z ${PLANARG} ]; then
    echo "undefined plan"
    help
    exit 1
  fi
  if [ -z ${DURATIONARG} ]; then
    echo "undefined duration"
    help
    exit 1
  fi
fi

if [ "${TESTARG}" == "non_steady_state_multi_vols" ]; then
	IMAGEARG="non_steady_state"
else
	IMAGEARG=${TESTARG}
fi

SCRIPTDIR=$(dirname "$(realpath "$0")")
DEPLOYDIR="${SCRIPTDIR}/../deploy/"
if [ "${OPERATION}" == "delete" ]; then
  kubectl delete configmap tc-config -n mayastor-e2e
  kubectl delete -f ${DEPLOYDIR}/workload_monitor/workload_monitor.yaml
  kubectl delete configmap td-config -n mayastor-e2e
  kubectl delete -f ${DEPLOYDIR}/test_director/test_director.yaml
  kubectl delete pod -n mayastor-e2e test-conductor
  kubectl delete -f ${DEPLOYDIR}/test_conductor/test_conductor.yaml
  kubectl delete -f ${DEPLOYDIR}/test_namespace.yaml
else
  kubectl create -f ${DEPLOYDIR}/test_namespace.yaml

  tmpfile=$(mktemp /tmp/tmp.XXXX)
  PLAN=${PLANARG} envsubst < ${DEPLOYDIR}/test_director/config.yaml.template > $tmpfile
  kubectl create configmap td-config -n mayastor-e2e --from-file=config-local.yaml=${tmpfile}
  rm ${tmpfile}

  kubectl create -f ${DEPLOYDIR}/test_director/test_director.yaml

  kubectl create configmap tc-config -n mayastor-e2e --from-file=${DEPLOYDIR}/test_conductor/${TESTARG}/config.yaml

  kubectl create -f ${DEPLOYDIR}/test_conductor/test_conductor.yaml
  TEST=${IMAGEARG} DURATION=${DURATIONARG} envsubst -no-unset < ${DEPLOYDIR}/test_conductor/test_conductor_pod.yaml.template | kubectl apply -f -

  kubectl create -f ${DEPLOYDIR}/workload_monitor/workload_monitor.yaml
fi

