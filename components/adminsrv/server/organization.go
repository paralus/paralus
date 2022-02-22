package server

import (
	"context"

	"github.com/RafaySystems/rcloud-base/components/adminsrv/pkg/service"
	adminrpc "github.com/RafaySystems/rcloud-base/components/adminsrv/proto/rpc"
	systempbv3 "github.com/RafaySystems/rcloud-base/components/adminsrv/proto/types/systempb/v3"
)

type organizationServer struct {
	service.OrganizationService
}

// NewOrganizationServer returns new organization server implementation
func NewOrganizationServer(ps service.OrganizationService) adminrpc.OrganizationServer {
	return &organizationServer{ps}
}

func (s *organizationServer) CreateOrganization(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error) {
	organization, err := s.Create(ctx, o)
	if err != nil {
		return o, err
	}
	return organization, nil
}

func (s *organizationServer) GetOrganizations(ctx context.Context, o *systempbv3.Organization) (*systempbv3.OrganizationList, error) {
	organizations, err := s.List(ctx, o)
	if err != nil {
		return nil, err
	}
	return organizations, nil
}

func (s *organizationServer) GetOrganization(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error) {

	organization, err := s.GetByName(ctx, o.Metadata.Name)
	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (s *organizationServer) DeleteOrganization(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error) {
	organization, err := s.Delete(ctx, o)
	if err != nil {
		return o, err
	}
	return organization, nil
}

func (s *organizationServer) UpdateOrganization(ctx context.Context, o *systempbv3.Organization) (*systempbv3.Organization, error) {
	organization, err := s.Update(ctx, o)
	if err != nil {
		return o, err
	}
	return organization, nil
}
