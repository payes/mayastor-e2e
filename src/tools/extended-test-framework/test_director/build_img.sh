REGISTRY="ci-registry.mayastor-ci.mayadata.io"
TAG="latest"
APP="test_director"
cd docker && cp ../${APP} . && docker build -t ${REGISTRY}/mayadata/${APP}:${TAG} .
docker push ${REGISTRY}/mayadata/${APP}:${TAG}

