package proxy


/*
// ProxyServer represents all the parameters required to start the Kubernetes proxy server. All
// fields are required.
type ProxyServer struct {
	Client                 clientset.Interface
	EventClient            v1core.EventsGetter
	MetricsBindAddress     string
	BindAddressHardFail    bool
	HealthzServer          healthcheck.ProxierHealthUpdater // interface 指针

	Recorder record.EventRecorder
	NodeRef  *corev1.ObjectReference
}

func newProxyServer(config *AgentConfig, cleanupAndExit bool, master string) (*ProxyServer, error) {
	// Create K8s Clientset, CRD Clientset and SharedInformerFactory for the given config.
	//k8sClient, _, err := k8s.CreateClients(o.config.ClientConnection, o.config.KubeAPIServerOverride)
	k8sClient, eventClient, err := k8s.CreateClients(config.ClientConnection, master)
	if err != nil {
		klog.Errorln("Error: CreateClients ", err)
		return nil, err
	}
	//informerFactory := informers.NewSharedInformerFactory(k8sClient, informerDefaultResync)

	hostname, err := node.GetHostname("")
	if err != nil {
		fmt.Println("Error: GetHostname ", err)
		return nil, err
	}
	// Creates a new event broadcaster.
	eventBroadcaster := record.NewBroadcaster()
	// NewRecorder 返回一个EventRecorder，该EventRecorder可用于向该EventBroadcaster发送事件，事件源设置为给定的事件源。
	recorder := eventBroadcaster.NewRecorder(Scheme, corev1.EventSource{Component: "kuryr-agent", Host: hostname})

	nodeRef := &corev1.ObjectReference{
		Kind:      "Node",
		Name:      hostname,
		UID:       types.UID(hostname),
		Namespace: "",
	}

	klog.Infof("newProxyServer: nodeRef: %+v\n", *nodeRef)

	// 实例化一个 proxierHealthServer
	var healthzServer healthcheck.ProxierHealthUpdater // 声明 healthzServer 为 ProxierHealthUpdater interface 指针。
	// 多路复用
	healthzServer = healthcheck.NewProxierHealthServer(config.HealthzBindAddress, 2*time.Second, recorder, nodeRef)

	serviceChange := &ServiceChangeTracker{
		items:						make(map[types.NamespacedName]*serviceChange),
		ipFamily:					corev1.IPv4Protocol,
		recorder:					recorder,
		processServiceMapChange: 	nil,

	}
	var proxier Provider
	proxier = &Proxier{
		serviceChanges: serviceChange,
	}

	return &ProxyServer{
		Client:                 k8sClient,
		EventClient:			eventClient,
		HealthzServer:          healthzServer,
		BindAddressHardFail: 	false,
		MetricsBindAddress:     config.MetricsBindAddress,
		Recorder:               recorder,
		NodeRef:                nodeRef,
		Proxier: 				proxier,
	}, nil
}

func (s *ProxyServer) birthCry() {
	fmt.Println("#### ProxyServer.birthCry")
	s.Recorder.Eventf(s.NodeRef, EventTypeNormal, "Starting", "Starting kuryr-agent.")
}

func (s *ProxyServer) CleanupAndExit() error {
	fmt.Println("\there is Options.CleanupAndExit")
	return nil
}

func (s *ProxyServer) Run() error {
	// 线程1： 轮询监视 apiserver
	// 线程2:  监听端口 响应 kubelet 的 addNetwork 和 delNetwork
	// 线程3： 健康检查

	var errCh chan error
	if s.BindAddressHardFail {
		errCh = make(chan error)
	}

	fmt.Println("\tProxyServer.Run(): Start up a healthz server ")
	// Start up a healthz server if requested
	serveHealthz(s.HealthzServer, errCh)

	fmt.Println("\tProxyServer.Run(): Start up a Metrics server on ", s.MetricsBindAddress)
	// Start up a metrics server if requested
	serveMetrics(s.MetricsBindAddress, "/metrics", true, errCh)

	kuryrPortRequirement, err := labels.NewRequirement("KuryrPort", selection.Exists, []string{})
	if err != nil{
		fmt.Println("Error!!! kuryrPortRequirement", err)
	}

	podRequirement, err := labels.NewRequirement("Pod", selection.Exists, []string{})
	if err != nil{
		fmt.Println("Error!!! kuryrPortRequirement", err)
	}

	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*podRequirement, *kuryrPortRequirement)

	// Make informers that filter out objects that want a non-default service proxy.
	//informerFactory := informers.NewSharedInformerFactoryWithOptions(s.Client, time.Second,
	//	informers.WithTweakListOptions(func(options *metav1.ListOptions) {
	//		options.LabelSelector = labelSelector.String()
	//	}))
	informerFactory := informers.NewSharedInformerFactoryWithOptions(s.Client, time.Second)

	serviceConfig := NewServiceConfig(informerFactory.Core().V1().Services(), time.Second)
	serviceConfig.RegisterEventHandler(s.Proxier)
	go serviceConfig.Run(wait.NeverStop)

	// Birth Cry after the birth is successful
	s.birthCry()
	return nil
}

type Provider interface {
	ServiceHandler

	// Sync immediately synchronizes the Provider's current state to proxy rules.
	Sync()
	// SyncLoop runs periodic work.
	// This is expected to run as a goroutine or as the main loop of the app.
	// It does not return.
	SyncLoop()
}

type ServiceHandler interface {
	// OnServiceAdd is called whenever creation of new service object
	// is observed.
	OnServiceAdd(service *corev1.Service)
	// OnServiceUpdate is called whenever modification of an existing
	// service object is observed.
	OnServiceUpdate(oldService, service *corev1.Service)
	// OnServiceDelete is called whenever deletion of an existing service
	// object is observed.
	OnServiceDelete(service *corev1.Service)
	// OnServiceSynced is called once all the initial event handlers were
	// called and the state is fully propagated to local cache.
	OnServiceSynced()
}

// proxyRun defines the interface to run a specified ProxyServer
type proxyRun interface {
	Run() error
	CleanupAndExit() error
}

// NewServiceConfig creates a new ServiceConfig.
func NewServiceConfig(serviceInformer coreinformers.ServiceInformer, resyncPeriod time.Duration) *ServiceConfig {
	result := &ServiceConfig{
		listerSynced: serviceInformer.Informer().HasSynced,
	}

	serviceInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    result.handleAddService,
			UpdateFunc: result.handleUpdateService,
			DeleteFunc: result.handleDeleteService,
		},
		resyncPeriod,
	)

	return result
}

// RegisterEventHandler registers a handler which is called on every service change.
func (c *ServiceConfig) RegisterEventHandler(handler ServiceHandler) {
	c.eventHandlers = append(c.eventHandlers, handler)
}

// Run waits for cache synced and invokes handlers after syncing.
func (c *ServiceConfig) Run(stopCh <-chan struct{}) {
	klog.Info("Starting service config controller")

	if !cache.WaitForNamedCacheSync("service config", stopCh, c.listerSynced) {
		return
	}

	for i := range c.eventHandlers {
		klog.V(3).Info("Calling handler.OnServiceSynced()")
		c.eventHandlers[i].OnServiceSynced()
	}
}

func (c *ServiceConfig) handleAddService(obj interface{}) {
	service, ok := obj.(*corev1.Service)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
		return
	}
	//for i := range c.eventHandlers {
	//	klog.V(4).Info("Calling handler.OnServiceAdd")
	//	c.eventHandlers[i].OnServiceAdd(service)
	//}
	fmt.Println("############ handleAddService> service:", service)
}

func (c *ServiceConfig) handleUpdateService(oldObj, newObj interface{}) {
	oldService, ok := oldObj.(*corev1.Service)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("unexpected object type: %v", oldObj))
		return
	}
	service, ok := newObj.(*corev1.Service)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("unexpected object type: %v", newObj))
		return
	}

	fmt.Println("############ handleUpdateService> oldService:", oldService, "\nservice: ", service)
}

func (c *ServiceConfig) handleDeleteService(obj interface{}) {
	service, ok := obj.(*corev1.Service)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("unexpected object type: %v", obj))
		return
	}
	fmt.Println("############ handleDeleteService> service:", service)
}

*/
