package k8s

import (
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	pgoperator "pgoperator/pkg/client/clientset/versioned"
)

type FakeClient struct {
	K8sClient kubernetes.Interface

	DiscoveryClient *discovery.DiscoveryClient

	ApiExtensionClient clientset.Interface

	KubeConfig *rest.Config

	PgOperatorCli pgoperator.Interface
}

func NewFakeClientSets(k8sClient kubernetes.Interface, opCli pgoperator.Interface, discoveryClient *discovery.DiscoveryClient,
	apiextensionsClient clientset.Interface, kubeConfig *rest.Config) Client {
	return &FakeClient{
		K8sClient:          k8sClient,
		DiscoveryClient:    discoveryClient,
		ApiExtensionClient: apiextensionsClient,
		KubeConfig:         kubeConfig,
		PgOperatorCli:      opCli,
	}
}

func (n *FakeClient) Kubernetes() kubernetes.Interface {
	return n.K8sClient
}

func (n *FakeClient) ApiExtensions() clientset.Interface {
	return n.ApiExtensionClient
}

func (n *FakeClient) Discovery() discovery.DiscoveryInterface {
	return n.DiscoveryClient
}

func (n *FakeClient) Config() *rest.Config {
	return n.KubeConfig
}

func (n *FakeClient) PgOperator() pgoperator.Interface { return n.PgOperatorCli }
