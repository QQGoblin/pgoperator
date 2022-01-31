package cluster

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"os"
	"pgoperator/cmd/controller/app/options"
	clusterv1alpha1 "pgoperator/pkg/apis/cluster/v1alpha1"
	pgOperatorCli "pgoperator/pkg/client/clientset/versioned"
	"pgoperator/pkg/client/clientset/versioned/scheme"
	clusterInformer "pgoperator/pkg/client/informers/externalversions/cluster/v1alpha1"
	clusterLister "pgoperator/pkg/client/listers/cluster/v1alpha1"
	"pgoperator/pkg/constants"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

type patroniClusterController struct {
	eventBroadcaster record.EventBroadcaster
	eventRecorder    record.EventRecorder

	kubernetesCli kubernetes.Interface

	pgOperatorCli pgOperatorCli.Interface
	clusterLister clusterLister.PatroniClusterLister
	clusterSynced cache.InformerSynced
	clusterQueue  workqueue.RateLimitingInterface

	workerCount int
	retryCount  int
	period      time.Duration
	waitPeriod  time.Duration

	mrgConfig *options.Config
}

func NewPatroniClusterController(kubernetesCli kubernetes.Interface, pgOperatorCli pgOperatorCli.Interface,
	clusterInformer clusterInformer.PatroniClusterInformer, mgrConfig *options.Config) *patroniClusterController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		klog.Info(fmt.Sprintf(format, args))
	})

	controllerNamespace := os.Getenv(constants.ControllerNamespaceEnvironment)

	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: kubernetesCli.CoreV1().Events(controllerNamespace)})
	r := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "patroni-controller"})

	c := &patroniClusterController{
		eventBroadcaster: broadcaster,
		eventRecorder:    r,
		kubernetesCli:    kubernetesCli,
		pgOperatorCli:    pgOperatorCli,
		clusterLister:    clusterInformer.Lister(),
		clusterSynced:    clusterInformer.Informer().HasSynced,
		clusterQueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "patroni-cluster"),
		workerCount:      5,
		retryCount:       3,
		period:           1 * time.Second,
		waitPeriod:       2 * time.Second,
		mrgConfig:        mgrConfig,
	}

	// 安装调谐函数
	clusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			c.enqueueCluster(newObj)
		},
		AddFunc:    c.enqueueCluster,
		DeleteFunc: c.enqueueCluster,
	})

	return c
}

func (c *patroniClusterController) enqueueCluster(obj interface{}) {

	clusterObj := obj.(*clusterv1alpha1.PatroniCluster)
	key, err := cache.MetaNamespaceKeyFunc(clusterObj)
	if err != nil {
		utilruntime.HandleError(errors.Errorf("get patroni cluster key %s failed", clusterObj.Name))
		return
	}
	c.clusterQueue.Add(key)
}

func (c *patroniClusterController) clusterWork() {
	for c.processNextItem() {
	}
}

func (c *patroniClusterController) processNextItem() bool {
	key, quit := c.clusterQueue.Get()

	if quit {
		return false
	}

	defer c.clusterQueue.Done(key)

	// 调用业务处理逻辑
	result, err := c.handleCluster(key.(string))

	// 异常处理
	if err != nil {
		// 重试次数判断
		if c.clusterQueue.NumRequeues(key) < c.retryCount {
			klog.Errorf("Error syncing PatroniCluster %s, retrying, %v", key, err)
			c.clusterQueue.AddRateLimited(key)
		} else {
			// 超出重试次数时丢弃
			c.clusterQueue.Forget(key)
			utilruntime.HandleError(err)
		}
		return true
	}

	// 重新入队逻辑
	if result.RequeueAfter > 0 {
		// 指定时间间隔后重新入队，时间间隔
		c.clusterQueue.Forget(key)
		c.clusterQueue.AddAfter(key, result.RequeueAfter)
		return true
	} else if result.Requeue {
		// 立即重新入队
		c.clusterQueue.AddRateLimited(key)
		return true
	}

	c.clusterQueue.Forget(key)
	return true

}

func (c *patroniClusterController) Run(ctx context.Context) error {
	defer func() {
		utilruntime.HandleCrash()
		c.clusterQueue.ShutDown()
		klog.V(2).Infof("shutting down patroni cluster controller")
	}()

	klog.V(0).Infof("starting patroni cluster controller")
	if !cache.WaitForCacheSync(ctx.Done(), c.clusterSynced) {
		return errors.New("failed to wait for cached to sync")
	}

	go wait.Until(c.clusterWork, c.period, ctx.Done())
	for i := 0; i < c.workerCount; i++ {
		go wait.Until(c.clusterWork, c.period, ctx.Done())
	}

	<-ctx.Done()

	return nil
}

func (c *patroniClusterController) Start(ctx context.Context) error {
	return c.Run(ctx)
}

func (c *patroniClusterController) handleCluster(key string) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
