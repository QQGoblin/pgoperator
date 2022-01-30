#!/usr/bin/env bash


# shellcheck disable=SC2006
GOPATH=`go env GOPATH`

OUTPUT_PKG=pgoperator/pkg/client
APIS_PKG=pgoperator/pkg/apis

# 需要生成代码的pkg，用逗号分隔
FQ_APIS=${APIS_PKG}/cluster/v1alpha1

# 生成clientset
echo "Generating clientset for ${FQ_APIS}"
${GOPATH}/bin/client-gen --clientset-name "${CLIENTSET_NAME_VERSIONED:-versioned}" --input-base "" --input ${FQ_APIS} --output-package ${OUTPUT_PKG}/"${CLIENTSET_PKG_NAME:-clientset}" -h "$PWD/hack/boilerplate.go.txt"



# 生成lister
echo "Generating listers for ${FQ_APIS}"
${GOPATH}/bin/lister-gen --input-dirs ${FQ_APIS} --output-package ${OUTPUT_PKG}/listers -h "$PWD/hack/boilerplate.go.txt"

# 生成informers
echo "Generating informers for ${FQ_APIS}"
${GOPATH}/bin/informer-gen \
           --input-dirs ${FQ_APIS} \
           --versioned-clientset-package ${OUTPUT_PKG}/${CLIENTSET_PKG_NAME:-clientset}/${CLIENTSET_NAME_VERSIONED:-versioned} \
           --listers-package ${OUTPUT_PKG}/listers \
           --output-package ${OUTPUT_PKG}/informers \
           -h "$PWD/hack/boilerplate.go.txt"

