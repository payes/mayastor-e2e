#!/usr/bin/env bash

set -eu

echo $#


if [ "$#" == "0" ]; then
	echo "usage $0 actor"
	exit 1
fi

SCRIPT_DIR=$(dirname "$0")

swagger generate server -A $1 -f ${SCRIPT_DIR}/swagger_full_oas2.yaml
