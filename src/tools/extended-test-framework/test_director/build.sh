#!/usr/bin/env bash

set -e pipefail

CGO_ENABLED=0 go build -a -installsuffix cgo

./build_img.sh
