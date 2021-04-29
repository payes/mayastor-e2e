#!/usr/bin/env bash
go build
docker build -t mayadata/e2e-agent .
