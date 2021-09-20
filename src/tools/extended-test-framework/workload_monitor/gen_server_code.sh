#!/usr/bin/env bash

set -eu

SCRIPT_DIR=$(dirname "$0")

swagger generate server -A etfw -f ${SCRIPT_DIR}/swagger_workload_monitor_oas2.yaml
