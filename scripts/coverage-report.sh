#!/usr/bin/env bash

help() {
  cat <<EOF
This script generates coverage reports for mayastor

Usage: $(basename "$0") [OPTIONS]

Options:
  -h, --help            Display this text
  --archive             path to jenkins run archive
  --tag                 tag of the build (required with --archive)
  --data_dir            path to coverage dir, contains .profraw files
  --binaries_dir        path to mayastor binaries
  --report_dir          path to generate the coverage report
  --mayastor            path to mayastor sources
  --control-plane       path to mayastor control plane sources

Either --archive or --data_dir and --binaries_dir can be specified,
but no both sets of options.
EOF
}

# Cursory check that CWD is mayastor/control plane repo root.
if [ -f shell.nix ] ;
then
    # The "correct" way to run this script is from
    # without nix
    if [[ -z $IN_NIX_SHELL ]];
    then
        nix-shell -p rustup zlib --run "$0 $*"
        exit
    fi
    set -euxo pipefail
    work_dir=''
    tmpdir=''
    archive=''
    tag=''
    data_dir=''
    binaries_dir=''
    report_dir=''
    mayastor_srcs=''
    mcp_srcs=''

    # Parse arguments
    while [ "$#" -gt 0 ]; do
      case $1 in
        -h|--help)
          help
          exit 0
          shift
          ;;
        -a|--archive)
          shift
          archive=$1
          shift
          ;;
        -t|--tag)
          shift
          tag=$1
          shift
          ;;
        -d|--data_dir)
          shift
          data_dir=$1
          shift
          ;;
        -b|--binaries_dir)
          shift
          binaries_dir=$1
          shift
          ;;
        -r|--report_dir)
          shift
          report_dir=$1
          shift
          ;;
        -w|--workspace)
          shift
          work_dir=$1
          shift
          ;;
        -M|--mayastor)
          shift
          mayastor_srcs=$1
          shift
          ;;
        -C|--control-plane)
          shift
          mcp_srcs=$1
          shift
          ;;
        *)
          echo "Unknown option: $1"
          exit 1
          ;;
      esac
    done
    tmpdir="${work_dir}"

    if [ -z "${report_dir}" ] \
        || [ -z "${mayastor_srcs}" ] \
        || [ -z "${mcp_srcs}" ]; then
        help
        exit 2
    fi

    if [ -n "${archive}" ]; then
        if [ -n "${data_dir}" ] \
            || [ -n "${binaries_dir}" ]; then
            echo "--archive cannot be used with --data_dir or --binaries_dir"
            exit 3
        fi
        if [ -z "${tag}" ]; then
            echo "--archive requires --tag"
            exit 4
        fi
        if [ -z "${tmpdir}" ]; then
            tmpdir=$(mktemp -d)
        fi

        data_dir="${tmpdir}/archive/artifacts/coverage"
        binaries_dir="${tmpdir}/archive/artifacts/binaries/${tag}"

        unzip -q "$archive" -d "$tmpdir"
    else
        if [ -z "${data_dir}" ] \
            || [ -z "${binaries_dir}" ]; then
            echo "--data_dir and --binaries_dir must be used in lieu of --archive"
            exit 4
        fi
    fi

    if [ -z "${tmpdir}" ]; then
        tmpdir=$(mktemp -d)
    fi

    cargo_dir="${tmpdir}/cargo"

    if [ ! -d "${data_dir}" ]; then
        echo  "${data_dir} does not exist"
        exit 5
    fi

    if [ ! -d "${binaries_dir}" ]; then
        echo  "${binaries_dir} does not exist"
        exit 6
    fi

    rm -rf "{$report_dir}"
    mkdir -p "${report_dir}/mayastor"
    mkdir -p "${report_dir}/mayastor-control-plane"

    rustup toolchain install nightly-2021-06-22
    rustup component add llvm-tools-preview
    cargo +nightly-2021-06-22 install --root "${cargo_dir}"  grcov
    export PATH="$PATH:${cargo_dir}/bin"

    find "$data_dir" -type f -name mayastor\*.profraw -print0 | xargs -0 grcov -s "${mayastor_srcs}" --binary-path "${binaries_dir}" -t html --branch --ignore-not-existing -o "$report_dir/mayastor"

    find "$data_dir" -type f -not -name mayastor\*.profraw -print0 | xargs -0 grcov -s "${mcp_srcs}" --binary-path "${binaries_dir}" -t html --branch --ignore-not-existing -o "$report_dir/mayastor-control-plane"

    if [ -z "${work_dir}" ] ; then
        rm -rf "$tmpdir"
    fi
else
    echo "cd Mayastor (repo) and then run this script"
fi
