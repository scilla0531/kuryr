package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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

// RuleAction describes the action to be applied on traffic matching a rule.
type RuleAction string

const (
	// RuleActionAllow describes that the traffic matching the rule must be allowed.
	RuleActionAllow RuleAction = "Allow"
	// RuleActionDrop describes that the traffic matching the rule must be dropped.
	RuleActionDrop RuleAction = "Drop"
	// RuleActionReject indicates that the traffic matching the rule must be rejected and the
	// client will receive a response.
	RuleActionReject RuleAction = "Reject"
)

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
	// IP version, either `4' or `6'.
	IPVersion int `json:"ip_version"`
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


// +genclient
// +genclient:nonNamespaced
// +genclient:onlyVerbs=list,get,watch
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type AppliedToGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// GroupMembers is list of resources selected by this group.
	GroupMembers []GroupMember `json:"groupMembers,omitempty" protobuf:"bytes,2,rep,name=groupMembers"`
}

// PodReference represents a Pod Reference.
type PodReference struct {
	// The name of this Pod.
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// The Namespace of this Pod.
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,2,opt,name=namespace"`
}

// ServiceReference represents reference to a v1.Service.
type ServiceReference struct {
	// The name of this Service.
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// The Namespace of this Service.
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,2,opt,name=namespace"`
}

// NamedPort represents a Port with a name on Pod.
type NamedPort struct {
	// Port represents the Port number.
	Port int32 `json:"port,omitempty" protobuf:"varint,1,opt,name=port"`
	// Name represents the associated name with this Port number.
	Name string `json:"name,omitempty" protobuf:"bytes,2,opt,name=name"`
	// Protocol for port. Must be UDP, TCP, or SCTP.
	Protocol Protocol `json:"protocol,omitempty" protobuf:"bytes,3,opt,name=protocol"`
}

// ExternalEntityReference represents a ExternalEntity Reference.
type ExternalEntityReference struct {
	// The name of this ExternalEntity.
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	// The Namespace of this ExternalEntity.
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,2,opt,name=namespace"`
}

