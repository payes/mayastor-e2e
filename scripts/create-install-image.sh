#!/usr/bin/env bash

# get the generated deployment files and push them to the CI repo tagged as 'nightly-stable-images'

set -euo pipefail

REGISTRY="ci-registry.mayastor-ci.mayadata.io"
OUTPUT_TAG="nightly-stable"
SCRIPTDIR=$(dirname "$(realpath "$0")")
E2EROOT=$(realpath "$SCRIPTDIR/..")
MAYASTOR_DIR=""
MOAC_DIR=""

help() {
  cat <<EOF
Create a mayastor install image, this contains all templates and files required for E2E to install mayastor

Usage: $(basename $0) [OPTIONS]

Options:
  -h, --help                 Display this text.
  --registry <host[:port]>   registry to push the generate docker image to
                             default is ${REGISTRY}
  --alias-tag                Tag to give CI image, default is ${OUTPUT_TAG}

  --mayastor                 Path to root Mayastor directory
  --moac                     Path to root MOAC directory
Examples:
  $(basename $0) --registry 127.0.0.1:5000 --alias-tag customized-tag
EOF
}

# Parse arguments
while [ "$#" -gt 0 ]; do
  case $1 in
    -h|--help)
      help
      exit 0
      shift
      ;;
    --registry)
      shift
      REGISTRY=$1
      shift
      ;;
    --alias-tag)
      shift
      OUTPUT_TAG=$1
      shift
      ;;
    --mayastor)
      shift
      MAYASTOR_DIR=$1
      shift
      ;;
    --moac)
      shift
      MOAC_DIR=$1
      shift
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

if [ -z "$MAYASTOR_DIR" ] ; then
    echo "no mayastor repository directory specified"
    exit 127
fi

if [ -z "$MOAC_DIR" ] ; then
    echo "no moac repository directory specified"
    exit 127
fi

image="mayadata/install-images:${OUTPUT_TAG}"
reg_image="${REGISTRY}/${image}"

DockerfileTxt="FROM scratch
copy install.tar /
CMD [\"ls\"]"

tmpdir=$(mktemp -d)
pushd "${MAYASTOR_DIR}" \
    && tar cf "$tmpdir/install.tar" scripts/generate-deploy-yamls.sh chart/ deploy rpc/proto/mayastor.proto \
    && git rev-parse HEAD > "$tmpdir/git-revision.mayastor" \
    && git rev-parse --short=12 HEAD >> "$tmpdir/git-revision.mayastor" \
    && popd

pushd "${MOAC_DIR}" \
    && mkdir -p "$tmpdir/csi/moac/crds" \
    && cp crds/* "$tmpdir/csi/moac/crds/" \
    && git rev-parse HEAD > "$tmpdir/git-revision.moac" \
    && git rev-parse --short=12 HEAD >> "$tmpdir/git-revision.moac" \
    && popd

echo "$DockerfileTxt" > "$tmpdir/Dockerfile"
pushd "$tmpdir" \
    && tar -rvf install.tar git-revision* csi/moac/crds/ \
    && docker build -t "${image}" . \
    && popd
rm -rf "$tmpdir"

docker tag "${image}" "${reg_image}"
docker push "${reg_image}"


