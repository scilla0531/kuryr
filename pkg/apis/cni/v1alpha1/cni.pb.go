// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pkg/apis/cni/v1alpha1/cni.proto

//package kuryrcni;

package v1alpha1

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	any "github.com/golang/protobuf/ptypes/any"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ErrorCode int32

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

var ErrorCode_name = map[int32]string{
	0:   "UNKNOWN",
	1:   "INCOMPATIBLE_CNI_VERSION",
	2:   "UNSUPPORTED_FIELD",
	3:   "UNKNOWN_CONTAINER",
	4:   "INVALID_ENVIRONMENT_VARIABLES",
	5:   "IO_FAILURE",
	6:   "DECODING_FAILURE",
	7:   "INVALID_NETWORK_CONFIG",
	11:  "TRY_AGAIN_LATER",
	101: "IPAM_FAILURE",
	102: "CONFIG_INTERFACE_FAILURE",
	103: "CHECK_INTERFACE_FAILURE",
	201: "UNKNOWN_RPC_ERROR",
	202: "INCOMPATIBLE_API_VERSION",
}

var ErrorCode_value = map[string]int32{
	"UNKNOWN":                       0,
	"INCOMPATIBLE_CNI_VERSION":      1,
	"UNSUPPORTED_FIELD":             2,
	"UNKNOWN_CONTAINER":             3,
	"INVALID_ENVIRONMENT_VARIABLES": 4,
	"IO_FAILURE":                    5,
	"DECODING_FAILURE":              6,
	"INVALID_NETWORK_CONFIG":        7,
	"TRY_AGAIN_LATER":               11,
	"IPAM_FAILURE":                  101,
	"CONFIG_INTERFACE_FAILURE":      102,
	"CHECK_INTERFACE_FAILURE":       103,
	"UNKNOWN_RPC_ERROR":             201,
	"INCOMPATIBLE_API_VERSION":      202,
}

func (x ErrorCode) String() string {
	return proto.EnumName(ErrorCode_name, int32(x))
}

func (ErrorCode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_e9dc6d0da63f0ae4, []int{0}
}

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

func (m *CniCmdArgs) Reset()         { *m = CniCmdArgs{} }
func (m *CniCmdArgs) String() string { return proto.CompactTextString(m) }
func (*CniCmdArgs) ProtoMessage()    {}
func (*CniCmdArgs) Descriptor() ([]byte, []int) {
	return fileDescriptor_e9dc6d0da63f0ae4, []int{0}
}

func (m *CniCmdArgs) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CniCmdArgs.Unmarshal(m, b)
}
func (m *CniCmdArgs) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CniCmdArgs.Marshal(b, m, deterministic)
}
func (m *CniCmdArgs) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CniCmdArgs.Merge(m, src)
}
func (m *CniCmdArgs) XXX_Size() int {
	return xxx_messageInfo_CniCmdArgs.Size(m)
}
func (m *CniCmdArgs) XXX_DiscardUnknown() {
	xxx_messageInfo_CniCmdArgs.DiscardUnknown(m)
}

var xxx_messageInfo_CniCmdArgs proto.InternalMessageInfo

func (m *CniCmdArgs) GetContainerId() string {
	if m != nil {
		return m.ContainerId
	}
	return ""
}

func (m *CniCmdArgs) GetNetns() string {
	if m != nil {
		return m.Netns
	}
	return ""
}

func (m *CniCmdArgs) GetIfname() string {
	if m != nil {
		return m.Ifname
	}
	return ""
}

func (m *CniCmdArgs) GetArgs() string {
	if m != nil {
		return m.Args
	}
	return ""
}

func (m *CniCmdArgs) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *CniCmdArgs) GetNetworkConfiguration() []byte {
	if m != nil {
		return m.NetworkConfiguration
	}
	return nil
}

