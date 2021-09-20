#!/usr/bin/env bash

set -e pipefail

SCRIPT_DIR=$(dirname "$0")
cd ${SCRIPT_DIR}

pushd test_conductor && ./build.sh; popd
pushd test_director && ./build.sh;  popd
pushd workload_monitor && ./build.sh; popd


