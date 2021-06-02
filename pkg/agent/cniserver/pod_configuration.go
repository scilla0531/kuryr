package cniserver

import (
	"fmt"
	cnitypes "github.com/containernetworking/cni/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"projectkuryr/kuryr/pkg/k8s"

	"github.com/containernetworking/cni/pkg/types/current"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"net"
	"projectkuryr/kuryr/pkg/agent/interfacestore"
	"projectkuryr/kuryr/pkg/agent/openflow"
	"projectkuryr/kuryr/pkg/ovs/ovsconfig"
	"strings"
)

type vethPair struct {
	name      string
	ifIndex   int
	peerIndex int
}

type k8sArgs struct {
	cnitypes.CommonArgs
	K8S_POD_NAME               cnitypes.UnmarshallableString
	K8S_POD_NAMESPACE          cnitypes.UnmarshallableString
	K8S_POD_INFRA_CONTAINER_ID cnitypes.UnmarshallableString
}

const (
	ovsExternalIDMAC          = "attached-mac"
	ovsExternalIDIP           = "ip-address"
	ovsExternalIDContainerID  = "container-id"
	ovsExternalIDPodName      = "pod-name"
	ovsExternalIDPodNamespace = "pod-namespace"
)

const (
	defaultOVSInterfaceType int = iota //nolint suppress deadcode check for windows
	internalOVSInterfaceType
)

type podConfigurator struct {
	ovsBridgeClient ovsconfig.OVSBridgeClient
	ofClient        openflow.Client
	//routeClient     route.Interface
	ifaceStore      interfacestore.InterfaceStore
	gatewayMAC      net.HardwareAddr
	ifConfigurator  *ifConfigurator
	// entityUpdates is a channel for notifying updates of local endpoints / entities (most notably Pod)
	// to other components which may benefit from this information, i.e NetworkPolicyController.
	//entityUpdates chan<- types.EntityReference
}

func newPodConfigurator(
	ovsBridgeClient ovsconfig.OVSBridgeClient,
	ofClient openflow.Client,
	//routeClient route.Interface,
	//ifaceStore interfacestore.InterfaceStore,
	gatewayMAC net.HardwareAddr,
	ovsDatapathType ovsconfig.OVSDatapathType,
	isOvsHardwareOffloadEnabled bool,
	//entityUpdates chan<- types.EntityReference,
) (*podConfigurator, error) {
	ifConfigurator, err := newInterfaceConfigurator(ovsDatapathType, isOvsHardwareOffloadEnabled)
	if err != nil {
		return nil, err
	}
	return &podConfigurator{
		ovsBridgeClient: ovsBridgeClient,
		ofClient:        ofClient,
		gatewayMAC:      gatewayMAC,
		ifConfigurator:  ifConfigurator,

	}, nil
}

func parseContainerIPs(ipcs []*current.IPConfig) ([]net.IP, error) {
	var ips []net.IP
	for _, ipc := range ipcs {
		ips = append(ips, ipc.Address.IP)
	}
	if len(ips) > 0 {
		return ips, nil
	} else {
		return nil, fmt.Errorf("failed to find a valid IP address")
	}
}

func buildContainerConfig(
	interfaceName, containerID, podName, podNamespace string,
	containerIface *current.Interface,
	ips []*current.IPConfig) *interfacestore.InterfaceConfig {
	containerIPs, err := parseContainerIPs(ips)
	if err != nil {
		klog.Errorf("Failed to find container %s IP", containerID)
	}
	// containerIface.Mac should be a valid MAC string, otherwise it should throw error before
	containerMAC, _ := net.ParseMAC(containerIface.Mac)
	return interfacestore.NewContainerInterface(
		interfaceName,
		containerID,
		podName,
		podNamespace,
		containerMAC,
		containerIPs)
}

func getContainerIPsString(ips []net.IP) string {
	var containerIPs []string
	for _, ip := range ips {
		containerIPs = append(containerIPs, ip.String())
	}
	return strings.Join(containerIPs, ",")
}

