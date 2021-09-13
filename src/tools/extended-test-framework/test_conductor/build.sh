#!/usr/bin/env bash

set -e pipefail

go get github.com/go-openapi/errors
go get github.com/go-openapi/runtime
go get github.com/go-openapi/runtime/client
go get github.com/go-openapi/strfmt
go get github.com/go-openapi/validate

go get github.com/jessevdk/go-flags

CGO_ENABLED=0 go build -a -installsuffix cgo

./build_img.sh

