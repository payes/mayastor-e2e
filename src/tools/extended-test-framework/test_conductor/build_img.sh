#!/usr/bin/env bash

set -e

REGISTRY="ci-registry.mayastor-ci.mayadata.io"
APP="test_conductor"
TAG="latest"

SCRIPT_DIR=$(dirname "$0")
cd ${SCRIPT_DIR}

pushd docker
cp ../${APP} .
#cp ../config.yaml .
../../../../../scripts/extract-install-image.sh --alias-tag nightly-stable --installroot .
install/nightly-stable/scripts/generate-deploy-yamls.sh -t nightly-stable -o . test

docker build -t ${REGISTRY}/mayadata/${APP}_cwd:${TAG} .
docker push ${REGISTRY}/mayadata/${APP}_cwd:${TAG}

rm *.yaml
rm -Rf etcd
rm -Rf install
rm ${APP}
