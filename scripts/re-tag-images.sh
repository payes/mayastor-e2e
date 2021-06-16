#!/usr/bin/env bash

# get the devlop images and push them to the CI repo tagged as 'nightly-stable'

set -euo pipefail

IMAGES="mayastor mayastor-csi mayastor-client moac"

REGISTRY="ci-registry.mayastor-ci.mayadata.io"
SRC_TAG=""
ALIAS_TAG=""

help() {
  cat <<EOF
Usage: $(basename $0) [OPTIONS]

Options:
  -h, --help                 Display this text.
  --registry <host[:port]>   Push the built images to the provided registry,
                             default is ${REGISTRY}
  --src-tag                  Tag to retag
  --alias-tag                Tag to give CI image

Examples:
  $(basename $0) --registry 127.0.0.1:5000 --src-tag 755c435fdb0a --alias-tag customized-tag
EOF
}

# Parse arguments
while [ "$#" -gt 0 ]; do
  case $1 in
    -h|--help)
      help
      exit 0
      shift
      ;;
    --registry)
      shift
      DESTINATION_REGISTRY=$1
      shift
      ;;
    --src-tag)
      shift
      SRC_TAG=$1
      shift
      ;;
    --alias-tag)
      shift
      ALIAS_TAG=$1
      shift
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

if [ -z "$SRC_TAG" ] ; then
    echo "source tag not specified"
    help
    exit 1
fi

if [ -z "$ALIAS_TAG" ] ; then
    echo "alias tag not specified"
    help
    exit 1
fi

for name in $IMAGES; do
  input_image="${REGISTRY}/mayadata/${name}:${SRC_TAG}"

  docker pull "${input_image}"

  if [ "$DESTINATION_REGISTRY" == "dockerhub" ]; then
    output_image="mayadata/${name}:${ALIAS_TAG}"
  else
    output_image="${DESTINATION_REGISTRY}/mayadata/${name}:${ALIAS_TAG}"
  fi

  docker tag "${input_image}" "${output_image}"

  docker push "${output_image}"
done

