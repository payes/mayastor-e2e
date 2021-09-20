#!/usr/bin/env bash

set -e pipefail

SCRIPT_DIR=$(dirname "$0")
cd ${SCRIPT_DIR}

./gen_server_code.sh

CGO_ENABLED=0 go build -a -installsuffix cgo

./build_img.sh
