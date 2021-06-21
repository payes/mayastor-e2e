#!/usr/bin/env bash
docker build -t mayadata/e2e-agent --build-arg GO_VERSION=1.16.3 .
