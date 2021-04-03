package app

import (
	"context"
	"fmt"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/portsbinding"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	v1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	kuryrv1alpha1 "projectkuryr/kuryr/pkg/apis/openstack/v1alpha1"
	kuryrclientset "projectkuryr/kuryr/pkg/client/clientset/versioned"
	kuryrinformers "projectkuryr/kuryr/pkg/client/informers/externalversions/openstack/v1alpha1"
	kuryrlisters "projectkuryr/kuryr/pkg/client/listers/openstack/v1alpha1"
	"projectkuryr/kuryr/pkg/openstack/openstackConfig"
	"reflect"
	"time"
)

const (
	SuccessSynced = "Created"
	FailedSynced = "Failed"
	MessageResourceKnsSynced = "KuryrNetwork synced successfully"
	minRetryDelay = 5 * time.Second
	maxRetryDelay = 300 * time.Second

	// NetworkPolicyController is the only writer of the antrea network policy
	// storages and will keep re-enqueuing failed items until they succeed.
	// Set resyncPeriod to 0 to disable resyncing.
	resyncPeriod time.Duration = 0
)

// Controller is the controller implementation for Kns resources
type NsController struct {
	config *ControllerConfig
	// a standard kubernetes clientset
	kubeclientset 		kubernetes.Interface
	// a clientset for our own API group
	crdclientset 		kuryrclientset.Interface
	osClient	 		*openstackConfig.OSClient

	nsLister 		 	corelisters.NamespaceLister
	nsSynced 		 	cache.InformerSynced

	podLister 		 	corelisters.PodLister
	podSynced 		 	cache.InformerSynced

	knsLister       	kuryrlisters.KuryrNetworkLister
	knsSynced        	cache.InformerSynced

	kpLister        	kuryrlisters.KuryrPortLister
	kpSynced        	cache.InformerSynced

	//namespaceStore storage.Interface
	//internalKuryrNetworkStore storage.Interface

	// workqueue是一个速率有限的工作队列。它用于排队处理工作，而不是在发生变化时立即执行。这意味着我们可以确保每次只处理固定数量的资源，并且很容易确保我们不会在两个不同的工作器中同时处理同一个项目。
	podQueue 			workqueue.RateLimitingInterface
	nsQueue 			workqueue.RateLimitingInterface
	// internalKuryrPortQueue maintains the KuryrPort objects that
	// need to be synced.
	internalKuryrPortQueue 		workqueue.RateLimitingInterface
	internalKuryrNetworkQueue 	workqueue.RateLimitingInterface

	// recorder is an event recorder for recording Event resources to the Kubernetes API.
	recorder record.EventRecorder
}

