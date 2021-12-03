#!/usr/bin/env bash

set -e pipefail

ETFWTAG="latest"

build_img () {
	APP=$1
	TAG=$2
	REGISTRY="ci-registry.mayastor-ci.mayadata.io"

	mkdir -p ../docker
	pushd ../docker
	cp ../cmd/${APP}/${APP} .
	cp ../cmd/${APP}/Dockerfile .

	docker build -t ${REGISTRY}/mayadata/${APP}:${ETFWTAG} .
	docker push ${REGISTRY}/mayadata/${APP}:${ETFWTAG}

	rm -Rf *
	popd
}

build () {

	#./gen_client_code.sh

	pushd ../cmd/$1 && CGO_ENABLED=0 go build -a -installsuffix cgo && popd

	build_img $1 $2
}

SCRIPT_DIR=$(dirname "$0")

pushd ${SCRIPT_DIR}

build test_conductor_steady_state
build test_conductor_non_steady_state
build test_conductor_replica_perturbation

popd
