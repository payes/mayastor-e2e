#!/usr/bin/env bash

# get the generated deployment files and push them to the CI repo tagged as 'nightly-stable-images'

set -euo pipefail

REGISTRY="ci-registry.mayastor-ci.mayadata.io"
OUTPUT_TAG="nightly-stable"
#SCRIPTDIR=$(dirname "$(realpath "$0")")
#E2EROOT=$(realpath "$SCRIPTDIR/..")
MAYASTOR_DIR=""
MCP_DIR=""
declare -a build_info

# render build_info as a json file
function build_info_to_json {
    printf "{\n"
    alen=$(( ${#build_info[*]}-1 ))

    for (( c=0; c <= alen; c++ ))
    do
        if (( c < alen )) ; then
            printf "    %s,\n" "${build_info[$c]}"
        else
            printf "    %s\n" "${build_info[$c]}"
        fi
    done
    printf "}\n"
}

help() {
  cat <<EOF
Create a mayastor install image, this contains all templates and files required for E2E to install mayastor

Usage: $(basename "$0") [OPTIONS]

Options:
  -h, --help                 Display this text.
  --registry <host[:port]>   registry to push the generate docker image to
                             default is ${REGISTRY}
  --alias-tag                Tag to give CI image, default is ${OUTPUT_TAG}

  --mayastor                 Path to root Mayastor directory
  --mcp                      Path to root Mayastor control plane directory
  --coverage                 Build is a coverage build
  --debug                    Build is debug build
Examples:
  $(basename "$0") --registry 127.0.0.1:5000 --alias-tag customized-tag
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
    --mcp)
      shift
      MCP_DIR=$1
      shift
      ;;
    --coverage)
      shift
      build_info+=('"coverage": true')
      ;;
    --debug)
      shift
      build_info+=('"debug": true')
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

if [ -z "$MCP_DIR" ]; then
    echo "mcp repository directory not specified"
    exit 127
fi

image="mayadata/install-images:${OUTPUT_TAG}"
reg_image="${REGISTRY}/${image}"

DockerfileTxt="FROM scratch
copy install.tar /
CMD [\"ls\"]"

tmpdir=$(mktemp -d)
workdir=$tmpdir/work
mkdir -p "$workdir"
pushd "${MAYASTOR_DIR}" \
    && tar cf "$tmpdir/install.tar" scripts/generate-deploy-yamls.sh rpc/mayastor-api/protobuf/mayastor.proto \
    && build_info+=("\"mayastor-revision\": \"$(git rev-parse HEAD)\"") \
    && build_info+=("\"mayastor-short-revision\": \"$(git rev-parse --short=12 HEAD)\"") \
    && cp -R chart/ "$workdir" \
    && popd

# Note we ensure that $workdir/mcp/bin/kubectl-mayastor is writable,
# to avoid failure in jenkins archiving where we overwrite
# artifacts/install-bundle returned by each parallel run
if [ -n "$MCP_DIR" ]; then
    pushd "${MCP_DIR}" \
        && mkdir -p "$workdir/mcp/scripts" \
        && cp scripts/generate-deploy-yamls.sh "$workdir/mcp/scripts" \
        && cp -R chart "$workdir/mcp" \
        && mkdir -p "$workdir/mcp/bin" \
        && cp "$(nix-build -A utils.release.kubectl-plugin --no-out-link)/bin/kubectl-mayastor" "$workdir/mcp/bin" \
        && chmod a+w "$workdir/mcp/bin/kubectl-mayastor" \
        && mkdir -p "$workdir/mcp/control-plane/rest/openapi-specs" \
        && cp control-plane/rest/openapi-specs/* "$workdir/mcp/control-plane/rest/openapi-specs" \
        && build_info+=("\"mayastor-control-plane-revision\": \"$(git rev-parse HEAD)\"") \
        && build_info+=("\"mayastor-control-plane-short-revision\": \"$(git rev-parse --short=12 HEAD)\"") \
        && popd
    # FIXME: dangling CRD yaml files break deployment using helm
    pushd "$workdir" \
        && rm -f chart/crds/* \
        && popd
fi

build_info_to_json > "$workdir/build_info.json"

pushd "$workdir" \
    && tar -rf "$tmpdir/install.tar" . \
    && popd

rm -rf "$workdir"

echo "$DockerfileTxt" > "$tmpdir/Dockerfile"
    pushd "$tmpdir" \
        && docker build -t "${image}" . \
        && popd

rm -rf "$tmpdir"

docker tag "${image}" "${reg_image}"
docker push "${reg_image}"
