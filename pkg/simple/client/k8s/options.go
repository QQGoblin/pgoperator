package k8s

import (
	"github.com/spf13/pflag"
	"os"
	"pgoperator/pkg/utils/reflectutils"
)

type KubernetesOptions struct {
	Kubeconfig string `json:"kubeconfig" yaml:"kubeconfig"`

	// kubernetes clientset qps
	// +optional
	QPS float32 `json:"qps,omitempty" yaml:"qps"`

	// kubernetes clientset burst
	// +optional
	Burst int `json:"burst,omitempty" yaml:"burst"`
}

func NewKubernetesOptions() *KubernetesOptions {
	return &KubernetesOptions{
		Kubeconfig: "",
	}
}

func (k *KubernetesOptions) Validate() []error {
	errors := []error{}
	if k.Kubeconfig != "" {
		if _, err := os.Stat(k.Kubeconfig); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func (k *KubernetesOptions) ApplyTo(options *KubernetesOptions) {
	reflectutils.Override(options, k)
}

func (k *KubernetesOptions) AddFlags(fs *pflag.FlagSet, c *KubernetesOptions) {
	fs.StringVar(&k.Kubeconfig, "kubeconfig", c.Kubeconfig, ""+
		"Path for kubernetes kubeconfig file, if left blank, will use "+
		"in cluster way.")
}
