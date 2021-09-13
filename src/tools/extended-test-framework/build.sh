#!/usr/bin/env bash

set -e pipefail

./gen_server_code.sh extended
./gen_client_code.sh extended

pushd test_conductor && ./build.sh; popd
pushd test_director && ./build.sh;  popd
pushd workload_monitor && ./build.sh; popd


