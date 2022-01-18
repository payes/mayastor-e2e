#!/usr/bin/env bash
set -eo pipefail
help() {
  cat <<EOF
This script will put a mark to loki logs, add optional string and add dump of
selected CRs if given -c.

Usage: $0 [-h|--help] [-c|--dump-cr] [<message>]
example: ./loki-marker.sh -c Node1 shutdown test
Options:
      -c|--dump-cr         dump selected CR output
EOF
}

#To check the clusterrolebinding for default service account
if ! kubectl get clusterrolebinding | grep  add-on-cluster-admin > /dev/null
then
  kubectl create clusterrolebinding add-on-cluster-admin --clusterrole=cluster-admin --serviceaccount=default:default
fi

TMPF=$(mktemp)
# shellcheck disable=SC2064
trap "[ -f '$TMPF' ] && rm '$TMPF'" TERM INT QUIT HUP EXIT

case "$1" in
  -h|--help)
      help
      exit
      ;;
  -c|--dump-crs)
    shift
    cat > "$TMPF" << EOF
                    apiVersion: batch/v1
                    kind: Job
                    metadata:
                      name: kubectl-$RANDOM
                      labels:
                        app: e2e-rest-agent
                    spec:
                      template:
                        metadata:
                          labels:
                            app: e2e-rest-agent
                        spec:
                          containers:
                          - name: kubectl-demo
                            image: bitnami/kubectl
                            command: ["/bin/bash","-c"]
                            args:
                              - |
                                echo '[TEST_EVENT]${*:+-$*}' &&
                                kubectl get nodes -o wide &&
                                kubectl get msn -n mayastor -o wide &&
                                kubectl get msp -n mayastor -o wide &&
                                kubectl get msv -n mayastor -o wide &&
                                kubectl get pvc -n default -o wide &&
                                kubectl get pv -o wide
                          restartPolicy: Never
                      backoffLimit: 4
EOF

   echo "apply the batch job yaml"
   kubectl apply -f "$TMPF"
   ;;
  *)
   cat > "$TMPF" << EOF
                    apiVersion: batch/v1
                    kind: Job
                    metadata:
                      name: kubectl-$RANDOM
                      labels:
                        app: e2e-rest-agent
                    spec:
                      template:
                        metadata:
                          labels:
                            app: e2e-rest-agent
                        spec:
                          containers:
                          - name: kubectl-demo
                            image: bitnami/kubectl
                            command: ["/bin/bash", "-c", "echo '[TEST_EVENT]${*:+-$*}'"]
                          restartPolicy: Never
                      backoffLimit: 4
EOF

   echo "apply the batch job yaml"
   kubectl apply -f "$TMPF"
   ;;
esac
