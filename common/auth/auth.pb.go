// Code generated by protoc-gen-go. DO NOT EDIT.
// source: auth.proto

/*
Package auth is a generated protocol buffer package.

It is generated from these files:
	auth.proto

It has these top-level messages:
	Credentials
	JWT
	LogoutResponse
*/
package auth

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Credentials struct {
	Username string `protobuf:"bytes,1,opt,name=username" json:"username,omitempty"`
	Password string `protobuf:"bytes,2,opt,name=password" json:"password,omitempty"`
	// Optionally instead of a password, provide a certificate issued by the auth servers CA
	Certificate string `protobuf:"bytes,3,opt,name=certificate" json:"certificate,omitempty"`
}

func (m *Credentials) Reset()                    { *m = Credentials{} }
func (m *Credentials) String() string            { return proto.CompactTextString(m) }
func (*Credentials) ProtoMessage()               {}
func (*Credentials) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Credentials) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func (m *Credentials) GetPassword() string {
	if m != nil {
		return m.Password
	}
	return ""
}

func (m *Credentials) GetCertificate() string {
	if m != nil {
		return m.Certificate
	}
	return ""
}

// A token signed with the private key of the auth server containing user identification
type JWT struct {
	Token string `protobuf:"bytes,1,opt,name=token" json:"token,omitempty"`
}

func (m *JWT) Reset()                    { *m = JWT{} }
func (m *JWT) String() string            { return proto.CompactTextString(m) }
func (*JWT) ProtoMessage()               {}
func (*JWT) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *JWT) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

// Confirmation of a session being logged out
type LogoutResponse struct {
}

func (m *LogoutResponse) Reset()                    { *m = LogoutResponse{} }
func (m *LogoutResponse) String() string            { return proto.CompactTextString(m) }
func (*LogoutResponse) ProtoMessage()               {}
func (*LogoutResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func init() {
	proto.RegisterType((*Credentials)(nil), "auth.Credentials")
	proto.RegisterType((*JWT)(nil), "auth.JWT")
	proto.RegisterType((*LogoutResponse)(nil), "auth.LogoutResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for AuthServer service

type AuthServerClient interface {
	// Provide credentials, A stream of tokens will be returned with new tokens being issued as previous ones expire
	Authorize(ctx context.Context, in *Credentials, opts ...grpc.CallOption) (AuthServer_AuthorizeClient, error)
	// Stop issuing tokens for the given user until they re-authorize
	Logout(ctx context.Context, in *JWT, opts ...grpc.CallOption) (*LogoutResponse, error)
}

type authServerClient struct {
	cc *grpc.ClientConn
}

func NewAuthServerClient(cc *grpc.ClientConn) AuthServerClient {
	return &authServerClient{cc}
}

func (c *authServerClient) Authorize(ctx context.Context, in *Credentials, opts ...grpc.CallOption) (AuthServer_AuthorizeClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_AuthServer_serviceDesc.Streams[0], c.cc, "/auth.AuthServer/Authorize", opts...)
	if err != nil {
		return nil, err
	}
	x := &authServerAuthorizeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type AuthServer_AuthorizeClient interface {
	Recv() (*JWT, error)
	grpc.ClientStream
}

type authServerAuthorizeClient struct {
	grpc.ClientStream
}

func (x *authServerAuthorizeClient) Recv() (*JWT, error) {
	m := new(JWT)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *authServerClient) Logout(ctx context.Context, in *JWT, opts ...grpc.CallOption) (*LogoutResponse, error) {
	out := new(LogoutResponse)
	err := grpc.Invoke(ctx, "/auth.AuthServer/Logout", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for AuthServer service

type AuthServerServer interface {
	// Provide credentials, A stream of tokens will be returned with new tokens being issued as previous ones expire
	Authorize(*Credentials, AuthServer_AuthorizeServer) error
	// Stop issuing tokens for the given user until they re-authorize
	Logout(context.Context, *JWT) (*LogoutResponse, error)
}

func RegisterAuthServerServer(s *grpc.Server, srv AuthServerServer) {
	s.RegisterService(&_AuthServer_serviceDesc, srv)
}

func _AuthServer_Authorize_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Credentials)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(AuthServerServer).Authorize(m, &authServerAuthorizeServer{stream})
}

type AuthServer_AuthorizeServer interface {
	Send(*JWT) error
	grpc.ServerStream
}

type authServerAuthorizeServer struct {
	grpc.ServerStream
}

func (x *authServerAuthorizeServer) Send(m *JWT) error {
	return x.ServerStream.SendMsg(m)
}

func _AuthServer_Logout_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JWT)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServerServer).Logout(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/auth.AuthServer/Logout",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServerServer).Logout(ctx, req.(*JWT))
	}
	return interceptor(ctx, in, info, handler)
}