type CniCmdRequest struct {
	CniArgs              *CniCmdArgs `protobuf:"bytes,1,opt,name=cni_args,json=cniArgs,proto3" json:"cni_args,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *CniCmdRequest) Reset()         { *m = CniCmdRequest{} }
func (m *CniCmdRequest) String() string { return proto.CompactTextString(m) }
func (*CniCmdRequest) ProtoMessage()    {}
func (*CniCmdRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_e9dc6d0da63f0ae4, []int{1}
}

func (m *CniCmdRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CniCmdRequest.Unmarshal(m, b)
}
func (m *CniCmdRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CniCmdRequest.Marshal(b, m, deterministic)
}
func (m *CniCmdRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CniCmdRequest.Merge(m, src)
}
func (m *CniCmdRequest) XXX_Size() int {
	return xxx_messageInfo_CniCmdRequest.Size(m)
}
func (m *CniCmdRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CniCmdRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CniCmdRequest proto.InternalMessageInfo

func (m *CniCmdRequest) GetCniArgs() *CniCmdArgs {
	if m != nil {
		return m.CniArgs
	}
	return nil
}

type Error struct {
	Code                 ErrorCode  `protobuf:"varint,1,opt,name=code,proto3,enum=kuryr.pkg.apis.cni.v1alpha1.ErrorCode" json:"code,omitempty"`
	Message              string     `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Details              []*any.Any `protobuf:"bytes,3,rep,name=details,proto3" json:"details,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *Error) Reset()         { *m = Error{} }
func (m *Error) String() string { return proto.CompactTextString(m) }
func (*Error) ProtoMessage()    {}
func (*Error) Descriptor() ([]byte, []int) {
	return fileDescriptor_e9dc6d0da63f0ae4, []int{2}
}

func (m *Error) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Error.Unmarshal(m, b)
}
func (m *Error) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Error.Marshal(b, m, deterministic)
}
func (m *Error) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Error.Merge(m, src)
}
func (m *Error) XXX_Size() int {
	return xxx_messageInfo_Error.Size(m)
}
func (m *Error) XXX_DiscardUnknown() {
	xxx_messageInfo_Error.DiscardUnknown(m)
}

var xxx_messageInfo_Error proto.InternalMessageInfo

func (m *Error) GetCode() ErrorCode {
	if m != nil {
		return m.Code
	}
	return ErrorCode_UNKNOWN
}

func (m *Error) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func (m *Error) GetDetails() []*any.Any {
	if m != nil {
		return m.Details
	}
	return nil
}

type CniCmdResponse struct {
	CniResult            []byte   `protobuf:"bytes,1,opt,name=cni_result,json=cniResult,proto3" json:"cni_result,omitempty"`
	Error                *Error   `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CniCmdResponse) Reset()         { *m = CniCmdResponse{} }
func (m *CniCmdResponse) String() string { return proto.CompactTextString(m) }
func (*CniCmdResponse) ProtoMessage()    {}
func (*CniCmdResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_e9dc6d0da63f0ae4, []int{3}
}

func (m *CniCmdResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CniCmdResponse.Unmarshal(m, b)
}
func (m *CniCmdResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CniCmdResponse.Marshal(b, m, deterministic)
}
func (m *CniCmdResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CniCmdResponse.Merge(m, src)
}
func (m *CniCmdResponse) XXX_Size() int {
	return xxx_messageInfo_CniCmdResponse.Size(m)
}
func (m *CniCmdResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_CniCmdResponse.DiscardUnknown(m)
}

var xxx_messageInfo_CniCmdResponse proto.InternalMessageInfo

func (m *CniCmdResponse) GetCniResult() []byte {
	if m != nil {
		return m.CniResult
	}
	return nil
}

func (m *CniCmdResponse) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}

func init() {
	proto.RegisterEnum("kuryr.pkg.apis.cni.v1alpha1.ErrorCode", ErrorCode_name, ErrorCode_value)
	proto.RegisterType((*CniCmdArgs)(nil), "kuryr.pkg.apis.cni.v1alpha1.CniCmdArgs")
	proto.RegisterType((*CniCmdRequest)(nil), "kuryr.pkg.apis.cni.v1alpha1.CniCmdRequest")
	proto.RegisterType((*Error)(nil), "kuryr.pkg.apis.cni.v1alpha1.Error")
	proto.RegisterType((*CniCmdResponse)(nil), "kuryr.pkg.apis.cni.v1alpha1.CniCmdResponse")
}

func init() { proto.RegisterFile("pkg/apis/cni/v1alpha1/cni.proto", fileDescriptor_e9dc6d0da63f0ae4) }

