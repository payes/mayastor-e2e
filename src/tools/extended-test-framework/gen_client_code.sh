#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(dirname "$0")

swagger generate client -A $1 -f ${SCRIPT_DIR}/swagger_full_oas2.yaml