var _AuthServer_serviceDesc = grpc.ServiceDesc{
	ServiceName: "auth.AuthServer",
	HandlerType: (*AuthServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Logout",
			Handler:    _AuthServer_Logout_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Authorize",
			Handler:       _AuthServer_Authorize_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "auth.proto",
}

func init() { proto.RegisterFile("auth.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 213 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x90, 0x4d, 0x4b, 0xc3, 0x40,
	0x10, 0x86, 0x1b, 0xab, 0xc5, 0x4c, 0x41, 0x74, 0xe8, 0x21, 0xc4, 0x4b, 0xd9, 0x93, 0x20, 0x16,
	0xd1, 0x5f, 0x20, 0xde, 0x8a, 0xa7, 0x5a, 0xc8, 0x79, 0x4d, 0xc6, 0x64, 0x51, 0x77, 0xc2, 0xec,
	0xac, 0x82, 0xbf, 0x5e, 0x92, 0xf5, 0x23, 0xde, 0xde, 0xe7, 0x7d, 0xe0, 0xdd, 0x0f, 0x00, 0x1b,
	0xb5, 0xdb, 0xf4, 0xc2, 0xca, 0x78, 0x38, 0x64, 0xd3, 0xc2, 0xf2, 0x5e, 0xa8, 0x21, 0xaf, 0xce,
	0xbe, 0x06, 0x2c, 0xe1, 0x38, 0x06, 0x12, 0x6f, 0xdf, 0xa8, 0xc8, 0xd6, 0xd9, 0x45, 0xbe, 0xfb,
	0xe5, 0xc1, 0xf5, 0x36, 0x84, 0x0f, 0x96, 0xa6, 0x38, 0x48, 0xee, 0x87, 0x71, 0x0d, 0xcb, 0x9a,
	0x44, 0xdd, 0xb3, 0xab, 0xad, 0x52, 0x31, 0x1f, 0xf5, 0xb4, 0x32, 0xe7, 0x30, 0xdf, 0x56, 0x7b,
	0x5c, 0xc1, 0x91, 0xf2, 0x0b, 0xf9, 0xef, 0xf5, 0x04, 0xe6, 0x14, 0x4e, 0x1e, 0xb8, 0xe5, 0xa8,
	0x3b, 0x0a, 0x3d, 0xfb, 0x40, 0x37, 0x1d, 0xc0, 0x5d, 0xd4, 0xee, 0x91, 0xe4, 0x9d, 0x04, 0xaf,
	0x20, 0x1f, 0x88, 0xc5, 0x7d, 0x12, 0x9e, 0x6d, 0xc6, 0x57, 0x4c, 0xae, 0x5d, 0xe6, 0xa9, 0xda,
	0x56, 0x7b, 0x33, 0xbb, 0xce, 0xf0, 0x12, 0x16, 0x69, 0x0e, 0xff, 0x44, 0xb9, 0x4a, 0xf1, 0xff,
	0x39, 0x66, 0xf6, 0xb4, 0x18, 0xbf, 0xe3, 0xf6, 0x2b, 0x00, 0x00, 0xff, 0xff, 0x31, 0x7c, 0x9a,
	0x5e, 0x1c, 0x01, 0x00, 0x00,
}
