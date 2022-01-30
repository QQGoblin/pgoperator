package informers

import (
	"k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	"k8s.io/client-go/informers"
	k8sFake "k8s.io/client-go/kubernetes/fake"
	pgOperatorFake "pgoperator/pkg/client/clientset/versioned/fake"
	pgOperatorInformer "pgoperator/pkg/client/informers/externalversions"
	"time"
)

type nullInformerFactory struct {
	fakeK8sInformerFactory        informers.SharedInformerFactory
	fakePgOperatorInformerFactory pgOperatorInformer.SharedInformerFactory
}

func NewNullInformerFactory() InformerFactory {
	fakeClient := k8sFake.NewSimpleClientset()
	fakeInformerFactory := informers.NewSharedInformerFactory(fakeClient, time.Minute*10)

	fakePgOperatorClient := pgOperatorFake.NewSimpleClientset()

	fakePgOperatorFactory := pgOperatorInformer.NewSharedInformerFactory(fakePgOperatorClient, time.Minute*10)

	return &nullInformerFactory{
		fakeK8sInformerFactory:        fakeInformerFactory,
		fakePgOperatorInformerFactory: fakePgOperatorFactory,
	}
}

func (i *nullInformerFactory) KubernetesSharedInformerFactory() informers.SharedInformerFactory {
	return i.fakeK8sInformerFactory
}

func (i *nullInformerFactory) PgOperatorInformerFactory() pgOperatorInformer.SharedInformerFactory {
	return i.fakePgOperatorInformerFactory
}

func (n nullInformerFactory) ApiExtensionSharedInformerFactory() externalversions.SharedInformerFactory {
	return nil
}

func (n nullInformerFactory) Start(stopCh <-chan struct{}) {
}
