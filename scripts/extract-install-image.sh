#!/usr/bin/env bash

# get the generated deployment files and push them to the CI repo tagged as 'nightly-stable-images'

set -euo pipefail

REGISTRY="ci-registry.mayastor-ci.mayadata.io"
TAG=""
#SCRIPTDIR=$(dirname "$(realpath "$0")")
#E2EROOT=$(realpath "$SCRIPTDIR/..")
INSTALLROOT=""


help() {
  cat <<EOF
This script extracts the templates and files required for E2E to install mayastor
from a mayastor install image.

Usage: $(basename "$0") [OPTIONS]

Options:
  -h, --help                 Display this text.
  --registry <host[:port]>   Registry to pull the install-image from
                             default is ${REGISTRY}
  --alias-tag                Tag of install image to use, default is ${TAG}
  --installroot              install root directory
  --product                  Product key [mayastor, bolt]

Examples:
  $(basename "$0") --registry 127.0.0.1:5000 --alias-tag customized-tag --installroot <path> --product bolt
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
    --installroot)
      shift
      if [ -n "$1" ]; then
            INSTALLROOT="$1"
      fi
      shift
      ;;
    --product)
      shift
      case $1 in
          mayastor)
             registry_subdir='mayadata'
             ;;
          bolt)
             registry_subdir='datacore'
             ;;
          *)
              echo "Unknown product: $1"
              exit 1
              ;;
      esac
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

if [ -z "$INSTALLROOT" ] ; then
    echo "install root not specified"
    help
fi

image=${REGISTRY}/$registry_subdir/install-images:${TAG}
docker pull "${image}"

# if the install bundle directory exists, we should
# not assume contents are valid for this installation
# remove the directory
if [ -d "${INSTALLROOT}" ]; then
    echo "Deleting contents of ${INSTALLROOT}"
    (cd "${INSTALLROOT}" && rm -rf "*") || exit 255
fi
mkdir -p "$INSTALLROOT"

tmpdir=$(mktemp -d)
contnr=$(docker create "${image}")
docker cp "${contnr}:/install.tar" "$tmpdir"
echo "Extracting install files to $INSTALLROOT"
pushd "$INSTALLROOT" && tar xf "$tmpdir/install.tar" && popd
docker rm "${contnr}" >& /dev/null
rm -rf "$tmpdir"
