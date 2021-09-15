#!/usr/bin/env bash
set -euo pipefail

bin_dir="."
tag=""
registry=""

while [ "$#" -gt 0 ]; do
  case "$1" in
    -o|--outputdir)
      shift
      bin_dir=$1
      ;;
    -t|--tag)
      shift
      tag=$1
      ;;
    -r|--registry)
      shift
      registry=$1
      ;;
    *)
      echo "Unknown option: $1"
      help
      exit 1
      ;;
  esac
  shift
done


rm -rf /tmp/Mayastor
mkdir -p /tmp/Mayastor/mayastor

# FIXME this script works with coverage builds where
# the same image is used for mayastor and mayastor-csi
image="${registry}/mayadata/mayastor:$tag"
echo "$bin_dir"
echo "$image"
contnr=$(docker create "${image}")
docker cp "${contnr}:/bin/mayastor" "$bin_dir"
docker cp "${contnr}:/bin/mayastor-csi" "$bin_dir"
docker rm "${contnr}"
