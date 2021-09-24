REGISTRY="ci-registry.mayastor-ci.mayadata.io"
TAG="latest"
APP="test_director"
IMG_NAME=mayadata/${APP}_cwd
SCRIPT_DIR=$(dirname "$0")
cd ${SCRIPT_DIR}

cd docker
cp ../cmd/test-framework-server/${APP} .
cp ../config/config-local.yaml .
docker build -t ${REGISTRY}/${IMG_NAME}:${TAG} .
docker push ${REGISTRY}/${IMG_NAME}:${TAG}
rm ${APP}
rm config-local.yaml

