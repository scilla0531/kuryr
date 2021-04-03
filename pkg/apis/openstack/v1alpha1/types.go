package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// List the supported protocols and their codes in traceflow.
// According to code in Kuryr agent and controller, default protocol is ICMP if protocol is not inputted by users.
const (
	ICMPProtocol int32 = 1
	TCPProtocol  int32 = 6
	UDPProtocol  int32 = 17
	SCTPProtocol int32 = 132
)

var SupportedProtocols = map[string]int32{
	"TCP":  TCPProtocol,
	"UDP":  UDPProtocol,
	"ICMP": ICMPProtocol,
}

var ProtocolsToString = map[int32]string{
	TCPProtocol:  "TCP",
	UDPProtocol:  "UDP",
	ICMPProtocol: "ICMP",
	SCTPProtocol: "SCTP",
}

// List the supported destination types in traceflow.
const (
	DstTypePod     = "Pod"
	DstTypeService = "Service"
	DstTypeIPv4    = "IPv4"
)

var SupportedDestinationTypes = []string{
	DstTypePod,
	DstTypeService,
	DstTypeIPv4,
}

// List the ethernet types.
const (
	EtherTypeIPv4 uint16 = 0x0800
	EtherTypeIPv6 uint16 = 0x86DD
)

type KuryrNetworkSpec struct {
	ProjectId string `json:"projectId"`
	IsTenant bool `json:"isTenant"`
}

type KuryrNetworkStatus struct {
	PodNetId string `json:"podNet,omitempty"`
	PodSubnetId string `json:"podSubnet,omitempty"`
	PodSubnetPool string `json:"podSubnetPool,omitempty"`
	PodSubnetCIDR string `json:"podSubnetCIDR,omitempty"`
	PodSgs []string `json:"podSecurityGroups,omitempty"`
	PodRouterId string `json:"podRouter,omitempty"`

	SvcSubnetId string `json:"svcSubnet,omitempty"`
	SvcSubnetCIDR string `json:"svcSubnetCIDR,omitempty"`
}

type KuryrPortSpec struct {
	PodUid string `json:"podUid"`
	PodNodeName string `json:"podNodeName,omitempty"`
}

type KuryrPortStatus struct {
	ProjectId string `json:"projectId"`
	Vifs []KuryrVif `json:"vifs"`
}

type KuryrVif struct {
	IfName string `json:"if_name"`
	IsDefault bool `json:"default"`
	Vif VIF `json:"vif"`
}

type VIF struct {
	// Unique identifier of the VIF port.
	ID string `json:"id"`

	// Indicates whether network is currently operational. Possible values include
	// `ACTIVE', `DOWN', `BUILD', or `ERROR'. Plug-ins might define additional
	// values.
	Status string `json:"status"`

	// Mac address to use on this port.
	MACAddress string `json:"mac_address"`

	// Identifies the device (e.g., virtual server) using this port.
	DeviceID string `json:"device_id,omitempty"`

	// Specifies the IDs of any security groups associated with a port.
	SecurityGroups []string `json:"security_groups"`

	PortProfile PortProfile `json:"port_profile"`
	// The network to which the VIF is connected.
	Network Network `json:"network"`

	Qos QosPolicy `json:"qos"`

	//Name of the registered os_vif plugin.
	Plugin string `json:"plugin"`
	BridgeName string `json:"bridge_name"`
	VifName string  `json:"vif_name"`
	//VifType string `json:"binding:vif_type"`

	//TenantID string `json:"tenant_id"`

	//ProjectID string `json:"project_id"`

	// Identifies the list of IP addresses the port will recognize/accept
	//AllowedAddressPairs []AddressPair `json:"allowed_address_pairs"`

	// Tags optionally set via extensions/attributestags
	Tags []string `json:"tags"`
}

type PortProfile struct {

}

type QosPolicy struct {
	ID string `json:"id"`
	Rules []QosRule  `json:"qos_rule"`
}

type QosRule struct {
	ID string `json:"id"`
	Direction string `json:"direction"`
	MaxBurstKbps string `json:"max_burst_kbps"`
	MaxKbps string `json:"max_kbps"`
}

type Network struct {
	ID string `json:"id"`
	Bridge string `json:"bridge"`
	MTU int `json:"mtu"`
	Subnets []Subnet `json:"subnets"`
	MultiHost string `json:"multi_host,omitempty"`
}

type Subnet struct {
	//ID string `json:"id"`
	Routes  []Route  `json:"routes"`
	Ips     []IP     `json:"ips"`
	Cidr    string   `json:"cidr"`
	Gateway string   `json:"gateway"`
	DNS     []string `json:"dns"`
}

