#!/usr/bin/env bash

set -eu

SCRIPTDIR=$(dirname "$(realpath "$0")")
E2EROOT=$(realpath "$SCRIPTDIR/..")
TESTDIR=$(realpath "$SCRIPTDIR/../src")
ARTIFACTSDIR=$(realpath "$SCRIPTDIR/../artifacts")
#reportsdir=$(realpath "$SCRIPTDIR/..")

# List and Sequence of tests.
#tests="install basic_volume_io csi replica rebuild node_disconnect/replica_pod_remove uninstall"
# Restrictions:
#   1. resource_check MUST follow csi
#       resource_check is a follow up check for the 3rd party CSI test suite.
#   2. ms_pod_disruption SHOULD be the last test before uninstall
#
DEFAULT_TESTS="install basic_volume_io csi resource_check replica rebuild ms_pod_disruption uninstall"
ONDEMAND_TESTS="install basic_volume_io csi resource_check uninstall"
EXTENDED_TESTS="install basic_volume_io csi resource_check rebuild io_soak multiple_vols_pod_io pool_modify ms_pod_disruption uninstall" # replica removed
CONTINUOUS_TESTS="install basic_volume_io csi resource_check rebuild io_soak ms_pod_disruption uninstall" # replica removed
SELF_CI_TESTS="install basic_volume_io csi resource_check io_soak multiple_vols_pod_io pvc_stress_fio pool_modify ms_pod_disruption uninstall" # replica removed, # rebuild removed
SOAK_TESTS="install io_soak uninstall"


#exit values
EXITV_OK=0
EXITV_INVALID_OPTION=1
EXITV_MISSING_OPTION=2
EXITV_FAILED=4
EXITV_FAILED_CLUSTER_OK=255

# Global state variables
#  test configuration state variables
loki_run_id=
device=
registry="ci-registry.mayastor-ci.mayadata.io"
tag="nightly"
#  script state variables
tests=""
custom_tests=""
profile="default"
on_fail="stop"
uninstall_cleanup="n"
generate_logs=0
logsdir="$ARTIFACTSDIR/logs"
resportsdir="$ARTIFACTSDIR/reports"
mayastor_root_dir=""

help() {
  cat <<EOF
Usage: $0 [OPTIONS]

Options:
  --build_number <number>   Build number, for use when sending Loki markers
  --loki_run_id <Loki run id>  ID string, for use when sending Loki markers
  --device <path>           Device path to use for storage pools.
  --registry <host[:port]>  Registry to pull the mayastor images from. (default: "ci-registry.mayastor-ci.mayadata.io")
                            'dockerhub' means use DockerHub
  --tag <name>              Docker image tag of mayastor images (default "nightly")
  --tests <list of tests>   Lists of tests to run, delimited by spaces (default: "$tests")
                            Note: the last 2 tests should be (if they are to be run)
                                - ms_pod_disruption
                                - uninstall
  --profile <continuous|extended|ondemand|self_ci|soak>
                            Run the tests corresponding to the profile (default: run all tests)
  --resportsdir <path>       Path to use for junit xml test reports (default: repo root)
  --logs                    Generate logs and cluster state dump at the end of successful test run,
                            prior to uninstall.
  --logsdir <path>          Location to generate logs (default: emit to stdout).
  --onfail <stop|uninstall|continue>
                            On fail, stop immediately,uninstall or continue default($on_fail)
                            Behaviour for "uninstall" only differs if uninstall is in the list of tests (the default).
                            If set to "continue" on failure, all resources are cleaned up and mayastor is re-installed.
  --uninstall_cleanup <y|n> On uninstall cleanup for reusable cluster. default($uninstall_cleanup)
  --config                  config name or configuration file default(test/e2e/configurations/ci_e2e_config.yaml)
  --mayastor                path to the mayastor source tree to use for testing.
                            This is required so that the install test uses the yaml files as defined for that
                            revision of mayastor under test.

Examples:
  $0 --device /dev/nvme0n1 --registry 127.0.0.1:5000 --tag a80ce0c
EOF
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
      custom_tests="$1"
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
      profile="$1"
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
            continue)
                on_fail=$1
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
        export e2e_config_file="$1"
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
    echo "Root directory for mayastor is required"
    exit $EXITV_MISSING_OPTION
fi
export e2e_mayastor_root_dir=$mayastor_root_dir

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

if [ -n "$custom_tests" ]; then
  if [ "$profile" != "default" ]; then
    echo "cannot specify --profile with --tests"
    help
    exit $EXITV_INVALID_OPTION
  fi
  profile="custom"
fi

case "$profile" in
  continuous)
    tests="$CONTINUOUS_TESTS"
    ;;
  extended)
    tests="$EXTENDED_TESTS"
    ;;
  ondemand)
    tests="$ONDEMAND_TESTS"
    ;;
  custom)
    tests="$custom_tests"
    ;;
  default)
    tests="$DEFAULT_TESTS"
    ;;
  selfci|self_ci)
    tests="$SELF_CI_TESTS"
    echo "Overriding config file to selfci_config.yaml"
    export e2e_config_file="selfci_config.yaml"
    ;;
  soak)
    tests="$SOAK_TESTS"
    echo "Overriding config file to soak_config.yaml"
    export e2e_config_file="soak_config.yaml"
    ;;
  *)
    echo "Unknown profile: $profile"
    help
    exit $EXITV_INVALID_OPTION
    ;;
esac

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
          if [ "$on_fail" == "continue" ] && [ "$testname" != "install" ] ; then
              # continue is only possible if install was successful
              echo "Attempting to continue....., cleanup"
              if ! runGoTest "cleanup" ; then
                  echo "\"cleanup\" failed"
                  exit $EXITV_FAILED
              fi
              echo "Mayastor pods were restarted.. re-installing"
              if ! runGoTest "uninstall"; then
                  echo "uninstall failed, abandoning attempt to continue"
                  echo $EXITV_FAILED
              fi
              if ! runGoTest "install"; then
                  echo "(re)install failed, abandoning attempt to continue"
                  echo $EXITV_FAILED
              fi
          else
              break
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
