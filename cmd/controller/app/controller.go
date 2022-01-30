package app

import (
	"k8s.io/klog/v2"
	"pgoperator/cmd/controller/app/options"
	"pgoperator/pkg/informers"
	"pgoperator/pkg/simple/client/k8s"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func addControllers(mgr manager.Manager, client k8s.Client, mgrConfig *options.Config, informerFactory informers.InformerFactory, options *k8s.KubernetesOptions) error {

	controllers := map[string]manager.Runnable{}

	for name, ctrl := range controllers {
		if ctrl == nil {
			klog.V(4).Info("%s is not going to run due to dependent component disabled.", name)
			continue
		}

		if err := mgr.Add(ctrl); err != nil {
			klog.Error(err, "add controller to manager failed", "name", name)
			return err
		}
	}

	return nil
}
