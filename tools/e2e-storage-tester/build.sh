#!/bin/env bash

set -e

SCRIPT_DIR=$(dirname "$0")
pushd ${SCRIPT_DIR}

docker build -t mayadata/e2e-storage-tester .

popd
