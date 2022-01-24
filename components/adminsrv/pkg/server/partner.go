package server

import (
	"context"

	"github.com/RafaySystems/rcloud-base/components/adminsrv/pkg/service"
	rpcv3 "github.com/RafaySystems/rcloud-base/components/adminsrv/proto/rpc/v3"
	systempbv3 "github.com/RafaySystems/rcloud-base/components/adminsrv/proto/types/systempb/v3"
)

type partnerServer struct {
	service.PartnerService
}

// NewPartnerServer returns new partner server implementation
func NewPartnerServer(ps service.PartnerService) rpcv3.PartnerServer {
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

	partner, err := s.GetByName(ctx, p.Metadata.Name)
	if err != nil {
		partner, err = s.GetByID(ctx, p.Metadata.Id)
		if err != nil {
			return nil, err
		}
	}

	return partner, nil
}

func (s *partnerServer) DeletePartner(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error) {
	partner, err := s.Delete(ctx, p)
	if err != nil {
		return nil, err
	}
	return partner, nil
}

func (s *partnerServer) UpdatePartner(ctx context.Context, p *systempbv3.Partner) (*systempbv3.Partner, error) {
	partner, err := s.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return partner, nil
}
