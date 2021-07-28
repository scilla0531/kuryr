package cniserver

import (
	"github.com/containernetworking/cni/pkg/types/current"
	"k8s.io/klog"
	"net"
	"os/exec"
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

func createOvsVifCmd(bridge, dev, ifaceId, mac, instanceId string) []string{
	// ovs-vsctl -- --if-exists del-port $tapname -- add-port br-int $tapname -- set Interface $tapname external-ids:iface-id=$neutron_port external-ids:iface-status=active external-ids:attached-mac=$port_mac external-ids:vm-uuid=kuryr
	cmd := []string{"--", "--if-exists", "del-port", dev, "--",
		"add-port", bridge, dev,
		"--", "set", "Interface", dev,
		"external-ids:iface-id=" + ifaceId,
		"external-ids:iface-status=active",
		"external-ids:attached-mac=" + mac,
		"external-ids:vm-uuid=" + instanceId}
	return cmd
}

func (kc *kpConfigurator) configureTap(
	portId string,
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

	hostIface := result.Interfaces[0]
	//containerIface := result.Interfaces[1]

	// Delete veth pair if any failure occurs in later manipulation.
	success := false
	defer func() {
		if !success {
			_ = kc.ifConfigurator.removeContainerLink(containerID, hostIface.Name)
		}
	}()

	cmdArgs := createOvsVifCmd("br-int", hostIfaceName, portId, mac, "kuryr")
	cmd := exec.Command("ovs-vsctl", cmdArgs...)
	err = cmd.Run()
	if err != nil {
		klog.Error(err)
	}else{
		success = true
	}

	defer func() {
		if !success {
			klog.Error("rollback ..........")
		}
	}()

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


	klog.Infof("Configured interfaces for container %s", containerID)

	return nil
}
