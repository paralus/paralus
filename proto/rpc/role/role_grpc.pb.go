// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: proto/rpc/role/role.proto

package rpcv3

import (
	context "context"
	v3 "github.com/paralus/paralus/proto/types/rolepb/v3"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// RoleClient is the client API for Role service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RoleClient interface {
	CreateRole(ctx context.Context, in *v3.Role, opts ...grpc.CallOption) (*v3.Role, error)
	GetRoles(ctx context.Context, in *v3.Role, opts ...grpc.CallOption) (*v3.RoleList, error)
	GetRole(ctx context.Context, in *v3.Role, opts ...grpc.CallOption) (*v3.Role, error)
	UpdateRole(ctx context.Context, in *v3.Role, opts ...grpc.CallOption) (*v3.Role, error)
	DeleteRole(ctx context.Context, in *v3.Role, opts ...grpc.CallOption) (*v3.Role, error)
}

type roleClient struct {
	cc grpc.ClientConnInterface
}

func NewRoleClient(cc grpc.ClientConnInterface) RoleClient {
	return &roleClient{cc}
}

func (c *roleClient) CreateRole(ctx context.Context, in *v3.Role, opts ...grpc.CallOption) (*v3.Role, error) {
	out := new(v3.Role)
	err := c.cc.Invoke(ctx, "/paralus.dev.rpc.v3.Role/CreateRole", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *roleClient) GetRoles(ctx context.Context, in *v3.Role, opts ...grpc.CallOption) (*v3.RoleList, error) {
	out := new(v3.RoleList)
	err := c.cc.Invoke(ctx, "/paralus.dev.rpc.v3.Role/GetRoles", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *roleClient) GetRole(ctx context.Context, in *v3.Role, opts ...grpc.CallOption) (*v3.Role, error) {
	out := new(v3.Role)
	err := c.cc.Invoke(ctx, "/paralus.dev.rpc.v3.Role/GetRole", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *roleClient) UpdateRole(ctx context.Context, in *v3.Role, opts ...grpc.CallOption) (*v3.Role, error) {
	out := new(v3.Role)
	err := c.cc.Invoke(ctx, "/paralus.dev.rpc.v3.Role/UpdateRole", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *roleClient) DeleteRole(ctx context.Context, in *v3.Role, opts ...grpc.CallOption) (*v3.Role, error) {
	out := new(v3.Role)
	err := c.cc.Invoke(ctx, "/paralus.dev.rpc.v3.Role/DeleteRole", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RoleServer is the server API for Role service.
// All implementations should embed UnimplementedRoleServer
// for forward compatibility
type RoleServer interface {
	CreateRole(context.Context, *v3.Role) (*v3.Role, error)
	GetRoles(context.Context, *v3.Role) (*v3.RoleList, error)
	GetRole(context.Context, *v3.Role) (*v3.Role, error)
	UpdateRole(context.Context, *v3.Role) (*v3.Role, error)
	DeleteRole(context.Context, *v3.Role) (*v3.Role, error)
}

// UnimplementedRoleServer should be embedded to have forward compatible implementations.
type UnimplementedRoleServer struct {
}

func (UnimplementedRoleServer) CreateRole(context.Context, *v3.Role) (*v3.Role, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateRole not implemented")
}
func (UnimplementedRoleServer) GetRoles(context.Context, *v3.Role) (*v3.RoleList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRoles not implemented")
}
func (UnimplementedRoleServer) GetRole(context.Context, *v3.Role) (*v3.Role, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRole not implemented")
}
func (UnimplementedRoleServer) UpdateRole(context.Context, *v3.Role) (*v3.Role, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateRole not implemented")
}
func (UnimplementedRoleServer) DeleteRole(context.Context, *v3.Role) (*v3.Role, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteRole not implemented")
}

// UnsafeRoleServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RoleServer will
// result in compilation errors.
type UnsafeRoleServer interface {
	mustEmbedUnimplementedRoleServer()
}

func RegisterRoleServer(s grpc.ServiceRegistrar, srv RoleServer) {
	s.RegisterService(&Role_ServiceDesc, srv)
}

func _Role_CreateRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v3.Role)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoleServer).CreateRole(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paralus.dev.rpc.v3.Role/CreateRole",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoleServer).CreateRole(ctx, req.(*v3.Role))
	}
	return interceptor(ctx, in, info, handler)
}

func _Role_GetRoles_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v3.Role)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoleServer).GetRoles(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paralus.dev.rpc.v3.Role/GetRoles",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoleServer).GetRoles(ctx, req.(*v3.Role))
	}
	return interceptor(ctx, in, info, handler)
}

func _Role_GetRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v3.Role)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoleServer).GetRole(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paralus.dev.rpc.v3.Role/GetRole",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoleServer).GetRole(ctx, req.(*v3.Role))
	}
	return interceptor(ctx, in, info, handler)
}

func _Role_UpdateRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v3.Role)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoleServer).UpdateRole(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paralus.dev.rpc.v3.Role/UpdateRole",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoleServer).UpdateRole(ctx, req.(*v3.Role))
	}
	return interceptor(ctx, in, info, handler)
}

func _Role_DeleteRole_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v3.Role)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RoleServer).DeleteRole(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paralus.dev.rpc.v3.Role/DeleteRole",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RoleServer).DeleteRole(ctx, req.(*v3.Role))
	}
	return interceptor(ctx, in, info, handler)
}

// Role_ServiceDesc is the grpc.ServiceDesc for Role service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Role_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "paralus.dev.rpc.v3.Role",
	HandlerType: (*RoleServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateRole",
			Handler:    _Role_CreateRole_Handler,
		},
		{
			MethodName: "GetRoles",
			Handler:    _Role_GetRoles_Handler,
		},
		{
			MethodName: "GetRole",
			Handler:    _Role_GetRole_Handler,
		},
		{
			MethodName: "UpdateRole",
			Handler:    _Role_UpdateRole_Handler,
		},
		{
			MethodName: "DeleteRole",
			Handler:    _Role_DeleteRole_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/rpc/role/role.proto",
}
