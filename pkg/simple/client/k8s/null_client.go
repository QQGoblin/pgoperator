package k8s

import (
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	pgoperator "pgoperator/pkg/client/clientset/versioned"
)

type nullClient struct {
}

func NewNullClient() Client {
	return &nullClient{}
}

func (n nullClient) Kubernetes() kubernetes.Interface {
	return nil
}

func (n nullClient) ApiExtensions() clientset.Interface {
	return nil
}

func (n nullClient) Discovery() discovery.DiscoveryInterface {
	return nil
}

func (n nullClient) Config() *rest.Config {
	return nil
}

func (n nullClient) PgOperator() pgoperator.Interface { return nil }