func NewNsController(
	config *ControllerConfig,
	kubeClientset kubernetes.Interface,
	crdClientset kuryrclientset.Interface,
	osClient 	*openstackConfig.OSClient,
	nsInformer v1.NamespaceInformer,
	knsInformer kuryrinformers.KuryrNetworkInformer,
	podInformer 	v1.PodInformer,
	kpInformer 		kuryrinformers.KuryrPortInformer,
	recorder record.EventRecorder) *NsController {

	c := &NsController{
		config: config,
		osClient: 		osClient,
		kubeclientset:  kubeClientset,
		crdclientset:   crdClientset,
		nsLister:       nsInformer.Lister(),
		nsSynced:       nsInformer.Informer().HasSynced,

		podLister: 		podInformer.Lister(),
		podSynced: 		podInformer.Informer().HasSynced,

		knsLister:      knsInformer.Lister(),
		knsSynced:      knsInformer.Informer().HasSynced,
		kpLister: 		kpInformer.Lister(),
		kpSynced: 		kpInformer.Informer().HasSynced,

		internalKuryrNetworkQueue: 	workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(minRetryDelay, maxRetryDelay), "KuryrNetwork"),
		internalKuryrPortQueue:    	workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(minRetryDelay, maxRetryDelay), "KuryrPort"),
		podQueue: 					workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(minRetryDelay, maxRetryDelay), "Pod"),
		nsQueue:         			workqueue.NewNamedRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(minRetryDelay, maxRetryDelay), "Namespace"),

		recorder:       recorder,
	}

	klog.Info("Setting up event handlers for ns")
	nsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.AddNs,
		UpdateFunc: c.UpdateNs,
	})

	klog.Info("Setting up event handlers for kns") // 认为除了 kuryr-controller 之外不会操作 kns
	knsInformer.Informer().AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		UpdateFunc: c.UpdateKns,
	},
	resyncPeriod)

	klog.Info("Setting up event handlers for pod")
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:	c.AddPod,
		// update delete event is not called repeatedly?
		UpdateFunc: c.UpdatePod,

		DeleteFunc:  func(obj interface{}) {
			if IsKuryrCniType(obj.(*corev1.Pod).GetAnnotations()) {
				klog.Infof("Delete Event -> Pod(%s) !!!\n", obj.(*corev1.Pod).GetName())
				//nsController.enqueue(obj)
			}
		},
	})

	klog.Info("Setting up event handlers for kuryrport")
	kpInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}){
			klog.Infof("Add Event -> KP(%s), ignore !!!\n", obj.(*kuryrv1alpha1.KuryrPort).GetName())
		},
		UpdateFunc: func(oldObj, newObj interface{}){
			klog.Infof("Update Event -> KP(%s) !!!\n", newObj.(*kuryrv1alpha1.KuryrPort).GetName())
			newKp := newObj.(*kuryrv1alpha1.KuryrPort)
			oldKp := oldObj.(*kuryrv1alpha1.KuryrPort)
			if newKp.ResourceVersion == oldKp.ResourceVersion {
				return
			}
			if isFinalize(newKp) && containsString(newKp.Finalizers, FinalizerKuryrPort ) {
				c.kpOnFinalize(newKp)
			}
		},
	})

	return c
}

func (c *NsController) AddNs(obj interface{}){
	ns := obj.(*corev1.Namespace)

	if IsKuryrCniType(ns.GetAnnotations()) {
		var key string
		var err error
		if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
			utilruntime.HandleError(err)
			return
		}
		klog.Infof("Add Event -> Namespace(%s) !!!\n", key)
		c.nsQueue.Add(key)
	}
}

func (c *NsController) UpdateNs(oldObj, newObj interface{}){
	new := newObj.(*corev1.Namespace)
	old := oldObj.(*corev1.Namespace)
	if new.ResourceVersion == old.ResourceVersion {
		klog.Infof("Update Event but version equal")
		return
	}
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(new); err != nil {
		utilruntime.HandleError(err)
		return
	}
	klog.Infof("Update Event -> Namespace(%s) !!!\n", key)
	c.nsQueue.Add(key)
}

func (c *NsController) AddPod(obj interface{}){
	pod := obj.(*corev1.Pod)
	if IsKuryrCniType(pod.GetAnnotations()) {
		var key string
		var err error
		if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
			utilruntime.HandleError(err)
			return
		}
		klog.Infof("Add Event -> Pod(%s) !!!\n", key)
		c.podQueue.Add(key)
	}
}

func (c *NsController) UpdatePod(oldObj, newObj interface{}){
	new := newObj.(*corev1.Pod)
	old := oldObj.(*corev1.Pod)
	if new.ResourceVersion == old.ResourceVersion {
		klog.Infof("Update Event but version equal")
		return
	}

	if IsKuryrCniType(new.GetAnnotations()) {
		var key string
		var err error
		if key, err = cache.MetaNamespaceKeyFunc(new); err != nil {
			utilruntime.HandleError(err)
			return
		}
		klog.Infof("Update Event -> Pod(%s) !!!\n", key)
		c.podQueue.Add(key)
	}
}

func (c *NsController) UpdateKns(oldObj, newObj interface{}){
	kns := newObj.(*kuryrv1alpha1.KuryrNetwork)
	klog.Infof("Update Event -> Kns(%s).\n", kns.Name)
	if isFinalize(kns){
		c.knsOnFinalize(kns)
	}
}

