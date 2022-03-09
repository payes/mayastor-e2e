#!/usr/bin/env bash

set -euo pipefail

SCRIPTDIR=$(dirname "$(realpath "$0")")
REGISTRY="ci-registry.mayastor-ci.mayadata.io"
DESTINATION_REGISTRY="$REGISTRY"
SRC_TAG=""
ALIAS_TAG=""

help() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Options:
  -h, --help                 Display this text.
  --registry <host[:port]>   Push the built images to the provided registry,
                             default is ${REGISTRY}
  --src-tag                  Tag to retag
  --alias-tag                Tag to give CI image

Examples:
  $(basename "$0") --registry 127.0.0.1:5000 --src-tag 755c435fdb0a --alias-tag customized-tag
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
      DESTINATION_REGISTRY=$1
      shift
      ;;
    --src-tag)
      shift
      SRC_TAG=$1
      shift
      ;;
    --alias-tag)
      shift
      ALIAS_TAG=$1
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
      product="$1"
      shift
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

if [ -z "$SRC_TAG" ] ; then
    echo "source tag not specified"
    help
    exit 1
fi

if [ -z "$ALIAS_TAG" ] ; then
    echo "alias tag not specified"
    help
    exit 1
fi

# extract the install bundle - and collect the names of mayadata/datacore images
# from helm chart yaml files.
# convert lines like
#   {{ .Values.mayastorCP.registry }}datacore/bolt-rest:{{ .Values.mayastorCP.tag }}
# to
#   bolt-rest
tmpdir=$(mktemp -d)
"$SCRIPTDIR/extract-install-image.sh" --alias-tag "$SRC_TAG" --installroot "$tmpdir" --product "$product"
images=$(find "$tmpdir" -type f -name '*.yaml' -print0 \
    | xargs -0 grep -w image \
    | grep -w -e datacore -e mayadata \
    | sed  \
    -e 's#.*datacore/##' \
    -e 's#.*mayadata/##' \
    -e 's#:{{.*##' \
)
rm -rf "$tmpdir"

for name in $images; do
  input_image="${REGISTRY}/$registry_subdir/${name}:${SRC_TAG}"

  docker pull "${input_image}"

  if [ "$DESTINATION_REGISTRY" == "dockerhub" ]; then
    output_image="$registry_subdir/${name}:${ALIAS_TAG}"
    # do not upload install-images to dockerhub
    if [ "$name" == "install-images" ]; then
        continue
    fi
  else
    output_image="${DESTINATION_REGISTRY}/$registry_subdir/${name}:${ALIAS_TAG}"
  fi

  docker tag "${input_image}" "${output_image}"

  docker push "${output_image}"
done

