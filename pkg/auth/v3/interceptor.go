package authv3

import (
	context "context"
	"reflect"
	"strings"

	"github.com/RafayLabs/rcloud-base/pkg/gateway"
	commonv3 "github.com/RafayLabs/rcloud-base/proto/types/commonpb/v3"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (ac authContext) NewAuthUnaryInterceptor(opt Option) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// TODO: Optimize authentication for a session/gRPC
		// channel
		for _, ex := range opt.ExcludeRPCMethods {
			if ex == info.FullMethod {
				return handler(ctx, req)
			}
		}

		// We have to get the value of org, and project (namespace in
		// future) as we will be using this inorder to authorize the
		// user's access to different resources
		reqValue := reflect.ValueOf(req).Elem()
		field := reqValue.FieldByName("Metadata")
		var org string
		var project string
		if field != (reflect.Value{}) {
			org = field.Interface().(*commonv3.Metadata).Organization
			project = field.Interface().(*commonv3.Metadata).Project
		}

		// overrides for picking up info when not in default metadata locations
		// XXX: This requires any new items which does not follow metadata convention added here
		switch strings.Split(info.FullMethod, "/")[1] {
		case "rafay.dev.rpc.v3.Project":
			project = field.Interface().(*commonv3.Metadata).Name
		case "rafay.dev.rpc.v3.Organization":
			org = field.Interface().(*commonv3.Metadata).Name
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "grpc metadata not exist")
		}
		var (
			url    string
			method string
			token  string
			cookie string
		)
		if len(md.Get(gateway.GatewayURL)) != 0 {
			url = md.Get(gateway.GatewayURL)[0]
		}
		if len(md.Get(gateway.GatewayMethod)) != 0 {
			method = md.Get(gateway.GatewayMethod)[0]
		}
		if len(md.Get(gateway.GatewayAPIKey)) != 0 {
			token = md.Get(gateway.GatewayAPIKey)[0]
		}
		if len(md.Get("grpcgateway-cookie")) != 0 {
			cookie = md.Get("grpcgateway-cookie")[0]
		}
		acReq := &commonv3.IsRequestAllowedRequest{
			Url:           url,
			Method:        method,
			XSessionToken: token,
			Cookie:        cookie,
			Org:           org,
			Project:       project,
		}
		res, err := ac.IsRequestAllowed(ctx, nil, acReq)
		if err != nil {
			_log.Errorf("Failed to authenticate a request: %s", err)
			return nil, status.Error(codes.Internal, codes.Internal.String())
		}

		s := res.GetStatus()
		switch s {
		case commonv3.RequestStatus_RequestAllowed:
			ctx := NewSessionContext(ctx, res.SessionData)
			return handler(ctx, req)
		case commonv3.RequestStatus_RequestMethodOrURLNotAllowed:
			return nil, status.Error(codes.PermissionDenied, res.GetReason())
		case commonv3.RequestStatus_RequestNotAuthenticated:
			return nil, status.Error(codes.Unauthenticated, res.GetReason())
		}

		// status should be any of three above.
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}
}
