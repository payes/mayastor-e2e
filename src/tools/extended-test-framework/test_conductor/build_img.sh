REGISTRY="ci-registry.mayastor-ci.mayadata.io"

APP="test_conductor"
TAG="latest"
cd docker && cp ../${APP} . && cp ../config.yaml . && docker build -t ${REGISTRY}/mayadata/${APP}:${TAG} .
docker push ${REGISTRY}/mayadata/${APP}:${TAG}

