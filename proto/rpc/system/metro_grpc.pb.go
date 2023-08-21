// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: proto/rpc/system/metro.proto

package systemv3

import (
	context "context"
	v3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	LocationService_CreateLocation_FullMethodName = "/paralus.dev.rpc.system.v3.LocationService/CreateLocation"
	LocationService_GetLocations_FullMethodName   = "/paralus.dev.rpc.system.v3.LocationService/GetLocations"
	LocationService_GetLocation_FullMethodName    = "/paralus.dev.rpc.system.v3.LocationService/GetLocation"
	LocationService_UpdateLocation_FullMethodName = "/paralus.dev.rpc.system.v3.LocationService/UpdateLocation"
	LocationService_DeleteLocation_FullMethodName = "/paralus.dev.rpc.system.v3.LocationService/DeleteLocation"
)

// LocationServiceClient is the client API for LocationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LocationServiceClient interface {
	CreateLocation(ctx context.Context, in *v3.Location, opts ...grpc.CallOption) (*v3.Location, error)
	GetLocations(ctx context.Context, in *v3.Location, opts ...grpc.CallOption) (*v3.LocationList, error)
	GetLocation(ctx context.Context, in *v3.Location, opts ...grpc.CallOption) (*v3.Location, error)
	UpdateLocation(ctx context.Context, in *v3.Location, opts ...grpc.CallOption) (*v3.Location, error)
	DeleteLocation(ctx context.Context, in *v3.Location, opts ...grpc.CallOption) (*v3.Location, error)
}

type locationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewLocationServiceClient(cc grpc.ClientConnInterface) LocationServiceClient {
	return &locationServiceClient{cc}
}

func (c *locationServiceClient) CreateLocation(ctx context.Context, in *v3.Location, opts ...grpc.CallOption) (*v3.Location, error) {
	out := new(v3.Location)
	err := c.cc.Invoke(ctx, LocationService_CreateLocation_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *locationServiceClient) GetLocations(ctx context.Context, in *v3.Location, opts ...grpc.CallOption) (*v3.LocationList, error) {
	out := new(v3.LocationList)
	err := c.cc.Invoke(ctx, LocationService_GetLocations_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *locationServiceClient) GetLocation(ctx context.Context, in *v3.Location, opts ...grpc.CallOption) (*v3.Location, error) {
	out := new(v3.Location)
	err := c.cc.Invoke(ctx, LocationService_GetLocation_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *locationServiceClient) UpdateLocation(ctx context.Context, in *v3.Location, opts ...grpc.CallOption) (*v3.Location, error) {
	out := new(v3.Location)
	err := c.cc.Invoke(ctx, LocationService_UpdateLocation_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *locationServiceClient) DeleteLocation(ctx context.Context, in *v3.Location, opts ...grpc.CallOption) (*v3.Location, error) {
	out := new(v3.Location)
	err := c.cc.Invoke(ctx, LocationService_DeleteLocation_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LocationServiceServer is the server API for LocationService service.
// All implementations should embed UnimplementedLocationServiceServer
// for forward compatibility
type LocationServiceServer interface {
	CreateLocation(context.Context, *v3.Location) (*v3.Location, error)
	GetLocations(context.Context, *v3.Location) (*v3.LocationList, error)
	GetLocation(context.Context, *v3.Location) (*v3.Location, error)
	UpdateLocation(context.Context, *v3.Location) (*v3.Location, error)
	DeleteLocation(context.Context, *v3.Location) (*v3.Location, error)
}

// UnimplementedLocationServiceServer should be embedded to have forward compatible implementations.
type UnimplementedLocationServiceServer struct {
}

func (UnimplementedLocationServiceServer) CreateLocation(context.Context, *v3.Location) (*v3.Location, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateLocation not implemented")
}
func (UnimplementedLocationServiceServer) GetLocations(context.Context, *v3.Location) (*v3.LocationList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLocations not implemented")
}
func (UnimplementedLocationServiceServer) GetLocation(context.Context, *v3.Location) (*v3.Location, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLocation not implemented")
}
func (UnimplementedLocationServiceServer) UpdateLocation(context.Context, *v3.Location) (*v3.Location, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateLocation not implemented")
}
func (UnimplementedLocationServiceServer) DeleteLocation(context.Context, *v3.Location) (*v3.Location, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteLocation not implemented")
}

// UnsafeLocationServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LocationServiceServer will
// result in compilation errors.
type UnsafeLocationServiceServer interface {
	mustEmbedUnimplementedLocationServiceServer()
}

func RegisterLocationServiceServer(s grpc.ServiceRegistrar, srv LocationServiceServer) {
	s.RegisterService(&LocationService_ServiceDesc, srv)
}

func _LocationService_CreateLocation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v3.Location)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocationServiceServer).CreateLocation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LocationService_CreateLocation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocationServiceServer).CreateLocation(ctx, req.(*v3.Location))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocationService_GetLocations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v3.Location)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocationServiceServer).GetLocations(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LocationService_GetLocations_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocationServiceServer).GetLocations(ctx, req.(*v3.Location))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocationService_GetLocation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v3.Location)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocationServiceServer).GetLocation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LocationService_GetLocation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocationServiceServer).GetLocation(ctx, req.(*v3.Location))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocationService_UpdateLocation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v3.Location)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocationServiceServer).UpdateLocation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LocationService_UpdateLocation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocationServiceServer).UpdateLocation(ctx, req.(*v3.Location))
	}
	return interceptor(ctx, in, info, handler)
}

func _LocationService_DeleteLocation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v3.Location)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LocationServiceServer).DeleteLocation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LocationService_DeleteLocation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LocationServiceServer).DeleteLocation(ctx, req.(*v3.Location))
	}
	return interceptor(ctx, in, info, handler)
}

// LocationService_ServiceDesc is the grpc.ServiceDesc for LocationService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LocationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "paralus.dev.rpc.system.v3.LocationService",
	HandlerType: (*LocationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateLocation",
			Handler:    _LocationService_CreateLocation_Handler,
		},
		{
			MethodName: "GetLocations",
			Handler:    _LocationService_GetLocations_Handler,
		},
		{
			MethodName: "GetLocation",
			Handler:    _LocationService_GetLocation_Handler,
		},
		{
			MethodName: "UpdateLocation",
			Handler:    _LocationService_UpdateLocation_Handler,
		},
		{
			MethodName: "DeleteLocation",
			Handler:    _LocationService_DeleteLocation_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/rpc/system/metro.proto",
}
