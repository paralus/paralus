// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: proto/rpc/system/oidc_provider.proto

package systemv3

import (
	context "context"
	v31 "github.com/paralus/paralus/proto/types/commonpb/v3"
	v3 "github.com/paralus/paralus/proto/types/systempb/v3"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	OIDCProviderService_CreateOIDCProvider_FullMethodName = "/paralus.dev.rpc.system.v3.OIDCProviderService/CreateOIDCProvider"
	OIDCProviderService_GetOIDCProvider_FullMethodName    = "/paralus.dev.rpc.system.v3.OIDCProviderService/GetOIDCProvider"
	OIDCProviderService_ListOIDCProvider_FullMethodName   = "/paralus.dev.rpc.system.v3.OIDCProviderService/ListOIDCProvider"
	OIDCProviderService_UpdateOIDCProvider_FullMethodName = "/paralus.dev.rpc.system.v3.OIDCProviderService/UpdateOIDCProvider"
	OIDCProviderService_DeleteOIDCProvider_FullMethodName = "/paralus.dev.rpc.system.v3.OIDCProviderService/DeleteOIDCProvider"
)

// OIDCProviderServiceClient is the client API for OIDCProviderService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type OIDCProviderServiceClient interface {
	CreateOIDCProvider(ctx context.Context, in *v3.OIDCProvider, opts ...grpc.CallOption) (*v3.OIDCProvider, error)
	GetOIDCProvider(ctx context.Context, in *v3.OIDCProvider, opts ...grpc.CallOption) (*v3.OIDCProvider, error)
	ListOIDCProvider(ctx context.Context, in *v31.Empty, opts ...grpc.CallOption) (*v3.OIDCProviderList, error)
	UpdateOIDCProvider(ctx context.Context, in *v3.OIDCProvider, opts ...grpc.CallOption) (*v3.OIDCProvider, error)
	DeleteOIDCProvider(ctx context.Context, in *v3.OIDCProvider, opts ...grpc.CallOption) (*v31.Empty, error)
}

type oIDCProviderServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewOIDCProviderServiceClient(cc grpc.ClientConnInterface) OIDCProviderServiceClient {
	return &oIDCProviderServiceClient{cc}
}

