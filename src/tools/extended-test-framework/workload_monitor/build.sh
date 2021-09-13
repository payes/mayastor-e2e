#!/usr/bin/env bash

set -e

SCRIPT_DIR=$(dirname "$0")
cd ${SCRIPT_DIR}

CGO_ENABLED=0 go build -a -installsuffix cgo

./build_img.sh