func (c *NsController) Run(threadiness int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.nsQueue.ShutDown()
	defer c.podQueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Namespaces&Pod controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.nsSynced, c.knsSynced); !ok {
		klog.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker4Ns, time.Second, stopCh)
		go wait.Until(c.runWorker4Pod, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")
}

func (c *NsController) runWorker4Pod() {
	for c.processNextPodWorkItem() {
	}
}

func (c *NsController) runWorker4Ns() {
	for c.processNextNamespaceWorkItem() {
	}
}

func (c *NsController) processNextNamespaceWorkItem() bool {
	key, shutdown := c.nsQueue.Get()
	if shutdown {
		return false
	}
	defer c.nsQueue.Done(key)

	err := c.syncNamespace(key.(string))
	if err != nil {
		// Put the item back on the workqueue to handle any transient errors.
		c.nsQueue.AddRateLimited(key)
		klog.Errorf("Failed to sync internal Namespace %s: %v", key, err)
		return true
	}

	// If no error occurs we Forget this item so it does not get queued again until
	// another change happens.
	c.nsQueue.Forget(key)
	return true
}

func (c *NsController) processNextPodWorkItem() bool {
	obj, shutdown := c.podQueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.podQueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.podQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.SyncPod(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.podQueue.AddRateLimited(obj)
			return fmt.Errorf("error syncing %s, requeuing. Error: %s", obj, err.Error())
		}
		c.podQueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *NsController) syncNamespace(key string) (err error) {
	/*
		删除 ns 流程(K8S EVENT)：
			Update Event -> Namespace(k8s add DeletionTimestamp)
		 	Update Event -> KuryrNetwork(k8s add DeletionTimestamp)
			Delete Event -> KuryrNetwork
			Update Event -> Namespace
		关键性问题:
			中途修改ns（走ns的update流程，之前的kns资源（及其子资源pod/port）处理）
	*/

	// Convert the namespace/name string into a distinct namespace and name
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	klog.Infof("get form nsQueue : %s\n", name)
	ns, err := c.nsLister.Get(name)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Namespace '%s' in work queue no longer exists", name))
		return err
	}

	if err = c.knsSyncFromNs(ns); err != nil{
		klog.Errorf("Failed synced KuryrNetwork(%s/%s)\n", ns.Name, ns.Name)
	}else{
		c.recorder.Event(ns, corev1.EventTypeNormal, SuccessSynced, MessageResourceKnsSynced)
		klog.Infof("Successfully synced KuryrNetwork(%s/%s)\n", ns.Name, ns.Name)
	}

	return err
}

func (c *NsController) SyncPod(key string) (err error) {// ns *corev1.Namespace
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	klog.Infof("SyncPod (%s/%s)\n", namespace, name)
	pod, err := c.kubeclientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("Pod '%s' in work queue no longer exists", name))
		return err
	}
	err = c.kpSyncFromPod(pod)
	if err != nil {
		klog.Infof("\tFailed synced KuryrPort(%s).\n", pod.GetName())
	}

	return err
}

// 监测到 kns 有变化，重新走 ns 流程（检查是否有，没有创建，有对比信息完整性，进行更新。
func (c *NsController) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}

	// 判断如果是kns变化，则获取对应的 ns obj 放入队列中
	switch v := object.(type) {
	case *kuryrv1alpha1.KuryrNetwork:
		ns, err := c.nsLister.Get(object.GetNamespace())
		if err != nil {
			klog.Infof("Ignoring error object %s", object.GetNamespace())
			return
		}
		klog.Infof("Update Event -> kns(%s) and push ns-obj to queue\n", ns.Name)
		//c.enqueue(ns)
	default:
		klog.Infof("UNKNOWN object type: %s", v)
	}
}

