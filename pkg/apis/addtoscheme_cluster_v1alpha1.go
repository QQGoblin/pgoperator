package apis

import "pgoperator/pkg/apis/cluster/v1alpha1"

func init() {
	// 注册对应CRD到Schemes
	AddToSchemes = append(AddToSchemes, v1alpha1.AddToScheme)
}
