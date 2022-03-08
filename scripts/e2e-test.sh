#!/usr/bin/env bash

set -eu

SCRIPTDIR=$(dirname "$(realpath "$0")")
E2EROOT=$(realpath "$SCRIPTDIR/..")
TESTDIR=$(realpath "$SCRIPTDIR/../src")
ARTIFACTSDIR=$(realpath "$SCRIPTDIR/../artifacts")
CONFIGSDIR=$(realpath "$SCRIPTDIR/../configurations/products/")

# removed: pvc_stress_fio temporarily mayastor bugs

#exit values
EXITV_OK=0
EXITV_INVALID_OPTION=1
EXITV_MISSING_OPTION=2
EXITV_FAILED=4
EXITV_FILE_MISMATCH=5
EXITV_CRD_GO_GEN=6
EXITV_VERSION_MISMATCH=7
EXITV_MISSING_KUBECTL_PLUGIN=8
EXITV_FAILED_CLUSTER_OK=255

platform_config_file="hetzner.yaml"
config_file="hcloudci_config.yaml"
mayastor_version=""

# Global state variables
#  test configuration state variables
loki_run_id=
loki_test_label=
device=
session="$(date +%Y%m%d-%H%M%S-)$(uuidgen -r)"
registry="ci-registry.mayastor-ci.mayadata.io"
tag="nightly-stable"
#  script state variables
tests=""
profile="default"
on_fail="stop"
uninstall_cleanup="n"
generate_logs=0
logsdir="$ARTIFACTSDIR/logs"
reportsdir="$ARTIFACTSDIR/reports"
coveragedir="$ARTIFACTSDIR/coverage/data"
mayastor_root_dir=""
policy_cleanup_before="${e2e_policy_cleanup_before:-false}"
profile_test_list=""
ssh_identity=""
grpc_code_gen=
crd_code_gen=

declare -A profiles
# List and Sequence of tests.
source "$SCRIPTDIR/test_lists.sh"

help() {
  cat <<EOF
Usage: $0 [OPTIONS]

Options:
  --build_number <number>   Build number, for use when sending Loki markers
  --loki_run_id <Loki run id>  ID string, for use when sending Loki markers
  --loki_test_label <Loki custom test label> Test label value, for use when sending Loki markers
  --device <path>           Device path to use for storage pools.
  --product                 Product which needs to be validated against e2e
  --registry <host[:port]>  Registry to pull the mayastor images from. (default: "ci-registry.mayastor-ci.mayadata.io")
                            'dockerhub' means use DockerHub
  --tests <list of tests>   Lists of tests to run, delimited by spaces (default: "$tests")
                            Note: the last test should be uninstall (if it is to be run)
  --profile <c1|nightly-stable|ondemand|self_ci|staging|validation>
                            Run the tests corresponding to the profile (default: run all tests)
  --reportsdir <path>       Path to use for junit xml test reports (default: repo root)
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
  --session                 session name, adds a subdirectory with session name to artifacts, logs and reports
                            directories to facilitate concurrent execution of test runs (default timestamp-uuid)
  --version                 Mayastor version, 0 => MOAC, > 1 => restful control plane
  --grpc_code_gen <true|false>
                            On true, grpc server and clinet code will be generated
                            On false, grpc server and clinet code will not be generated
  --crd_code_gen <true|false>
                            On true, custom resource clinet code will be generated
                            On false, custom resource clinet code will not be generated
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
    -g|--grpc_code_gen)
      shift
      grpc_code_gen="$1"
      ;;
    -c|--crd_code_gen)
      shift
      crd_code_gen="$1"
      ;;
    -T|--tests)
      shift
      set_profile "custom"
      profiles[custom]="$1"
      ;;
    -R|--reportsdir)
      shift
      reportsdir="$1"
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
    --loki_test_label)
      shift
      loki_test_label="$1"
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
    -p|--product)
        shift
        product="$1"
        ;;
    --session)
        shift
        session="$1"
        ;;
    --ssh_identity)
        shift
        ssh_identity="$1"
        ;;
    --version)
        shift
            case "$1" in
                0|1)
                    mayastor_version=$1
                    ;;
                *)
                    echo "Unknown control plane: $1"
                    help
                    exit $EXITV_INVALID_OPTION
                    ;;
            esac
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
export loki_test_label="$loki_test_label"

