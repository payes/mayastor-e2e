#!/usr/bin/env bash

set -e pipefail

SCRIPT_DIR=$(dirname "$0")
pushd ${SCRIPT_DIR}

pushd ../cmd/main && CGO_ENABLED=0 go build -a -installsuffix cgo -o log_monitor && popd

./build_img.sh

popd