#!/usr/bin/env bash

set -eu

SCRIPTDIR=$(dirname "$(realpath "$0")")
E2EROOT=$(realpath "$SCRIPTDIR/..")
TESTDIR=$(realpath "$SCRIPTDIR/../src")
ARTIFACTSDIR=$(realpath "$SCRIPTDIR/../artifacts")
#reportsdir=$(realpath "$SCRIPTDIR/..")

# removed: pvc_stress_fio temporarily mayastor bugs

#exit values
EXITV_OK=0
EXITV_INVALID_OPTION=1
EXITV_MISSING_OPTION=2
EXITV_FAILED=4
EXITV_FILE_MISMATCH=5
EXITV_FAILED_CLUSTER_OK=255

platform_config_file="hetzner.yaml"
config_file="mayastor_ci_hcloud_e2e_config.yaml"

# Global state variables
#  test configuration state variables
loki_run_id=
device=
registry="ci-registry.mayastor-ci.mayadata.io"
tag="nightly-stable"
#  script state variables
tests=""
profile="default"
on_fail="stop"
uninstall_cleanup="n"
generate_logs=0
logsdir="$ARTIFACTSDIR/logs"
resportsdir="$ARTIFACTSDIR/reports"
mayastor_root_dir=""
policy_cleanup_before="${e2e_policy_cleanup_before:-false}"
profile_test_list=""

declare -A profiles
# List and Sequence of tests.
source "$SCRIPTDIR/test_lists.sh"

help() {
  cat <<EOF
Usage: $0 [OPTIONS]

Options:
  --build_number <number>   Build number, for use when sending Loki markers
  --loki_run_id <Loki run id>  ID string, for use when sending Loki markers
  --device <path>           Device path to use for storage pools.
  --registry <host[:port]>  Registry to pull the mayastor images from. (default: "ci-registry.mayastor-ci.mayadata.io")
                            'dockerhub' means use DockerHub
  --tests <list of tests>   Lists of tests to run, delimited by spaces (default: "$tests")
                            Note: the last 2 tests should be (if they are to be run)
                                - ms_pod_disruption
                                - uninstall
  --profile <continuous|nightly|nightly_full|ondemand|self_ci|soak|validation>
                            Run the tests corresponding to the profile (default: run all tests)
  --resportsdir <path>       Path to use for junit xml test reports (default: repo root)
  --logs                    Generate logs and cluster state dump at the end of successful test run,
                            prior to uninstall.
  --logsdir <path>          Location to generate logs (default: emit to stdout).
  --onfail <stop|uninstall|reinstall|restart>
                            On fail, stop immediately,uninstall, reinstall and continue or restart and continue default($on_fail)
                            Behaviour for "uninstall" only differs if uninstall is in the list of tests (the default).
                            If set to "reinstall" on failure, all resources are cleaned up and mayastor is re-installed.
                            If set to "restart" on failure, all resources are cleaned up and mayastor pods are restarted by deleting.
  --uninstall_cleanup <y|n> On uninstall cleanup for reusable cluster. default($uninstall_cleanup)
  --config                  config name or configuration file default($config_file)
  --platform_config         test platform configuration file default($platform_config_file)
  --tag <name>              Docker image tag of mayastor images (default "$tag")
                            install files are retrieved from the CI registry using the appropriately
                            tagged docker image :- mayadata/install-images
  --mayastor                path to the mayastor source tree to use for testing.
                            If this is specified the install test uses the install yaml files from this tree
                            instead of the tagged image.

Examples:
  $0 --device /dev/nvme0n1 --registry 127.0.0.1:5000 --tag a80ce0c
EOF
}

function setup_profile_testlist {
    for key in "${!profiles[@]}"; do
        if [ "$key" == "$1" ] ; then
            profile_test_list="${profiles[$1]}"
            return 0
        fi
    done
    return 1
}

function set_profile {
    if [ "$profile" == "$1" ]; then
        return
    fi

    # Can only sensibly override the default profile
    if [ "$profile" != "default" ]; then
        echo "--profile and --tests contradict"
        help
        exit $EXITV_INVALID_OPTION
    fi
    profile="$1"
}

