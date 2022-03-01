#!/usr/bin/env bash

set -e

DURATIONARG=""
RUNNAMEARG=""
OPERATION=""
PATHARG=""
PLANARG=""
SENDXRAYTESTARG="1"
SENDEVENTARG="1"
TAG="latest"
TESTARG=""

help() {
  cat <<EOF
Usage: $0 [OPTIONS]
or:    $0 --remove

Options:
  --duration <duration>      set the overal run time for the test with units,
                             e.g. 12d, 34h, 56m27s etc. The default is 14d.
  --name     <name>          (optional) string passed to log output to
                             identify the test
  --plan     <test plan ID>  specify the test plan to receive the test runs
  --secure-file-path <path>  file path for k8s sealed secrets
  --tag      <name>          (optional) deploy ETFW images having this tag,
                             default "latest"
  --test     <name>          specify test_conductor test to run, steady_state,
                             non_steady_state, non_steady_state_multi_vols,
                             replica_perturbation or replica_elimination
or:
  --remove                   remove instead of deploy
Examples:
  $0 --test steady_state --plan AB-123 --duration 6d12h --secure-file-path ./secret.yaml
  $0 --remove
EOF
}

# Parse arguments
while [ "$#" -gt 0 ]; do
  case "$1" in

    -d|--duration)
      shift
      DURATIONARG=$1
      ;;
    --noevent)
      SENDEVENTARG=0
      ;;
    --noxraytest)
      SENDXRAYTESTARG=0
      ;;
    -n|--name)
      shift
      RUNNAMEARG=$1
      ;;
    -p|--plan)
      shift
      PLANARG=$1
      ;;
    -r|--remove)
      OPERATION="delete"
      ;;
    -s|--secure-file-path)
      shift
      PATHARG=$(realpath "$1")
      ;;
    -t|--test)
      shift
      case $1 in
            steady_state|non_steady_state|non_steady_state_multi_vols|replica_perturbation|replica_elimination|primitive_pool_deletion)
                TESTARG=$1
                ;;
            *)
                echo "unrecognized test"
                help
                exit 1
		;;
      esac
      ;;
    --tag)
      shift
      TAG=$1
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

SCRIPTDIR=$(dirname "$(realpath "$0")")
DEPLOYDIR="${SCRIPTDIR}/../deploy/"

deploy_test_conductor() {
  kubectl create configmap tc-config -n mayastor-e2e --from-file=./test_conductor/${TESTARG}/config.yaml
  helm template chart \
	  --set duration=${DURATIONARG} \
	  --set name="${RUNNAMEARG}" \
	  --set sendxraytest=${SENDXRAYTESTARG} \
	  --set sendevent=${SENDEVENTARG} \
	  --set tag=${TAG} \
	  -s templates/test_conductor/test_conductor_pod.yaml > ./test_conductor/test_conductor_pod.yaml
  kubectl create -f ./test_conductor/test_conductor.yaml
  kubectl create -f ./test_conductor/test_conductor_pod.yaml
}

undeploy_test_conductor() {
  kubectl delete pod -n mayastor-e2e test-conductor || true
  kubectl delete -f ./test_conductor/test_conductor.yaml || true
  kubectl delete configmap tc-config -n mayastor-e2e || true
}

deploy_test_director() {
  helm template chart --set plan=${PLANARG} -s templates/test_director/config.yaml > test_director/config.yaml
  kubectl create configmap td-config -n mayastor-e2e --from-file=config-local.yaml=test_director/config.yaml
  kubectl create -f ${PATHARG}
  kubectl create -f ./test_director/controller.yaml
  kubectl create -f ./test_director/test_director_sealed_secret.yaml
  helm template chart --set tag=${TAG} \
          -s templates/test_director/test_director_pod.yaml > ./test_director/test_director_pod.yaml
  kubectl create -f ./test_director/test_director.yaml
  kubectl create -f ./test_director/test_director_pod.yaml
}

undeploy_test_director() {
  kubectl delete pod -n mayastor-e2e test-director || true
  kubectl delete configmap td-config -n mayastor-e2e || true
  kubectl delete secret test-director-secret -n mayastor-e2e || true
  kubectl delete sealedsecret test-director-secret -n mayastor-e2e || true
  kubectl delete secret -n kube-system -l sealedsecrets.bitnami.com/sealed-secrets-key || true
  kubectl delete -f ./test_director/controller.yaml || true
  kubectl delete -f ./test_director/test_director.yaml || true
}

deploy_log_monitor() {
  kubectl create -f ./log_monitor/fluentd_namespace.yaml
  kubectl create -f ./log_monitor/fluentd_configmap.yaml
  kubectl create -f ./log_monitor/fluentd_rbac.yaml
  kubectl create -f ./log_monitor/fluentd.yaml
  helm template chart --set tag=${TAG} \
          -s templates/log_monitor/log_monitor_pod.yaml > ./log_monitor/log_monitor_pod.yaml
  kubectl create -f ./log_monitor/log_monitor.yaml
  kubectl create -f ./log_monitor/log_monitor_pod.yaml
}

undeploy_log_monitor() {
  kubectl delete pod -n mayastor-e2e log-monitor || true
  kubectl delete -f ./log_monitor/log_monitor.yaml || true
  kubectl delete -f ./log_monitor/fluentd.yaml || true
  kubectl delete -f ./log_monitor/fluentd_rbac.yaml || true
  kubectl delete -f ./log_monitor/fluentd_configmap.yaml || true
  kubectl delete -f ./log_monitor/fluentd_namespace.yaml || true
}

deploy_workload_monitor() {
  helm template chart --set tag=${TAG} \
          -s templates/workload_monitor/workload_monitor_pod.yaml > ./workload_monitor/workload_monitor_pod.yaml
  kubectl create -f ./workload_monitor/workload_monitor.yaml
  kubectl create -f ./workload_monitor/workload_monitor_pod.yaml
}

undeploy_workload_monitor() {
  kubectl delete pod -n mayastor-e2e workload-monitor || true
  kubectl delete -f ./workload_monitor/workload_monitor.yaml || true
}

pushd ${DEPLOYDIR}

if [ "${OPERATION}" == "delete" ]; then
  undeploy_workload_monitor
  undeploy_test_conductor
  undeploy_test_director
  undeploy_log_monitor
  kubectl delete -f ./test_namespace.yaml
else
  kubectl create -f ./test_namespace.yaml
  deploy_log_monitor
  deploy_test_director
  deploy_test_conductor
  deploy_workload_monitor
fi

popd

