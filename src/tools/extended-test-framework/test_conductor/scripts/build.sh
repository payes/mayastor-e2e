#!/usr/bin/env bash

set -euo pipefail

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


build_img () {
	APP=$1
	REGISTRY="ci-registry.mayastor-ci.mayadata.io"

	mkdir -p ../docker
	pushd ../docker
	cp ../cmd/${APP}/${APP} .
	cp ../cmd/${APP}/Dockerfile .

	docker build -t ${REGISTRY}/mayadata/${APP}:${TAG} .
	docker push ${REGISTRY}/mayadata/${APP}:${TAG}

	rm -Rf *
	popd
}

build () {

	#./gen_client_code.sh

	pushd ../cmd/$1 && CGO_ENABLED=0 go build -a -installsuffix cgo && popd

	build_img $1
}

SCRIPT_DIR=$(dirname "$0")

pushd ${SCRIPT_DIR}

build test_conductor_steady_state
build test_conductor_non_steady_state
build test_conductor_replica_perturbation
build test_conductor_replica_elimination
build test_conductor_primitive_pool_deletion

popd
