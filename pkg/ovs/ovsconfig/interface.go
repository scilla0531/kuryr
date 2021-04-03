package ovsconfig

type TunnelType string

type OVSDatapathType string

const (
	GeneveTunnel = "geneve"
	VXLANTunnel  = "vxlan"
	GRETunnel    = "gre"
	STTTunnel    = "stt"

	OVSDatapathSystem OVSDatapathType = "system"
	OVSDatapathNetdev OVSDatapathType = "netdev"
)

type OVSBridgeClient interface {
	Create() Error
	Delete() Error
	GetExternalIDs() (map[string]string, Error)
	SetExternalIDs(externalIDs map[string]interface{}) Error
	SetDatapathID(datapathID string) Error
	GetInterfaceOptions(name string) (map[string]string, Error)
	SetInterfaceOptions(name string, options map[string]interface{}) Error
	CreatePort(name, ifDev string, externalIDs map[string]interface{}) (string, Error)
	CreateInternalPort(name string, ofPortRequest int32, externalIDs map[string]interface{}) (string, Error)
	CreateTunnelPort(name string, tunnelType TunnelType, ofPortRequest int32) (string, Error)
	CreateTunnelPortExt(name string, tunnelType TunnelType, ofPortRequest int32, csum bool, localIP string, remoteIP string, psk string, externalIDs map[string]interface{}) (string, Error)
	CreateUplinkPort(name string, ofPortRequest int32, externalIDs map[string]interface{}) (string, Error)
	DeletePort(portUUID string) Error
	DeletePorts(portUUIDList []string) Error
	GetOFPort(ifName string) (int32, Error)
	GetPortData(portUUID, ifName string) (*OVSPortData, Error)
	GetPortList() ([]OVSPortData, Error)
	SetInterfaceMTU(name string, MTU int) error
	GetOVSVersion() (string, Error)
	AddOVSOtherConfig(configs map[string]interface{}) Error
	GetOVSOtherConfig() (map[string]string, Error)
	DeleteOVSOtherConfig(configs map[string]interface{}) Error
	GetBridgeName() string
	IsHardwareOffloadEnabled() bool
	GetOVSDatapathType() OVSDatapathType
}