#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(dirname "$0")
cd ${SCRIPT_DIR}

TAG="latest"

help() {
  cat <<EOF
Usage: $0 [OPTIONS]

Options:
  --tag <name> 		tag to assign to built images
Examples:
  $0 --tag my_custom_tag
EOF
}

# Parse arguments
while [ "$#" -gt 0 ]; do
  case "$1" in
    -t|--tag)
      shift
      echo "building with tag $1"
      TAG="$1"
      ;;
    -h)
      help
      exit 0
      ;;
    *)
      echo "unrecognized parameter"
      help
      exit 1
      ;;
  esac
  shift
done

#./gen_server_code.sh
#./gen_client_code.sh

pushd ../cmd/workload_monitor && CGO_ENABLED=0 go build -a -installsuffix cgo; popd

./build_img.sh ${TAG}

