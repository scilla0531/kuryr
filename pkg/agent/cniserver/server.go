package cniserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	cnitypes "github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"projectkuryr/kuryr/pkg/agent/config"
	"projectkuryr/kuryr/pkg/agent/interfacestore"
	"projectkuryr/kuryr/pkg/agent/openflow"
	"projectkuryr/kuryr/pkg/agent/route"
	"projectkuryr/kuryr/pkg/agent/types"
	"projectkuryr/kuryr/pkg/agent/util"
	crdclientset "projectkuryr/kuryr/pkg/client/clientset/versioned"
	cniv1alpha1 "projectkuryr/kuryr/pkg/apis/cni/v1alpha1"
	cnipb "projectkuryr/kuryr/pkg/apis/cni/v1alpha1"
	"sync"

	"projectkuryr/kuryr/pkg/cni"
	"projectkuryr/kuryr/pkg/ovs/ovsconfig"
	"time"
)

const (
	// networkReadyTimeout is the maximum time the CNI server will wait for network ready when processing CNI Add
	// requests. If timeout occurs, tryAgainLaterResponse will be returned.
	// The default runtime request timeout of kubelet is 2 minutes.
	// https://github.com/kubernetes/kubernetes/blob/v1.19.3/staging/src/k8s.io/kubelet/config/v1beta1/types.go#L451
	// networkReadyTimeout is set to a shorter time so it returns a clear message to the runtime.
	networkReadyTimeout = 30 * time.Second
)

// containerAccessArbitrator is used to ensure that concurrent goroutines cannot perfom operations
// on the same containerID. Other parts of the code make this assumption (in particular the
// InstallPodFlows / UninstallPodFlows methods of the OpenFlow client, which are invoked
// respectively by CmdAdd and CmdDel). The idea is to simply the locking requirements for the rest
// of the code by ensuring that all the requests for a given container are serialized.
type containerAccessArbitrator struct {
	mutex             sync.Mutex
	cond              *sync.Cond
	busyContainerKeys map[string]bool // used as a set of container keys
}

// lockContainer prevents other goroutines from accessing containerKey. If containerKey is already
// locked by another goroutine, this function will block until the container is available. Every
// call to lockContainer must be followed by a call to unlockContainer on the same containerKey.
func (arbitrator *containerAccessArbitrator) lockContainer(containerKey string) {
	arbitrator.cond.L.Lock()
	defer arbitrator.cond.L.Unlock()
	for {
		_, ok := arbitrator.busyContainerKeys[containerKey]
		if !ok {
			break
		}
		arbitrator.cond.Wait()
	}
	arbitrator.busyContainerKeys[containerKey] = true
}

// unlockContainer releases access to containerKey.
func (arbitrator *containerAccessArbitrator) unlockContainer(containerKey string) {
	arbitrator.cond.L.Lock()
	defer arbitrator.cond.L.Unlock()
	delete(arbitrator.busyContainerKeys, containerKey)
	arbitrator.cond.Broadcast()
}

type CNIServer struct {
	cniSocket            string
	supportedCNIVersions map[string]bool
	serverVersion        string
	crdClient 			crdclientset.Interface

	nodeConfig           *config.NodeConfig
	hostProcPathPrefix   string
	kubeClient           clientset.Interface
	containerAccess      *containerAccessArbitrator
	podConfigurator      *podConfigurator
	isChaining           bool
	routeClient          route.Interface
	//networkReadyCh notifies that the network is ready so new Pods can be created. Therefore, CmdAdd waits for it.
	networkReadyCh <-chan struct{}
}

var supportedCNIVersionSet map[string]bool

type RuntimeDNS struct {
	Nameservers []string `json:"servers,omitempty"`
	Search      []string `json:"searches,omitempty"`
}

type RuntimeConfig struct {
	DNS RuntimeDNS `json:"dns"`
}

type NetworkConfig struct {
	CNIVersion string          `json:"cniVersion,omitempty"`
	Name       string          `json:"name,omitempty"`
	Type       string          `json:"type,omitempty"`
	DeviceID   string          `json:"deviceID"` // PCI address of a VF
	MTU        int             `json:"mtu,omitempty"`
	DNS        cnitypes.DNS    `json:"dns"`
	//IPAM       ipam.IPAMConfig `json:"ipam,omitempty"`
	// Options to be passed in by the runtime.
	//RuntimeConfig RuntimeConfig `json:"runtimeConfig"`

	//RawPrevResult map[string]interface{} `json:"prevResult,omitempty"`
	//PrevResult    cnitypes.Result        `json:"-"`
}