var fileDescriptor_e9dc6d0da63f0ae4 = []byte{
	// 671 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x53, 0xdd, 0x4e, 0xdb, 0x48,
	0x14, 0xc6, 0xf9, 0x85, 0x93, 0x2c, 0xeb, 0x9d, 0x0d, 0xac, 0x17, 0x16, 0x2d, 0xe4, 0x62, 0x17,
	0xb1, 0x92, 0x23, 0xc2, 0xcd, 0xaa, 0x77, 0xce, 0x64, 0x42, 0x47, 0x84, 0x71, 0x34, 0x71, 0x82,
	0xda, 0x1b, 0xcb, 0xd8, 0x13, 0x63, 0x25, 0x19, 0xa7, 0xb6, 0xd3, 0x8a, 0x87, 0xe8, 0xeb, 0xf4,
	0xa2, 0x6f, 0xd0, 0x3e, 0x46, 0x9f, 0xa4, 0xb2, 0x9d, 0x84, 0xaa, 0xaa, 0x28, 0x37, 0xdc, 0x9d,
	0xf3, 0x7d, 0xe7, 0xef, 0x3b, 0x73, 0x06, 0xfe, 0x5e, 0x4c, 0xfd, 0x96, 0xb3, 0x08, 0xe2, 0x96,
	0x2b, 0x83, 0xd6, 0xdb, 0x73, 0x67, 0xb6, 0xb8, 0x73, 0xce, 0x53, 0x47, 0x5f, 0x44, 0x61, 0x12,
	0xa2, 0xc3, 0xe9, 0x32, 0xba, 0x8f, 0xf4, 0xc5, 0xd4, 0xd7, 0xd3, 0x30, 0x3d, 0x65, 0xd6, 0x61,
	0x07, 0x7f, 0xfa, 0x61, 0xe8, 0xcf, 0x44, 0x2b, 0x0b, 0xbd, 0x5d, 0x4e, 0x5a, 0x8e, 0xbc, 0xcf,
	0xf3, 0x9a, 0x1f, 0x15, 0x00, 0x2c, 0x03, 0x3c, 0xf7, 0x8c, 0xc8, 0x8f, 0xd1, 0x09, 0xd4, 0xdd,
	0x50, 0x26, 0x4e, 0x20, 0x45, 0x64, 0x07, 0x9e, 0xa6, 0x1c, 0x2b, 0xa7, 0x3b, 0xbc, 0xb6, 0xc1,
	0xa8, 0x87, 0x1a, 0x50, 0x96, 0x22, 0x91, 0xb1, 0x56, 0xc8, 0xb8, 0xdc, 0x41, 0xfb, 0x50, 0x09,
	0x26, 0xd2, 0x99, 0x0b, 0xad, 0x98, 0xc1, 0x2b, 0x0f, 0x21, 0x28, 0x39, 0x91, 0x1f, 0x6b, 0xa5,
	0x0c, 0xcd, 0xec, 0x14, 0x5b, 0x38, 0xc9, 0x9d, 0x56, 0xce, 0xb1, 0xd4, 0x46, 0x17, 0xb0, 0x27,
	0x45, 0xf2, 0x2e, 0x8c, 0xa6, 0xb6, 0x1b, 0xca, 0x49, 0xe0, 0x2f, 0x23, 0x27, 0x09, 0x42, 0xa9,
	0x55, 0x8e, 0x95, 0xd3, 0x3a, 0x6f, 0xac, 0x48, 0xfc, 0x2d, 0xd7, 0x1c, 0xc2, 0x2f, 0xf9, 0xec,
	0x5c, 0xbc, 0x59, 0x8a, 0x38, 0x41, 0x1d, 0xd8, 0x76, 0x65, 0x60, 0x67, 0x1d, 0xd3, 0xd1, 0x6b,
	0xed, 0x7f, 0xf5, 0x47, 0x16, 0xa3, 0x3f, 0x28, 0xe7, 0x55, 0x57, 0x06, 0xa9, 0xd1, 0x7c, 0xaf,
	0x40, 0x99, 0x44, 0x51, 0x18, 0xa1, 0x17, 0x50, 0x72, 0x43, 0x4f, 0x64, 0x95, 0x76, 0xdb, 0xff,
	0x3c, 0x5a, 0x29, 0xcb, 0xc0, 0xa1, 0x27, 0x78, 0x96, 0x83, 0x34, 0xa8, 0xce, 0x45, 0x1c, 0x3b,
	0xbe, 0x58, 0xed, 0x69, 0xed, 0x22, 0x1d, 0xaa, 0x9e, 0x48, 0x9c, 0x60, 0x16, 0x6b, 0xc5, 0xe3,
	0xe2, 0x69, 0xad, 0xdd, 0xd0, 0xf3, 0xe7, 0xd1, 0xd7, 0xcf, 0xa3, 0x1b, 0xf2, 0x9e, 0xaf, 0x83,
	0x9a, 0x01, 0xec, 0xae, 0x45, 0xc6, 0x8b, 0x50, 0xc6, 0x02, 0x1d, 0x01, 0xa4, 0x2a, 0x23, 0x11,
	0x2f, 0x67, 0x49, 0x36, 0x5d, 0x9d, 0xef, 0xb8, 0x32, 0xe0, 0x19, 0x80, 0xfe, 0x87, 0xb2, 0x48,
	0xa7, 0xc9, 0x1a, 0xd7, 0xda, 0xcd, 0x9f, 0xcf, 0xcd, 0xf3, 0x84, 0xb3, 0x2f, 0x05, 0xd8, 0xd9,
	0x08, 0x41, 0x35, 0xa8, 0x8e, 0xd8, 0x15, 0x33, 0x6f, 0x98, 0xba, 0x85, 0xfe, 0x02, 0x8d, 0x32,
	0x6c, 0x5e, 0x0f, 0x0c, 0x8b, 0x76, 0xfa, 0xc4, 0xc6, 0x8c, 0xda, 0x63, 0xc2, 0x87, 0xd4, 0x64,
	0xaa, 0x82, 0xf6, 0xe0, 0xb7, 0x11, 0x1b, 0x8e, 0x06, 0x03, 0x93, 0x5b, 0xa4, 0x6b, 0xf7, 0x28,
	0xe9, 0x77, 0xd5, 0x42, 0x0e, 0x67, 0x15, 0x6c, 0x6c, 0x32, 0xcb, 0xa0, 0x8c, 0x70, 0xb5, 0x88,
	0x4e, 0xe0, 0x88, 0xb2, 0xb1, 0xd1, 0xa7, 0x5d, 0x9b, 0xb0, 0x31, 0xe5, 0x26, 0xbb, 0x26, 0xcc,
	0xb2, 0xc7, 0x06, 0xa7, 0x46, 0xa7, 0x4f, 0x86, 0x6a, 0x09, 0xed, 0x02, 0x50, 0xd3, 0xee, 0x19,
	0xb4, 0x3f, 0xe2, 0x44, 0x2d, 0xa3, 0x06, 0xa8, 0x5d, 0x82, 0xcd, 0x2e, 0x65, 0x97, 0x1b, 0xb4,
	0x82, 0x0e, 0x60, 0x7f, 0x5d, 0x88, 0x11, 0xeb, 0xc6, 0xe4, 0x57, 0x69, 0x9f, 0x1e, 0xbd, 0x54,
	0xab, 0xe8, 0x77, 0xf8, 0xd5, 0xe2, 0xaf, 0x6c, 0xe3, 0xd2, 0xa0, 0xcc, 0xee, 0x1b, 0x16, 0xe1,
	0x6a, 0x0d, 0xa9, 0x50, 0xa7, 0x03, 0xe3, 0x7a, 0x53, 0x42, 0xa4, 0xba, 0xf2, 0x14, 0x9b, 0x32,
	0x8b, 0xf0, 0x9e, 0x81, 0xc9, 0x86, 0x9d, 0xa0, 0x43, 0xf8, 0x03, 0xbf, 0x24, 0xf8, 0xea, 0x07,
	0xa4, 0x8f, 0xf6, 0x1f, 0xd4, 0xf1, 0x01, 0xb6, 0x09, 0xe7, 0x26, 0x57, 0x3f, 0x29, 0xe8, 0xe8,
	0xbb, 0x55, 0x19, 0x83, 0x87, 0x55, 0x7d, 0x56, 0xda, 0x1f, 0x0a, 0x50, 0xc4, 0x32, 0x40, 0x2e,
	0x54, 0xd2, 0xdb, 0xf3, 0x3c, 0x74, 0xf6, 0x84, 0x1b, 0x5d, 0x5d, 0xf8, 0xc1, 0x7f, 0x4f, 0x8a,
	0xcd, 0x0f, 0xa5, 0xb9, 0x85, 0x04, 0x6c, 0xe3, 0xb9, 0x87, 0xef, 0x84, 0x3b, 0x7d, 0xce, 0x36,
	0xb9, 0x96, 0xae, 0x98, 0x3d, 0x63, 0x93, 0x0e, 0xbc, 0xde, 0x5e, 0x93, 0xb7, 0x95, 0xec, 0xaf,
	0x5c, 0x7c, 0x0d, 0x00, 0x00, 0xff, 0xff, 0xfd, 0x11, 0xeb, 0x67, 0x18, 0x05, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// CniClient is the client API for Cni service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type CniClient interface {
	CmdAdd(ctx context.Context, in *CniCmdRequest, opts ...grpc.CallOption) (*CniCmdResponse, error)
	CmdCheck(ctx context.Context, in *CniCmdRequest, opts ...grpc.CallOption) (*CniCmdResponse, error)
	CmdDel(ctx context.Context, in *CniCmdRequest, opts ...grpc.CallOption) (*CniCmdResponse, error)
}

