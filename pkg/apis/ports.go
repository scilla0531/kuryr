package apis

const (

	// the default port for the kuryr-agent APIServer.
	KuryrAgentAPIPort = 10350

	defaultAgentMetricsBindAddress = ":8036"
	defaultAgentHealthzBindAddress = ":8037"

	// the default port for the kuryr-controller APIServer.
	KuryrControllerAPIPort = 10349
	DefaultServiceCIDR             = "10.96.0.0/12"
)
