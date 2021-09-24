#!/usr/bin/env bash

set -e pipefail

SCRIPT_DIR=$(dirname "$0")
cd ${SCRIPT_DIR}

pushd cmd/test-framework-server && CGO_ENABLED=0 go build -a -installsuffix cgo -o test_director && popd

./build_img.sh
