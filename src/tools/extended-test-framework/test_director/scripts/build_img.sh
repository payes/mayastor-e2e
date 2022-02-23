#!/usr/bin/env bash

set -euo pipefail

REGISTRY="ci-registry.mayastor-ci.mayadata.io"
TAG="latest"
APP="test_director"
IMG_NAME=mayadata/${APP}
SCRIPT_DIR=$(dirname "$0")

if (($# == 1)); then
        TAG=$1
fi

pushd ${SCRIPT_DIR}

pushd ../docker
cp ../cmd/test-framework-server/${APP} .
docker build -t ${REGISTRY}/${IMG_NAME}:${TAG} .
docker push ${REGISTRY}/${IMG_NAME}:${TAG}
rm ${APP}
popd

popd

