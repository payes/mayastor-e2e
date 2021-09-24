#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(dirname "$0")

swagger generate server -A test_framework -f ${SCRIPT_DIR}/swagger_test_director_oas2.yaml

