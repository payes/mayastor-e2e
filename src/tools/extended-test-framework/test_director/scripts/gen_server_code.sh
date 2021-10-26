#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(dirname "$0")

pushd ${SCRIPT_DIR}/..

swagger generate server -A test_framework -f swagger_test_director_oas2.yaml

popd
