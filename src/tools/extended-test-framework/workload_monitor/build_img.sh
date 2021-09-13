#!/usr/bin/env bash

set -e pipefail

REGISTRY="ci-registry.mayastor-ci.mayadata.io"
TAG="cwd"
APP="workload_monitor"

SCRIPT_DIR=$(dirname "$0")
cd ${SCRIPT_DIR}

cd docker && cp ../${APP} . && docker build -t ${REGISTRY}/mayadata/${APP}:${TAG} .
docker push ${REGISTRY}/mayadata/${APP}:${TAG}

rm -f ${APP}
