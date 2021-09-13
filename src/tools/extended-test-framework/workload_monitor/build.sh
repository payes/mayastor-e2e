#!/usr/bin/env bash

set -e

APP="workload_monitor"
go get github.com/go-openapi/errors
go get github.com/go-openapi/runtime
go get github.com/go-openapi/runtime/client
go get github.com/go-openapi/strfmt

go get github.com/jessevdk/go-flags

../gen_server_code.sh ${APP}
../gen_client_code.sh ${APP}

CGO_ENABLED=0 go build -a -installsuffix cgo

./build_img.sh

