package app

import (
	componentbaseconfig "k8s.io/component-base/config"
)

type AgentConfig struct {
	// clientConnection specifies the kubeconfig file and client connection settings for the agent
	// to communicate with the apiserver.
	ClientConnection componentbaseconfig.ClientConnectionConfiguration `yaml:"clientConnection"`

	// healthzBindAddress is the IP address and port for the health check server to serve on,
	// defaulting to 0.0.0.0:8088
	HealthzBindAddress string `yaml:"healthzBindAddress,omitempty""`
	// metricsBindAddress is the IP address and port for the metrics server to serve on,
	// defaulting to 0.0.0.0:8089
	MetricsBindAddress string `yaml:"metricsBindAddress,omitempty""`

	// BindAddressHardFail, if true, kuryr-agent will treat failure to bind to a port as fatal and exit
	BindAddressHardFail bool

	CNISocket string `yaml:"cniSocket,omitempty"`

	// Name of the OpenVSwitch bridge kuryr-agent will create and use.
	// Make sure it doesn't conflict with your existing OpenVSwitch bridges.
	// Defaults to br-int.
	OVSBridge string `yaml:"ovsBridge,omitempty"`
	// Datapath type to use for the OpenVSwitch bridge created by kuryr. Supported values are:
	// - system
	// - netdev
	// 'system' is the default value and corresponds to the kernel datapath. Use 'netdev' to run
	// OVS in userspace mode. Userspace mode requires the tun device driver to be available.
	OVSDatapathType string `yaml:"ovsDatapathType,omitempty"`
	// Runtime data directory used by Open vSwitch.
	// Default value:
	// - On Linux platform: /var/run/openvswitch
	// - On Windows platform: C:\openvswitch\var\run\openvswitch
	OVSRunDir string `yaml:"ovsRunDir,omitempty"`
	// Name of the interface kuryr-agent will create and use for host <--> pod communication.
	// Make sure it doesn't conflict with your existing interfaces.
	// Defaults to kuryr-gw0.
	HostGateway string `yaml:"hostGateway,omitempty"`
	// Determines how traffic is encapsulated. It has the following options:
	// encap(default):    Inter-node Pod traffic is always encapsulated and Pod to external network
	//                    traffic is SNAT'd.
	// noEncap:           Inter-node Pod traffic is not encapsulated; Pod to external network traffic is
	//                    SNAT'd if noSNAT is not set to true. Underlying network must be capable of
	//                    supporting Pod traffic across IP subnets.
	// hybrid:            noEncap if source and destination Nodes are on the same subnet, otherwise encap.
	// networkPolicyOnly: kuryr enforces NetworkPolicy only, and utilizes CNI chaining and delegates Pod
	//                    IPAM and connectivity to the primary CNI.
	TrafficEncapMode string `yaml:"trafficEncapMode,omitempty"`
	// Whether or not to SNAT (using the Node IP) the egress traffic from a Pod to the external network.
	// This option is for the noEncap traffic mode only, and the default value is false. In the noEncap
	// mode, if the cluster's Pod CIDR is reachable from the external network, then the Pod traffic to
	// the external network needs not be SNAT'd. In the networkPolicyOnly mode, kuryr-agent never
	// performs SNAT and this option will be ignored; for other modes it must be set to false.
	NoSNAT bool `yaml:"noSNAT,omitempty"`
	// Tunnel protocols used for encapsulating traffic across Nodes. Supported values:
	// - geneve (default)
	// - vxlan
	// - gre
	// - stt
	TunnelType string `yaml:"tunnelType,omitempty"`
	// Default MTU to use for the host gateway interface and the network interface of each Pod.
	// If omitted, kuryr-agent will discover the MTU of the Node's primary interface and
	// also adjust MTU to accommodate for tunnel encapsulation overhead (if applicable).
	DefaultMTU int `yaml:"defaultMTU,omitempty"`
	// Mount location of the /proc directory. The default is "/host", which is appropriate when
	// kuryr-agent is run as part of the kuryr DaemonSet (and the host's /proc directory is mounted
	// as /host/proc in the kuryr-agent container). When running kuryr-agent as a process,
	// hostProcPathPrefix should be set to "/" in the YAML config.
	HostProcPathPrefix string `yaml:"hostProcPathPrefix,omitempty"`
	// ClusterIP CIDR range for Services. It's required when kuryrProxy is not enabled, and should be
	// set to the same value as the one specified by --service-cluster-ip-range for kube-apiserver. When
	// kuryrProxy is enabled, this parameter is not needed and will be ignored if provided.
	// Default is 10.96.0.0/12
	ServiceCIDR string `yaml:"serviceCIDR,omitempty"`
	// ClusterIP CIDR range for IPv6 Services. It's required when using kuryr-agent to provide IPv6 Service in a Dual-Stack
	// cluster or an IPv6 only cluster. The value should be the same as the configuration for kube-apiserver specified by
	// --service-cluster-ip-range. When kuryrProxy is enabled, this parameter is not needed.
	// No default value for this field.
	ServiceCIDRv6 string `yaml:"serviceCIDRv6,omitempty"`
	// Whether or not to enable IPSec (ESP) encryption for Pod traffic across Nodes. IPSec encryption
	// is supported only for the GRE tunnel type. kuryr uses Preshared Key (PSK) for IKE
	// authentication. When IPSec tunnel is enabled, the PSK value must be passed to kuryr Agent
	// through an environment variable: kuryr_IPSEC_PSK.
	// Defaults to false.
	EnableIPSecTunnel bool `yaml:"enableIPSecTunnel,omitempty"`

	// Enable metrics exposure via Prometheus. Initializes Prometheus metrics listener
	// Defaults to true.
	EnablePrometheusMetrics bool `yaml:"enablePrometheusMetrics,omitempty"`
	// Provide the IPFIX collector address as a string with format <HOST>:[<PORT>][:<PROTO>].
	// HOST can either be the DNS name or the IP of the Flow Collector. For example,
	// "flow-aggregator.flow-aggregator.svc" can be provided as DNS name to connect
	// to the kuryr Flow Aggregator service. If IP, it can be either IPv4 or IPv6.
	// However, IPv6 address should be wrapped with [].
	// If PORT is empty, we default to 4739, the standard IPFIX port.
	// If no PROTO is given, we consider "tcp" as default. We support "tcp" and
	// "udp" L4 transport protocols.
	// Defaults to "flow-aggregator.flow-aggregator.svc:4739:tcp".
	FlowCollectorAddr string `yaml:"flowCollectorAddr,omitempty"`
	// Provide flow poll interval in format "0s". This determines how often flow
	// exporter dumps connections in conntrack module. Flow poll interval should
	// be greater than or equal to 1s(one second).
	// Defaults to "5s". Follow the time units of duration type.
	FlowPollInterval string `yaml:"flowPollInterval,omitempty"`
	// Provide the active flow export timeout, which is the timeout after which
	// a flow record is sent to the collector for active flows. Thus, for flows
	// with a continuous stream of packets, a flow record will be exported to the
	// collector once the elapsed time since the last export event is equal to the
	// value of this timeout.
	// Defaults to "60s". Follow the time units of duration type.
	ActiveFlowExportTimeout string `yaml:"activeFlowExportTimeout,omitempty"`
	// Provide the idle flow export timeout, which is the timeout after which a
	// flow record is sent to the collector for idle flows. A flow is considered
	// idle if no packet matching this flow has been observed since the last export
	// event.
	// Defaults to "15s". Follow the time units of duration type.
	IdleFlowExportTimeout string `yaml:"idleFlowExportTimeout,omitempty"`
	// Enable TLS communication from flow exporter to flow aggregator.
	// Defaults to true.
	EnableTLSToFlowAggregator bool `yaml:"enableTLSToFlowAggregator,omitempty"`
	// Provide the port range used by NodePortLocal. When the NodePortLocal feature is enabled, a port from that range will be assigned
	// whenever a Pod's container defines a specific port to be exposed (each container can define a list of ports as pod.spec.containers[].ports),
	// and all Node traffic directed to that port will be forwarded to the Pod.
	NPLPortRange string `yaml:"nplPortRange,omitempty"`
	// Provide the address of Kubernetes apiserver, to override any value provided in kubeconfig or InClusterConfig.
	// Defaults to "". It must be a host string, a host:port pair, or a URL to the base of the apiserver.
	KubeAPIServerOverride string `yaml:"kubeAPIServerOverride,omitempty"`
	// Cipher suites to use.
	TLSCipherSuites string `yaml:"tlsCipherSuites,omitempty"`
	// TLS min version.
	TLSMinVersion string `yaml:"tlsMinVersion,omitempty"`
}
