#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(dirname "$0")

pushd ${SCRIPT_DIR}/../swagger && swagger generate client -A etfw -f ${SCRIPT_DIR}/../../test_director/swagger_test_director_oas2.yaml; popd

