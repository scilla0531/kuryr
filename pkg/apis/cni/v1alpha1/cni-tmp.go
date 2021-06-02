package v1alpha1

import "google.golang.org/grpc"

type KuryrDaemonData struct {
	ContainerID string      `json:"CNI_CONTAINERID"`
	Netns       string      `json:"CNI_NETNS"`
	IfName      string      `json:"CNI_IFNAME"`
	Args        string      `json:"CNI_ARGS"`
	Path        string      `json:"CNI_PATH"`
	Command     string      `json:"CNI_COMMAND"`
	KuryrConf   interface{} `json:"config_kuryr"`
}


func RegisterCniServer(s *grpc.Server, srv CniServer) {
	s.RegisterService(&_Cni_serviceDesc, srv)
}