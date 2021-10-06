#!/usr/bin/env bash

set -e pipefail

#MAYASTORTAG="nightly-stable"
MAYASTORTAG="mcp-2021-10-05-00-17-04"
ETFWTAG="latest"

build_img () {
	APP=$1
	TAG=$2
	REGISTRY="ci-registry.mayastor-ci.mayadata.io"

	mkdir -p ../docker
	pushd ../docker
	cp ../cmd/${APP}/${APP} .
	cp ../cmd/${APP}/Dockerfile .

	#../../../../../scripts/extract-install-image.sh --alias-tag ${MAYASTORTAG} --installroot .
	#install-bundle/${MAYASTORTAG}/scripts/generate-deploy-yamls.sh -t ${MAYASTORTAG} -o . test
	#cp install-bundle/*/mcp/target/release/kubectl-mayastor .
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
cd ${SCRIPT_DIR}

build test_conductor_steady_state
build test_conductor_non_steady_state
build test_conductor_replica_perturbation

