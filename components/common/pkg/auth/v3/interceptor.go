package authv3

import (
	context "context"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/gateway"
	commonpbv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
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
		acReq := &commonpbv3.IsRequestAllowedRequest{
			Url:           url,
			Method:        method,
			XSessionToken: token,
			Cookie:        cookie,
		}
		res, err := ac.IsRequestAllowed(ctx, acReq)
		if err != nil {
			_log.Errorf("Failed to authenticate a request: %s", err)
			return nil, status.Error(codes.Internal, codes.Internal.String())
		}

		s := res.GetStatus()
		switch s {
		case commonpbv3.RequestStatus_RequestAllowed:
			ctx := newSessionContext(ctx, res.SessionData)
			return handler(ctx, req)
		case commonpbv3.RequestStatus_RequestMethodOrURLNotAllowed:
			return nil, status.Error(codes.PermissionDenied, res.GetReason())
		case commonpbv3.RequestStatus_RequestNotAuthenticated:
			return nil, status.Error(codes.Unauthenticated, res.GetReason())
		}

		// status should be any of three above.
		return nil, status.Error(codes.Internal, codes.Internal.String())
	}
}
