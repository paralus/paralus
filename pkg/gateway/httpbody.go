package gateway

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	common "github.com/paralus/paralus/proto/types/commonpb/v3"
)

// httpBodyMarshaler is a Marshaler which supports marshaling of a
// paralus.dev.common.types.v2.HttpBody message as the full response body if it is
// the actual message used as the response. If not, then this will
// simply fallback to the Marshaler specified as its default Marshaler.
type httpBodyMarshaler struct {
	runtime.Marshaler
}

// NewHTTPBodyMarshaler returns new http body marshaler
func NewHTTPBodyMarshaler() runtime.Marshaler {
	return &httpBodyMarshaler{
		Marshaler: &paralusJSON{},
	}
}

// ContentType implementation to keep backwards compatibility with marshal interface
func (h *httpBodyMarshaler) ContentType(v interface{}) string {
	return h.ContentTypeFromMessage(nil)
}

// ContentTypeFromMessage in case v is a google.api.HttpBody message it returns
// its specified content type otherwise fall back to the default Marshaler.
func (h *httpBodyMarshaler) ContentTypeFromMessage(v interface{}) string {
	if httpBody, ok := v.(*common.HttpBody); ok {
		return httpBody.GetContentType()
	}
	return h.Marshaler.ContentType(v)
}

// Marshal marshals "v" by returning the body bytes if v is a
// google.api.HttpBody message, otherwise it falls back to the default Marshaler.
func (h *httpBodyMarshaler) Marshal(v interface{}) ([]byte, error) {
	if httpBody, ok := v.(*common.HttpBody); ok {
		return httpBody.Data, nil
	}
	return h.Marshaler.Marshal(v)
}
