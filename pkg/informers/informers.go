package informers

import (
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"pgoperator/pkg/client/clientset/versioned"
	pgoperatorInformers "pgoperator/pkg/client/informers/externalversions"
	"time"
)

const defaultResync = 600 * time.Second

type InformerFactory interface {
	KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory
	PgOperatorInformerFactory() pgoperatorInformers.SharedInformerFactory
	ApiExtensionSharedInformerFactory() apiextensions.SharedInformerFactory
	Start(stopCh <-chan struct{})
}

type informerFactories struct {
	informerFactory              k8sinformers.SharedInformerFactory
	apiextensionsInformerFactory apiextensions.SharedInformerFactory
	pgOperatorInformerFactory    pgoperatorInformers.SharedInformerFactory
}

func NewInformerFactories(client kubernetes.Interface, opCli versioned.Interface, apiextensionsClient clientset.Interface) InformerFactory {

	factory := &informerFactories{}

	if client != nil {
		factory.informerFactory = k8sinformers.NewSharedInformerFactory(client, defaultResync)
	}

	if opCli != nil {
		factory.pgOperatorInformerFactory = pgoperatorInformers.NewSharedInformerFactory(opCli, defaultResync)
	}

	if apiextensionsClient != nil {
		factory.apiextensionsInformerFactory = apiextensions.NewSharedInformerFactory(apiextensionsClient, defaultResync)
	}

	return factory
}

func (i *informerFactories) KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory {
	return i.informerFactory
}

func (i *informerFactories) PgOperatorInformerFactory() pgoperatorInformers.SharedInformerFactory {
	return i.pgOperatorInformerFactory
}

func (i *informerFactories) ApiExtensionSharedInformerFactory() apiextensions.SharedInformerFactory {
	return i.apiextensionsInformerFactory
}

func (i *informerFactories) Start(stopCh <-chan struct{}) {
	if i.informerFactory != nil {
		i.informerFactory.Start(stopCh)
	}

	if i.pgOperatorInformerFactory != nil {
		i.pgOperatorInformerFactory.Start(stopCh)
	}

	if i.apiextensionsInformerFactory != nil {
		i.apiextensionsInformerFactory.Start(stopCh)
	}

}