// GroupMember represents resource member to be populated in Groups.
type GroupMember struct {
	// Pod maintains the reference to the Pod.
	Pod *PodReference `json:"pod,omitempty" protobuf:"bytes,1,opt,name=pod"`
	// ExternalEntity maintains the reference to the ExternalEntity.
	ExternalEntity *ExternalEntityReference `json:"externalEntity,omitempty" protobuf:"bytes,2,opt,name=externalEntity"`
	// IP is the IP address of the Endpoints associated with the GroupMember.
	IPs []IPAddress `json:"ips,omitempty" protobuf:"bytes,3,rep,name=ips"`
	// Ports is the list NamedPort of the GroupMember.
	Ports []NamedPort `json:"ports,omitempty" protobuf:"bytes,4,rep,name=ports"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ClusterGroupMembers is a list of GroupMember objects that are currently selected by a ClusterGroup.
type ClusterGroupMembers struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	EffectiveMembers  []GroupMember `json:"effectiveMembers" protobuf:"bytes,2,rep,name=effectiveMembers"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// AppliedToGroupPatch describes the incremental update of an AppliedToGroup.
type AppliedToGroupPatch struct {
	metav1.TypeMeta     `json:",inline"`
	metav1.ObjectMeta   `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	AddedGroupMembers   []GroupMember `json:"addedGroupMembers,omitempty" protobuf:"bytes,2,rep,name=addedGroupMembers"`
	RemovedGroupMembers []GroupMember `json:"removedGroupMembers,omitempty" protobuf:"bytes,3,rep,name=removedGroupMembers"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// AppliedToGroupList is a list of AppliedToGroup objects.
type AppliedToGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []AppliedToGroup `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:nonNamespaced
// +genclient:onlyVerbs=list,get,watch
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// AddressGroup is the message format of antrea/pkg/controller/types.AddressGroup in an API response.
type AddressGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	GroupMembers      []GroupMember `json:"groupMembers,omitempty" protobuf:"bytes,2,rep,name=groupMembers"`
}

// IPAddress describes a single IP address. Either an IPv4 or IPv6 address must be set.
type IPAddress []byte

// IPNet describes an IP network.
type IPNet struct {
	IP           IPAddress `json:"ip,omitempty" protobuf:"bytes,1,opt,name=ip"`
	PrefixLength int32     `json:"prefixLength,omitempty" protobuf:"varint,2,opt,name=prefixLength"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// AddressGroupPatch describes the incremental update of an AddressGroup.
type AddressGroupPatch struct {
	metav1.TypeMeta     `json:",inline"`
	metav1.ObjectMeta   `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	AddedGroupMembers   []GroupMember `json:"addedGroupMembers,omitempty" protobuf:"bytes,2,rep,name=addedGroupMembers"`
	RemovedGroupMembers []GroupMember `json:"removedGroupMembers,omitempty" protobuf:"bytes,3,rep,name=removedGroupMembers"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// AddressGroupList is a list of AddressGroup objects.
type AddressGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []AddressGroup `json:"items" protobuf:"bytes,2,rep,name=items"`
}

type NetworkPolicyType string

const (
	K8sNetworkPolicy           NetworkPolicyType = "K8sNetworkPolicy"
	AntreaClusterNetworkPolicy NetworkPolicyType = "AntreaClusterNetworkPolicy"
	AntreaNetworkPolicy        NetworkPolicyType = "AntreaNetworkPolicy"
)

type NetworkPolicyReference struct {
	// Type of the NetworkPolicy.
	Type NetworkPolicyType `json:"type,omitempty" protobuf:"bytes,1,opt,name=type,casttype=NetworkPolicyType"`
	// Namespace of the NetworkPolicy. It's empty for Antrea ClusterNetworkPolicy.
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,2,opt,name=namespace"`
	// Name of the NetworkPolicy.
	Name string `json:"name,omitempty" protobuf:"bytes,3,opt,name=name"`
	// UID of the NetworkPolicy.
	UID types.UID `json:"uid,omitempty" protobuf:"bytes,4,opt,name=uid,casttype=k8s.io/apimachinery/pkg/types.UID"`
}

// +genclient
// +genclient:nonNamespaced
// +genclient:onlyVerbs=list,get,watch
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// NetworkPolicy is the message format of antrea/pkg/controller/types.NetworkPolicy in an API response.
type NetworkPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Rules is a list of rules to be applied to the selected GroupMembers.
	Rules []NetworkPolicyRule `json:"rules,omitempty" protobuf:"bytes,2,rep,name=rules"`
	// AppliedToGroups is a list of names of AppliedToGroups to which this policy applies.
	// Cannot be set in conjunction with any NetworkPolicyRule.AppliedToGroups in Rules.
	AppliedToGroups []string `json:"appliedToGroups,omitempty" protobuf:"bytes,3,rep,name=appliedToGroups"`
	// Priority represents the relative priority of this Network Policy as compared to
	// other Network Policies. Priority will be unset (nil) for K8s NetworkPolicy.
	Priority *float64 `json:"priority,omitempty" protobuf:"fixed64,4,opt,name=priority"`
	// TierPriority represents the priority of the Tier associated with this Network
	// Policy. The TierPriority will remain nil for K8s NetworkPolicy.
	TierPriority *int32 `json:"tierPriority,omitempty" protobuf:"varint,5,opt,name=tierPriority"`
	// Reference to the original NetworkPolicy that the internal NetworkPolicy is created for.
	SourceRef *NetworkPolicyReference `json:"sourceRef,omitempty" protobuf:"bytes,6,opt,name=sourceRef"`
}

// Direction defines traffic direction of NetworkPolicyRule.
type Direction string

const (
	DirectionIn  Direction = "In"
	DirectionOut Direction = "Out"
)

// NetworkPolicyRule describes a particular set of traffic that is allowed.
type NetworkPolicyRule struct {
	// The direction of this rule.
	// If it's set to In, From must be set and To must not be set.
	// If it's set to Out, To must be set and From must not be set.
	Direction Direction `json:"direction,omitempty" protobuf:"bytes,1,opt,name=direction"`
	// From represents sources which should be able to access the GroupMembers selected by the policy.
	From NetworkPolicyPeer `json:"from,omitempty" protobuf:"bytes,2,opt,name=from"`
	// To represents destinations which should be able to be accessed by the GroupMembers selected by the policy.
	To NetworkPolicyPeer `json:"to,omitempty" protobuf:"bytes,3,opt,name=to"`
	// Services is a list of services which should be matched.
	Services []Service `json:"services,omitempty" protobuf:"bytes,4,rep,name=services"`
	// Priority defines the priority of the Rule as compared to other rules in the
	// NetworkPolicy.
	Priority int32 `json:"priority,omitempty" protobuf:"varint,5,opt,name=priority"`
	// Action specifies the action to be applied on the rule. i.e. Allow/Drop. An empty
	// action “nil” defaults to Allow action, which would be the case for rules created for
	// K8s Network Policy.
	Action *RuleAction `json:"action,omitempty" protobuf:"bytes,6,opt,name=action,casttype=github.com/vmware-tanzu/antrea/pkg/apis/security/v1alpha1.RuleAction"`
	// EnableLogging indicates whether or not to generate logs when rules are matched. Default to false.
	EnableLogging bool `json:"enableLogging" protobuf:"varint,7,opt,name=enableLogging"`
	// AppliedToGroups is a list of names of AppliedToGroups to which this rule applies.
	// Cannot be set in conjunction with NetworkPolicy.AppliedToGroups of the NetworkPolicy
	// that this Rule is referred to.
	AppliedToGroups []string `json:"appliedToGroups,omitempty" protobuf:"bytes,8,opt,name=appliedToGroups"`
	// Name describes the intention of this rule.
	// Name should be unique within the policy.
	Name string `json:"name,omitempty" protobuf:"bytes,9,opt,name=name"`
}

// Protocol defines network protocols supported for things like container ports.
type Protocol string

const (
	// ProtocolTCP is the TCP protocol.
	ProtocolTCP Protocol = "TCP"
	// ProtocolUDP is the UDP protocol.
	ProtocolUDP Protocol = "UDP"
	// ProtocolSCTP is the SCTP protocol.
	ProtocolSCTP Protocol = "SCTP"
)

// Service describes a port to allow traffic on.
type Service struct {
	// The protocol (TCP, UDP, or SCTP) which traffic must match. If not specified, this
	// field defaults to TCP.
	// +optional
	Protocol *Protocol `json:"protocol,omitempty" protobuf:"bytes,1,opt,name=protocol"`
	// The port name or number on the given protocol. If not specified, this matches all port numbers.
	// +optional
	Port *intstr.IntOrString `json:"port,omitempty" protobuf:"bytes,2,opt,name=port"`
	// EndPort defines the end of the port range, being the end included within the range.
	// It can only be specified when a numerical `port` is specified.
	// +optional
	EndPort *int32 `json:"endPort,omitempty" protobuf:"bytes,3,opt,name=endPort"`
}

// NetworkPolicyPeer describes a peer of NetworkPolicyRules.
// It could be a list of names of AddressGroups and/or a list of IPBlock.
type NetworkPolicyPeer struct {
	// A list of names of AddressGroups.
	AddressGroups []string `json:"addressGroups,omitempty" protobuf:"bytes,1,rep,name=addressGroups"`
	// A list of IPBlock.
	IPBlocks []IPBlock `json:"ipBlocks,omitempty" protobuf:"bytes,2,rep,name=ipBlocks"`
}

// IPBlock describes a particular CIDR (Ex. "192.168.1.1/24"). The except entry describes CIDRs that should
// not be included within this rule.
type IPBlock struct {
	// CIDR is an IPNet represents the IP Block.
	CIDR IPNet `json:"cidr" protobuf:"bytes,1,name=cidr"`
	// Except is a slice of IPNets that should not be included within an IP Block.
	// Except values will be rejected if they are outside the CIDR range.
	// +optional
	Except []IPNet `json:"except,omitempty" protobuf:"bytes,2,rep,name=except"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// NetworkPolicyList is a list of NetworkPolicy objects.
type NetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []NetworkPolicy `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// +genclient
// +genclient:nonNamespaced
// +genclient:onlyVerbs=create
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeStatsSummary contains stats produced on a Node. It's used by the antrea-agents to report stats to the antrea-controller.
type NodeStatsSummary struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// The TrafficStats of K8s NetworkPolicies collected from the Node.
	NetworkPolicies []NetworkPolicyStats `json:"networkPolicies,omitempty" protobuf:"bytes,2,rep,name=networkPolicies"`
	// The TrafficStats of Antrea ClusterNetworkPolicies collected from the Node.
	AntreaClusterNetworkPolicies []NetworkPolicyStats `json:"antreaClusterNetworkPolicies,omitempty" protobuf:"bytes,3,rep,name=antreaClusterNetworkPolicies"`
	// The TrafficStats of Antrea NetworkPolicies collected from the Node.
	AntreaNetworkPolicies []NetworkPolicyStats `json:"antreaNetworkPolicies,omitempty" protobuf:"bytes,4,rep,name=antreaNetworkPolicies"`
}

// NetworkPolicyStats contains the information and traffic stats of a NetworkPolicy.
type NetworkPolicyStats struct {
	// The reference of the NetworkPolicy.
	NetworkPolicy NetworkPolicyReference `json:"networkPolicy,omitempty" protobuf:"bytes,1,opt,name=networkPolicy"`
	// The stats of the NetworkPolicy.
	TrafficStats TrafficStats `json:"trafficStats,omitempty" protobuf:"bytes,2,opt,name=trafficStats"`
	// The stats of the NetworkPolicy rules. It's empty for K8s NetworkPolicies as they don't have rule name to identify a rule.
	RuleTrafficStats []RuleTrafficStats `json:"ruleTrafficStats,omitempty" protobuf:"bytes,3,rep,name=ruleTrafficStats"`
}

// TrafficStats contains the traffic stats of a NetworkPolicy.
type TrafficStats struct {
	// Packets is the packets count hit by the NetworkPolicy.
	Packets int64 `json:"packets,omitempty" protobuf:"varint,1,opt,name=packets"`
	// Bytes is the bytes count hit by the NetworkPolicy.
	Bytes int64 `json:"bytes,omitempty" protobuf:"varint,2,opt,name=bytes"`
	// Sessions is the sessions count hit by the NetworkPolicy.
	Sessions int64 `json:"sessions,omitempty" protobuf:"varint,3,opt,name=sessions"`
}

// RuleTrafficStats contains TrafficStats of single rule inside a NetworkPolicy.
type RuleTrafficStats struct {
	Name         string       `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	TrafficStats TrafficStats `json:"trafficStats,omitempty" protobuf:"bytes,2,opt,name=trafficStats"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// NetworkPolicyStatus is the status of a NetworkPolicy.
type NetworkPolicyStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Nodes contains statuses produced on a list of Nodes.
	Nodes []NetworkPolicyNodeStatus `json:"nodes,omitempty" protobuf:"bytes,2,rep,name=nodes"`
}

// NetworkPolicyNodeStatus is the status of a NetworkPolicy on a Node.
type NetworkPolicyNodeStatus struct {
	// The name of the Node that produces the status.
	NodeName string `json:"nodeName,omitempty" protobuf:"bytes,1,opt,name=nodeName"`
	// The generation realized by the Node.
	Generation int64 `json:"generation,omitempty" protobuf:"varint,2,opt,name=generation"`
}

type GroupReference struct {
	// Namespace of the Group. Empty for ClusterGroup.
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,1,opt,name=namespace"`
	// Name of the Group.
	Name string `json:"name,omitempty" protobuf:"bytes,2,opt,name=name"`
	// UID of the Group.
	UID types.UID `json:"uid,omitempty" protobuf:"bytes,3,opt,name=uid,casttype=k8s.io/apimachinery/pkg/types.UID"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// GroupAssociation is the message format in an API response for groupassociation queries.
type GroupAssociation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// AssociatedGroups is a list of GroupReferences that is associated with the
	// Pod/ExternalEntity being queried.
	AssociatedGroups []GroupReference `json:"associatedGroups" protobuf:"bytes,2,rep,name=associatedGroups"`
}