func (c *NsController) kpSyncFromPod(pod *corev1.Pod) error{
	if isPodCompleted(pod) {
		klog.Infof("\tPod(%s/%s) is completed, removing the KuryrPort", pod.GetNamespace(), pod.GetName())
		return c.PodOnFinalize(pod)
	}

	if isFinalize(pod) {
		if containsString(pod.Finalizers, FinalizerPod){
			klog.Infof("\tPod(%s/%s) is Finalize, removing the KuryrPort", pod.GetNamespace(), pod.GetName())
			return c.PodOnFinalize(pod)
		}
		klog.Infof("\tPod(%s/%s) is Finalize, No processing of KuryrPort", pod.GetNamespace(), pod.GetName())
		return nil
	}

	kp, err := c.kpLister.KuryrPorts(pod.Namespace).Get(pod.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return c.newKuryrPort(pod)
		}
	}else{
		klog.Infof("\tGot KuryrPort:(%s/%s)\n", kp.GetNamespace(), kp.GetName())
		// 是否需要更新 labels？
	}

	return nil
}

func (c *NsController) PodOnFinalize(pod *corev1.Pod) error{
	kp, err := c.kpLister.KuryrPorts(pod.Namespace).Get(pod.Name)
	if err != nil {
		if errors.IsNotFound(err){
			klog.Infof(">>>>>>> Get KuryrPort(%s/%s) is not found.\n", pod.Namespace, pod.Name)
			return nil
		}else{
			klog.Errorf("Get KuryrPort(%s/%s) Error: %s\n", pod.Namespace, pod.Name, err)
			return err
		}
	}
	// pod 删除时， k8s 并不知道要删除 kp。所以要在 pod 删除事件中删除或者设置 kp 的删除
	return c.crdclientset.OpenstackV1alpha1().KuryrPorts(kp.Namespace).Delete(context.TODO(), kp.Name, metav1.DeleteOptions{})
}

func isPodCompleted(pod *corev1.Pod) bool {
	if pod.Status.Phase == k8sPodStatusSucceeded || pod.Status.Phase == k8sPodStatusFailed {
		return true
	}
	return false
}

func (c *NsController) knsSyncFromNs(ns *corev1.Namespace) error{
	/*
		1. 查询当前是否有kns，没有创建
		2. 有则只更新 labels （遗留： 始终以最新的ns为准（包含用户手动变更删除kns等）拉取创建kns还是以第一次创建为准->如果以第一次为准后续ns修改不做更新，那么当用户手动删除kns时，一样存在重新拉取数据的问题，如果更改，就要处理之前kns创建的资源，比如端口缓存等）
	*/
	if isFinalize(ns) {
		klog.Infof("\tNamespace(%s) is Finalize, returned directly.\n", ns.GetName())
		return nil
	}

	knsCur, err := c.knsLister.KuryrNetworks(ns.Name).Get(ns.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			knsNew, err := c.newKuryrNetwork(ns)
			if err != nil {
				klog.Errorln("Invoke newKuryrNetwork Failed. %v", err)
				return err
			}
			if err = c.createKns(knsNew); err != nil{
				klog.Errorf("Invoke createKns Failed. %v\n", err)
			}
		}else{
			klog.Errorf("Get current kns(%s) Failed. %v\n", ns.Name, err)
			return err
		}
	}else{ //update labels only
		if reflect.DeepEqual(knsCur.Labels, ns.Labels) {
			return nil
		}

		knsCur.Labels = ns.Labels
		if err = c.updateKns(knsCur); err != nil{
			klog.Errorf("Update kns(%s) labels(%v) Failed. %v\n", ns.Name, ns.Labels, err)
			return err
		}
	}

	klog.Infof("Ended of knsSyncFromNs(%v)\n", ns.Name)
	return nil
}

