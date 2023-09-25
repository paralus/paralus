package authv3

import (
	context "context"
	"strings"

	"github.com/paralus/paralus/pkg/common"
	"github.com/paralus/paralus/pkg/gateway"
	"github.com/paralus/paralus/pkg/utils"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type hasMetadata interface {
	GetMetadata() *commonv3.Metadata
}

func (ac authContext) NewAuthUnaryInterceptor(opt Option) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// TODO: Optimize authentication for a session/gRPC
		// channel
		for _, ex := range opt.ExcludeRPCMethods {
			if ex == info.FullMethod {
				return handler(ctx, req)
			}
		}

		// We have to get the value of org and project (namespace in
		// future) as we will be using this inorder to authorize the
		// user's access to different resources
		var org string
		var project string
		resource, ok := req.(hasMetadata)
		if ok {
			meta := resource.GetMetadata()
			if meta != nil {
				org = meta.Organization
				project = meta.Project
			}

			// Overrides for picking up info when not in default
			// metadata locations
			// XXX: This requires any new items which does not follow
			// metadata convention to be added here
			switch strings.Split(info.FullMethod, "/")[1] {
			case "paralus.dev.rpc.v3.Project":
				project = meta.Name
			case "paralus.dev.rpc.v3.Organization":
				org = meta.Name
			}
		}

		noAuthz := utils.Contains(opt.ExcludeAuthzMethods, info.FullMethod)

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "grpc metadata not exist")
		}
		var (
			url    string
			method string
			token  string
			apiKey string
			apiTkn string
			cookie string
			host   string
			ua     string
			ip     string
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
		if len(md.Get("X-API-KEYID")) != 0 {
			apiKey = md.Get("X-API-KEYID")[0]
		}
		if len(md.Get("X-API-TOKEN")) != 0 {
			apiTkn = md.Get("X-API-TOKEN")[0]
		}
		if len(md.Get("grpcgateway-cookie")) != 0 {
			cookie = md.Get("grpcgateway-cookie")[0]
		}
		if len(md.Get("x-gateway-host")) != 0 {
			host = md.Get("x-gateway-host")[0]
		}
		if len(md.Get("x-gateway-user-agent")) != 0 {
			ua = md.Get("x-gateway-user-agent")[0]
		}
		if len(md.Get("x-gateway-remote-addr")) != 0 {
			ip = md.Get("x-gateway-remote-addr")[0]
		}

		acReq := &commonv3.IsRequestAllowedRequest{
			Url:           url,
			Method:        method,
			XSessionToken: token,
			XApiKey:       apiKey,
			XApiToken:     apiTkn,
			Cookie:        cookie,
			Org:           org,
			Project:       project,
			NoAuthz:       noAuthz, // FIXME: any better way to do this?
		}

		res, err := ac.IsRequestAllowed(ctx, acReq)
		if err != nil {
			_log.Errorf("Failed to authenticate a request: %s", err)
			return nil, status.Error(codes.Internal, codes.Internal.String())
		}

		s := res.GetStatus()
		_log.Debug("user authentication status ", s)
		switch s {
		case commonv3.RequestStatus_RequestAllowed:
			sd := res.SessionData
			sd.ClientIp = ip
			sd.ClientHost = host
			sd.ClientUa = ua
			ctx := context.WithValue(ctx, common.SessionDataKey, sd)
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
