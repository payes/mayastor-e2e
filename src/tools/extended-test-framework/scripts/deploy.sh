#!/usr/bin/env bash

set -e

TESTARG=""
OPERATION=""
PLANARG=""
DURATIONARG=""
SENDXRAYTESTARG="1"
SENDEVENTARG="1"
PATHARG=""

help() {
  cat <<EOF
Usage: $0 [OPTIONS]
or:    $0 --remove

Options:
  --test <name>          test_conductor test to run, steady_state, non_steady_state,
                         non_steady_state_multi_vols, or replica_perturbation
  --plan <test plan ID>  specify the test plan to receive the test runs
  --duration <duration>  set the overal run time for the test with units, e.g. 12d, 34h, 56m27s etc
  --secure-file-path	 file path for k8s sealed secrets
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
            steady_state|non_steady_state|non_steady_state_multi_vols|replica_perturbation)
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
    -s|--secure-file-path)
      shift
      PATHARG=$1
      ;;
    -r|--remove)
      OPERATION="delete"
      set +e # we can ignore errors when undeploying
      ;;
    --noevent)
      SENDEVENTARG=0
      ;;
    --noxraytest)
      SENDXRAYTESTARG=0
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
  if [ -z ${PATHARG} ]; then
    echo "undefined secure file path"
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
  kubectl delete secret test-director-secret -n mayastor-e2e
  kubectl delete sealedsecret test-director-secret -n mayastor-e2e
  kubectl delete secret -n kube-system -l sealedsecrets.bitnami.com/sealed-secrets-key
  kubectl delete -f ${DEPLOYDIR}/test_director/controller.yaml
  kubectl delete -f ${DEPLOYDIR}/test_director/test_director.yaml
  kubectl delete pod -n mayastor-e2e test-conductor
  kubectl delete -f ${DEPLOYDIR}/test_conductor/test_conductor.yaml
  kubectl delete -f ${DEPLOYDIR}/log-monitor/log_monitor.yaml
  kubectl delete -f ${DEPLOYDIR}/test_namespace.yaml
  kubectl delete -f ${DEPLOYDIR}/log-monitor/fluentd.yaml
  kubectl delete -f ${DEPLOYDIR}/log-monitor/fluentd_rbac.yaml
  kubectl delete -f ${DEPLOYDIR}/log-monitor/fluentd_configmap.yaml
  kubectl delete -f ${DEPLOYDIR}/log-monitor/fluentd_namespace.yaml
else
  kubectl create -f ${DEPLOYDIR}/test_namespace.yaml
  kubectl create -f ${DEPLOYDIR}/log-monitor/fluentd_namespace.yaml
  kubectl create -f ${DEPLOYDIR}/log-monitor/fluentd_configmap.yaml
  kubectl create -f ${DEPLOYDIR}/log-monitor/fluentd_rbac.yaml
  kubectl create -f ${DEPLOYDIR}/log-monitor/fluentd.yaml
  kubectl create -f ${DEPLOYDIR}/log-monitor/log_monitor.yaml

  tmpfile=$(mktemp /tmp/tmp.XXXX)
  PLAN=${PLANARG} envsubst < ${DEPLOYDIR}/test_director/config.yaml.template > $tmpfile
  kubectl create configmap td-config -n mayastor-e2e --from-file=config-local.yaml=${tmpfile}
  rm ${tmpfile}
  kubectl create -f ${PATHARG}
  kubectl create -f ${DEPLOYDIR}/test_director/controller.yaml
  kubectl create -f ${DEPLOYDIR}/test_director/test_director_sealed_secret.yaml
  kubectl create -f ${DEPLOYDIR}/test_director/test_director.yaml

  kubectl create configmap tc-config -n mayastor-e2e --from-file=${DEPLOYDIR}/test_conductor/${TESTARG}/config.yaml

  kubectl create -f ${DEPLOYDIR}/test_conductor/test_conductor.yaml
  TEST=${IMAGEARG} DURATION=${DURATIONARG} SENDXRAYTEST=${SENDXRAYTESTARG} SENDEVENT=${SENDEVENTARG} envsubst  < ${DEPLOYDIR}/test_conductor/test_conductor_pod.yaml.template | kubectl apply -f -

  kubectl create -f ${DEPLOYDIR}/workload_monitor/workload_monitor.yaml
fi

