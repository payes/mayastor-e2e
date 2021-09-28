#!/usr/bin/env bash

set -e pipefail

SCRIPT_DIR=$(dirname "$0")
cd ${SCRIPT_DIR}

#./gen_server_code.sh
#./gen_client_code.sh

pushd ../cmd/workload_monitor && CGO_ENABLED=0 go build -a -installsuffix cgo; popd

./build_img.sh
