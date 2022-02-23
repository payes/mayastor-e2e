#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(dirname "$0")
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

pushd ${SCRIPT_DIR}

pushd ../log_monitor/scripts && ./build.sh -t ${TAG}; popd
pushd ../test_conductor/scripts && ./build.sh -t ${TAG}; popd
pushd ../test_director/scripts && ./build.sh -t ${TAG}; popd
pushd ../workload_monitor/scripts && ./build.sh -t ${TAG}; popd

popd

