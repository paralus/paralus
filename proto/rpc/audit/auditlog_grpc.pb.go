// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: proto/rpc/audit/auditlog.proto

package eventv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// AuditLogClient is the client API for AuditLog service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AuditLogClient interface {
	GetAuditLog(ctx context.Context, in *AuditLogSearchRequest, opts ...grpc.CallOption) (*AuditLogSearchResponse, error)
	GetAuditLogByProjects(ctx context.Context, in *AuditLogSearchRequest, opts ...grpc.CallOption) (*AuditLogSearchResponse, error)
}

type auditLogClient struct {
	cc grpc.ClientConnInterface
}

func NewAuditLogClient(cc grpc.ClientConnInterface) AuditLogClient {
	return &auditLogClient{cc}
}

func (c *auditLogClient) GetAuditLog(ctx context.Context, in *AuditLogSearchRequest, opts ...grpc.CallOption) (*AuditLogSearchResponse, error) {
	out := new(AuditLogSearchResponse)
	err := c.cc.Invoke(ctx, "/rep.framework.event.v1.AuditLog/getAuditLog", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *auditLogClient) GetAuditLogByProjects(ctx context.Context, in *AuditLogSearchRequest, opts ...grpc.CallOption) (*AuditLogSearchResponse, error) {
	out := new(AuditLogSearchResponse)
	err := c.cc.Invoke(ctx, "/rep.framework.event.v1.AuditLog/getAuditLogByProjects", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuditLogServer is the server API for AuditLog service.
// All implementations should embed UnimplementedAuditLogServer
// for forward compatibility
type AuditLogServer interface {
	GetAuditLog(context.Context, *AuditLogSearchRequest) (*AuditLogSearchResponse, error)
	GetAuditLogByProjects(context.Context, *AuditLogSearchRequest) (*AuditLogSearchResponse, error)
}

// UnimplementedAuditLogServer should be embedded to have forward compatible implementations.
type UnimplementedAuditLogServer struct {
}

func (UnimplementedAuditLogServer) GetAuditLog(context.Context, *AuditLogSearchRequest) (*AuditLogSearchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAuditLog not implemented")
}
func (UnimplementedAuditLogServer) GetAuditLogByProjects(context.Context, *AuditLogSearchRequest) (*AuditLogSearchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAuditLogByProjects not implemented")
}

// UnsafeAuditLogServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuditLogServer will
// result in compilation errors.
type UnsafeAuditLogServer interface {
	mustEmbedUnimplementedAuditLogServer()
}

func RegisterAuditLogServer(s grpc.ServiceRegistrar, srv AuditLogServer) {
	s.RegisterService(&AuditLog_ServiceDesc, srv)
}

func _AuditLog_GetAuditLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuditLogSearchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuditLogServer).GetAuditLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rep.framework.event.v1.AuditLog/getAuditLog",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuditLogServer).GetAuditLog(ctx, req.(*AuditLogSearchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuditLog_GetAuditLogByProjects_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuditLogSearchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuditLogServer).GetAuditLogByProjects(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rep.framework.event.v1.AuditLog/getAuditLogByProjects",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuditLogServer).GetAuditLogByProjects(ctx, req.(*AuditLogSearchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AuditLog_ServiceDesc is the grpc.ServiceDesc for AuditLog service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AuditLog_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rep.framework.event.v1.AuditLog",
	HandlerType: (*AuditLogServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "getAuditLog",
			Handler:    _AuditLog_GetAuditLog_Handler,
		},
		{
			MethodName: "getAuditLogByProjects",
			Handler:    _AuditLog_GetAuditLogByProjects_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/rpc/audit/auditlog.proto",
}
