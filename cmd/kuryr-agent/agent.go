package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog"
	"net"
	"net/http"
	"projectkuryr/kuryr/pkg/healthcheck"
	"projectkuryr/kuryr/pkg/k8s"
	"projectkuryr/kuryr/pkg/utils/node"
	"projectkuryr/kuryr/pkg/version"
	"sync"
	"time"
)

const informerDefaultResync = 10*time.Minute

// Valid values for event types (new types could be added in future)
const (
	// Information only and will not cause any problems
	EventTypeNormal string = "Normal"
	// These events are to warn that something might go wrong
	EventTypeWarning string = "Warning"
)


var (
	// Scheme defines methods for serializing and deserializing API objects.
	Scheme = runtime.NewScheme()
	// Codecs provides methods for retrieving codecs and serializers for specific
	// versions and content types.
	Codecs = serializer.NewCodecFactory(Scheme, serializer.EnableStrict)
)

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
	Proxier  Provider

}

// NewProxyServer returns a new ProxyServer.
func NewProxyServer(o *Options) (*ProxyServer, error) {
	return newProxyServer(o.config, o.CleanupAndExit, o.master)
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

var _ Provider = &Proxier{}

// ServiceConfig tracks a set of service configurations.
type ServiceConfig struct {
	listerSynced  cache.InformerSynced
	eventHandlers []ServiceHandler
}

type Proxier struct {
	serviceChanges   *ServiceChangeTracker
}

func (proxier *Proxier) Sync() {
	fmt.Println("Proxier.Sync")
}

func (proxier *Proxier) SyncLoop() {
	fmt.Println("Proxier.SyncLoop")
}

func (proxier *Proxier) OnServiceUpdate(oldService, service *corev1.Service) {
	fmt.Println("Proxier.OnServiceUpdate")
}

func (proxier *Proxier) OnServiceDelete(service *corev1.Service) {
	proxier.OnServiceUpdate(service, nil)
}

func (proxier *Proxier) OnServiceAdd(service *corev1.Service) {
	proxier.OnServiceUpdate(nil, service)
}

func (proxier *Proxier) OnServiceSynced() {
	fmt.Println("Proxier.OnServiceUpdate")
}

// serviceChange contains all changes to services that happened since proxy rules were synced.  For a single object,
// changes are accumulated, i.e. previous is state from before applying the changes,
// current is state after applying all of the changes.
type serviceChange struct {
	previous ServiceMap
	current  ServiceMap
}

// ServiceMap maps a service to its ServicePort.
type ServiceMap map[ServicePortName]ServicePort
// ServicePortName carries a namespace + name + portname.  This is the unique
// identifier for a load-balanced service.
type ServicePortName struct {
	types.NamespacedName
	Port     string
	Protocol corev1.Protocol
}

// ServicePort is an interface which abstracts information about a service.
type ServicePort interface {
	// String returns service string.  An example format can be: `IP:Port/Protocol`.
	String() string
	// GetClusterIP returns service cluster IP in net.IP format.
	ClusterIP() net.IP
	// GetPort returns service port if present. If return 0 means not present.
	Port() int
	// GetSessionAffinityType returns service session affinity type
	SessionAffinityType() corev1.ServiceAffinity
	// GetStickyMaxAgeSeconds returns service max connection age
	StickyMaxAgeSeconds() int
	// ExternalIPStrings returns service ExternalIPs as a string array.
	ExternalIPStrings() []string
	// LoadBalancerIPStrings returns service LoadBalancerIPs as a string array.
	LoadBalancerIPStrings() []string
	// GetProtocol returns service protocol.
	Protocol() corev1.Protocol
	// LoadBalancerSourceRanges returns service LoadBalancerSourceRanges if present empty array if not
	LoadBalancerSourceRanges() []string
	// GetHealthCheckNodePort returns service health check node port if present.  If return 0, it means not present.
	HealthCheckNodePort() int
	// GetNodePort returns a service Node port if present. If return 0, it means not present.
	NodePort() int
	// GetOnlyNodeLocalEndpoints returns if a service has only node local endpoints
	OnlyNodeLocalEndpoints() bool
	// TopologyKeys returns service TopologyKeys as a string array.
	TopologyKeys() []string
}

// This handler is invoked by the apply function on every change. This function should not modify the
// ServiceMap's but just use the changes for any Proxier specific cleanup.
type processServiceMapChangeFunc func(previous, current ServiceMap)

// ServiceChangeTracker carries state about uncommitted changes to an arbitrary number of
// Services, keyed by their namespace and name.
type ServiceChangeTracker struct {
	// lock protects items.
	lock sync.Mutex
	// items maps a service to its serviceChange.
	items map[types.NamespacedName]*serviceChange
	// makeServiceInfo allows proxier to inject customized information when processing service.
	//makeServiceInfo         makeServicePortFunc
	processServiceMapChange processServiceMapChangeFunc
	ipFamily                corev1.IPFamily

	recorder record.EventRecorder
}

// proxyRun defines the interface to run a specified ProxyServer
type proxyRun interface {
	Run() error
	CleanupAndExit() error
}


var ErrNotInCluster = errors.New("unable to load in-cluster configuration.")

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

func serveHealthz(hz healthcheck.ProxierHealthUpdater, errCh chan error) {
	if hz == nil {
		return
	}

	fn := func() {
		err := hz.Run()
		if err != nil {
			klog.Errorf("healthz server failed: %v", err)
			if errCh != nil {
				errCh <- fmt.Errorf("healthz server failed: %v", err)
				// if in hardfail mode, never retry again
				blockCh := make(chan error)
				<-blockCh
			}
		} else {
			klog.Errorf("healthz server returned without error")
		}
	}
	go wait.Until(fn, 5*time.Second, wait.NeverStop)
}

func serveMetrics(bindAddress, kuryrAgentMetrics string, enableProfiling bool, errCh chan error) {
	if len(bindAddress) == 0 {
		return
	}

	proxyMux := mux.NewPathRecorderMux("kuryr-agent")
	healthz.InstallHandler(proxyMux)
	healthz.InstallReadyzHandler(proxyMux)
	healthz.InstallLivezHandler(proxyMux)

	proxyMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		fmt.Fprintf(w, "%s\n", kuryrAgentMetrics)
	})

	//lint:ignore SA1019 See the Metrics Stability Migration KEP
	proxyMux.Handle("/metrics", legacyregistry.Handler())

	if enableProfiling {
		routes.Profiling{}.Install(proxyMux)
	}

	fn := func() {
		err := http.ListenAndServe(bindAddress, proxyMux)
		if err != nil {
			err = fmt.Errorf("starting metrics server failed: %v", err)
			if err != nil {
				klog.Errorf("metrics server failed: %v", err)
			}
			if errCh != nil {
				errCh <- err
				// if in hardfail mode, never retry again
				blockCh := make(chan error)
				<-blockCh
			}
		}
	}
	go wait.Until(fn, 5*time.Second, wait.NeverStop)
}

func newAgentCommand() *cobra.Command {
	opts := newOptions()

	cmd := &cobra.Command{
		Use:  "kuryr-agent",
		Long: "The kuryr agent runs on each node.",
		Run: func(cmd *cobra.Command, args []string) {
			klog.Infoln("newAgentCommand Run")

			if err := opts.complete(args); err != nil {
				klog.Fatalf("Failed to complete: %v", err)
			}
			if err := opts.validate(args); err != nil {
				klog.Fatalf("Failed to validate: %v", err)
			}
			if err := opts.Run(); err != nil {
				klog.Fatalf("Error running agent: %v", err)
			}
		},
		Version: version.GetFullVersionWithRuntimeInfo(),
	}

	flags := cmd.Flags()
	opts.addFlags(flags)
	//// Install log flags
	//flags.AddGoFlagSet(flag.CommandLine)
	return cmd
}