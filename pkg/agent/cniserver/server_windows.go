// +build windows

package cniserver

import (
	"strings"

	"github.com/containernetworking/cni/pkg/types/current"
	"k8s.io/klog"
)

const dockerInfraContainerNetNS = "none"

// updateResultDNSConfig update the DNS config from CNIConfig.
// For windows platform, if runtime dns values are there use that else use cni conf supplied dns.
// See PR: https://github.com/kubernetes/kubernetes/pull/63905
// Note: For windows node, DNS Capability is needed to be set to enable DNS config can be passed to CNI.
// See PR: https://github.com/kubernetes/kubernetes/pull/67435
func updateResultDNSConfig(result *current.Result, cniConfig *CNIConfig) {
	result.DNS = cniConfig.DNS
	if len(cniConfig.RuntimeConfig.DNS.Nameservers) > 0 {
		result.DNS.Nameservers = cniConfig.RuntimeConfig.DNS.Nameservers
	}
	if len(cniConfig.RuntimeConfig.DNS.Search) > 0 {
		result.DNS.Search = cniConfig.RuntimeConfig.DNS.Search
	}
	klog.Infof("Got runtime DNS configuration: %v", result.DNS)
}

// On windows platform netNS is not used, return it directly.
func (s *CNIServer) hostNetNsPath(netNS string) string {
	return netNS
}

// isInfraContainer returns true if a container is infra container according to the network namespace path.
// On Windows platform:
//   - When using Docker as CRI runtime, the network namespace of infra container is "none".
//   - When using Containerd as CRI runtime, the network namespace of infra container is
//     a string which does not contains ":".
func isInfraContainer(netNS string) bool {
	return netNS == dockerInfraContainerNetNS || !strings.Contains(netNS, ":")
}

// isDockerContainer returns true if a container is created by Docker with the provided network namespace.
// The network namespace format of Docker container is:
//   - Infra container: "none"
//   - Workload container: "container:$infra_container_id"
func isDockerContainer(netNS string) bool {
	return netNS == dockerInfraContainerNetNS || strings.Contains(netNS, ":")
}

func getInfraContainer(containerID, netNS string) string {
	if isInfraContainer(netNS) {
		return containerID
	}
	parts := strings.Split(netNS, ":")
	if len(parts) != 2 {
		klog.Errorf("Cannot get infra container ID, unexpected netNS: %v, fallback to containerID", netNS)
		return containerID
	}
	return strings.TrimSpace(parts[1])
}

// getInfraContainer returns the infra (sandbox) container ID of a Pod.
// On Windows, kubelet sends two kinds of CNI ADD requests for each Pod:
// 1. <container_id:"067e66fa59ade9c36552aeedac4f1420fe8efe0d2a4061ecdac45f67c5ef035c" netns:"none" ifname:"eth0"
// args:"IgnoreUnknown=1;K8S_POD_NAMESPACE=default;K8S_POD_NAME=win-webserver-6c7bdbf9fc-lswt2;K8S_POD_INFRA_CONTAINER_ID=067e66fa59ade9c36552aeedac4f1420fe8efe0d2a4061ecdac45f67c5ef035c" >
// 2. <container_id:"0cd7ab3df88aa15f6a9b7f5fc2008ef4cb5740e9ded8ede3633dca7344fd58ca" netns:"container:067e66fa59ade9c36552aeedac4f1420fe8efe0d2a4061ecdac45f67c5ef035c" ifname:"eth0"
// args:"IgnoreUnknown=1;K8S_POD_NAMESPACE=default;K8S_POD_NAME=win-webserver-6c7bdbf9fc-lswt2;K8S_POD_INFRA_CONTAINER_ID=0cd7ab3df88aa15f6a9b7f5fc2008ef4cb5740e9ded8ede3633dca7344fd58ca" >
//
// The first request uses infra container ID as "container_id", while subsequent requests use workload container ID as
// "container_id" and have infra container ID in "netns" in the form of "container:<INFRA CONTAINER ID>".
func (c *CNIConfig) getInfraContainer() string {
	return getInfraContainer(c.ContainerId, c.Netns)
}