// BuildOVSPortExternalIDs parses OVS port external_ids from InterfaceConfig.
// external_ids are used to compare and sync container interface configuration.
func BuildOVSPortExternalIDs(containerConfig *interfacestore.InterfaceConfig) map[string]interface{} {
	externalIDs := make(map[string]interface{})
	externalIDs[ovsExternalIDMAC] = containerConfig.MAC.String()
	externalIDs[ovsExternalIDContainerID] = containerConfig.ContainerID
	externalIDs[ovsExternalIDIP] = getContainerIPsString(containerConfig.IPs)
	externalIDs[ovsExternalIDPodName] = containerConfig.PodName
	externalIDs[ovsExternalIDPodNamespace] = containerConfig.PodNamespace
	return externalIDs
}

func (pc *podConfigurator) createOVSPort(ovsPortName string, ovsAttachInfo map[string]interface{}) (string, error) {
	var portUUID string
	var err error
	switch pc.ifConfigurator.getOVSInterfaceType() {
	case internalOVSInterfaceType:
		portUUID, err = pc.ovsBridgeClient.CreateInternalPort(ovsPortName, 0, ovsAttachInfo)
	default:
		portUUID, err = pc.ovsBridgeClient.CreatePort(ovsPortName, ovsPortName, ovsAttachInfo)
	}
	if err != nil {
		klog.Errorf("Failed to add OVS port %s, remove from local cache: %v", ovsPortName, err)
		return "", err
	} else {
		return portUUID, nil
	}
}


func (pc *podConfigurator) reconcile(pods []corev1.Pod, containerAccess *containerAccessArbitrator) error {
	// desiredPods is the set of Pods that should be present, based on the
	// current list of Pods got from the Kubernetes API.
	desiredPods := sets.NewString()
	// actualPods is the set of Pods that are present, based on the container
	// interfaces got from the OVSDB.
	actualPods := sets.NewString()
	// knownInterfaces is the list of interfaces currently in the local cache.
	knownInterfaces := pc.ifaceStore.GetInterfacesByType(interfacestore.ContainerInterface)

	for _, pod := range pods {
		// Skip Pods for which we are not in charge of the networking.
		if pod.Spec.HostNetwork {
			continue
		}
		desiredPods.Insert(k8s.NamespacedName(pod.Namespace, pod.Name))
	}
	for _, containerConfig := range knownInterfaces {
		namespacedName := k8s.NamespacedName(containerConfig.PodNamespace, containerConfig.PodName)
		actualPods.Insert(namespacedName)
		if desiredPods.Has(namespacedName) {
			// This interface matches an existing Pod.
			// We rely on the interface cache / store - which is initialized from the persistent
			// OVSDB - to map the Pod to its interface configuration. The interface
			// configuration includes the parameters we need to replay the flows.
			klog.V(4).Infof("Syncing interface %s for Pod %s", containerConfig.InterfaceName, namespacedName)
			if err := pc.ofClient.InstallPodFlows(
				containerConfig.InterfaceName,
				containerConfig.IPs,
				containerConfig.MAC,
				uint32(containerConfig.OFPort),
			); err != nil {
				klog.Errorf("Error when re-installing flows for Pod %s", namespacedName)
			}
		} else {
			// clean-up and delete interface
			klog.V(4).Infof("Deleting interface %s", containerConfig.InterfaceName)
			if err := pc.removeInterfaces(containerConfig.ContainerID); err != nil {
				klog.Errorf("Failed to delete interface %s: %v", containerConfig.InterfaceName, err)
			}
			// interface should no longer be in store after the call to removeInterfaces
		}
	}

	missingPods := desiredPods.Difference(actualPods)
	pc.reconcileMissingPods(missingPods, containerAccess)
	return nil
}

