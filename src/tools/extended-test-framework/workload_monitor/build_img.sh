REGISTRY="ci-registry.mayastor-ci.mayadata.io"

TAG="cwd"
APP="workload_monitor"
cd docker && cp ../${APP} . && docker build -t ${REGISTRY}/mayadata/${APP}:${TAG} .
docker push ${REGISTRY}/mayadata/${APP}:${TAG}

