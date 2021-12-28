package gateway

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// ErrNoHandlers returned when not handlers are passed to gateway
var ErrNoHandlers = errors.New("no handlers defined")

// HandlerFromEndpoint defines the function for registering grpc gateway handlers to grpc endpoint
type HandlerFromEndpoint func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)

// NewGateway returns new grpc gateway
func NewGateway(ctx context.Context, endpoint string, serveMuxOptions []runtime.ServeMuxOption, handlers ...HandlerFromEndpoint) (http.Handler, error) {

	rafayJSON := NewRafayJSON()
	rafayYAML := NewRafayYAML()
	httpBody := NewHTTPBodyMarshaler()
	serveMuxOptions = append(serveMuxOptions,
		runtime.WithMarshalerOption(runtime.MIMEWildcard, httpBody),
		runtime.WithMarshalerOption(jsonContentType, rafayJSON),
		runtime.WithMarshalerOption(yamlContentType, rafayYAML),
		//runtime.WithProtoErrorHandler(customErrorHandler),
		//runtime.WithIncomingHeaderMatcher(customHeaderMatcher),
		runtime.WithMetadata(rafayGatewayAnnotator),
	)

	mux := runtime.NewServeMux(serveMuxOptions...)

	if len(handlers) < 1 {
		return nil, ErrNoHandlers
	}

	opts := []grpc.DialOption{grpc.WithInsecure()}

	for _, handler := range handlers {
		err := handler(ctx, mux, endpoint, opts)
		if err != nil {
			return nil, err
		}
	}

	return mux, nil
}