type cniClient struct {
	cc *grpc.ClientConn
}

func NewCniClient(cc *grpc.ClientConn) CniClient {
	return &cniClient{cc}
}

func (c *cniClient) CmdAdd(ctx context.Context, in *CniCmdRequest, opts ...grpc.CallOption) (*CniCmdResponse, error) {
	out := new(CniCmdResponse)
	err := c.cc.Invoke(ctx, "/kuryr.pkg.apis.cni.v1alpha1.Cni/CmdAdd", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cniClient) CmdCheck(ctx context.Context, in *CniCmdRequest, opts ...grpc.CallOption) (*CniCmdResponse, error) {
	out := new(CniCmdResponse)
	err := c.cc.Invoke(ctx, "/kuryr.pkg.apis.cni.v1alpha1.Cni/CmdCheck", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cniClient) CmdDel(ctx context.Context, in *CniCmdRequest, opts ...grpc.CallOption) (*CniCmdResponse, error) {
	out := new(CniCmdResponse)
	err := c.cc.Invoke(ctx, "/kuryr.pkg.apis.cni.v1alpha1.Cni/CmdDel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CniServer is the server API for Cni service.
type CniServer interface {
	CmdAdd(context.Context, *CniCmdRequest) (*CniCmdResponse, error)
	CmdCheck(context.Context, *CniCmdRequest) (*CniCmdResponse, error)
	CmdDel(context.Context, *CniCmdRequest) (*CniCmdResponse, error)
}

// UnimplementedCniServer can be embedded to have forward compatible implementations.
type UnimplementedCniServer struct {
}

func (*UnimplementedCniServer) CmdAdd(ctx context.Context, req *CniCmdRequest) (*CniCmdResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CmdAdd not implemented")
}
func (*UnimplementedCniServer) CmdCheck(ctx context.Context, req *CniCmdRequest) (*CniCmdResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CmdCheck not implemented")
}
func (*UnimplementedCniServer) CmdDel(ctx context.Context, req *CniCmdRequest) (*CniCmdResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CmdDel not implemented")
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
		FullMethod: "/kuryr.pkg.apis.cni.v1alpha1.Cni/CmdAdd",
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
		FullMethod: "/kuryr.pkg.apis.cni.v1alpha1.Cni/CmdCheck",
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
		FullMethod: "/kuryr.pkg.apis.cni.v1alpha1.Cni/CmdDel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CniServer).CmdDel(ctx, req.(*CniCmdRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Cni_serviceDesc = grpc.ServiceDesc{
	ServiceName: "kuryr.pkg.apis.cni.v1alpha1.Cni",
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
	Metadata: "pkg/apis/cni/v1alpha1/cni.proto",
}
