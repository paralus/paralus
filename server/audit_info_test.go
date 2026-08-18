package server

import (
	"context"
	"errors"
	"testing"

	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/pkg/query"
	sentryrpc "github.com/paralus/paralus/proto/rpc/sentry"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/paralus/paralus/proto/types/sentry"
	systempbv3 "github.com/paralus/paralus/proto/types/systempb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeAccountPermissionService implements service.AccountPermissionService
// with function fields so each test wires only the method it needs.
type fakeAccountPermissionService struct {
	getAccountFn func(ctx context.Context, accountID string) (*models.Account, error)
}

func (f *fakeAccountPermissionService) GetAccountPermissions(ctx context.Context, accountID string, orgID, partnerID string) ([]sentry.AccountPermission, error) {
	panic("not implemented")
}
func (f *fakeAccountPermissionService) IsPartnerSuperAdmin(ctx context.Context, accountID, partnerID string) (bool, bool, error) {
	panic("not implemented")
}
func (f *fakeAccountPermissionService) GetAccountProjectsByPermission(ctx context.Context, accountID, orgID, partnerID string, permission string) ([]sentry.AccountPermission, error) {
	panic("not implemented")
}
func (f *fakeAccountPermissionService) GetAccountPermissionsByProjectIDPermissions(ctx context.Context, accountID, orgID, partnerID string, projects, permissions []string) ([]sentry.AccountPermission, error) {
	panic("not implemented")
}
func (f *fakeAccountPermissionService) GetAcccountsWithApprovalPermission(ctx context.Context, orgID, partnerID string) ([]string, error) {
	panic("not implemented")
}
func (f *fakeAccountPermissionService) GetSSOAcccountsWithApprovalPermission(ctx context.Context, orgID, partnerID string) ([]string, error) {
	panic("not implemented")
}
func (f *fakeAccountPermissionService) IsOrgAdmin(ctx context.Context, accountID, partnerID string) (bool, error) {
	panic("not implemented")
}
func (f *fakeAccountPermissionService) GetAccount(ctx context.Context, accountID string) (*models.Account, error) {
	return f.getAccountFn(ctx, accountID)
}
func (f *fakeAccountPermissionService) GetAccountGroups(ctx context.Context, accountID string) ([]string, error) {
	panic("not implemented")
}
func (f *fakeAccountPermissionService) IsAccountActive(ctx context.Context, accountID, orgID string) (bool, error) {
	panic("not implemented")
}
func (f *fakeAccountPermissionService) IsSSOAccount(ctx context.Context, accountID string) (bool, error) {
	panic("not implemented")
}

// auditInfoBootstrapService embeds fakeBootstrapService and adds the one
// extra method (GetBootstrapAgentTemplateForHost) that LookupCluster needs
// and fakeBootstrapService otherwise panics on.
type auditInfoBootstrapService struct {
	fakeBootstrapService
	getBootstrapAgentTemplateForHostFn func(ctx context.Context, host string) (*sentry.BootstrapAgentTemplate, error)
}

func (f *auditInfoBootstrapService) GetBootstrapAgentTemplateForHost(ctx context.Context, host string) (*sentry.BootstrapAgentTemplate, error) {
	return f.getBootstrapAgentTemplateForHostFn(ctx, host)
}

func TestAuditInfoServer_LookupUser(t *testing.T) {
	t.Run("system user is resolved entirely from the CN without a service call", func(t *testing.T) {
		s := NewAuditInfoServer(&fakeBootstrapService{}, &fakeAccountPermissionService{}, &fakeProjectService{})

		cn := "su=true/u=sysuser/a=acc1/o=org1/p=part1/st=rs"
		resp, err := s.LookupUser(context.Background(), &sentryrpc.LookupUserRequest{UserCN: cn})
		require.NoError(t, err)
		assert.Equal(t, "sysuser", resp.UserName)
		assert.Equal(t, "acc1", resp.AccountID)
		assert.Equal(t, "org1", resp.OrganizationID)
		assert.Equal(t, "part1", resp.PartnerID)
		assert.Equal(t, "paralus system", resp.SessionType)
	})

	t.Run("non-system user is resolved via GetAccount", func(t *testing.T) {
		aps := &fakeAccountPermissionService{getAccountFn: func(ctx context.Context, accountID string) (*models.Account, error) {
			assert.Equal(t, "acc1", accountID)
			return &models.Account{Username: "real-user"}, nil
		}}
		s := NewAuditInfoServer(&fakeBootstrapService{}, aps, &fakeProjectService{})

		cn := "a=acc1/o=org1/p=part1/st=ts"
		resp, err := s.LookupUser(context.Background(), &sentryrpc.LookupUserRequest{UserCN: cn})
		require.NoError(t, err)
		assert.Equal(t, "real-user", resp.UserName)
		assert.Equal(t, "kubectl cli", resp.SessionType)
	})

	t.Run("propagates a GetAccount error", func(t *testing.T) {
		want := errors.New("boom")
		aps := &fakeAccountPermissionService{getAccountFn: func(ctx context.Context, accountID string) (*models.Account, error) {
			return nil, want
		}}
		s := NewAuditInfoServer(&fakeBootstrapService{}, aps, &fakeProjectService{})

		_, err := s.LookupUser(context.Background(), &sentryrpc.LookupUserRequest{UserCN: "a=acc1"})
		assert.Same(t, want, err)
	})
}

func TestAuditInfoServer_LookupCluster(t *testing.T) {
	t.Run("rejects a cluster SNI without a dot separator", func(t *testing.T) {
		s := NewAuditInfoServer(&auditInfoBootstrapService{}, &fakeAccountPermissionService{}, &fakeProjectService{})

		_, err := s.LookupCluster(context.Background(), &sentryrpc.LookupClusterRequest{ClusterSNI: "no-dot-here"})
		assert.Error(t, err)
	})

	t.Run("propagates a GetBootstrapAgentTemplateForHost error", func(t *testing.T) {
		want := errors.New("template not found")
		bs := &auditInfoBootstrapService{
			getBootstrapAgentTemplateForHostFn: func(ctx context.Context, host string) (*sentry.BootstrapAgentTemplate, error) {
				return nil, want
			},
		}
		s := NewAuditInfoServer(bs, &fakeAccountPermissionService{}, &fakeProjectService{})

		_, err := s.LookupCluster(context.Background(), &sentryrpc.LookupClusterRequest{ClusterSNI: "cluster1.relay.example.com"})
		assert.Same(t, want, err)
	})

	t.Run("propagates a GetBootstrapAgent error", func(t *testing.T) {
		want := errors.New("agent not found")
		bs := &auditInfoBootstrapService{
			getBootstrapAgentTemplateForHostFn: func(ctx context.Context, host string) (*sentry.BootstrapAgentTemplate, error) {
				return &sentry.BootstrapAgentTemplate{Metadata: &commonv3.Metadata{Labels: map[string]string{
					"paralus.dev/connectorAgentTemplate": "tmpl1",
				}}}, nil
			},
		}
		bs.fakeBootstrapService.getBootstrapAgentFn = func(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgent, error) {
			return nil, want
		}
		s := NewAuditInfoServer(bs, &fakeAccountPermissionService{}, &fakeProjectService{})

		_, err := s.LookupCluster(context.Background(), &sentryrpc.LookupClusterRequest{ClusterSNI: "cluster1.relay.example.com"})
		assert.Same(t, want, err)
	})

	t.Run("propagates a project lookup error", func(t *testing.T) {
		want := errors.New("project not found")
		bs := &auditInfoBootstrapService{
			getBootstrapAgentTemplateForHostFn: func(ctx context.Context, host string) (*sentry.BootstrapAgentTemplate, error) {
				return &sentry.BootstrapAgentTemplate{Metadata: &commonv3.Metadata{Labels: map[string]string{
					"paralus.dev/connectorAgentTemplate": "tmpl1",
				}}}, nil
			},
		}
		bs.fakeBootstrapService.getBootstrapAgentFn = func(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgent, error) {
			return &sentry.BootstrapAgent{Metadata: &commonv3.Metadata{
				Project: "proj1",
				Labels:  map[string]string{"paralus.dev/clusterName": "cluster1"},
			}}, nil
		}
		prs := &fakeProjectService{getByIDFn: func(ctx context.Context, id string) (*systempbv3.Project, error) {
			return nil, want
		}}
		s := NewAuditInfoServer(bs, &fakeAccountPermissionService{}, prs)

		_, err := s.LookupCluster(context.Background(), &sentryrpc.LookupClusterRequest{ClusterSNI: "cluster1.relay.example.com"})
		assert.Same(t, want, err)
	})

	t.Run("success resolves the cluster name and project name", func(t *testing.T) {
		bs := &auditInfoBootstrapService{
			getBootstrapAgentTemplateForHostFn: func(ctx context.Context, host string) (*sentry.BootstrapAgentTemplate, error) {
				assert.Equal(t, "*.relay.example.com", host)
				return &sentry.BootstrapAgentTemplate{Metadata: &commonv3.Metadata{Labels: map[string]string{
					"paralus.dev/connectorAgentTemplate": "tmpl1",
				}}}, nil
			},
		}
		bs.fakeBootstrapService.getBootstrapAgentFn = func(ctx context.Context, templateRef string, opts ...query.Option) (*sentry.BootstrapAgent, error) {
			assert.Equal(t, "tmpl1", templateRef)
			return &sentry.BootstrapAgent{Metadata: &commonv3.Metadata{
				Project: "proj1",
				Labels:  map[string]string{"paralus.dev/clusterName": "cluster1"},
			}}, nil
		}
		prs := &fakeProjectService{getByIDFn: func(ctx context.Context, id string) (*systempbv3.Project, error) {
			assert.Equal(t, "proj1", id)
			return &systempbv3.Project{Metadata: &commonv3.Metadata{Name: "project-one"}}, nil
		}}
		s := NewAuditInfoServer(bs, &fakeAccountPermissionService{}, prs)

		resp, err := s.LookupCluster(context.Background(), &sentryrpc.LookupClusterRequest{ClusterSNI: "cluster1.relay.example.com"})
		require.NoError(t, err)
		assert.Equal(t, "cluster1", resp.Name)
		assert.Equal(t, "project-one", resp.Project)
	})
}
