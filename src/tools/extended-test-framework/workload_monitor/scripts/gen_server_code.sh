#!/usr/bin/env bash

set -eu

SCRIPT_DIR=$(dirname "$0")

pushd ${SCRIPT_DIR}/../swagger && swagger generate server -A etfw -f ../swagger_workload_monitor_oas2.yaml; popd
