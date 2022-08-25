package server

import (
	"context"

	"github.com/paralus/paralus/pkg/service"
	systemrpc "github.com/paralus/paralus/proto/rpc/system"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	systempbv3 "github.com/paralus/paralus/proto/types/systempb/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type partnerServer struct {
	service.PartnerService
}

// NewPartnerServer returns new partner server implementation
func NewPartnerServer(ps service.PartnerService) systemrpc.PartnerServiceServer {
	return &partnerServer{ps}
}

func updatePartnerStatus(req *systempbv3.Partner, resp *systempbv3.Partner, err error) *systempbv3.Partner {
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

func (s *partnerServer) CreatePartner(ctx context.Context, req *systempbv3.Partner) (*systempbv3.Partner, error) {
	resp, err := s.Create(ctx, req)
	return updatePartnerStatus(req, resp, err), err
}
func (s *partnerServer) GetPartner(ctx context.Context, req *systempbv3.Partner) (*systempbv3.Partner, error) {
	resp, err := s.GetByName(ctx, req.Metadata.Name)
	return updatePartnerStatus(req, resp, err), err
}

func (s *partnerServer) DeletePartner(ctx context.Context, req *systempbv3.Partner) (*systempbv3.Partner, error) {
	resp, err := s.Delete(ctx, req)
	return updatePartnerStatus(req, resp, err), err
}

func (s *partnerServer) GetInitPartner(ctx context.Context, req *systemrpc.EmptyRequest) (*systempbv3.Partner, error) {

	partner, err := s.GetOnlyPartner(ctx)
	if err != nil {
		return nil, err
	}
	return partner, nil
}

func (s *partnerServer) UpdatePartner(ctx context.Context, req *systempbv3.Partner) (*systempbv3.Partner, error) {
	resp, err := s.Update(ctx, req)
	return updatePartnerStatus(req, resp, err), err
}
