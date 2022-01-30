package options

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"
	"os"
	"pgoperator/pkg/constants"
	"pgoperator/pkg/simple/client/k8s"
	"strings"
)

// Config 配置对象
type Config struct {
	// 指定kubernetes集群的配置文件，None时使用容器内的配置
	KubernetesOptions *k8s.KubernetesOptions `yaml:"kubernetes"`
}

func New() *Config {
	s := &Config{
		KubernetesOptions: k8s.NewKubernetesOptions(),
	}
	return s
}

// Validate 验证kubernetes配置是否有效
func (c *Config) Validate() []error {
	var errs []error
	errs = append(errs, c.KubernetesOptions.Validate()...)
	return errs
}

func (c *Config) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}

	c.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), c.KubernetesOptions)

	kfs := fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		kfs.AddGoFlag(fl)
	})

	return fss
}

func TryLoadFromDisk() (*Config, error) {
	cf, err := os.Open(constants.DefaultConfigurationName)
	if err != nil {
		klog.Warning("load config file failed, roll back to use default...")
		return New(), nil
	}
	defer cf.Close()

	bs, err := ioutil.ReadAll(cf)
	if err != nil {
		return nil, err
	}

	opt := &Config{}
	err = yaml.Unmarshal(bs, opt)

	if opt.KubernetesOptions == nil {
		opt.KubernetesOptions = &k8s.KubernetesOptions{}
	}

	if err != nil {
		return nil, err
	}
	return opt, nil
}
