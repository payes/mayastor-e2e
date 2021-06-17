#!/usr/bin/env bash

# get the generated deployment files and push them to the CI repo tagged as 'nightly-stable-images'

set -euo pipefail

REGISTRY="ci-registry.mayastor-ci.mayadata.io"
TAG=""
SCRIPTDIR=$(dirname "$(realpath "$0")")
E2EROOT=$(realpath "$SCRIPTDIR/..")
ARTIFACTSDIR=$(realpath "$E2EROOT/artifacts")
INSTALLROOT="$ARTIFACTSDIR/install"


help() {
  cat <<EOF
This scriopt extracts the templates and files required for E2E to install mayastor
from a mayastor install image.
The files will be extracted to a location under "$INSTALLROOT"

Usage: $(basename $0) [OPTIONS]

Options:
  -h, --help                 Display this text.
  --registry <host[:port]>   Registry to pull the install-image from
                             default is ${REGISTRY}
  --alias-tag                Tag of install image to use, default is ${TAG}

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
      TAG=$1
      shift
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

if [ -z "$TAG" ] ; then
    echo "tag not specified"
    help
fi

image=${REGISTRY}/mayadata/install-images:${TAG}
docker pull "${image}"

rm -rf "${INSTALLROOT:?}/${TAG}"
mkdir -p "$INSTALLROOT/${TAG}"

tmpdir=$(mktemp -d)
contnr=$(docker create "${image}")
docker cp "${contnr}:/install.tar" "$tmpdir"
echo "Extracting install files to $INSTALLROOT/${TAG}"
pushd "$INSTALLROOT/${TAG}" && tar xf "$tmpdir/install.tar" && popd
docker rm "${contnr}" >& /dev/null
rm -rf "$tmpdir"
