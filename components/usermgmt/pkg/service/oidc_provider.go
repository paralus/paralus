package service

import (
	"context"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	userv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
	bun "github.com/uptrace/bun"
)

type OIDCProviderService interface {
	Create(context.Context, *userv3.OIDCProvider) (*userv3.OIDCProvider, error)
	GetByID(context.Context, *userv3.OIDCProvider) (*userv3.OIDCProvider, error)
	List(context.Context) (*userv3.OIDCProviderList, error)
	Update(context.Context, *userv3.OIDCProvider) (*userv3.OIDCProvider, error)
	Delete(context.Context, *userv3.OIDCProvider) error
}

type oidcProvider struct {
	dao pg.EntityDAO
}

func NewOIDCProviderService(db *bun.DB) OIDCProviderService {
	return &oidcProvider{
		dao: pg.NewEntityDAO(db),
	}
}

func (s *oidcProvider) Create(ctx context.Context, provider *userv3.OIDCProvider) (*userv3.OIDCProvider, error) {
	return &userv3.OIDCProvider{}, nil
}

func (s *oidcProvider) GetByID(ctx context.Context, provider *userv3.OIDCProvider) (*userv3.OIDCProvider, error) {
	return &userv3.OIDCProvider{}, nil
}

func (s *oidcProvider) List(ctx context.Context) (*userv3.OIDCProviderList, error) {
	return &userv3.OIDCProviderList{}, nil
}

func (s *oidcProvider) Update(ctx context.Context, provider *userv3.OIDCProvider) (*userv3.OIDCProvider, error) {
	return &userv3.OIDCProvider{}, nil
}

func (s *oidcProvider) Delete(ctx context.Context, provider *userv3.OIDCProvider) error {
	return nil
}
