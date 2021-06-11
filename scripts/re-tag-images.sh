#!/usr/bin/env bash

# get the devlop images and push them to the CI repo tagged as 'nightly-stable'

set -euo pipefail

IMAGES="mayastor mayastor-csi mayastor-client moac"

REGISTRY="ci-registry.mayastor-ci.mayadata.io"
INPUT_TAG="develop"
OUTPUT_TAG="nightly-stable"

help() {
  cat <<EOF
Usage: $(basename $0) [OPTIONS]

Options:
  -h, --help                 Display this text.
  --registry <host[:port]>   Push the built images to the provided registry,
                             default is ${REGISTRY}
  --alias-tag                Tag to give CI image, default is ${OUTPUT_TAG}

Examples:
  $(basename $0) --registry 127.0.0.1:5000 --alias-tag customized-tag
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
      REGISTRY=$1
      shift
      ;;
    --alias-tag)
      shift
      OUTPUT_TAG=$1
      shift
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

for name in $IMAGES; do
  input_image="mayadata/${name}:${INPUT_TAG}"

  docker pull ${input_image}

  output_image="${REGISTRY}/mayadata/${name}:${OUTPUT_TAG}"

  docker tag ${input_image} ${output_image}

  docker push ${output_image}
done

