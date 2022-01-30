package k8s

import (
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	pgoperator "pgoperator/pkg/client/clientset/versioned"
)

type Client interface {
	Kubernetes() kubernetes.Interface
	PgOperator() pgoperator.Interface
	Discovery() discovery.DiscoveryInterface
	ApiExtensions() clientset.Interface
	Config() *rest.Config
}

type kubernetesClient struct {
	k8s             kubernetes.Interface
	pgOperator      pgoperator.Interface
	discoveryClient *discovery.DiscoveryClient
	apiextensions   clientset.Interface
	config          *rest.Config
}

func NewKubernetesClientOrDie(options *KubernetesOptions) Client {
	config, err := clientcmd.BuildConfigFromFlags("", options.Kubeconfig)
	if err != nil {
		panic(err)
	}

	config.QPS = options.QPS
	config.Burst = options.Burst

	k := &kubernetesClient{
		k8s:             kubernetes.NewForConfigOrDie(config),
		discoveryClient: discovery.NewDiscoveryClientForConfigOrDie(config),
		pgOperator:      pgoperator.NewForConfigOrDie(config),
		apiextensions:   clientset.NewForConfigOrDie(config),
		config:          config,
	}

	return k
}

func NewKubernetesClient(options *KubernetesOptions) (Client, error) {

	config, err := clientcmd.BuildConfigFromFlags("", options.Kubeconfig)
	if err != nil {
		return nil, err
	}

	config.QPS = options.QPS
	config.Burst = options.Burst

	var k kubernetesClient
	k.k8s, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.discoveryClient, err = discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	k.apiextensions, err = clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.pgOperator, err = pgoperator.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	k.config = config

	return &k, nil
}

func (k *kubernetesClient) Kubernetes() kubernetes.Interface        { return k.k8s }
func (k *kubernetesClient) Discovery() discovery.DiscoveryInterface { return k.discoveryClient }
func (k *kubernetesClient) ApiExtensions() clientset.Interface      { return k.apiextensions }
func (k *kubernetesClient) Config() *rest.Config                    { return k.config }
func (k *kubernetesClient) PgOperator() pgoperator.Interface        { return k.pgOperator }
