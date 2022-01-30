GOBIN=$(shell go env GOPATH)/bin
CONTROLLER_GEN=$(GOBIN)/controller-gen

## 安装代码生成工具
tools:
	./hack/generate_tools.sh

## 生成 CRD 文件
manifests: tools
	$(CONTROLLER_GEN) $(CRD_OPTIONS) crd:allowDangerousTypes=false paths="./pkg/apis/..." output:crd:artifacts:config=config/crd

## 生成 DeepCopy
deepcopy: tools
	$(CONTROLLER_GEN) object:headerFile=hack/boilerplate.go.txt paths="./pkg/apis/..."

## 生成 Clientset
clientset: tools
