package server

import (
	"context"
	"errors"
	"testing"

	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/stretchr/testify/assert"
)

// fakeAuthService implements authv3.AuthService with a single function
// field; IsRequestAllowed is its only method.
type fakeAuthService struct {
	isRequestAllowedFn func(ctx context.Context, ira *v3.IsRequestAllowedRequest) (*v3.IsRequestAllowedResponse, error)
}

func (f *fakeAuthService) IsRequestAllowed(ctx context.Context, ira *v3.IsRequestAllowedRequest) (*v3.IsRequestAllowedResponse, error) {
	return f.isRequestAllowedFn(ctx, ira)
}

func TestAuthServer_IsRequestAllowed(t *testing.T) {
	t.Run("delegates to the auth service and returns its response", func(t *testing.T) {
		want := &v3.IsRequestAllowedResponse{Status: v3.RequestStatus_RequestAllowed}
		as := &fakeAuthService{isRequestAllowedFn: func(ctx context.Context, ira *v3.IsRequestAllowedRequest) (*v3.IsRequestAllowedResponse, error) {
			return want, nil
		}}
		s := NewAuthServer(as)

		got, err := s.IsRequestAllowed(context.Background(), &v3.IsRequestAllowedRequest{})
		assert.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("propagates the auth service error unchanged", func(t *testing.T) {
		want := errors.New("boom")
		as := &fakeAuthService{isRequestAllowedFn: func(ctx context.Context, ira *v3.IsRequestAllowedRequest) (*v3.IsRequestAllowedResponse, error) {
			return nil, want
		}}
		s := NewAuthServer(as)

		_, err := s.IsRequestAllowed(context.Background(), &v3.IsRequestAllowedRequest{})
		assert.Same(t, want, err)
	})
}