func (c *NsController) newKuryrPort(pod *corev1.Pod) error{
	kns, err := c.knsLister.KuryrNetworks(pod.Namespace).Get(pod.Namespace)
	if err != nil {
		klog.Errorf("get kns(%s/%s) failed %v", pod.Namespace, pod.Namespace, err)
		return err
	}

	podProject := kns.Spec.ProjectId
	podNetwork := kns.Status.PodNetId
	fixedIP := ports.IP{
		SubnetID: kns.Status.PodSubnetId,
	}

	annotations := pod.GetAnnotations()

	if annotations[AnnotationPodSubnet] != "" && annotations[AnnotationPodFixedIP] != "" {
		fixedIP.SubnetID = annotations[AnnotationPodSubnet]
		fixedIP.IPAddress = annotations[AnnotationPodFixedIP]
	}

	var podSgs []string

	if "" != annotations[AnnotationPodSg] {
		podSgs = []string{annotations[AnnotationPodSg]}
	}else if len(kns.Status.PodSgs) > 0 {
		klog.Infof("Get sgs from kns: %v", kns.Status.PodSgs)
		podSgs = kns.Status.PodSgs
	}

	podIfNameIsDefault := true
	podIfName := c.config.Openstack.LinkIface
	if ifName, ok := annotations[AnnotationPodIfName]; ok {
		podIfName = ifName
		podIfNameIsDefault = false
	}

	portCreateOpts := &ports.CreateOpts{
		ProjectID: 		podProject,
		NetworkID: 		podNetwork,
		DeviceOwner:	OpenstackPortDeviceOwner,
		AdminStateUp: gophercloud.Enabled,
		FixedIPs:     []ports.IP{fixedIP},
		//SecurityGroups: &[]string{},
	}
	if len(podSgs) > 0 {
		klog.Infof("Create port with sgS: %v", podSgs)
		portCreateOpts.SecurityGroups = &podSgs
	}
	//profile := map[string]interface{}{"foo": "bar"}
	createOpts := portsbinding.CreateOptsExt{
		CreateOptsBuilder: portCreateOpts,
		HostID:            pod.Spec.NodeName,
		VNICType:		   "normal",
		//Profile:           profile,
	}

	netExt, err :=  c.osClient.GetNetwork(portCreateOpts.NetworkID)
	if err != nil{
		klog.Errorf("Get network(%s) Error: %v\n", portCreateOpts.NetworkID, err)
		return err
	}

	subnet, err := c.osClient.GetSubnet(fixedIP.SubnetID)
	if err != nil {
		klog.Errorf("Get subnet(%v) Error: %v\n", fixedIP.SubnetID, err)

		return err
	}

	portExt, err := c.osClient.CreatePort(createOpts)
	if err != nil{
		klog.Errorf("Create port (%v) Error: %v\n", createOpts, err)
		return err
	}

	var ips []kuryrv1alpha1.IP
	for _, ip := range  portExt.FixedIPs {
		ips = append(ips, kuryrv1alpha1.IP{IPAddress: ip.IPAddress, SubnetID: ip.SubnetID})
	}

	vif := kuryrv1alpha1.KuryrVif{
		IsDefault: podIfNameIsDefault,
		IfName:    podIfName,
		Vif: kuryrv1alpha1.VIF{
			VifName:        "tap" + portExt.ID[:11],
			BridgeName:     c.config.Openstack.OvsBridge,
			Status:         portExt.Status,
			ID:             portExt.ID,
			MACAddress:     portExt.MACAddress,
			Plugin:         portExt.VIFType,
			//VIFType: portExt.VIFType,
			SecurityGroups: portExt.SecurityGroups,
			Network: kuryrv1alpha1.Network{
				ID:  portExt.NetworkID,
				MTU: netExt.MTU,
				Subnets: []kuryrv1alpha1.Subnet{{
					//ID:      subnet.ID,
					Cidr:    subnet.CIDR,
					Gateway: subnet.GatewayIP,
					DNS:     subnet.DNSNameservers,
					Ips: ips,
				}},
			},
		},
	}
	kp := &kuryrv1alpha1.KuryrPort{
		ObjectMeta: metav1.ObjectMeta{
			Name:        pod.Name,
			Namespace:   pod.Namespace,
			Annotations: pod.Annotations,
			Labels: 	pod.Labels,
			Finalizers: []string{FinalizerKuryrPort},
		},
		Spec: kuryrv1alpha1.KuryrPortSpec{
			PodUid: string(pod.UID),
			PodNodeName: pod.Spec.NodeName,
		},
		Status: kuryrv1alpha1.KuryrPortStatus{
			ProjectId: kns.Spec.ProjectId,
		},
	}

	kp.Status.Vifs = append(kp.Status.Vifs, vif)
	_, err = c.crdclientset.OpenstackV1alpha1().KuryrPorts(kp.Namespace).Create(context.TODO(), kp, metav1.CreateOptions{})
	if err != nil {
		c.osClient.DeletePort(portExt.ID) // 多线程导致的冲突。1、为什么有多次连续同一个pod的update event 2、队列去重
		klog.Errorf("Create KuryrPorts Error: %s", err)
		return err
	}

	if !containsString(pod.Finalizers, FinalizerPod){
		pod.Finalizers = append(pod.Finalizers, FinalizerPod)
		_, err = c.kubeclientset.CoreV1().Pods(pod.Namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("Create KuryrPorts Error: %s", err)
		}
	}

	return err
}

