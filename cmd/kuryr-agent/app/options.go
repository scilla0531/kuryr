package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/klog"
	"projectkuryr/kuryr/pkg/agent/config"
	"projectkuryr/kuryr/pkg/ovs/ovsconfig"
	"projectkuryr/kuryr/pkg/cni"
	"projectkuryr/kuryr/pkg/version"
	"time"
)

const (
	defaultOVSBridge               = "br-int"
	defaultHostGateway             = "kuryr-gw0"
	defaultHostProcPathPrefix      = "/host"
	defaultServiceCIDR             = "10.96.0.0/12"
	defaultTunnelType              = ovsconfig.VXLANTunnel
	defaultFlowCollectorAddress    = "flow-aggregator.flow-aggregator.svc:4739:tcp"
	defaultFlowCollectorTransport  = "tcp"
	defaultFlowCollectorPort       = "4739"
	defaultFlowPollInterval        = 5 * time.Second
	defaultActiveFlowExportTimeout = 60 * time.Second
	defaultIdleFlowExportTimeout   = 15 * time.Second
	defaultNPLPortRange            = "40000-41000"
	defaultAgentBindAddress		   = ":5036"
	defaultAgentMetricsBindAddress = ":8036"
	defaultAgentHealthzBindAddress = ":8037"
)

type Options struct {
	// The path of configuration file.
	configFile string

	// The configuration object
	config         *AgentConfig
	// errCh is the channel that errors will be sent
	errCh chan error

	// master is used to override the kubeconfig's URL to the apiserver.
	//master string

	//proxyServer    proxyRun
	CleanupAndExit bool
}

func newOptions() *Options {
	return &Options{
		config: &AgentConfig{
			EnablePrometheusMetrics:   true,
			EnableTLSToFlowAggregator: true,
		},
	}
}

func (o *Options) loadConfigFromFile() error {
	data, err := ioutil.ReadFile(o.configFile)
	if err != nil {
		return err
	}

	//return yaml.UnmarshalStrict(data, &o.config)
	return yaml.Unmarshal(data, &o.config)
}

// addFlags adds flags to fs and binds them to options.
func (o *Options) addFlags(fs *pflag.FlagSet) {
	// 仅支持配置文件的方式进行参数设置，这样后续可以加入监控配置文件变化的功能
	fs.StringVar(&o.configFile, "config", o.configFile, "The path to the configuration file")
}

func (o *Options) setDefaults() {
	if o.config.OVSBridge == "" {
		o.config.OVSBridge = defaultOVSBridge
	}

	if o.config.CNISocket == "" {
		o.config.CNISocket = cni.KuryrCNISocketAddr
	}
	if o.config.OVSBridge == "" {
		o.config.OVSBridge = defaultOVSBridge
	}
	if o.config.OVSDatapathType == "" {
		o.config.OVSDatapathType = string(ovsconfig.OVSDatapathSystem)
	}
	if o.config.OVSRunDir == "" {
		o.config.OVSRunDir = ovsconfig.DefaultOVSRunDir
	}
	if o.config.HostGateway == "" {
		o.config.HostGateway = defaultHostGateway
	}
	if o.config.TrafficEncapMode == "" {
		o.config.TrafficEncapMode = config.TrafficEncapModeEncap.String()
	}
	if o.config.TunnelType == "" {
		o.config.TunnelType = defaultTunnelType
	}
	if o.config.HostProcPathPrefix == "" {
		o.config.HostProcPathPrefix = defaultHostProcPathPrefix
	}
	//if o.config.ServiceCIDR == "" {
	//	o.config.ServiceCIDR = defaultServiceCIDR
	//}
	if o.config.BindAddress == "" {
		o.config.BindAddress = defaultAgentBindAddress
	}
	if o.config.MetricsBindAddress == "" {
		o.config.MetricsBindAddress = defaultAgentMetricsBindAddress
	}
	if o.config.HealthzBindAddress == "" {
		o.config.HealthzBindAddress = defaultAgentHealthzBindAddress
	}
}

// complete completes all the required options.
func (o *Options) complete(args []string) error {
	if len(o.configFile) > 0 {
		if err := o.loadConfigFromFile(); err != nil {
			return err
		}
	}
	klog.Infof("complete > config from yaml:\n%+v\n\n", *o.config)

	o.setDefaults()
	return nil
}

// validate validates all the required options. It must be called after complete.
func (o *Options) validate(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("no positional arguments are supported")
	}

	// 检查 o.config.HealthzBindAddress 和 o.config.MetricsBindAddress 是否符合ipPort

	return nil
}

func NewAgentCommand() *cobra.Command {
	opts := newOptions()

	cmd := &cobra.Command{
		Use:  "kuryr-agent",
		Long: "The kuryr agent runs on each node.",
		Run: func(cmd *cobra.Command, args []string) {
			klog.Infoln("newAgentCommand Run")

			if err := opts.complete(args); err != nil {
				klog.Fatalf("Failed to complete: %v", err)
			}
			if err := opts.validate(args); err != nil {
				klog.Fatalf("Failed to validate: %v", err)
			}
			if err := run(opts); err != nil {
				klog.Fatalf("Error running agent: %v", err)
			}
		},
		Version: version.GetFullVersionWithRuntimeInfo(),
	}

	flags := cmd.Flags()
	opts.addFlags(flags)
	//// Install log flags
	//flags.AddGoFlagSet(flag.CommandLine)
	return cmd
}
