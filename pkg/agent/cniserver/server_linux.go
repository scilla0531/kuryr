package cniserver

import "github.com/containernetworking/cni/pkg/types/current"

// updateResultDNSConfig updates the DNS config from CNIConfig.
func updateResultDNSConfig(result *current.Result, cniConfig *CNIConfig) {
	result.DNS = cniConfig.DNS
}

// When running in a container, the host's /proc directory is mounted under s.hostProcPathPrefix, so
// we need to prepend s.hostProcPathPrefix to the network namespace path provided by the cni. When
// running as a simple process, s.hostProcPathPrefix will be empty.
func (s *CNIServer) hostNetNsPath(netNS string) string {
	if netNS == "" {
		return ""
	}
	return s.hostProcPathPrefix + netNS
}

// isInfraContainer returns true if a container is infra container according to the network namespace path.
// Always return true on Linux platform, because kubelet only call CNI request for infra container.
func isInfraContainer(netNS string) bool {
	return true
}

// getInfraContainer returns the sandbox container ID of a Pod.
// On Linux, it's always the ContainerID in the request.
func (c *CNIConfig) getInfraContainer() string {
	return c.ContainerId
}
