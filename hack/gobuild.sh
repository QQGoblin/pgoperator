#!/bin/bash

GIT_COMMIT=$(git rev-parse HEAD)
GIT_SHA=$(git rev-parse --short HEAD)
GIT_TAG=$(git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY=$(test -n "$(git status --porcelain)" && echo "dirty" || echo "clean")


VERSION_METADATA=unreleased
# Clear the "unreleased" string in BuildMetadata
if [[ -n $GIT_TAG ]]; then
  VERSION_METADATA=
fi

rm -rf  output/*

# GOOS=linux GOARCH=amd64 CGO_ENABLED=0  GO111MODULE=on GOPROXY=https://goproxy.cn go mod vendor

GOOS=linux GOARCH=amd64 CGO_ENABLED=0  GO111MODULE=on GOPROXY=https://goproxy.cn go build -o output/manager cmd/controller/main.go

