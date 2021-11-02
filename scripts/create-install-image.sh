#!/usr/bin/env bash

# get the generated deployment files and push them to the CI repo tagged as 'nightly-stable-images'

set -euo pipefail

REGISTRY="ci-registry.mayastor-ci.mayadata.io"
OUTPUT_TAG="nightly-stable"
SCRIPTDIR=$(dirname "$(realpath "$0")")
E2EROOT=$(realpath "$SCRIPTDIR/..")
MAYASTOR_DIR=""
MOAC_DIR=""
MCP_DIR=""

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
  --mcp                      Path to root Mayastor control plane directory
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
    --mcp)
      shift
      MCP_DIR=$1
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

if [ -z "$MOAC_DIR" ] && [ -z "$MCP_DIR" ]; then
    echo "neither moac nor mcp repository directory specified"
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
    && tar cf "$tmpdir/install.tar" scripts/generate-deploy-yamls.sh rpc/mayastor-api/protobuf/mayastor.proto deploy \
    && git rev-parse HEAD > "$workdir/git-revision.mayastor" \
    && git rev-parse --short=12 HEAD >> "$workdir/git-revision.mayastor" \
    && cp -R chart/ "$workdir" \
    && popd

if [ -n "$MOAC_DIR" ]; then
    pushd "${MOAC_DIR}" \
        && mkdir -p "$workdir/csi/moac/crds" \
        && cp crds/* "$workdir/csi/moac/crds/" \
        && git rev-parse HEAD > "$workdir/git-revision.moac" \
        && git rev-parse --short=12 HEAD >> "$workdir/git-revision.moac" \
        && popd
fi

# Note we ensure that $workdir/mcp/bin/kubectl-mayastor is writable,
# to avoid failure in jenkins archiving where we overwrite
# artifacts/install-bundle returned by each parallel run
if [ -n "$MCP_DIR" ]; then
    pushd "${MCP_DIR}" \
        && mkdir -p "$workdir/mcp/scripts" \
        && cp scripts/generate-deploy-yamls.sh "$workdir/mcp/scripts" \
        && cp -R chart "$workdir/mcp" \
        && nix-build -A utils.release.kubectl-plugin \
        && mkdir -p "$workdir/mcp/bin" \
        && cp "$(nix-build -A utils.release.kubectl-plugin --no-out-link)/bin/kubectl-mayastor" "$workdir/mcp/bin" \
        && chmod a+w "$workdir/mcp/bin/kubectl-mayastor" \
        && mkdir -p "$workdir/mcp/control-plane/rest/openapi-specs" \
        && cp control-plane/rest/openapi-specs/* "$workdir/mcp/control-plane/rest/openapi-specs" \
        && git rev-parse HEAD > "$workdir/git-revision.mcp" \
        && git rev-parse --short=12 HEAD >> "$workdir/git-revision.mcp" \
        && popd
    # FIXME: dangling CRD yaml files break deployment using helm
    pushd "$workdir" \
        && rm -f chart/crds/* \
        && popd
    # FIXME: remove moac yaml files which are pulled in from mayastor
    pushd "$workdir" \
        && find chart/ -name moac\*.yaml -print | xargs rm -f \
        && popd
fi

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
