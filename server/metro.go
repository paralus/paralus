package server

import (
	"context"

	"github.com/google/uuid"
	"github.com/paralus/paralus/pkg/service"
	systemrpc "github.com/paralus/paralus/proto/rpc/system"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
)

type locationServer struct {
	ms service.MetroService
}

// NewLocationServer returns new location server implementation
func NewLocationServer(ms service.MetroService) systemrpc.LocationServiceServer {
	return &locationServer{ms}
}

func (s *locationServer) CreateLocation(ctx context.Context, p *infrav3.Location) (*infrav3.Location, error) {
	p, err := s.ms.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *locationServer) GetLocation(ctx context.Context, p *infrav3.Location) (*infrav3.Location, error) {

	partner, err := s.ms.GetByName(ctx, p.Metadata.Name)
	if err != nil {
		id, err := uuid.Parse(p.Metadata.Id)
		if err != nil {
			return nil, err
		}
		partner, err = s.ms.GetById(ctx, id)
		if err != nil {
			return nil, err
		}
	}

	return partner, nil
}

func (s *locationServer) DeleteLocation(ctx context.Context, p *infrav3.Location) (*infrav3.Location, error) {
	partner, err := s.ms.Delete(ctx, p)
	if err != nil {
		return nil, err
	}
	return partner, nil
}

func (s *locationServer) UpdateLocation(ctx context.Context, p *infrav3.Location) (*infrav3.Location, error) {
	partner, err := s.ms.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return partner, nil
}

func (s *locationServer) GetLocations(ctx context.Context, m *infrav3.Location) (*infrav3.LocationList, error) {
	organizations, err := s.ms.List(ctx, m.Metadata.Partner)
	if err != nil {
		return nil, err
	}
	return organizations, nil
}
