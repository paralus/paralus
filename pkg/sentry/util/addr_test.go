package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAddr(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		wantHost string
		wantPort int32
	}{
		{"host and port", "example.com:8080", "example.com", 8080},
		{"no colon returns empty host and zero port", "example.com", "", 0},
		{"empty string", "", "", 0},
		{"invalid port text yields zero port", "example.com:abc", "example.com", 0},
		{"negative port yields zero port", "example.com:-1", "example.com", 0},
		{"zero port yields zero port", "example.com:0", "example.com", 0},
		{"ipv4 with port", "127.0.0.1:9090", "127.0.0.1", 9090},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port := ParseAddr(tt.addr)
			assert.Equal(t, tt.wantHost, host)
			assert.Equal(t, tt.wantPort, port)
		})
	}
}
