#!/usr/bin/env bash

SCRIPTDIR=$(dirname "$(realpath "$0")")
GOSRCDIR=$(realpath "$SCRIPTDIR/../src")
exitv=0

reformat=1
if [ "$1" == "-n" ]; then
    reformat=0
    shift
fi

if [ -n "$*" ]; then
    # go fmt
    unformatted=$(gofmt -l "$@")
fi

if [ $reformat -ne 0 ]; then
    for file in $unformatted; do
        gofmt -w "$file"
        echo "Reformatted $file"
    done
else
    if [ -n "$unformatted" ]; then
        printf "Please run\n\tgofmt -w %s\n" "$unformatted"
        exitv=1
    fi
fi

if golangci-lint > /dev/null 2>&1 ; then
    cd "$GOSRCDIR" || exit 1
# running golangci-lint on individual files throws resolve errors
# for example undeclared name
#    for file in "$@"; do
#        relfile=${file//src/.}
#        if ! golangci-lint run "$relfile"; then
#            exitv=1
#        fi
#        :
#    done
# so we run it on the whole go src tree
    if ! golangci-lint run -v ; then
        exitv=1
    fi
else
    exitv=1
    echo "Please install golangci-lint"
fi

exit $exitv
