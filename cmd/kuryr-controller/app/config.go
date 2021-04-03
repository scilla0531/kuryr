package app

import (
	componentbaseconfig "k8s.io/component-base/config"
)

type ControllerConfig struct {
	// FeatureGates is a map of feature names to bools that enable or disable experimental features.
	FeatureGates map[string]bool `yaml:"featureGates,omitempty"`

	ClientConnection componentbaseconfig.ClientConnectionConfiguration `yaml:"clientConnection"`
	// APIPort is the port for the kuryr-controller APIServer to serve on.
	// Defaults to 10349.
	APIPort int `yaml:"apiPort,omitempty"`
	// Enable metrics exposure via Prometheus. Initializes Prometheus metrics listener
	// Defaults to true.
	EnablePrometheusMetrics bool `yaml:"enablePrometheusMetrics,omitempty"`

	ServiceCIDR   string    `yaml:"serviceCIDR,omitempty"`
	ServiceCIDRv6 string    `yaml:"serviceCIDRv6,omitempty"`
	Openstack     Openstack `yaml:"openstack"`
}

type Openstack struct {
	AuthUrl string `yaml:"authUrl,omitempty"`
	AuthType string `yaml:"authType,omitempty"`
	UserName string `yaml:"userName,omitempty"`
	PassWord string `yaml:"passWord,omitempty"`
	UserDomainName string `yaml:"userDomainName,omitempty"`
	ProjectName string `yaml:"projectName,omitempty"`
	ProjectDomainName string `yaml:"projectDomainName,omitempty"`
	Region string `yaml:"region,omitempty"`

	EnabledDefaultNetworkResources bool `yaml:"enabledDefaultNetworkResources"`
	//kuryrv1alpha1.KuryrNetworkStatus `yaml:""`
	NetworkResources `yaml:"openstack"`
}

type NetworkResources struct {
	PodNetId string `yaml:"podNet,omitempty"` // 不需要用户配置，通过 subnetid 反查
	PodSubnetId string `yaml:"podSubnet,omitempty"`
	PodSubnetPool string `yaml:"podSubnetPool,omitempty"` // 配置了 subnet 就不需要改字段
	PodSubnetCIDR string `yaml:"podSubnetCIDR,omitempty"` // 不需要用户配置，通过subnet反查
	PodSgIds []string `yaml:"podSg,omitempty"`  // 如果没有配置，kns中sgs为空，使用默认安全组
	PodRouterId string `yaml:"podRouter,omitempty"`      //
	ProjectId string `yaml:"projectId,omitempty"`

	SvcSubnetCIDR  string `yaml:"svcSubnetCIDR,omitempty"`
	SvcSubnetId string `yaml:"svcSubnet,omitempty"`
	OvsBridge string `yaml:"ovsBridge,omitempty"`
	LinkIface string `yaml:"linkIface,omitempty"`
}

//可以选择的控制字段有三种：
// -：不要解析这个字段
// omitempty：当字段为空（默认值）时，不要解析这个字段。比如 false、0、nil、长度为 0 的 array，map，slice，string
// FieldName：当解析 json 的时候，使用这个名字