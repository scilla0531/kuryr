package cniserver

import (
	"github.com/containernetworking/cni/pkg/types/current"
	"k8s.io/klog"
	"net"
	"projectkuryr/kuryr/pkg/ovs/ovsconfig"
)

type kpConfigurator struct {
	ovsBridgeClient ovsconfig.OVSBridgeClient
	gatewayMAC      net.HardwareAddr
	ifConfigurator  *ifConfigurator
}

func newKpConfigurator(
	ovsBridgeClient ovsconfig.OVSBridgeClient,
	ovsDatapathType ovsconfig.OVSDatapathType,
	isOvsHardwareOffloadEnabled bool,
) (*kpConfigurator, error) {
	ifConfigurator, err := newInterfaceConfigurator(ovsDatapathType, isOvsHardwareOffloadEnabled)
	if err != nil {
		return nil, err
	}
	return &kpConfigurator{
		ovsBridgeClient: ovsBridgeClient,
		ifConfigurator:  ifConfigurator,
	}, nil
}

func (kc *kpConfigurator) configureTap(
	containerID string,
	hostIfaceName string,
	containerNetNS string,
	containerIFDev string,
	mtu int,
	mac string,
	result *current.Result,
	createOVSPort bool,
	containerAccess *containerAccessArbitrator,
) error {
	err := kc.ifConfigurator.configureContainerTapLinkVeth(containerID, hostIfaceName, containerNetNS, containerIFDev, mtu, mac, result)
	if err != nil {
		klog.Errorf("configureContainerLink Error: %s\n", err)
		return err
	}

	if !createOVSPort {
		return nil
	}
	return nil
}

func (kc *kpConfigurator) configureInterfaces(
	podName string,
	podNameSpace string,
	containerID string,
	containerNetNS string,
	containerIFDev string,
	mtu int,
	sriovVFDeviceID string,
	result *current.Result,
	createOVSPort bool,
	containerAccess *containerAccessArbitrator,
) error {

	err := kc.ifConfigurator.configureContainerLink(podName, podNameSpace, containerID, containerNetNS, containerIFDev, mtu, sriovVFDeviceID, result)
	if err != nil {
		klog.Errorf("configureContainerLink Error: %s\n", err)
		return err
	}

	if !createOVSPort {
		return nil
	}

	/*
	hostIface := result.Interfaces[0]
	containerIface := result.Interfaces[1]

	// Delete veth pair if any failure occurs in later manipulation.
	success := false
	defer func() {
		if !success {
			_ = kc.ifConfigurator.removeContainerLink(containerID, hostIface.Name)
		}
	}()

	// ovs-vsctl -- --if-exists del-port $tapname -- add-port br-int $tapname -- set Interface $tapname external-ids:iface-id=$neutron_port external-ids:iface-status=active external-ids:attached-mac=$port_mac external-ids:vm-uuid=kuryr
	var containerConfig *interfacestore.InterfaceConfig
	if containerConfig, err = kc.connectInterfaceToOVS(podName, podNameSpace, containerID, hostIface, containerIface, result.IPs, containerAccess); err != nil {
		return fmt.Errorf("failed to connect to ovs for container %s: %v", containerID, err)
	} else {
		success = true
	}
	defer func() {
		if !success {
			_ = kc.disconnectInterfaceFromOVS(containerConfig)
		}
	}()

	// Note that the IP address should be advertised after Pod OpenFlow entries are installed, otherwise the packet might
	// be dropped by OVS.
	if err = kc.ifConfigurator.advertiseContainerAddr(containerNetNS, containerIface.Name, result); err != nil {
		klog.Errorf("Failed to advertise IP address for container %s: %v", containerID, err)
	}
	// Mark the manipulation as success to cancel deferred operations.
	success = true
	*/
	klog.Infof("Configured interfaces for container %s", containerID)

	return nil
}
