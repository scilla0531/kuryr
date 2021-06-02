package v1alpha1

import (
	"context"
	"github.com/gogo/protobuf/gogoproto"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/grpc"
)

/*
改文件应该使用protobuf方式进行版本管理，临时使用手动配置
*/
type CniCmdArgs struct {
	ContainerId          string   `protobuf:"bytes,1,opt,name=container_id,json=containerId,proto3" json:"container_id,omitempty"`
	Netns                string   `protobuf:"bytes,2,opt,name=netns,proto3" json:"netns,omitempty"`
	Ifname               string   `protobuf:"bytes,3,opt,name=ifname,proto3" json:"ifname,omitempty"`
	Args                 string   `protobuf:"bytes,4,opt,name=args,proto3" json:"args,omitempty"`
	Path                 string   `protobuf:"bytes,5,opt,name=path,proto3" json:"path,omitempty"`
	NetworkConfiguration []byte   `protobuf:"bytes,6,opt,name=network_configuration,json=networkConfiguration,proto3" json:"network_configuration,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

type KuryrDaemonData struct {
	ContainerID string      `json:"CNI_CONTAINERID"`
	Netns       string      `json:"CNI_NETNS"`
	IfName      string      `json:"CNI_IFNAME"`
	Args        string      `json:"CNI_ARGS"`
	Path        string      `json:"CNI_PATH"`
	Command     string      `json:"CNI_COMMAND"`
	KuryrConf   interface{} `json:"config_kuryr"`
}

type ErrorCode int32

type Error struct {
	Code                 ErrorCode  `protobuf:"varint,1,opt,name=code,proto3,enum=github.com.vmware_tanzu.antrea.pkg.apis.cni.v1beta1.ErrorCode" json:"code,omitempty"`
	Message              string     `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Details              []*any.Any `protobuf:"bytes,3,rep,name=details,proto3" json:"details,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}


const (
	ErrorCode_UNKNOWN                       ErrorCode = 0
	ErrorCode_INCOMPATIBLE_CNI_VERSION      ErrorCode = 1
	ErrorCode_UNSUPPORTED_FIELD             ErrorCode = 2
	ErrorCode_UNKNOWN_CONTAINER             ErrorCode = 3
	ErrorCode_INVALID_ENVIRONMENT_VARIABLES ErrorCode = 4
	ErrorCode_IO_FAILURE                    ErrorCode = 5
	ErrorCode_DECODING_FAILURE              ErrorCode = 6
	ErrorCode_INVALID_NETWORK_CONFIG        ErrorCode = 7
	ErrorCode_TRY_AGAIN_LATER               ErrorCode = 11
	ErrorCode_IPAM_FAILURE                  ErrorCode = 101
	ErrorCode_CONFIG_INTERFACE_FAILURE      ErrorCode = 102
	ErrorCode_CHECK_INTERFACE_FAILURE       ErrorCode = 103
	// these errors are not used by the servers, but we declare them here to
	// make sure they are reserved.
	ErrorCode_UNKNOWN_RPC_ERROR        ErrorCode = 201
	ErrorCode_INCOMPATIBLE_API_VERSION ErrorCode = 202
)

func test(){
	_ = gogoproto.E_GoprotoEnumPrefix
}

type CniCmdRequest struct {
	CniArgs              *CniCmdArgs `protobuf:"bytes,1,opt,name=cni_args,json=cniArgs,proto3" json:"cni_args,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}
type CniCmdResponse struct {
	CniResult            []byte   `protobuf:"bytes,1,opt,name=cni_result,json=cniResult,proto3" json:"cni_result,omitempty"`
	Error                *Error   `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}
// CniServer is the server API for Cni service.
type CniServer interface {
	CmdAdd(context.Context, *CniCmdRequest) (*CniCmdResponse, error)
	CmdCheck(context.Context, *CniCmdRequest) (*CniCmdResponse, error)
	CmdDel(context.Context, *CniCmdRequest) (*CniCmdResponse, error)
}


func RegisterCniServer(s *grpc.Server, srv CniServer) {
	s.RegisterService(&_Cni_serviceDesc, srv)
}


func _Cni_CmdAdd_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CniCmdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CniServer).CmdAdd(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/github.com.vmware_tanzu.antrea.pkg.apis.cni.v1beta1.Cni/CmdAdd",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CniServer).CmdAdd(ctx, req.(*CniCmdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cni_CmdCheck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CniCmdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CniServer).CmdCheck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/github.com.vmware_tanzu.antrea.pkg.apis.cni.v1beta1.Cni/CmdCheck",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CniServer).CmdCheck(ctx, req.(*CniCmdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Cni_CmdDel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CniCmdRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CniServer).CmdDel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/github.com.vmware_tanzu.antrea.pkg.apis.cni.v1beta1.Cni/CmdDel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CniServer).CmdDel(ctx, req.(*CniCmdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Cni_serviceDesc = grpc.ServiceDesc{
	ServiceName: "github.com.vmware_tanzu.antrea.pkg.apis.cni.v1beta1.Cni",
	HandlerType: (*CniServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CmdAdd",
			Handler:    _Cni_CmdAdd_Handler,
		},
		{
			MethodName: "CmdCheck",
			Handler:    _Cni_CmdCheck_Handler,
		},
		{
			MethodName: "CmdDel",
			Handler:    _Cni_CmdDel_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/apis/cni/v1beta1/cni.proto",
}