func (c *NsController) kpOnFinalize(kp *kuryrv1alpha1.KuryrPort) {
	klog.Infof("\tFinalizer KuryrPort(%s)\n", kp.GetName())

	// Release ports
	for _, vif := range kp.Status.Vifs {
		portId := vif.Vif.ID
		c.osClient.DeletePort(portId)
		klog.Infof("\tDelete Port(%s).", portId)
	}

	// Remove finalizer out of pod.
	pod, err := c.kubeclientset.CoreV1().Pods(kp.Namespace).Get(context.TODO(), kp.Name, metav1.GetOptions{})
	if err == nil {
		pod.Finalizers = removeString(pod.Finalizers, FinalizerPod)
		_, err := c.kubeclientset.CoreV1().Pods(pod.Namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
		klog.Infof("\tRemove Pod Finalizers. Ret: %v", err)
	}

	// Remove finalizer from KuryrPort.
	kp.Finalizers = removeString(kp.Finalizers, FinalizerKuryrPort)
	c.updateKp(kp)
	klog.Infof("\tRemove KuryrPort Finalizers.")
}

//func (c *NsController) delKuryrPort(pod *corev1.Pod) error{
//	kp, err := c.kpLister.KuryrPorts(pod.Namespace).Get(pod.Namespace)
//	if err != nil {
//		klog.Errorf("get kp(%s/%s) failed %v", pod.Namespace, pod.Namespace, err)
//		return err
//	}
//	kp.Finalizers = removeString(kns.Finalizers, FinalizerKuryrNetwork)
//
//	_, err = c.crdclientset.OpenstackV1alpha1().KuryrPorts(kp.Namespace).Delete(context.TODO(), kp.Name, metav1.DeleteOptions{})
//	return err
//}

func (c *NsController) updateKp(kp *kuryrv1alpha1.KuryrPort) error{
	_, err := c.crdclientset.OpenstackV1alpha1().KuryrPorts(kp.Namespace).Update(context.TODO(), kp, metav1.UpdateOptions{})
	return err
}

func (c *NsController) knsOnFinalize(kns *kuryrv1alpha1.KuryrNetwork) {
	/* (设置了 fininalizer ，预删除，相应的事件是 update 而不是 delete).该接口只处理 kns 删除的情况，ns 变更导致的更新放在 ns 中处理
	1. clear kns 衍生的资源
	2. if kns 创建的时候创建的 network: delete network
	3. if 清理成功: remove fininalizer
	*/

	klog.Infof("\tknsOnFinalize(%s) Started.\n", kns.Name)
	if err := c.deleteExternalResources(kns); err != nil {
		// 如果删除失败，则直接返回对应 err，controller 会自动执行重试逻辑
		klog.Errorf("kns deleteExternalResources failed. %v", err)
		return
	}
	// 如果对应 hook 执行成功，那么清空 finalizers， k8s 删除对应资源
	kns.Finalizers = removeString(kns.Finalizers, FinalizerKuryrNetwork)
	if err := c.updateKns(kns); err != nil{
		klog.Errorf("Update kns failed. %v", err)
	}

	klog.Infof("\tknsOnFinalize(%v) Ended.\n", kns.Name)
}

func (c *NsController) deleteExternalResources(kns *kuryrv1alpha1.KuryrNetwork) error {
	// 删除 Kns 关联的外部资源逻辑(ip pool)
	// 需要确保实现是幂等的
	klog.Infof("\t\tKns delete ExternalResources!!!!!")
	return nil
}

func (c *NsController) updateKns(kns *kuryrv1alpha1.KuryrNetwork) error {
	_, err := c.crdclientset.OpenstackV1alpha1().KuryrNetworks(kns.Namespace).Update(context.TODO(), kns, metav1.UpdateOptions{})
	return err
}

func (c *NsController) createKns(kns *kuryrv1alpha1.KuryrNetwork) error {
	_, err := c.crdclientset.OpenstackV1alpha1().KuryrNetworks(kns.Namespace).Create(context.TODO(), kns, metav1.CreateOptions{})
	return err
}

/*
	如果先创建了ns指定了 租户信息，创建了一批pods，后期修改了ns中租户信息，那么将ns中的网络信息进行更新（如果之前是默认创建的新的，删除所有的pod，删除网络资源。
	如果之前是指定的租户网络资源，直接更新kns信息，之前的pod保留，在ns删除时再删除pods/ports，因为租户指定的网络资源不做删除，只删除pod使用的ports））
	之所以监测kns的变化是防止有人手动修改了kns，这样保证kns的信息一直和ns保持一致
*/
func (c *NsController) newKuryrNetwork(ns *corev1.Namespace) (*kuryrv1alpha1.KuryrNetwork, error){
	annotations := ns.GetAnnotations()
	/*
		不创建 kns 的情况：
			!IsKuryrCniType or 全局共享网络资源没有 enable && 没有在创建ns时指定网络资源
	*/
	if !IsKuryrCniType(annotations){
		return nil,nil
	}
	if !c.config.Openstack.EnabledDefaultNetworkResources && !isNetworkResourceSpecifiedAndValid(annotations){
		return nil, fmt.Errorf("No available network resources to use! (%s)", annotations)
	}

	kns := kuryrv1alpha1.KuryrNetwork{
		ObjectMeta: metav1.ObjectMeta{
			Name:        ns.Name,
			Namespace:   ns.Name,
			Annotations: ns.Annotations,
			Labels: ns.Labels,
			Finalizers: []string{FinalizerKuryrNetwork},
		},
		Spec: kuryrv1alpha1.KuryrNetworkSpec{
			ProjectId: c.config.Openstack.ProjectId,
			IsTenant:  false,
		},
		Status: kuryrv1alpha1.KuryrNetworkStatus{
			PodRouterId: c.config.Openstack.PodRouterId,
			PodSgs: c.config.Openstack.PodSgIds,
			PodSubnetId : c.config.Openstack.PodSubnetId,
			PodSubnetCIDR: c.config.Openstack.PodSubnetCIDR,
			PodNetId : c.config.Openstack.PodNetId,
			SvcSubnetId: c.config.Openstack.SvcSubnetId,
			SvcSubnetCIDR: c.config.Openstack.SvcSubnetCIDR,
		},
	}

	if isNetworkResourceSpecifiedAndValid(annotations){
		kns.Status.PodRouterId = annotations[AnnotationPodRouter]
		kns.Status.PodSubnetId = annotations[AnnotationPodSubnet]
		kns.Status.SvcSubnetId = annotations[AnnotationSvcSubnet]
		kns.Status.PodSgs = []string{annotations[AnnotationPodSg]}

		subnet, err :=  c.osClient.GetSubnet(annotations[AnnotationPodSubnet])
		if err != nil{
			klog.Errorf("Get subnet by id(%s) Failed. %v\n", annotations[AnnotationPodSubnet], err)
			return nil, err
		}
		kns.Spec.IsTenant = true
		kns.Spec.ProjectId = subnet.ProjectID
		kns.Status.PodSubnetCIDR = subnet.CIDR
		kns.Status.PodNetId = subnet.NetworkID
		//kns.Status.SvcSubnetCIDR 同一个环境中CIDR是相同的
	}

	return &kns, nil
}


