#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(dirname "$0")

cat ${SCRIPT_DIR}/../test_director/swagger_test_director_oas2.yaml > /tmp/combined.yaml
cat ${SCRIPT_DIR}/../workload_monitor/swagger_workload_monitor_oas2.yaml >> /tmp/combined.yaml

swagger generate server -A etfw -f ${SCRIPT_DIR}/swagger_test_director_oas2.yaml

