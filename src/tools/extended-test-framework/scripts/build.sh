#!/usr/bin/env bash

set -e pipefail

SCRIPT_DIR=$(dirname "$0")
pushd ${SCRIPT_DIR}

pushd ../test_conductor/scripts && ./build.sh; popd
pushd ../test_director/scripts && ./build.sh;  popd
pushd ../workload_monitor/scripts && ./build.sh; popd

popd