type Route struct {
	Cidr    string `json:"cidr"`
	Gateway string `json:"gateway"`
}

// IP is a sub-struct that represents an individual IP.
type IP struct {
	SubnetID  string `json:"subnet_id"`
	IPAddress string `json:"ip_address,omitempty"`
}

// AddressPair contains the IP Address and the MAC address.
type AddressPair struct {
	IPAddress  string `json:"ip_address,omitempty"`
	MACAddress string `json:"mac_address,omitempty"`
}

type SecGroupRule struct {
	// The UUID for this security group rule.
	ID string `json:"id"`

	// The direction in which the security group rule is applied. The only values
	// allowed are "ingress" or "egress". For a compute instance, an ingress
	// security group rule is applied to incoming (ingress) traffic for that
	// instance. An egress rule is applied to traffic leaving the instance.
	Direction string `json:"direction"`

	// Descripton of the rule
	Description string `json:"description"`

	// Must be IPv4 or IPv6, and addresses represented in CIDR must match the
	// ingress or egress rules.
	EtherType string `json:"ethertype"`

	// The security group ID to associate with this security group rule.
	SecGroupID string `json:"security_group_id"`

	// The minimum port number in the range that is matched by the security group
	// rule. If the protocol is TCP or UDP, this value must be less than or equal
	// to the value of the PortRangeMax attribute. If the protocol is ICMP, this
	// value must be an ICMP type.
	PortRangeMin int `json:"port_range_min"`

	// The maximum port number in the range that is matched by the security group
	// rule. The PortRangeMin attribute constrains the PortRangeMax attribute. If
	// the protocol is ICMP, this value must be an ICMP type.
	PortRangeMax int `json:"port_range_max"`

	// The protocol that is matched by the security group rule. Valid values are
	// "tcp", "udp", "icmp" or an empty string.
	Protocol string `json:"protocol"`

	// The remote group ID to be associated with this security group rule. You
	// can specify either RemoteGroupID or RemoteIPPrefix.
	RemoteGroupID string `json:"remote_group_id"`

	// The remote IP prefix to be associated with this security group rule. You
	// can specify either RemoteGroupID or RemoteIPPrefix . This attribute
	// matches the specified IP prefix as the source IP address of the IP packet.
	RemoteIPPrefix string `json:"remote_ip_prefix"`

	// TenantID is the project owner of this security group rule.
	TenantID string `json:"tenant_id"`

	// ProjectID is the project owner of this security group rule.
	ProjectID string `json:"project_id"`
}

type KuryrNetworkPolicySpec struct {
	EgressSgRules string `json:"egressSgRules"`
	IngressSgRules string `json:"ingressSgRules"`
	PodSelector string `json:"podSelector,omitempty"`
	PolicyTypes string `json:"policyTypes,omitempty"`
}

type KuryrNetworkPolicyStatus struct {
	SecurityGroupId string `json:"securityGroupId,omitempty"`
	SecurityGroupRules []SecGroup `json:"securityGroupRules,omitempty"`
}

// SecGroup represents a container for security group rules.
type SecGroup struct {
	// The UUID for the security group.
	ID string `json:"id"`

	// Human-readable name for the security group. Might not be unique.
	// Cannot be named "default" as that is automatically created for a tenant.
	Name string `json:"name"`

	// A slice of security group rules that dictate the permitted behaviour for
	// traffic entering and leaving the group.
	Rules []SecGroupRule `json:"security_group_rules"`

	// TenantID is the project owner of the security group.
	TenantID string `json:"tenant_id"`

	// ProjectID is the project owner of the security group.
	ProjectID string `json:"project_id"`

	// Tags optionally set via extensions/attributestags
	Tags []string `json:"tags"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KuryrNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KuryrNetworkSpec `json:"spec"`
	Status KuryrNetworkStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KuryrNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []KuryrNetwork `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KuryrPort struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KuryrPortSpec `json:"spec"`
	Status KuryrPortStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KuryrPortPatch describes the incremental update of an KuryrPort.
type KuryrPortPatch struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	AddedGroupMembers   KuryrNetworkSpec
	RemovedGroupMembers KuryrNetworkStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KuryrPortList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []KuryrPort `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KuryrNetworkPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KuryrNetworkSpec `json:"spec"`
	Status KuryrNetworkStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KuryrNetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []KuryrNetworkPolicy `json:"items"`
}
