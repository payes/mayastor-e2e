#!/usr/bin/env bash

set -e pipefail

REGISTRY="ci-registry.mayastor-ci.mayadata.io"
TAG="latest"
APP="test_director"
IMG_NAME=mayadata/${APP}
SCRIPT_DIR=$(dirname "$0")
pushd ${SCRIPT_DIR}

cd ../docker
cp ../cmd/test-framework-server/${APP} .
cp ../config/config-local.yaml .
docker build -t ${REGISTRY}/${IMG_NAME}:${TAG} .
docker push ${REGISTRY}/${IMG_NAME}:${TAG}
rm ${APP}
rm config-local.yaml

popd