func (pc *podConfigurator) connectInterfaceToOVSCommon(ovsPortName string, containerConfig *interfacestore.InterfaceConfig) error {
	// create OVS Port and add attach container configuration into external_ids
	containerID := containerConfig.ContainerID
	klog.V(2).Infof("Adding OVS port %s for container %s", ovsPortName, containerID)
	ovsAttachInfo := BuildOVSPortExternalIDs(containerConfig)
	portUUID, err := pc.createOVSPort(ovsPortName, ovsAttachInfo)
	if err != nil {
		return fmt.Errorf("failed to add OVS port for container %s: %v", containerID, err)
	}
	// Remove OVS port if any failure occurs in later manipulation.
	defer func() {
		if err != nil {
			_ = pc.ovsBridgeClient.DeletePort(portUUID)
		}
	}()

	// GetOFPort will wait for up to 1 second for OVSDB to report the OFPort number.
	ofPort, err := pc.ovsBridgeClient.GetOFPort(ovsPortName)
	if err != nil {
		return fmt.Errorf("failed to get of_port of OVS port %s: %v", ovsPortName, err)
	}

	klog.V(2).Infof("Setting up Openflow entries for container %s", containerID)
	err = pc.ofClient.InstallPodFlows(ovsPortName, containerConfig.IPs, containerConfig.MAC, uint32(ofPort))
	if err != nil {
		return fmt.Errorf("failed to add Openflow entries for container %s: %v", containerID, err)
	}
	containerConfig.OVSPortConfig = &interfacestore.OVSPortConfig{PortUUID: portUUID, OFPort: ofPort}
	// Add containerConfig into local cache
	//pc.ifaceStore.AddInterface(containerConfig)
	// Notify the Pod update event to required components.
	//pc.entityUpdates <- types.EntityReference{
	//	Pod: &v1beta2.PodReference{Name: containerConfig.PodName, Namespace: containerConfig.PodNamespace},
	//}
	return nil
}
// disconnectInterfaceFromOVS disconnects an existing interface from ovs br-int.
func (pc *podConfigurator) disconnectInterfaceFromOVS(containerConfig *interfacestore.InterfaceConfig) error {
	containerID := containerConfig.ContainerID
	klog.V(2).Infof("Deleting Openflow entries for container %s", containerID)
	if err := pc.ofClient.UninstallPodFlows(containerConfig.InterfaceName); err != nil {
		return fmt.Errorf("failed to delete Openflow entries for container %s: %v", containerID, err)
		// We should not delete OVS port if Pod flows deletion fails, otherwise
		// it is possible a new Pod will reuse the reclaimed ofport number, and
		// the OVS flows added for the new Pod can conflict with the stale
		// flows of the deleted Pod.
	}

	klog.V(2).Infof("Deleting OVS port %s for container %s", containerConfig.PortUUID, containerID)
	// TODO: handle error and introduce garbage collection for failure on deletion
	if err := pc.ovsBridgeClient.DeletePort(containerConfig.PortUUID); err != nil {
		return fmt.Errorf("failed to delete OVS port for container %s: %v", containerID, err)
	}
	// Remove container configuration from cache.
	pc.ifaceStore.DeleteInterface(containerConfig)
	klog.Infof("Removed interfaces for container %s", containerID)
	return nil
}

func (pc *podConfigurator) removeInterfaces(containerID string) error {
	containerConfig, found := pc.ifaceStore.GetContainerInterface(containerID)
	if !found {
		klog.V(2).Infof("Did not find the port for container %s in local cache", containerID)
		return nil
	}

	// Deleting veth devices and OVS port must be called after Openflows are uninstalled.
	// Otherwise there could be a race condition:
	// 1. Pod A's ofport was released
	// 2. Pod B got the ofport released above
	// 3. Flows for Pod B were installed
	// 4. Flows for Pod A were uninstalled
	// Because Pod A and Pod B had same ofport, they had overlapping flows, e.g. the
	// classifier flow in table 0 which has only in_port as the match condition, then
	// step 4 can remove flows owned by Pod B by mistake.
	// Note that deleting the interface attached to an OVS port can release the ofport.
	if err := pc.disconnectInterfaceFromOVS(containerConfig); err != nil {
		return err
	}

	if err := pc.ifConfigurator.removeContainerLink(containerID, containerConfig.InterfaceName); err != nil {
		return err
	}
	return nil
}