package app

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	"projectkuryr/kuryr/pkg/agent/interfacestore"

	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog"
	"net/http"
	"projectkuryr/kuryr/pkg/agent/cniserver"
	"projectkuryr/kuryr/pkg/agent/openflow"
	ofconfig "projectkuryr/kuryr/pkg/ovs/openflow"
	kuryrinformers "projectkuryr/kuryr/pkg/client/informers/externalversions"
	"projectkuryr/kuryr/pkg/healthcheck"
	"projectkuryr/kuryr/pkg/k8s"
	"projectkuryr/kuryr/pkg/ovs/ovsconfig"
	"projectkuryr/kuryr/pkg/signals"
	"projectkuryr/kuryr/pkg/version"
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

var ErrNotInCluster = errors.New("unable to load in-cluster configuration.")

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

func run(o *Options) error {
	klog.Infof("Starting Kuryr agent (version %s)", version.GetFullVersion())
	// Create K8s Clientset, CRD Clientset and SharedInformerFactory for the given config.
	//k8sClient, _, crdClient, err := k8s.CreateClients(o.config.ClientConnection, o.config.KubeAPIServerOverride)
	_, _, crdClient, err := k8s.CreateClientsCrd(o.config.ClientConnection, "")
	if err != nil {
		return fmt.Errorf("error creating k8s clients: %v", err)
	}

	//informerFactory := informers.NewSharedInformerFactory(client, informerDefaultResync)
	//podInformer := informerFactory.Core().V1().Pods()
	crdInformerFactory := kuryrinformers.NewSharedInformerFactory(crdClient, informerDefaultResync)
	kpInformer := crdInformerFactory.Openstack().V1alpha1().KuryrPorts()

	// Create an ifaceStore that caches network interfaces managed by this node.
	ifaceStore := interfacestore.NewInterfaceStore()
	klog.Infof("GetContainerInterfaceNum: %+v\n", ifaceStore.GetContainerInterfaceNum())

	// Create ovsdb and openflow clients.

	ovsdbAddress := ovsconfig.GetConnAddress(o.config.OVSRunDir) // default is /var/run/openvswitch/db.sock
	ovsdbConnection, err := ovsconfig.NewOVSDBConnectionUDS(ovsdbAddress)
	if err != nil {
		// TODO: ovsconfig.NewOVSDBConnectionUDS might return timeout in the future, need to add retry
		return fmt.Errorf("error connecting OVSDB: %v", err)
	}
	defer ovsdbConnection.Close()

	ovsDatapathType := ovsconfig.OVSDatapathType(o.config.OVSDatapathType)
	ovsBridgeClient := ovsconfig.NewOVSBridge(o.config.OVSBridge, ovsDatapathType, ovsdbConnection)
	klog.Infof("<ovsBridgeClient>: %+v\n", *ovsBridgeClient)

	ovsPorts, err := ovsBridgeClient.GetPortList()
	if err != nil {
		klog.Infof("GetPortList Error: %s\n", err)
	}
	for _, ovsPort := range ovsPorts {
		klog.Infof("<ovsPort>: %+v\n", ovsPort)
	}

	ovsBridgeMgmtAddr := ofconfig.GetMgmtAddress(o.config.OVSRunDir, o.config.OVSBridge)
	klog.Infof("<config>: ovsBridgeMgmtAddr : %+v\n", ovsBridgeMgmtAddr)
	ofClient := openflow.NewClient(o.config.OVSBridge, ovsBridgeMgmtAddr, ovsDatapathType,
		false,
		false,
		false)

	ofPorts := ofClient.GetFlowTableStatus()
	for _, ofport := range ofPorts {
		klog.Infof("ofport: %+v\n", ofport)
	}

	cniServer := cniserver.New(
		o.config.CNISocket,
		o.config.HostProcPathPrefix,
		crdClient,
		kpInformer,
		nil,
		nil,
		nil)

	err = cniServer.InitializeCniServer(ovsBridgeClient)
	if err != nil {
		return fmt.Errorf("error initializing CNI server: %v", err)
	}
	// set up signal capture: the first SIGTERM / SIGINT signal is handled gracefully and will
	// cause the stopCh channel to be closed; if another signal is received before the program
	// exits, we will force exit.
	stopCh := signals.RegisterSignalHandlers()
	crdInformerFactory.Start(stopCh)
	go cniServer.Run(stopCh)

	<-stopCh
	klog.Info("Stopping Kuryr agent")
	return nil
}