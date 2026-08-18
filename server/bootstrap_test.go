package server

import (
	"context"
	"database/sql"
	"testing"

	"github.com/paralus/paralus/pkg/query"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fakeBootstrapService implements service.BootstrapService with function
// fields so each test wires only the method it needs; everything else
// panics to surface accidental extra calls.
type fakeBootstrapService struct {
	getBootstrapAgentFn    func(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgent, error)
	deleteBootstrapAgentFn func(ctx context.Context, templateRef string, opts ...query.Option) error
	patchBootstrapAgentFn  func(ctx context.Context, ba *sentry.BootstrapAgent, templateRef string, opts ...query.Option) error
}

func (f *fakeBootstrapService) PatchBootstrapInfra(ctx context.Context, infra *sentry.BootstrapInfra) error {
	panic("not implemented")
}
func (f *fakeBootstrapService) GetBootstrapInfra(ctx context.Context, name string) (*sentry.BootstrapInfra, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) PatchBootstrapAgentTemplate(ctx context.Context, template *sentry.BootstrapAgentTemplate) error {
	panic("not implemented")
}
func (f *fakeBootstrapService) GetBootstrapAgentTemplate(ctx context.Context, name string) (*sentry.BootstrapAgentTemplate, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) GetBootstrapAgentTemplateForToken(ctx context.Context, token string) (*sentry.BootstrapAgentTemplate, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) GetBootstrapAgentTemplateForHost(ctx context.Context, host string) (*sentry.BootstrapAgentTemplate, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) SelectBootstrapAgentTemplates(ctx context.Context, opts ...query.Option) (*sentry.BootstrapAgentTemplateList, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) CreateBootstrapAgent(ctx context.Context, agent *sentry.BootstrapAgent) error {
	panic("not implemented")
}
func (f *fakeBootstrapService) GetBootstrapAgent(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgent, error) {
	return f.getBootstrapAgentFn(ctx, templateRef, opts...)
}
func (f *fakeBootstrapService) GetBootstrapAgents(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgentList, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) GetBootstrapAgentForToken(ctx context.Context, token string) (*sentry.BootstrapAgent, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) GetBootstrapAgentCountForClusterID(ctx context.Context, clusterID string, orgID string) (int, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) GetBootstrapAgentForClusterID(ctx context.Context, clusterID string, orgID string) (*sentry.BootstrapAgent, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) SelectBootstrapAgents(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgentList, error) {
	panic("not implemented")
}
func (f *fakeBootstrapService) RegisterBootstrapAgent(ctx context.Context, token, ip, fingerprint string) error {
	panic("not implemented")
}
func (f *fakeBootstrapService) DeleteBootstrapAgent(ctx context.Context, templateRef string, opts ...query.Option) error {
	return f.deleteBootstrapAgentFn(ctx, templateRef, opts...)
}
func (f *fakeBootstrapService) PatchBootstrapAgent(ctx context.Context, ba *sentry.BootstrapAgent, templateRef string, opts ...query.Option) error {
	return f.patchBootstrapAgentFn(ctx, ba, templateRef, opts...)
}

func TestBootstrapServer_GetBootstrapAgent(t *testing.T) {
	t.Run("invalid template scope is rejected before touching the service", func(t *testing.T) {
		bs := &fakeBootstrapService{}
		s := NewBootstrapServer(bs, nil, nil)

		in := &sentry.BootstrapAgent{
			Metadata: &commonv3.Metadata{},
			Spec:     &sentry.BootstrapAgentSpec{TemplateRef: "not-a-template-scope"},
		}
		_, err := s.GetBootstrapAgent(context.Background(), in)
		assert.Error(t, err)
	})

	t.Run("translates sql.ErrNoRows into a gRPC NotFound status", func(t *testing.T) {
		bs := &fakeBootstrapService{
			getBootstrapAgentFn: func(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgent, error) {
				return nil, sql.ErrNoRows
			},
		}
		s := NewBootstrapServer(bs, nil, nil)

		in := &sentry.BootstrapAgent{
			Metadata: &commonv3.Metadata{},
			Spec:     &sentry.BootstrapAgentSpec{TemplateRef: "template/my-scope"},
		}
		_, err := s.GetBootstrapAgent(context.Background(), in)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("passes through the resolved scope and returns the agent on success", func(t *testing.T) {
		var gotScope string
		want := &sentry.BootstrapAgent{Metadata: &commonv3.Metadata{Name: "agent-1"}}
		bs := &fakeBootstrapService{
			getBootstrapAgentFn: func(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgent, error) {
				gotScope = templateRef
				return want, nil
			},
		}
		s := NewBootstrapServer(bs, nil, nil)

		in := &sentry.BootstrapAgent{
			Metadata: &commonv3.Metadata{},
			Spec:     &sentry.BootstrapAgentSpec{TemplateRef: "template/my-scope"},
		}
		got, err := s.GetBootstrapAgent(context.Background(), in)
		require.NoError(t, err)
		assert.Equal(t, "my-scope", gotScope)
		assert.Same(t, want, got)
	})

	t.Run("propagates a non-ErrNoRows service error unchanged", func(t *testing.T) {
		want := status.Error(codes.Internal, "db is down")
		bs := &fakeBootstrapService{
			getBootstrapAgentFn: func(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgent, error) {
				return nil, want
			},
		}
		s := NewBootstrapServer(bs, nil, nil)

		in := &sentry.BootstrapAgent{
			Metadata: &commonv3.Metadata{},
			Spec:     &sentry.BootstrapAgentSpec{TemplateRef: "template/my-scope"},
		}
		_, err := s.GetBootstrapAgent(context.Background(), in)
		assert.Same(t, want, err)
	})
}

func TestBootstrapServer_DeleteBootstrapAgent(t *testing.T) {
	t.Run("invalid template scope is rejected before touching the service", func(t *testing.T) {
		bs := &fakeBootstrapService{}
		s := NewBootstrapServer(bs, nil, nil)

		in := &sentry.BootstrapAgent{
			Metadata: &commonv3.Metadata{},
			Spec:     &sentry.BootstrapAgentSpec{TemplateRef: "bogus"},
		}
		resp, err := s.DeleteBootstrapAgent(context.Background(), in)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("translates sql.ErrNoRows into a gRPC NotFound status but still returns a response", func(t *testing.T) {
		bs := &fakeBootstrapService{
			deleteBootstrapAgentFn: func(ctx context.Context, templateRef string, opts ...query.Option) error {
				return sql.ErrNoRows
			},
		}
		s := NewBootstrapServer(bs, nil, nil)

		in := &sentry.BootstrapAgent{
			Metadata: &commonv3.Metadata{},
			Spec:     &sentry.BootstrapAgentSpec{TemplateRef: "template/my-scope"},
		}
		resp, err := s.DeleteBootstrapAgent(context.Background(), in)
		require.Error(t, err)
		require.NotNil(t, resp)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("succeeds when the service reports no error", func(t *testing.T) {
		bs := &fakeBootstrapService{
			deleteBootstrapAgentFn: func(ctx context.Context, templateRef string, opts ...query.Option) error {
				return nil
			},
		}
		s := NewBootstrapServer(bs, nil, nil)

		in := &sentry.BootstrapAgent{
			Metadata: &commonv3.Metadata{},
			Spec:     &sentry.BootstrapAgentSpec{TemplateRef: "template/my-scope"},
		}
		resp, err := s.DeleteBootstrapAgent(context.Background(), in)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}

func TestBootstrapServer_UpdateBootstrapAgent(t *testing.T) {
	t.Run("invalid template scope is rejected before touching the service", func(t *testing.T) {
		bs := &fakeBootstrapService{}
		s := NewBootstrapServer(bs, nil, nil)

		in := &sentry.BootstrapAgent{
			Metadata: &commonv3.Metadata{},
			Spec:     &sentry.BootstrapAgentSpec{TemplateRef: "bogus"},
		}
		_, err := s.UpdateBootstrapAgent(context.Background(), in)
		assert.Error(t, err)
	})

	t.Run("translates sql.ErrNoRows into a gRPC NotFound status and echoes the request back as ret", func(t *testing.T) {
		bs := &fakeBootstrapService{
			patchBootstrapAgentFn: func(ctx context.Context, ba *sentry.BootstrapAgent, templateRef string, opts ...query.Option) error {
				return sql.ErrNoRows
			},
		}
		s := NewBootstrapServer(bs, nil, nil)

		in := &sentry.BootstrapAgent{
			Metadata: &commonv3.Metadata{},
			Spec:     &sentry.BootstrapAgentSpec{TemplateRef: "template/my-scope"},
		}
		ret, err := s.UpdateBootstrapAgent(context.Background(), in)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Same(t, in, ret)
	})

	t.Run("succeeds and echoes the request back as ret", func(t *testing.T) {
		bs := &fakeBootstrapService{
			patchBootstrapAgentFn: func(ctx context.Context, ba *sentry.BootstrapAgent, templateRef string, opts ...query.Option) error {
				return nil
			},
		}
		s := NewBootstrapServer(bs, nil, nil)

		in := &sentry.BootstrapAgent{
			Metadata: &commonv3.Metadata{},
			Spec:     &sentry.BootstrapAgentSpec{TemplateRef: "template/my-scope"},
		}
		ret, err := s.UpdateBootstrapAgent(context.Background(), in)
		require.NoError(t, err)
		assert.Same(t, in, ret)
	})
}
