#!/usr/bin/env bash

# script to output to stdout the list of tests
# defined by the given profile as a space-separated string

SCRIPTDIR="$(dirname "$(realpath "$0")")"
set -eu

if [ "$#" != "1" ]; then
  echo "usage $0 <profile>"
  exit 1
fi

declare -A profiles

# import the set of profiles
source "$SCRIPTDIR/test_lists.sh"

if ! test "${profiles[$1]+isset}"; then
  echo "profile $1 not found"
  exit 1
fi

# hack to ensure csi and resource_check are tested together
set=${profiles[$1]}
outputset=""
for string in $set; do
   if [ "${outputset}" == "" ]; then
     outputset="${string}"
   elif [ "$string" == "resource_check" ]; then
     outputset="${outputset},${string}"
   else
     outputset="${outputset} ${string}"
   fi
done

echo "${outputset}"

