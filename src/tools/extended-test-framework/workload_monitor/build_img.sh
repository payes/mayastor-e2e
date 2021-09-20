#!/usr/bin/env bash

REGISTRY="ci-registry.mayastor-ci.mayadata.io"
TAG="latest"
APP="workload_monitor"

SCRIPT_DIR=$(dirname "$0")
cd ${SCRIPT_DIR}

cd docker && cp ../${APP} . && docker build -t ${REGISTRY}/mayadata/${APP}_cwd:${TAG} .
docker push ${REGISTRY}/mayadata/${APP}_cwd:${TAG}

rm -f ${APP}
