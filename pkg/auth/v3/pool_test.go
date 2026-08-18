package authv3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAuthPool(t *testing.T) {
	p := NewAuthPool("localhost:1", 1)
	assert.NotNil(t, p)
}

func TestAuthPool_Close_NilInnerPoolDoesNotPanic(t *testing.T) {
	p := &authPool{}
	assert.NotPanics(t, func() { p.Close() })
}
