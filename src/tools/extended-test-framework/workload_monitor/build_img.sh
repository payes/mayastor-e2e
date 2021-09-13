#!/usr/bin/env bash

SCRIPT_DIR=$(dirname "$0")
cd ${SCRIPT_DIR}

TAG="latest"
APP="workload_monitor"

cd docker && cp ../${APP} . && docker build -t ${REGISTRY}/mayadata/${APP}:${TAG} .

docker push ${REGISTRY}/mayadata/${APP}:${TAG}

rm -f ${APP}

