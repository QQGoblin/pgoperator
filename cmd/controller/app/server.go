package app

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/term"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"os"
	"pgoperator/cmd/controller/app/options"
	"pgoperator/pkg/apis"
	"pgoperator/pkg/informers"
	"pgoperator/pkg/simple/client/k8s"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

func NewControllerManagerCommand() *cobra.Command {

	s, err := options.TryLoadFromDisk()
	if err != nil {
		klog.Fatal("failed to load configuration from disk", err)
		return nil
	}

	cmd := &cobra.Command{
		Use:  "manager",
		Long: "patroni cluster controller manager",
		Run: func(cmd *cobra.Command, args []string) {
			if errs := s.Validate(); len(errs) != 0 {
				klog.Error(utilerrors.NewAggregate(errs))
				os.Exit(1)
			}

			if err = run(s, signals.SetupSignalHandler()); err != nil {
				klog.Error(err)
				os.Exit(1)
			}
		},
		SilenceUsage: true,
	}

	fs := cmd.Flags()
	namedFlagSets := s.Flags()

	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	// usage子命令
	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})

	// TODO：version 子命令后续实现
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of patroni cluster controller manager",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("v0.0.1")
		},
	}

	cmd.AddCommand(versionCmd)

	return cmd
}

// controller 启动函数
func run(mgrConfig *options.Config, ctx context.Context) error {

	// 重新封装 client
	// 可以像访问原生对象一样访问Custom Crd
	k8sClient, err := k8s.NewKubernetesClient(mgrConfig.KubernetesOptions)
	if err != nil {
		klog.Errorf("Failed to create kubernetes clientset %v", err)
		return err
	}

	// 重新封装 informer
	// 添加 Custom Crd
	informerFactory := informers.NewInformerFactories(
		k8sClient.Kubernetes(),
		k8sClient.PgOperator(),
		k8sClient.ApiExtensions(),
	)

	klog.V(0).Info("setting up manager")
	ctrl.SetLogger(klogr.New())

	// 通过 controller-runtime 提供的接口创建 controller manager
	mgrOptions := manager.Options{
		HealthProbeBindAddress: ":8118",
		LeaderElection:         true,
		LeaderElectionID:       "cb659ce9.rccp.patroni.controller",
	}
	mgr, err := manager.New(k8sClient.Config(), mgrOptions)
	if err != nil {
		klog.Fatalf("unable to set up overall controller manager: %v", err)
	}

	// 在 mgr 中注册 scheme
	if err = apis.AddToScheme(mgr.GetScheme()); err != nil {
		klog.Fatalf("unable add APIs to scheme: %v", err)
	}

	// register common meta types into schemas.
	metav1.AddToGroupVersion(mgr.GetScheme(), metav1.SchemeGroupVersion)

	// 注册controller
	if err = addControllers(mgr, k8sClient, informerFactory, mgrConfig); err != nil {
		klog.Fatalf("unable to register controllers to the manager: %v", err)
	}

	// 添加健康检查和ready
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		klog.Fatalf("unable to set up health check: %v", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		klog.Fatalf("unable to set up ready check: %v", err)
	}

	klog.V(0).Info("Starting cache resource from kube-apiserver...")
	informerFactory.Start(ctx.Done())

	klog.V(0).Info("Starting the controllers.")
	if err = mgr.Start(ctx); err != nil {
		klog.Fatalf("unable to run the manager: %v", err)
	}

	return nil
}
