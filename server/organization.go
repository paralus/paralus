package server

import (
	"context"

	"github.com/paralus/paralus/pkg/service"
	systemrpc "github.com/paralus/paralus/proto/rpc/system"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	systempbv3 "github.com/paralus/paralus/proto/types/systempb/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type organizationServer struct {
	service.OrganizationService
}

// NewOrganizationServer returns new organization server implementation
func NewOrganizationServer(ps service.OrganizationService) systemrpc.OrganizationServiceServer {
	return &organizationServer{ps}
}

func updateOrganizationStatus(req *systempbv3.Organization, resp *systempbv3.Organization, err error) *systempbv3.Organization {
	if err != nil {
		req.Status = &v3.Status{
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return req
	}
	resp.Status = &v3.Status{ConditionStatus: v3.ConditionStatus_StatusOK}
	return resp
}

func (s *organizationServer) CreateOrganization(ctx context.Context, req *systempbv3.Organization) (*systempbv3.Organization, error) {
	resp, err := s.Create(ctx, req)
	return updateOrganizationStatus(req, resp, err), err
}

func (s *organizationServer) GetOrganizations(ctx context.Context, req *systempbv3.Organization) (*systempbv3.OrganizationList, error) {
	return s.List(ctx, req)
}

func (s *organizationServer) GetOrganization(ctx context.Context, req *systempbv3.Organization) (*systempbv3.Organization, error) {
	resp, err := s.GetByName(ctx, req.Metadata.Name)
	return updateOrganizationStatus(req, resp, err), err
}

func (s *organizationServer) DeleteOrganization(ctx context.Context, req *systempbv3.Organization) (*systempbv3.Organization, error) {
	resp, err := s.Delete(ctx, req)
	return updateOrganizationStatus(req, resp, err), err
}

func (s *organizationServer) UpdateOrganization(ctx context.Context, req *systempbv3.Organization) (*systempbv3.Organization, error) {
	resp, err := s.Update(ctx, req)
	return updateOrganizationStatus(req, resp, err), err
}
