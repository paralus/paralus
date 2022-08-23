// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: proto/rpc/role/rolepermission.proto

package rolev3

import (
	context "context"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	v31 "github.com/paralus/paralus/proto/types/rolepb/v3"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// RolepermissionServiceClient is the client API for RolepermissionService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RolepermissionServiceClient interface {
	GetRolepermissions(ctx context.Context, in *v3.QueryOptions, opts ...grpc.CallOption) (*v31.RolePermissionList, error)
}

type rolepermissionServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRolepermissionServiceClient(cc grpc.ClientConnInterface) RolepermissionServiceClient {
	return &rolepermissionServiceClient{cc}
}

func (c *rolepermissionServiceClient) GetRolepermissions(ctx context.Context, in *v3.QueryOptions, opts ...grpc.CallOption) (*v31.RolePermissionList, error) {
	out := new(v31.RolePermissionList)
	err := c.cc.Invoke(ctx, "/paralus.dev.rpc.role.v3.RolepermissionService/GetRolepermissions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RolepermissionServiceServer is the server API for RolepermissionService service.
// All implementations should embed UnimplementedRolepermissionServiceServer
// for forward compatibility
type RolepermissionServiceServer interface {
	GetRolepermissions(context.Context, *v3.QueryOptions) (*v31.RolePermissionList, error)
}

// UnimplementedRolepermissionServiceServer should be embedded to have forward compatible implementations.
type UnimplementedRolepermissionServiceServer struct {
}

func (UnimplementedRolepermissionServiceServer) GetRolepermissions(context.Context, *v3.QueryOptions) (*v31.RolePermissionList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRolepermissions not implemented")
}

// UnsafeRolepermissionServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RolepermissionServiceServer will
// result in compilation errors.
type UnsafeRolepermissionServiceServer interface {
	mustEmbedUnimplementedRolepermissionServiceServer()
}

func RegisterRolepermissionServiceServer(s grpc.ServiceRegistrar, srv RolepermissionServiceServer) {
	s.RegisterService(&RolepermissionService_ServiceDesc, srv)
}

func _RolepermissionService_GetRolepermissions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v3.QueryOptions)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RolepermissionServiceServer).GetRolepermissions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paralus.dev.rpc.role.v3.RolepermissionService/GetRolepermissions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RolepermissionServiceServer).GetRolepermissions(ctx, req.(*v3.QueryOptions))
	}
	return interceptor(ctx, in, info, handler)
}

// RolepermissionService_ServiceDesc is the grpc.ServiceDesc for RolepermissionService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RolepermissionService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "paralus.dev.rpc.role.v3.RolepermissionService",
	HandlerType: (*RolepermissionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetRolepermissions",
			Handler:    _RolepermissionService_GetRolepermissions_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/rpc/role/rolepermission.proto",
}