type CNIConfig struct {
	*NetworkConfig
	*cniv1alpha1.CniCmdArgs
	*k8sArgs
}

func (s *CNIServer) loadNetworkConfig(request *cnipb.CniCmdRequest) (*CNIConfig, error) {
	cniConfig := &CNIConfig{}
	cniConfig.CniCmdArgs = request.CniArgs
	if err := json.Unmarshal(request.CniArgs.NetworkConfiguration, cniConfig); err != nil {
		return cniConfig, err
	}
	cniConfig.k8sArgs = &k8sArgs{}
	if err := cnitypes.LoadArgs(request.CniArgs.Args, cniConfig.k8sArgs); err != nil {
		return cniConfig, err
	}

	klog.V(3).Infof("Load network configurations: %v", cniConfig)
	return cniConfig, nil
}

func (s *CNIServer) generateCNIErrorResponse(cniErrorCode cnipb.ErrorCode, cniErrorMsg string) *cnipb.CniCmdResponse {
	return &cnipb.CniCmdResponse{
		Error: &cnipb.Error{
			Code:    cniErrorCode,
			Message: cniErrorMsg,
		},
	}
}

func (s *CNIServer) decodingFailureResponse(what string) *cnipb.CniCmdResponse {
	return s.generateCNIErrorResponse(
		cnipb.ErrorCode_DECODING_FAILURE,
		fmt.Sprintf("Failed to decode %s", what),
	)
}

func (s *CNIServer) isCNIVersionSupported(reqVersion string) bool {
	_, exist := s.supportedCNIVersions[reqVersion]
	return exist
}

func (s *CNIServer) incompatibleCniVersionResponse(cniVersion string) *cnipb.CniCmdResponse {
	cniErrorCode := cnipb.ErrorCode_INCOMPATIBLE_CNI_VERSION
	cniErrorMsg := fmt.Sprintf("Unsupported CNI version [%s], supported versions %s", cniVersion, version.All.SupportedVersions())
	return s.generateCNIErrorResponse(cniErrorCode, cniErrorMsg)
}

func (s *CNIServer) checkRequestMessage(request *cnipb.CniCmdRequest) (*CNIConfig, *cnipb.CniCmdResponse) {
	cniConfig, err := s.loadNetworkConfig(request)
	if err != nil {
		klog.Errorf("Failed to parse network configuration: %v", err)
		return nil, s.decodingFailureResponse("network config")
	}
	cniVersion := cniConfig.CNIVersion
	// Check if CNI version in the request is supported
	if !s.isCNIVersionSupported(cniVersion) {
		klog.Errorf(fmt.Sprintf("Unsupported CNI version [%s], supported CNI versions %s", cniVersion, version.All.SupportedVersions()))
		return cniConfig, s.incompatibleCniVersionResponse(cniVersion)
	}

	return cniConfig, nil
}

func (s *CNIServer) tryAgainLaterResponse() *cnipb.CniCmdResponse {
	cniErrorCode := cnipb.ErrorCode_TRY_AGAIN_LATER
	cniErrorMsg := "Server is busy, please retry later"
	return s.generateCNIErrorResponse(cniErrorCode, cniErrorMsg)
}

// reconcile performs startup reconciliation for the CNI server. The CNI server is in charge of
// installing Pod flows, so as part of this reconciliation process we retrieve the Pod list from the
// K8s apiserver and replay the necessary flows.
func (s *CNIServer) reconcile() error {
	klog.Infof("Reconciliation for CNI server")
	// For performance reasons, use ResourceVersion="0" in the ListOptions to ensure the request is served from
	// the watch cache in kube-apiserver.
	pods, err := s.kubeClient.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector:   "spec.nodeName=" + s.nodeConfig.Name,
		ResourceVersion: "0",
	})
	if err != nil {
		return fmt.Errorf("failed to list Pods running on Node %s: %v", s.nodeConfig.Name, err)
	}

	return s.podConfigurator.reconcile(pods.Items, s.containerAccess)
}