# Parse arguments
while [ "$#" -gt 0 ]; do
  case "$1" in
    -m|--mayastor)
      shift
      mayastor_root_dir=$1
      ;;
    -d|--device)
      shift
      device=$1
      ;;
    -r|--registry)
      shift
      if [[ "$1" == 'dockerhub' ]]; then
          registry=''
      else
          registry=$1
      fi
      ;;
    -t|--tag)
      shift
      tag=$1
      ;;
    -T|--tests)
      shift
      set_profile "custom"
      profiles[custom]="$1"
      ;;
    -R|--reportsdir)
      shift
      resportsdir="$1"
      ;;
    -h|--help)
      help
      exit $EXITV_OK
      ;;
    --build_number) # TODO remove this option
      shift
      loki_run_id="$1"
      ;;
    --loki_run_id)
      shift
      loki_run_id="$1"
      ;;
    --logs)
      generate_logs=1
      ;;
    --logsdir)
      shift
      logsdir="$1"
      if [[ "${logsdir:0:1}" == '.' ]]; then
          logsdir="$PWD/$logsdir"
      fi
      ;;
    --profile)
      shift
      set_profile "$1"
      ;;
    --onfail)
        shift
        case $1 in
            uninstall)
                on_fail=$1
                ;;
            stop)
                on_fail=$1
                ;;
            reinstall|continue)
                on_fail="reinstall"
                policy_cleanup_before='true'
                ;;
            restart)
                on_fail=$1
                policy_cleanup_before='true'
                ;;
            *)
                echo "invalid option for --onfail"
                help
                exit $EXITV_INVALID_OPTION
        esac
      ;;
    --uninstall_cleanup)
        shift
        case $1 in
            y|n)
                uninstall_cleanup=$1
                ;;
            *)
                echo "invalid option for --uninstall_cleanup"
                help
                exit $EXITV_INVALID_OPTION
        esac
      ;;
    --config)
        shift
        config_file="$1"
        ;;
    --platform_config)
        shift
        platform_config_file="$1"
        ;;
    *)
      echo "Unknown option: $1"
      help
      exit $EXITV_INVALID_OPTION
      ;;
  esac
  shift
done

export loki_run_id="$loki_run_id" # can be empty string

if [ -z "$mayastor_root_dir" ]; then
    if ! "$SCRIPTDIR/extract-install-image.sh" --alias-tag "$tag"
    then
        echo "Unable to extract install files for $tag"
        exit $EXITV_INVALID_OPTION
    fi
    export mayastor_root_dir="$ARTIFACTSDIR/install/$tag"
fi
export e2e_mayastor_root_dir=$mayastor_root_dir

# grpc proto compatibility check
if ! cmp src/common/mayastorclient/grpc/mayastor.proto "$mayastor_root_dir/rpc/proto/mayastor.proto"
then
    echo "src/common/mayastorclient/grpc/mayastor.proto != $mayastor_root_dir/rpc/proto/mayastor.proto"
    echo "see src/common/mayastorclient/grpc/README.md"
# 17/06/2021 temporarily mutate the check into warning
# to properly fix we need to generate the client code from the proto,
# and for that to work we need and install bundle which packages the proto
# file from mayastor.
#    exit $EXITV_FILE_MISMATCH
    echo "WARNING proto files mismatch: src/common/mayastorclient/grpc/mayastor.proto != $mayastor_root_dir/rpc/proto/mayastor.proto"
fi

# CRD compatibility checks
if ! cmp src/common/custom_resources/mayastorvolume.yaml "$mayastor_root_dir/csi/moac/crds/mayastorvolume.yaml"
then
    echo "src/common/custom_resources/mayastorvolume.yaml != $mayastor_root_dir/csi/moac/crds/mayastorvolume.yaml"
    echo "see src/common/custom_resources/README.md"
    exit $EXITV_FILE_MISMATCH
fi

if ! cmp src/common/custom_resources/mayastorpool.yaml "$mayastor_root_dir/csi/moac/crds/mayastorpool.yaml"
then
    echo "src/common/custom_resources/mayastorpool.yaml != $mayastor_root_dir/csi/moac/crds/mayastorpool.yaml"
    echo "see src/common/custom_resources/README.md"
# 24/06/2021 temporarily mutate the check into warning
# to properly fix we need to generate the client code from the proto,
# and for that to work we need and install bundle which packages the proto
# file from mayastor.
#    exit $EXITV_FILE_MISMATCH
    echo "WARNING CRD yaml mismatch: src/common/custom_resources/mayastorpool.yaml != $mayastor_root_dir/csi/moac/crds/mayastorpool.yaml"
fi

if ! cmp src/common/custom_resources/mayastornode.yaml "$mayastor_root_dir/csi/moac/crds/mayastornode.yaml"
then
    echo "src/common/custom_resources/mayastornode.yaml != $mayastor_root_dir/csi/moac/crds/mayastornode.yaml"
    echo "see src/common/custom_resources/README.md"
    exit $EXITV_FILE_MISMATCH
fi


if [ -z "$device" ]; then
  echo "Device for storage pools must be specified"
  help
  exit $EXITV_MISSING_OPTION
fi
export e2e_pool_device=$device

if [ -n "$tag" ]; then
  export e2e_image_tag="$tag"
fi

export e2e_docker_registry="$registry" # can be empty string
export e2e_root_dir="$E2EROOT"

case "$profile" in
  nightlyfull|nightly_full)
    profile="nightly_full"
    echo "Overriding config file to nightly_full_config.yaml"
    config_file="nightly_full_config.yaml"
    ;;
  selfci|self_ci)
    profile="self_ci"
    echo "Overriding config file to selfci_config.yaml"
    config_file="selfci_config.yaml"
    ;;
  soak)
    echo "Overriding config file to soak_config.yaml"
    config_file="soak_config.yaml"
    ;;
esac

