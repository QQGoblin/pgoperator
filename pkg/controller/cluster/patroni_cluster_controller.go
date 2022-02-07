package cluster

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	rbacV1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
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

const (
	patroniClusterFinalizerStr     = "patroni-cluster-controller"
	defaultServiceAccountName      = "patroni"
	defaultClusterRoleBinding      = "patroni-binding"
	defaultClusterRoleName         = "patroni-ep-access"
	defaultSuperUserName           = "postgres"
	defaultSuperUserPassword       = "Ruijie@rccp123"
	defaultReplicationUserName     = "standby"
	defaultReplicationUserPassword = "Ruijie@rccp123"
	defaultPgDataPath              = "/home/postgres/pgdata/pgroot/data"
	defaultPgPass                  = "/tmp/pgpass"
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

	ns, name, err := cache.SplitMetaNamespaceKey(key)

	if err != nil {
		klog.Error(errors.Wrapf(err, "not a valid controller key %s", key))
		return ctrl.Result{}, err
	}

	pCluster, err := c.clusterLister.PatroniClusters(ns).Get(name)

	if err != nil {
		klog.Error(errors.Wrapf(err, "Failed to get patroni-cluster object on cache %s/%s", ns, name))
		return ctrl.Result{}, err
	}

	pClusterFinalizer := sets.NewString(pCluster.ObjectMeta.Finalizers...)

	if pClusterFinalizer.Has(patroniClusterFinalizerStr) && !pCluster.ObjectMeta.DeletionTimestamp.IsZero() {

		// TODO: 删除逻辑

		// 执行完成删除逻辑后移除Finlizer，CRD正式被删除
		pClusterFinalizer.Delete(patroniClusterFinalizerStr)
		pCluster.ObjectMeta.Finalizers = pClusterFinalizer.List()
		if _, err := c.pgOperatorCli.RccpV1alpha1().PatroniClusters(ns).Update(context.Background(), pCluster, metav1.UpdateOptions{}); err != nil {
			klog.Error(errors.Wrap(err, "Delete finalizer failed..."))
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// ADD 控制：新创建的Obj没有对应 Finalizer
	if !pClusterFinalizer.Has(patroniClusterFinalizerStr) {
		pCluster.ObjectMeta.Finalizers = append(pCluster.ObjectMeta.Finalizers, patroniClusterFinalizerStr)
		pCluster.PatroniClusterStatus.Status = clusterv1alpha1.ClusterInit
		if _, err := c.pgOperatorCli.RccpV1alpha1().PatroniClusters(ns).Update(context.Background(), pCluster, metav1.UpdateOptions{}); err != nil {
			klog.Error(errors.Wrap(err, "Add finalizer hook failed..."))
			return ctrl.Result{}, err
		}

		// TODO: 创建集群逻辑

		return ctrl.Result{}, nil
	}

	//TODO: Update 逻辑，幂等逻辑主要功能包括：滚动更新、健康检查

	return ctrl.Result{}, nil
}

func (c *patroniClusterController) initCluster(pCluster *clusterv1alpha1.PatroniCluster) error {

	if err := c.grantPermission(pCluster.Namespace); err != nil {
		klog.Error(errors.Wrapf(err, "grant permission for namespace %s failed", pCluster.Namespace))
		return err
	}

	ns := pCluster.Namespace
	pClusterName := pCluster.Name
	for _, n := range pCluster.PatroniClusterSpec.NodeList {
		replName := fmt.Sprintf("%s-%s", pClusterName, n)
		_, err := c.kubernetesCli.AppsV1().StatefulSets(ns).Get(context.Background(), replName, metav1.GetOptions{})
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				klog.Error(errors.Wrapf(err, "init patroni cluster replicas statefelset %s/%s failed", ns, replName))
				return err
			}

			stsTpl := generatorStatefulset(n, pCluster)

			_, err = c.kubernetesCli.AppsV1().StatefulSets(ns).Create(context.Background(), &stsTpl, metav1.CreateOptions{})
			if err != nil {
				klog.Error(errors.Wrapf(err, "init patroni cluster replicas statefelset %s/%s failed, unable create statefulset", ns, replName))
				return err
			}
		}
	}

	return nil
}

func (c *patroniClusterController) grantPermission(namespace string) error {

	// 指定命名空间中创建 serviceaccount
	_, err := c.kubernetesCli.CoreV1().ServiceAccounts(namespace).Get(
		context.Background(),
		defaultServiceAccountName,
		metav1.GetOptions{},
	)

	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}

		saTpl := &v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: defaultServiceAccountName,
				Labels: map[string]string{
					"rccp.ruijie.com.cn": "patroni-cluster-controller",
				},
			},
		}
		_, err = c.kubernetesCli.CoreV1().ServiceAccounts(namespace).Create(context.Background(), saTpl, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	// 绑定clusterrole
	bindingName := fmt.Sprintf("%s:%s", namespace, defaultClusterRoleBinding)

	_, err = c.kubernetesCli.RbacV1().ClusterRoleBindings().Get(context.Background(), bindingName, metav1.GetOptions{})

	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
		bindingTpl := &rbacV1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: bindingName,
				Labels: map[string]string{
					"rccp.ruijie.com.cn": "patroni-cluster-controller",
				},
			},
			RoleRef: rbacV1.RoleRef{
				Kind: "ClusterRole",
				Name: defaultClusterRoleName,
			},
			Subjects: []rbacV1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      defaultServiceAccountName,
					Namespace: namespace,
				},
			},
		}

		_, err := c.kubernetesCli.RbacV1().ClusterRoleBindings().Create(context.Background(), bindingTpl, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}
