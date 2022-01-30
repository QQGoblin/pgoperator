#!/bin/bash

set -e

GOBIN='go env GOPATH'/bin

# PS: go 1.17 以上版本
if [ ! -f "$GOBIN"/controller-gen ]; then
  GOPROXY=https://goproxy.cn go install -mod=vendor sigs.k8s.io/controller-tools/cmd/controller-gen
  GOPROXY=https://goproxy.cn go install -mod=vendor k8s.io/code-generator/cmd/{deepcopy-gen,client-gen,lister-gen,informer-gen}
fi

# TODO: 离线环境