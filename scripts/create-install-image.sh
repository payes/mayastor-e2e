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
  --product                  Product key [mayastor, bolt]
Examples:
  $(basename "$0") --registry 127.0.0.1:5000 --alias-tag customized-tag --product bolt
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
    --product)
      shift
      case $1 in
          mayastor)
             regroot='mayadata'
             ;;
          bolt)
             regroot='datacore'
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

if [ -z "$MAYASTOR_DIR" ] ; then
    echo "mayastor repository directory not specified"
    exit 127
fi

if [ -z "$MCP_DIR" ]; then
    echo "mcp repository directory not specified"
    exit 127
fi

image="$regroot/install-images:${OUTPUT_TAG}"
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
    pushd "${MCP_DIR}"
        build_info+=("\"mayastor-control-plane-revision\": \"$(git rev-parse HEAD)\"")
        build_info+=("\"mayastor-control-plane-short-revision\": \"$(git rev-parse --short=12 HEAD)\"")
        # for release 1 compatibility
        mkdir -p "$workdir/mcp/scripts" || exit 1
        cp scripts/*.sh "$workdir/mcp/scripts" || exit 1
        # post release 1
        mkdir -p "$workdir/mcp/scripts/deploy" || exit 1
        if [ -d "scripts/deploy" ]; then
            cp scripts/deploy/* "$workdir/mcp/scripts/deploy" || exit 1
        fi
        cp -R chart "$workdir/mcp" || exit 1
        mkdir -p "$workdir/mcp/control-plane/rest/openapi-specs" || exit 1
        cp control-plane/rest/openapi-specs/* "$workdir/mcp/control-plane/rest/openapi-specs" || exit 1
        mkdir -p "$workdir/mcp/bin" || exit 1
        # try current rune for kubectl plugin build
        if ! plugin_nix_dir=$(nix-build -A utils.release.linux-musl.kubectl-plugin --no-out-link); then
        # try previous rune for  kubectl plugin build
            if ! plugin_nix_dir=$(nix-build -A utils.release.kubectl-plugin --no-out-link); then
                echo "failed to build kubectl plugin"
                exit 1
            fi
        fi
        cp "$plugin_nix_dir/bin/kubectl-mayastor" "$workdir/mcp/bin" || exit 1
        chmod a+wx "$workdir/mcp/bin/kubectl-mayastor" || exit 1
        find . -print
    popd
    # Release 1 compatibility {
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