func (s *CNIServer) CmdAdd(ctx context.Context, request *cnipb.CniCmdRequest) (*cnipb.CniCmdResponse, error) {
	klog.Infof("Received CmdAdd request %v", request)
	cniConfig, response := s.checkRequestMessage(request)
	if response != nil {
		return response, nil
	}

	select {
	case <-time.After(networkReadyTimeout):
		klog.Errorf("Cannot process CmdAdd request for container %v because network is not ready", cniConfig.ContainerId)
		return s.tryAgainLaterResponse(), nil
	case <-s.networkReadyCh:
	}

	cniVersion := cniConfig.CNIVersion
	result := &current.Result{CNIVersion: cniVersion}
	netNS := s.hostNetNsPath(cniConfig.Netns)
	isInfraContainer := isInfraContainer(netNS)

	success := false
	defer func() {
		// Rollback to delete configurations once ADD is failure.
		if !success {
			if isInfraContainer {
				klog.Warningf("CmdAdd for container %v failed, and try to rollback", cniConfig.ContainerId)
				if _, err := s.CmdDel(ctx, request); err != nil {
					klog.Warningf("Failed to rollback after CNI add failure: %v", err)
				}
			} else {
				klog.Warningf("CmdAdd for container %v failed", cniConfig.ContainerId)
			}
		}
	}()

	infraContainer := cniConfig.getInfraContainer()
	s.containerAccess.lockContainer(infraContainer)
	defer s.containerAccess.unlockContainer(infraContainer)

	var resultBytes bytes.Buffer
	_ = result.PrintTo(&resultBytes)
	klog.Infof("CmdAdd for container %v succeeded", cniConfig.ContainerId)
	// mark success as true to avoid rollback
	success = true
	return &cnipb.CniCmdResponse{CniResult: resultBytes.Bytes()}, nil
}

func (s *CNIServer) CmdDel(_ context.Context, request *cnipb.CniCmdRequest) (
	*cnipb.CniCmdResponse, error) {
	klog.Infof("Received CmdDel request %v", request)

	cniConfig, response := s.checkRequestMessage(request)
	if response != nil {
		return response, nil
	}

	infraContainer := cniConfig.getInfraContainer()
	s.containerAccess.lockContainer(infraContainer)
	defer s.containerAccess.unlockContainer(infraContainer)

	klog.Infof("CmdDel for container %v succeeded", cniConfig.ContainerId)
	return &cnipb.CniCmdResponse{CniResult: []byte("")}, nil
}

func (s *CNIServer) CmdCheck(_ context.Context, request *cnipb.CniCmdRequest) (
	*cnipb.CniCmdResponse, error) {
	klog.Infof("Received CmdCheck request %v", request)

	cniConfig, response := s.checkRequestMessage(request)
	if response != nil {
		return response, nil
	}

	infraContainer := cniConfig.getInfraContainer()
	s.containerAccess.lockContainer(infraContainer)
	defer s.containerAccess.unlockContainer(infraContainer)

	klog.Infof("CmdCheck for container %v succeeded", cniConfig.ContainerId)
	return &cnipb.CniCmdResponse{CniResult: []byte("")}, nil
}

func New(
	cniSocket, hostProcPathPrefix string,
	crdClient crdclientset.Interface,
	networkReadyCh <-chan struct{},
) *CNIServer {
	return &CNIServer{
		cniSocket:            cniSocket,
		supportedCNIVersions: supportedCNIVersionSet,
		serverVersion:        cni.KuryrCNIVersion,
		crdClient: 			crdClient,
		networkReadyCh:       networkReadyCh,
		hostProcPathPrefix: hostProcPathPrefix,
	}
}

func (s *CNIServer) Initialize(
	ovsBridgeClient ovsconfig.OVSBridgeClient,
	ofClient openflow.Client,
	ifaceStore interfacestore.InterfaceStore,
	entityUpdates chan<- types.EntityReference,
) error {
	var err error
	s.podConfigurator, err = newPodConfigurator(
		ovsBridgeClient,
		ofClient,
		s.nodeConfig.GatewayConfig.MAC,
		ovsBridgeClient.GetOVSDatapathType(), ovsBridgeClient.IsHardwareOffloadEnabled(),
		)
	if err != nil {
		return fmt.Errorf("error during initialize podConfigurator: %v", err)
	}
	if err := s.reconcile(); err != nil {
		return fmt.Errorf("error during initial reconciliation for CNI server: %v", err)
	}
	return nil
}

func (s *CNIServer) Run(stopCh <-chan struct{}) {
	klog.Info("Starting CNI server")
	defer klog.Info("Shutting down CNI server")

	listener, err := util.ListenLocalSocket(s.cniSocket)
	if err != nil {
		klog.Fatalf("Failed to bind on %s: %v", s.cniSocket, err)
	}
	rpcServer := grpc.NewServer()

	cnipb.RegisterCniServer(rpcServer, s)
	klog.Info("CNI server is listening ...")
	go func() {
		if err := rpcServer.Serve(listener); err != nil {
			klog.Errorf("Failed to serve connections: %v", err)
		}
	}()
	<-stopCh
}
