#!/usr/bin/env bash
#
# Usage: cd src/<testname> &&
#
USAGE="Usage: $0 <path_to_test_dir> [path_to_kubectl_mayastor]"

if [ -z "$KUBECONFIG" ]; then
    echo "KUBECONFIG not defined, tests are unable to access the test cluster"
    exit 1
fi

if [ "$1" == "-h" ] || [ -z "$1" ]; then
    echo "$USAGE"
    exit 1
fi

cd "$1" || exit 1
shift

if [ -z "$1" ]; then
    if ! tmp=$(which kubectl-mayastor); then
        :
    fi
    if [ -z "$tmp" ]; then
        echo "$USAGE"
        exit 1
    else
        echo "NOTE: Found and using $tmp"
        kcm="$tmp"
    fi
else
    kcm="$1"
fi

if [ -f "$kcm" ]; then
    echo "Checking $kcm get volumes"
    "$kcm" get volumes
else
    echo "$kcm is not a file"
    echo "$USAGE"
    exit 1
fi

kcdir=$(dirname "$kcm")
export e2e_kubectl_plugin_dir="${kcdir}"
# fail quick => if first It clause failed attempt to skip
# subsequent It clauses - works if on failure,
# resources created in It clause are left "dangling"
export e2e_fail_quick="true"
go test -v . -ginkgo.v -ginkgo.progress -timeout 0