if [ -z "$session" ]; then
    sessiondir="$ARTIFACTSDIR"
else
    sessiondir="$ARTIFACTSDIR/sessions/$session"
    logsdir="$logsdir/$session"
    reportsdir="$reportsdir/$session"
    coveragedir="$coveragedir/$session"
fi

if [ -z "$mayastor_root_dir" ]; then
    mkdir -p "$sessiondir"
    export mayastor_root_dir="$ARTIFACTSDIR/install-bundle/$tag"
    mkdir -p "$mayastor_root_dir"
    if ! "$SCRIPTDIR/extract-install-image.sh" --alias-tag "$tag" --installroot "$mayastor_root_dir"
    then
        echo "Unable to extract install files for $tag"
        exit $EXITV_INVALID_OPTION
    fi
    kbctl_plugin=$(find "$mayastor_root_dir" -name kubectl-mayastor)
    if [ -n "$kbctl_plugin" ]; then
        kbctl_plugin_dir=$(dirname "$kbctl_plugin")
        echo "Found kubectl plugin $kbctl_plugin"
        export e2e_kubectl_plugin_dir=$kbctl_plugin_dir
    else
        echo "Did not find mayastor kubectl-plugin"
        exit $EXITV_MISSING_KUBECTL_PLUGIN
    fi
fi
export e2e_mayastor_root_dir=$mayastor_root_dir
export e2e_session_dir=$sessiondir


if [ -z "$device" ]; then
  echo "Device for storage pools must be specified"
  help
  exit $EXITV_MISSING_OPTION
fi
export e2e_pool_device=$device

if [ -z "$product" ]; then
  echo "Product (Mayastor/Bolt) must be specified"
  help
  exit $EXITV_MISSING_OPTION
fi

export e2e_product_config_yaml="$CONFIGSDIR/${product}.yaml"

if [ -n "$tag" ]; then
  export e2e_image_tag="$tag"
fi

export e2e_docker_registry="$registry" # can be empty string
export e2e_root_dir="$E2EROOT"

case "$profile" in
  nightly|nightly-stable)
    echo "Overriding policy_cleanup_before=true"
    policy_cleanup_before="true"
    ;;
  selfci|self_ci)
    profile="self_ci"
    echo "Overriding config file to selfci_config.yaml"
    config_file="selfci_config.yaml"
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

export e2e_reports_dir="$reportsdir"

if [ "$uninstall_cleanup" == 'n' ] ; then
    export e2e_uninstall_cleanup=0
else
    export e2e_uninstall_cleanup=1
fi

mkdir -p "$sessiondir"
mkdir -p "$reportsdir"
mkdir -p "$logsdir"

kubectl get nodes -o yaml > "$reportsdir/k8s_nodes.yaml"

test_failed=0

# Generate gRPC server and client code from mayastor.proto file
if [ -n "$grpc_code_gen" -a "$grpc_code_gen"="true" ]; then
  echo "Generating gRPC client and server code: $PWD"
  #Update mayastor.proto file with option go_package = "github.com/openebs/mayastor-api/protobuf";
  path="$mayastor_root_dir/rpc/mayastor-api/protobuf"
  sed -i '/syntax = "proto3";/a option go_package = "github.com/openebs/mayastor-api/protobuf";' "$path/mayastor.proto"
  cmd="cd $mayastor_root_dir/rpc/mayastor-api && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative protobuf/mayastor.proto"
  echo "Command to execute in nix shell: $cmd"
  nix-shell --run "$cmd" ./ci.nix
  # Copy mayastor_grpc.pb.go and mayastor.pb.go to /src/common/mayastorclient/protobuf
  cp "$path/mayastor_grpc.pb.go" "$path/mayastor.pb.go" ./src/common/mayastorclient/protobuf
