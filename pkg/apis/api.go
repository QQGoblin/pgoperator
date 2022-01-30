package apis

import "k8s.io/apimachinery/pkg/runtime"

// AddToSchemes : 类型为函数列表 type SchemeBuilder []func(*Scheme) error
var AddToSchemes runtime.SchemeBuilder

func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
