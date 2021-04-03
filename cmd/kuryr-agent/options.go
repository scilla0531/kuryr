package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/klog"
	"net"
	"projectkuryr/kuryr/pkg/agent/config"
	"projectkuryr/kuryr/pkg/ovs/ovsconfig"
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
	proxyServer    proxyRun
	CleanupAndExit bool
	// errCh is the channel that errors will be sent
	errCh chan error
	// master is used to override the kubeconfig's URL to the apiserver.
	master string
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

	//err = yaml.Unmarshal(data, &o.config)
	//fmt.Printf("######## get struct from []byte: \n%+v\n########\n\n", *o.config)
	//fmt.Printf("######## get ClientConnection from []byte: \n%+v\n########\n\n", o.config.ClientConnection)
	//return err

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
		o.config.CNISocket = KuryrCNISocketAddr
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
	if o.config.ServiceCIDR == "" {
		o.config.ServiceCIDR = defaultServiceCIDR
	}
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

	_, _, err := net.ParseCIDR(o.config.ServiceCIDR)
	if err != nil {
		return fmt.Errorf("Service CIDR %s is invalid", o.config.ServiceCIDR)
	}
	if o.config.ServiceCIDRv6 != "" {
		_, _, err := net.ParseCIDR(o.config.ServiceCIDRv6)
		if err != nil {
			return fmt.Errorf("Service CIDR v6 %s is invalid", o.config.ServiceCIDRv6)
		}
	}

	// 检查 o.config.HealthzBindAddress 和 o.config.MetricsBindAddress 是否符合ipPort

	return nil
}

// runLoop will watch on the update change of the proxy server's configuration file.
// Return an error when updated
func (o *Options) runLoop() error {
	// run the proxy in goroutine
	go func() {
		err := o.proxyServer.Run()
		o.errCh <- err
	}()

	for {
		err := <-o.errCh
		if err != nil {
			return err
		}
	}
}

//Run runs the specified ProxyServer.
func (o *Options) Run() error {
	defer close(o.errCh)

	proxyServer, err := NewProxyServer(o)
	if err != nil {
		return err
	}

	if o.CleanupAndExit {
		klog.Infoln()
		return proxyServer.CleanupAndExit()
	}

	o.proxyServer = proxyServer
	return o.runLoop()
}
