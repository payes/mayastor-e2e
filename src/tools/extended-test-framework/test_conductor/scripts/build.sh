#!/usr/bin/env bash

set -e pipefail

build_img () {
	APP=$1
	TAG=$2
	REGISTRY="ci-registry.mayastor-ci.mayadata.io"

	mkdir -p ../docker
	pushd ../docker
	cp ../cmd/${APP}/${APP} .
	cp ../cmd/${APP}/Dockerfile .

	../../../../../scripts/extract-install-image.sh --alias-tag nightly-stable --installroot .
	install-bundle/nightly-stable/scripts/generate-deploy-yamls.sh -t nightly-stable -o . test

	docker build -t ${REGISTRY}/mayadata/${APP}:${TAG} .
	docker push ${REGISTRY}/mayadata/${APP}:${TAG}

	rm -Rf *
	popd
}

build () {

	#./gen_client_code.sh

	pushd ../cmd/$1 && CGO_ENABLED=0 go build -a -installsuffix cgo && popd

	build_img $1 $2
}

SCRIPT_DIR=$(dirname "$0")
cd ${SCRIPT_DIR}

build test_conductor_steady_state latest
build test_conductor_replica_perturbation latest

