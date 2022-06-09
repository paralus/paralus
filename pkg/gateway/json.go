package gateway

import (
	"io"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/segmentio/encoding/json"
)

const (
	jsonContentType string = "application/json"
)

// paralusJSON is the paralus object to json marshaller
type paralusJSON struct {
}

// NewParalusJSON returns new grpc gateway paralus json marshaller
func NewParalusJSON() runtime.Marshaler {
	return &paralusJSON{}
}

// Marshal marshals "v" into byte sequence.
func (m *paralusJSON) Marshal(v interface{}) ([]byte, error) {

	return json.Marshal(v)
}

// Unmarshal unmarshals "data" into "v".
// "v" must be a pointer value.
func (m *paralusJSON) Unmarshal(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}

// NewDecoder returns a Decoder which reads byte sequence from "r".
func (m *paralusJSON) NewDecoder(r io.Reader) runtime.Decoder {
	return json.NewDecoder(r)
}

// NewEncoder returns an Encoder which writes bytes sequence into "w".
func (m *paralusJSON) NewEncoder(w io.Writer) runtime.Encoder {
	return json.NewEncoder(w)

}

// ContentType returns the Content-Type which this marshaler is responsible for.
func (m *paralusJSON) ContentType(v interface{}) string {
	return jsonContentType
}