func (c *oIDCProviderServiceClient) CreateOIDCProvider(ctx context.Context, in *v3.OIDCProvider, opts ...grpc.CallOption) (*v3.OIDCProvider, error) {
	out := new(v3.OIDCProvider)
	err := c.cc.Invoke(ctx, OIDCProviderService_CreateOIDCProvider_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *oIDCProviderServiceClient) GetOIDCProvider(ctx context.Context, in *v3.OIDCProvider, opts ...grpc.CallOption) (*v3.OIDCProvider, error) {
	out := new(v3.OIDCProvider)
	err := c.cc.Invoke(ctx, OIDCProviderService_GetOIDCProvider_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *oIDCProviderServiceClient) ListOIDCProvider(ctx context.Context, in *v31.Empty, opts ...grpc.CallOption) (*v3.OIDCProviderList, error) {
	out := new(v3.OIDCProviderList)
	err := c.cc.Invoke(ctx, OIDCProviderService_ListOIDCProvider_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *oIDCProviderServiceClient) UpdateOIDCProvider(ctx context.Context, in *v3.OIDCProvider, opts ...grpc.CallOption) (*v3.OIDCProvider, error) {
	out := new(v3.OIDCProvider)
	err := c.cc.Invoke(ctx, OIDCProviderService_UpdateOIDCProvider_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *oIDCProviderServiceClient) DeleteOIDCProvider(ctx context.Context, in *v3.OIDCProvider, opts ...grpc.CallOption) (*v31.Empty, error) {
	out := new(v31.Empty)
	err := c.cc.Invoke(ctx, OIDCProviderService_DeleteOIDCProvider_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// OIDCProviderServiceServer is the server API for OIDCProviderService service.
// All implementations should embed UnimplementedOIDCProviderServiceServer
// for forward compatibility
type OIDCProviderServiceServer interface {
	CreateOIDCProvider(context.Context, *v3.OIDCProvider) (*v3.OIDCProvider, error)
	GetOIDCProvider(context.Context, *v3.OIDCProvider) (*v3.OIDCProvider, error)
	ListOIDCProvider(context.Context, *v31.Empty) (*v3.OIDCProviderList, error)
	UpdateOIDCProvider(context.Context, *v3.OIDCProvider) (*v3.OIDCProvider, error)
	DeleteOIDCProvider(context.Context, *v3.OIDCProvider) (*v31.Empty, error)
}

// UnimplementedOIDCProviderServiceServer should be embedded to have forward compatible implementations.
type UnimplementedOIDCProviderServiceServer struct {
}

func (UnimplementedOIDCProviderServiceServer) CreateOIDCProvider(context.Context, *v3.OIDCProvider) (*v3.OIDCProvider, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateOIDCProvider not implemented")
}
func (UnimplementedOIDCProviderServiceServer) GetOIDCProvider(context.Context, *v3.OIDCProvider) (*v3.OIDCProvider, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOIDCProvider not implemented")
}
func (UnimplementedOIDCProviderServiceServer) ListOIDCProvider(context.Context, *v31.Empty) (*v3.OIDCProviderList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListOIDCProvider not implemented")
}
func (UnimplementedOIDCProviderServiceServer) UpdateOIDCProvider(context.Context, *v3.OIDCProvider) (*v3.OIDCProvider, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateOIDCProvider not implemented")
}
func (UnimplementedOIDCProviderServiceServer) DeleteOIDCProvider(context.Context, *v3.OIDCProvider) (*v31.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteOIDCProvider not implemented")
}

// UnsafeOIDCProviderServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to OIDCProviderServiceServer will
// result in compilation errors.
type UnsafeOIDCProviderServiceServer interface {
	mustEmbedUnimplementedOIDCProviderServiceServer()
}

func RegisterOIDCProviderServiceServer(s grpc.ServiceRegistrar, srv OIDCProviderServiceServer) {
	s.RegisterService(&OIDCProviderService_ServiceDesc, srv)
}

func _OIDCProviderService_CreateOIDCProvider_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v3.OIDCProvider)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OIDCProviderServiceServer).CreateOIDCProvider(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OIDCProviderService_CreateOIDCProvider_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OIDCProviderServiceServer).CreateOIDCProvider(ctx, req.(*v3.OIDCProvider))
	}
	return interceptor(ctx, in, info, handler)
}

func _OIDCProviderService_GetOIDCProvider_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v3.OIDCProvider)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OIDCProviderServiceServer).GetOIDCProvider(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OIDCProviderService_GetOIDCProvider_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OIDCProviderServiceServer).GetOIDCProvider(ctx, req.(*v3.OIDCProvider))
	}
	return interceptor(ctx, in, info, handler)
}

func _OIDCProviderService_ListOIDCProvider_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v31.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OIDCProviderServiceServer).ListOIDCProvider(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OIDCProviderService_ListOIDCProvider_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OIDCProviderServiceServer).ListOIDCProvider(ctx, req.(*v31.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _OIDCProviderService_UpdateOIDCProvider_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v3.OIDCProvider)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OIDCProviderServiceServer).UpdateOIDCProvider(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OIDCProviderService_UpdateOIDCProvider_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OIDCProviderServiceServer).UpdateOIDCProvider(ctx, req.(*v3.OIDCProvider))
	}
	return interceptor(ctx, in, info, handler)
}

func _OIDCProviderService_DeleteOIDCProvider_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v3.OIDCProvider)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OIDCProviderServiceServer).DeleteOIDCProvider(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: OIDCProviderService_DeleteOIDCProvider_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(OIDCProviderServiceServer).DeleteOIDCProvider(ctx, req.(*v3.OIDCProvider))
	}
	return interceptor(ctx, in, info, handler)
}

// OIDCProviderService_ServiceDesc is the grpc.ServiceDesc for OIDCProviderService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var OIDCProviderService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "paralus.dev.rpc.system.v3.OIDCProviderService",
	HandlerType: (*OIDCProviderServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateOIDCProvider",
			Handler:    _OIDCProviderService_CreateOIDCProvider_Handler,
		},
		{
			MethodName: "GetOIDCProvider",
			Handler:    _OIDCProviderService_GetOIDCProvider_Handler,
		},
		{
			MethodName: "ListOIDCProvider",
			Handler:    _OIDCProviderService_ListOIDCProvider_Handler,
		},
		{
			MethodName: "UpdateOIDCProvider",
			Handler:    _OIDCProviderService_UpdateOIDCProvider_Handler,
		},
		{
			MethodName: "DeleteOIDCProvider",
			Handler:    _OIDCProviderService_DeleteOIDCProvider_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/rpc/system/oidc_provider.proto",
}
