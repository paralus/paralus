package rpcv3

import (
	"context"

	"github.com/RafaySystems/rcloud-base/components/adminsrv/pkg/service"
	systempbv3 "github.com/RafaySystems/rcloud-base/components/adminsrv/proto/types/systempb/v3"
)

type partnerServer struct {
	service.PartnerService
}

// NewPartnerServer returns new placement server implementation
func NewPartnerServer(ps service.PartnerService) PartnerServer {
	return &partnerServer{ps}
}

func (s *partnerServer) CreatePartner(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error) {
	p, err := s.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}
func (s *partnerServer) GetPartner(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error) {

	partner, err := s.GetByID(ctx, p.Metadata.Id)
	if err != nil {
		return nil, err
	}

	return partner, nil
}

func (s *partnerServer) DeletePartner(context.Context, *systempbv3.Partner) (*systempbv3.Partner, error) {
	return nil, nil
}

func (s *partnerServer) UpdatePartner(context.Context, *systempbv3.Partner) (*systempbv3.Partner, error) {
	return nil, nil
}