fi

# Generate CR client code from mayastorpoolcrd.yaml file
if [ -n "$crd_code_gen" -a "$crd_code_gen"="true" ]; then
  echo "Generating CR client code: $PWD"
  path="$mayastor_root_dir/mcp/chart/templates/mayastorpoolcrd.yaml"
  cmd="./scripts/genGoCrdTypes.py $path"
  echo "Command to execute in nix shell: $cmd"
  nix-shell --run "$cmd" ./ci.nix
fi

# Run go test in directory specified as $1 (relative path)
# maximum test runtime is 120 minutes
function runGoTest {
    pushd "$TESTDIR"
    echo "Running go test in $PWD/\"$1\""
    if [ -z "$1" ] || [ ! -d "$1" ]; then
        echo "Unable to locate test directory  $PWD/\"$1\""
        popd
        return 1
    fi

    cd "$1"
    # timeout test run after 3 hours
    if ! go test -v . -ginkgo.v -ginkgo.progress -timeout 120m; then
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

# The values of e2e_control_plane must match control plane values
# in src/common/constants.go
version_from_bundle="1.0.0"

if [ -n "$mayastor_version" ]; then
    if [ "$version_from_bundle" != "$mayastor_version" ]; then
        echo "Install bundle version ($version_from_bundle) does not match $mayastor_version"
        exit $EXITV_VERSION_MISMATCH
    fi
else
    mayastor_version="$version_from_bundle"
fi

export e2e_mayastor_version=$mayastor_version
export e2e_config_file="$config_file"
export e2e_platform_config_file="$platform_config_file"
export e2e_policy_cleanup_before="$policy_cleanup_before"

#preprocess tests so that command line can use commas as delimiters
tests=${tests//,/ }

echo "Environment:"
echo "    e2e_session_dir=$e2e_session_dir"
echo "    e2e_mayastor_root_dir=$e2e_mayastor_root_dir"
echo "    loki_run_id=$loki_run_id"
echo "    loki_test_label=$loki_test_label"
echo "    e2e_root_dir=$e2e_root_dir"
echo "    e2e_pool_device=$e2e_pool_device"
echo "    e2e_product_config_yaml=$e2e_product_config_yaml"
echo "    e2e_image_tag=$e2e_image_tag"
echo "    e2e_docker_registry=$e2e_docker_registry"
echo "    e2e_reports_dir=$e2e_reports_dir"
echo "    e2e_uninstall_cleanup=$e2e_uninstall_cleanup"
echo "    e2e_config_file=$e2e_config_file"
echo "    e2e_platform_config_file=$e2e_platform_config_file"
echo "    e2e_policy_cleanup_before=$e2e_policy_cleanup_before"
echo "    e2e_mayastor_version=$e2e_mayastor_version"
echo ""
echo "Script control settings:"
echo "    profile=$profile"
echo "    on_fail=$on_fail"
echo "    uninstall_cleanup=$uninstall_cleanup"
echo "    generate_logs=$generate_logs"
echo "    logsdir=$logsdir"
echo ""
echo "list of tests: $tests"

if contains "$tests" "install" ; then
    if ! "$SCRIPTDIR/remote-coverage-files.py" --clear --identity_file "$ssh_identity" ; then
        echo "***************************** failed to clear coverage files"
    fi
fi

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
    else
        if ! "$SCRIPTDIR/remote-coverage-files.py" --get --path "$coveragedir" --identity_file "$ssh_identity" ; then
            echo "Failed to retrieve coverage files"
        fi
        if  [ "$test_failed" -ne 0 ] ; then
            # tests failed, but uninstall was successful
            # so cluster is reusable
            echo "At least one test FAILED! Cluster is usable."
            exit $EXITV_FAILED_CLUSTER_OK
        fi
    fi
fi


if [ "$test_failed" -ne 0 ] ; then
    echo "At least one test FAILED!"
    exit $EXITV_FAILED
fi

echo "All tests have PASSED!"
exit $EXITV_OK