if ! setup_profile_testlist "$profile" ; then
    echo "Unknown profile: $profile"
    help
    exit $EXITV_INVALID_OPTION
fi

if [ "$profile" != "custom" ] ; then
    tests="install $profile_test_list uninstall"
else
    tests="$profile_test_list"
fi

export e2e_reports_dir="$resportsdir"

if [ "$uninstall_cleanup" == 'n' ] ; then
    export e2e_uninstall_cleanup=0
else
    export e2e_uninstall_cleanup=1
fi

mkdir -p "$ARTIFACTSDIR"
mkdir -p "$resportsdir"
mkdir -p "$logsdir"

test_failed=0

# Run go test in directory specified as $1 (relative path)
function runGoTest {
    pushd "$TESTDIR"
    echo "Running go test in $PWD/\"$1\""
    if [ -z "$1" ] || [ ! -d "$1" ]; then
        echo "Unable to locate test directory  $PWD/\"$1\""
        popd
        return 1
    fi

    cd "$1"
    if ! go test -v . -ginkgo.v -ginkgo.progress -timeout 0; then
        generate_logs=1
        popd
        return 1
    fi

    popd
    return 0
}

function emitLogs {
    if [ -z "$1" ]; then
        logPath="$logsdir"
    else
        logPath="$logsdir/$1"
    fi
    mkdir -p "$logPath"
    if ! "$SCRIPTDIR/e2e-cluster-dump.sh" --destdir "$logPath" ; then
        # ignore failures in the dump script
        :
    fi
    unset logPath
}

# Check if $2 is in $1
contains() {
    [[ $1 =~ (^|[[:space:]])$2($|[[:space:]]) ]] && return 0  || return 1
}

export e2e_config_file="$config_file"
export e2e_platform_config_file="$platform_config_file"
export e2e_policy_cleanup_before="$policy_cleanup_before"

#preprocess tests so that command line can use commas as delimiters
tests=${tests//,/ }

echo "Environment:"
echo "    e2e_mayastor_root_dir=$e2e_mayastor_root_dir"
echo "    loki_run_id=$loki_run_id"
echo "    e2e_root_dir=$e2e_root_dir"
echo "    e2e_pool_device=$e2e_pool_device"
echo "    e2e_image_tag=$e2e_image_tag"
echo "    e2e_docker_registry=$e2e_docker_registry"
echo "    e2e_reports_dir=$e2e_reports_dir"
echo "    e2e_uninstall_cleanup=$e2e_uninstall_cleanup"
echo "    e2e_config_file=$e2e_config_file"
echo "    e2e_platform_config_file=$e2e_platform_config_file"
echo "    e2e_policy_cleanup_before=$e2e_policy_cleanup_before"
echo ""
echo "Script control settings:"
echo "    profile=$profile"
echo "    on_fail=$on_fail"
echo "    uninstall_cleanup=$uninstall_cleanup"
echo "    generate_logs=$generate_logs"
echo "    logsdir=$logsdir"
echo ""
echo "list of tests: $tests"

for testname in $tests; do
  # defer uninstall till after other tests have been run.
  if [ "$testname" != "uninstall" ] ;  then
      if ! runGoTest "$testname" ; then
          echo "Test \"$testname\" FAILED!"
          test_failed=1
          emitLogs "$testname"
          if [ "$testname" != "install" ] ; then
              if [ "$on_fail" == "restart" ] ; then
                  echo "Attempting to continue by cleaning up and restarting mayastor pods........"
                  if ! runGoTest "tools/restart" ; then
                      echo "\"restart\" failed"
                      exit $EXITV_FAILED
                  fi
              elif [ "$on_fail" == "reinstall" ] ; then
                  echo "Attempting to continue by cleaning up and re-installing........"
                  runGoTest "tools/cleanup"
                  if ! runGoTest "uninstall"; then
                      echo "uninstall failed, abandoning attempt to continue"
                      exit $EXITV_FAILED
                  fi
                  if ! runGoTest "install"; then
                      echo "(re)install failed, abandoning attempt to continue"
                      exit $EXITV_FAILED
                  fi
              else
                  break
              fi
          fi
      fi
  fi
done

if [ "$generate_logs" -ne 0 ]; then
    emitLogs ""
fi

if [ "$test_failed" -ne 0 ] && [ "$on_fail" == "stop" ]; then
    echo "At least one test FAILED!"
    exit $EXITV_FAILED
fi

# Always run uninstall test if specified
if contains "$tests" "uninstall" ; then
    if ! runGoTest "uninstall" ; then
        echo "Test \"uninstall\" FAILED!"
        test_failed=1
        emitLogs "uninstall"
    elif  [ "$test_failed" -ne 0 ] ; then
        # tests failed, but uninstall was successful
        # so cluster is reusable
        echo "At least one test FAILED! Cluster is usable."
        exit $EXITV_FAILED_CLUSTER_OK
    fi
fi


if [ "$test_failed" -ne 0 ] ; then
    echo "At least one test FAILED!"
    exit $EXITV_FAILED
fi

echo "All tests have PASSED!"
exit $EXITV_OK
