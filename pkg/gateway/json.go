package gateway

import (
	"io"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/segmentio/encoding/json"
)

const (
	jsonContentType string = "application/json"
)

// rafayJSON is the rafay object to json marshaller
type rafayJSON struct {
}

// NewParalusJSON returns new grpc gateway rafay json marshaller
func NewParalusJSON() runtime.Marshaler {
	return &rafayJSON{}
}

// Marshal marshals "v" into byte sequence.
func (m *rafayJSON) Marshal(v interface{}) ([]byte, error) {

	return json.Marshal(v)
}

// Unmarshal unmarshals "data" into "v".
// "v" must be a pointer value.
func (m *rafayJSON) Unmarshal(b []byte, v interface{}) error {
	return json.Unmarshal(b, v)
}

// NewDecoder returns a Decoder which reads byte sequence from "r".
func (m *rafayJSON) NewDecoder(r io.Reader) runtime.Decoder {
	return json.NewDecoder(r)
}

// NewEncoder returns an Encoder which writes bytes sequence into "w".
func (m *rafayJSON) NewEncoder(w io.Writer) runtime.Encoder {
	return json.NewEncoder(w)

}

// ContentType returns the Content-Type which this marshaler is responsible for.
func (m *rafayJSON) ContentType(v interface{}) string {
	return jsonContentType
}
