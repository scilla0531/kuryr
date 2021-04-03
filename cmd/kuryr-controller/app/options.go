package app

import (
	"fmt"
	"github.com/gophercloud/gophercloud"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/klog"
	"net"
	"projectkuryr/kuryr/pkg/apis"
	"projectkuryr/kuryr/pkg/openstack/openstackConfig"
	"projectkuryr/kuryr/pkg/version"
)


const OpenStackResourceUnsetDefaultVal = "UNSET"
const DefaultIfName = "eth0"

const (
	AnnotationCniType = "k8s.v1.cni.cncf.io/networks"
	AnnotationCniTypeKuryr = "kube-system/kuryr" 	// k8s.v1.cni.cncf.io/networks: kube-system/kuryr)

	AnnotationPodSg = "podSg"
	AnnotationPodRouter = "podRouter"
	AnnotationPodSubnet = "podSubnet"
	AnnotationPodFixedIP = "fixedIP"
	AnnotationPodIfName = "ifName"

	AnnotationPodCIDR = "podSubnetCIDR"
	AnnotationPodNet = "podNet"
	AnnotationProject = "project"

	AnnotationSvcSubnet = "serviceSubnet"
	AnnotationSvcCIDR = "serviceCIDR"


	KURYR_FQDN = "kuryr.openstack.org"

	FinalizerPod = KURYR_FQDN + "/pod-finalizer" //不依赖kuryr，因为不动。让kuryr的资源依赖kuryr
	FinalizerSvc = KURYR_FQDN +"/service-finalizer"
	FinalizerNetworkPolicy = KURYR_FQDN +"/networkpolicy-finalizer"
	FinalizerKuryrNetwork = KURYR_FQDN +"/kuryrnetwork-finalizer"
	FinalizerKuryrPort = KURYR_FQDN +"/kuryrport-finalizer"
	FinalizerKuryrLB = KURYR_FQDN +"/kuryrloadbalancer-finalizers"

	OpenstackPortDeviceOwner = "compute:kuryr"
	k8sPodStatusPending = "Pending"
	k8sPodStatusSucceeded = "Succeeded"
	k8sPodStatusFailed = "Failed"
)


type Options struct {
	// The path of configuration file.
	configFile string
	// The configuration object
	config	*ControllerConfig
	// errCh is the channel that errors will be sent
	errCh chan error
}

func NewOptions() *Options {
	return &Options{
		config: &ControllerConfig{
			EnablePrometheusMetrics:   true,
		},
	}
}

func (o *Options) loadConfigFromFile() error {
	data, err := ioutil.ReadFile(o.configFile)
	if err != nil {
		return err
	}

	yaml.UnmarshalStrict(data, &o.config.Openstack)
	//klog.Infof("#### Openstack :%+v\n\n", o.config.Openstack.PodSgIds)

	//return yaml.UnmarshalStrict(data, &o.config)
	return yaml.Unmarshal(data, &o.config)
}

// addFlags adds flags to fs and binds them to options.
func (o *Options) addFlags(fs *pflag.FlagSet) {
	// 仅支持配置文件的方式进行参数设置，这样后续可以加入监控配置文件变化的功能
	fs.StringVar(&o.configFile, "config", o.configFile, "The path to the configuration file")
}

func (o *Options) setDefaults() {
	if o.config.ServiceCIDR == "" {
		o.config.ServiceCIDR = apis.DefaultServiceCIDR
	}

	if o.config.APIPort == 0 {
		o.config.APIPort = apis.KuryrControllerAPIPort
	}

	if o.config.Openstack.LinkIface == "" {
		o.config.Openstack.LinkIface = "eth0"
	}

	var podSubnetId string
	if o.config.Openstack.PodSubnetId != "" {
		podSubnetId = o.config.Openstack.PodSubnetId
	}else if o.config.Openstack.PodSubnetPool != "" {
		podSubnetId = o.config.Openstack.PodSubnetPool
	}

	if podSubnetId != "" {
		osClient, err := geOsClient(o.config.Openstack)
		if err != nil {
			klog.Errorf("Invoke NewOSClient Error: %v\n", err)
			return
		}

		subnet, err := osClient.GetSubnet(podSubnetId)
		if err != nil || subnet == nil{
			klog.Errorf("Get openstack subnet failed : %v\n", err)
			return
		}
		klog.Infof("Get subnet by id(%s) : NetworkID: %v, CIDR: %v\n", podSubnetId, subnet.NetworkID, subnet.CIDR)
		o.config.Openstack.PodSubnetCIDR = subnet.CIDR
		o.config.Openstack.PodNetId = subnet.NetworkID
		o.config.Openstack.ProjectId = subnet.ProjectID

		if o.config.Openstack.PodRouterId == "" {
			o.config.Openstack.PodRouterId = OpenStackResourceUnsetDefaultVal
		}
		o.config.Openstack.EnabledDefaultNetworkResources = true
	}
}

// complete completes all the required options.
func (o *Options) complete(args []string) error {
	if len(o.configFile) > 0 {
		if err := o.loadConfigFromFile(); err != nil {
			return err
		}
	}

	o.setDefaults()
	klog.Infof("complete > config from yaml:\n%+v\n\n", o.config.Openstack)

	return nil
}

func geOsClient(oscfg Openstack) (*openstackConfig.OSClient, error){
	authOpts := &gophercloud.AuthOptions{
		IdentityEndpoint: oscfg.AuthUrl,
		Username:         oscfg.UserName,
		Password:         oscfg.PassWord,
		DomainName:		  oscfg.UserDomainName,
		Scope: &gophercloud.AuthScope{
			//ProjectName: oscfg.ProjectName,
			//ProjectID: oscfg.ProjectId, // 根据 project 无法 list endpoint
			DomainName: oscfg.ProjectDomainName,
		},
	}
	//klog.Infof("%v, %v\n", authOpts, authOpts.Scope)
	return openstackConfig.NewOSClient(authOpts, oscfg.Region)
}

// validate validates all the required options. It must be called after complete.
func (o *Options) validate(args []string) (int, error) {
	if len(args) != 0 {
		return 0, fmt.Errorf("no positional arguments are supported")
	}

	_, _, err := net.ParseCIDR(o.config.ServiceCIDR)
	if err != nil {
		return 0, fmt.Errorf("Service CIDR %s is invalid", o.config.ServiceCIDR)
	}
	if o.config.ServiceCIDRv6 != "" {
		_, _, err := net.ParseCIDR(o.config.ServiceCIDRv6)
		if err != nil {
			return 0, fmt.Errorf("Service CIDR v6 %s is invalid", o.config.ServiceCIDRv6)
		}
	}

	// 检查 o.config.HealthzBindAddress 和 o.config.MetricsBindAddress 是否符合ipPort

	return 0, nil
}

func (o *Options) Run() error {
	//LearnClientGo(o.config)

	return nil
}

func NewControllerCommand() *cobra.Command {
	opts := NewOptions()

	cmd := &cobra.Command{
		Use:  "kuryr-controller",
		Long: "The Kuryr Controller.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.complete(args); err != nil {
				klog.Fatalf("Failed to complete: %v", err)
			}
			if _, err := opts.validate(args); err != nil {
				klog.Fatalf("Failed to validate: %v", err)
			}
			if err := run(opts); err != nil {
				klog.Fatalf("Error running controller: %v", err)
			}
		},
		Version: version.GetFullVersionWithRuntimeInfo(),
	}

	flags := cmd.Flags()
	opts.addFlags(flags)

	return cmd
}


