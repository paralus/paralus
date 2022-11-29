// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: proto/rpc/sentry/cluster_authz.proto

/*
Package sentry is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package sentry

import (
	"context"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Suppress "imported and not used" errors
var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray
var _ = metadata.Join

var (
	filter_ClusterAuthorizationService_GetUserAuthorization_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_ClusterAuthorizationService_GetUserAuthorization_0(ctx context.Context, marshaler runtime.Marshaler, client ClusterAuthorizationServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetUserAuthorizationRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_ClusterAuthorizationService_GetUserAuthorization_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.GetUserAuthorization(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_ClusterAuthorizationService_GetUserAuthorization_0(ctx context.Context, marshaler runtime.Marshaler, server ClusterAuthorizationServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetUserAuthorizationRequest
	var metadata runtime.ServerMetadata

	if err := req.ParseForm(); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err := runtime.PopulateQueryParameters(&protoReq, req.Form, filter_ClusterAuthorizationService_GetUserAuthorization_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := server.GetUserAuthorization(ctx, &protoReq)
	return msg, metadata, err

}

// RegisterClusterAuthorizationServiceHandlerServer registers the http handlers for service ClusterAuthorizationService to "mux".
// UnaryRPC     :call ClusterAuthorizationServiceServer directly.
// StreamingRPC :currently unsupported pending https://github.com/grpc/grpc-go/issues/906.
// Note that using this registration option will cause many gRPC library features to stop working. Consider using RegisterClusterAuthorizationServiceHandlerFromEndpoint instead.
func RegisterClusterAuthorizationServiceHandlerServer(ctx context.Context, mux *runtime.ServeMux, server ClusterAuthorizationServiceServer) error {

	mux.Handle("GET", pattern_ClusterAuthorizationService_GetUserAuthorization_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateIncomingContext(ctx, mux, req, "/paralus.dev.sentry.rpc.ClusterAuthorizationService/GetUserAuthorization", runtime.WithHTTPPathPattern("/v2/sentry/authorization/user"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_ClusterAuthorizationService_GetUserAuthorization_0(annotatedContext, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_ClusterAuthorizationService_GetUserAuthorization_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

// RegisterClusterAuthorizationServiceHandlerFromEndpoint is same as RegisterClusterAuthorizationServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterClusterAuthorizationServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterClusterAuthorizationServiceHandler(ctx, mux, conn)
}

// RegisterClusterAuthorizationServiceHandler registers the http handlers for service ClusterAuthorizationService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterClusterAuthorizationServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterClusterAuthorizationServiceHandlerClient(ctx, mux, NewClusterAuthorizationServiceClient(conn))
}

// RegisterClusterAuthorizationServiceHandlerClient registers the http handlers for service ClusterAuthorizationService
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "ClusterAuthorizationServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "ClusterAuthorizationServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "ClusterAuthorizationServiceClient" to call the correct interceptors.
func RegisterClusterAuthorizationServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client ClusterAuthorizationServiceClient) error {

	mux.Handle("GET", pattern_ClusterAuthorizationService_GetUserAuthorization_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		var err error
		var annotatedContext context.Context
		annotatedContext, err = runtime.AnnotateContext(ctx, mux, req, "/paralus.dev.sentry.rpc.ClusterAuthorizationService/GetUserAuthorization", runtime.WithHTTPPathPattern("/v2/sentry/authorization/user"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_ClusterAuthorizationService_GetUserAuthorization_0(annotatedContext, inboundMarshaler, client, req, pathParams)
		annotatedContext = runtime.NewServerMetadataContext(annotatedContext, md)
		if err != nil {
			runtime.HTTPError(annotatedContext, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_ClusterAuthorizationService_GetUserAuthorization_0(annotatedContext, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_ClusterAuthorizationService_GetUserAuthorization_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3}, []string{"v2", "sentry", "authorization", "user"}, ""))
)

var (
	forward_ClusterAuthorizationService_GetUserAuthorization_0 = runtime.ForwardResponseMessage
)