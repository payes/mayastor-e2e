#!/usr/bin/env bash

set -euo pipefail

REGISTRY="ci-registry.mayastor-ci.mayadata.io"
TAG="latest"
APP="workload_monitor"

if (( $# == 1 )); then
        TAG=$1
fi

SCRIPT_DIR=$(dirname "$0")
cd ${SCRIPT_DIR}

pushd ../docker && cp ../cmd/${APP}/${APP} . && docker build -t ${REGISTRY}/mayadata/${APP}:${TAG} .
docker push ${REGISTRY}/mayadata/${APP}:${TAG}

rm -f ${APP}
popd

