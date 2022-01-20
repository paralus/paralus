package service

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/internal/models"
	userv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
	"github.com/google/uuid"
	bun "github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/structpb"
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

func generateCallbackUrl() (string, error) {
	base, err := url.Parse(os.Getenv("APP_HOST_HTTP"))
	if err != nil {
		return "", err
	}
	uuid := uuid.New()
	return fmt.Sprintf("%s/auth/v3/sso/callback/%s", base, uuid), nil
}

func (s *oidcProvider) Create(ctx context.Context, provider *userv3.OIDCProvider) (*userv3.OIDCProvider, error) {
	// validate name
	name := provider.Metadata.GetName()
	if len(name) == 0 {
		return &userv3.OIDCProvider{}, fmt.Errorf("EMPTY NAME")
	}
	e := &models.OIDCProvider{}
	s.dao.GetByName(ctx, name, e)
	if e.Name == name {
		return &userv3.OIDCProvider{}, fmt.Errorf("DUPLICATE NAME")
	}

	callback, err := generateCallbackUrl()
	if err != nil {
		return &userv3.OIDCProvider{}, err
	}
	entity := &models.OIDCProvider{
		Name:            name,
		CreatedAt:       time.Time{},
		ModifiedAt:      time.Time{},
		ProviderName:    provider.Spec.GetProviderName(),
		MapperURL:       provider.Spec.GetMapperUrl(),
		MapperFilename:  provider.Spec.GetMapperFilename(),
		ClientId:        provider.Spec.GetClientId(),
		ClientSecret:    provider.Spec.GetClientSecret(),
		Scopes:          provider.Spec.GetScopes(),
		IssuerURL:       provider.Spec.GetIssuerUrl(),
		AuthURL:         provider.Spec.GetAuthUrl(),
		TokenURL:        provider.Spec.GetTokenUrl(),
		RequestedClaims: provider.Spec.GetRequestedClaims().AsMap(),
		Predefined:      provider.Spec.GetPredefined(),
		CallbackURL:     callback,
	}
	_, err = s.dao.Create(ctx, entity)
	if err != nil {
		return &userv3.OIDCProvider{}, err
	}

	rclaims, _ := structpb.NewStruct(entity.RequestedClaims)
	rv := &userv3.OIDCProvider{
		ApiVersion: "usermgmt.k8smgmt.io/v3",
		Kind:       "OIDCProvider",
		Metadata: &commonv3.Metadata{
			Name:        entity.Name,
			Description: entity.Description,
			Id:          entity.Id.String(),
		},
		Spec: &userv3.OIDCProviderSpec{
			ProviderName:    entity.ProviderName,
			MapperUrl:       entity.MapperURL,
			MapperFilename:  entity.MapperFilename,
			ClientId:        entity.ClientId,
			ClientSecret:    entity.ClientSecret,
			Scopes:          entity.Scopes,
			IssuerUrl:       entity.IssuerURL,
			AuthUrl:         entity.AuthURL,
			TokenUrl:        entity.TokenURL,
			RequestedClaims: rclaims,
			Predefined:      entity.Predefined,
			CallbackUrl:     entity.CallbackURL,
		},
	}
	return rv, nil
}

func (s *oidcProvider) GetByID(ctx context.Context, provider *userv3.OIDCProvider) (*userv3.OIDCProvider, error) {
	id, err := uuid.Parse(provider.Metadata.GetId())
	if err != nil {
		return &userv3.OIDCProvider{}, err
	}

	entity := &models.OIDCProvider{}
	_, err = s.dao.GetByID(ctx, id, entity)
	// TODO: Return proper error for Id not exist
	if err != nil {
		return &userv3.OIDCProvider{}, err
	}

	rclaims, _ := structpb.NewStruct(entity.RequestedClaims)
	rv := &userv3.OIDCProvider{
		ApiVersion: "usermgmt.k8smgmt.io/v3",
		Kind:       "OIDCProvider",
		Metadata: &commonv3.Metadata{
			Name:        entity.Name,
			Description: entity.Description,
			Id:          entity.Id.String(),
		},
		Spec: &userv3.OIDCProviderSpec{
			ProviderName:    entity.ProviderName,
			MapperUrl:       entity.MapperURL,
			MapperFilename:  entity.MapperFilename,
			ClientId:        entity.ClientId,
			Scopes:          entity.Scopes,
			IssuerUrl:       entity.IssuerURL,
			AuthUrl:         entity.AuthURL,
			TokenUrl:        entity.TokenURL,
			RequestedClaims: rclaims,
			Predefined:      entity.Predefined,
			CallbackUrl:     entity.CallbackURL,
		},
	}
	return rv, nil
}

func (s *oidcProvider) List(ctx context.Context) (*userv3.OIDCProviderList, error) {
	var (
		entities []models.OIDCProvider
		orgID    uuid.NullUUID
		parID    uuid.NullUUID
	)
	_, err := s.dao.List(ctx, parID, orgID, &entities)
	if err != nil {
		return &userv3.OIDCProviderList{}, nil
	}
	var result []*userv3.OIDCProvider
	for _, entity := range entities {
		rclaims, _ := structpb.NewStruct(entity.RequestedClaims)
		e := &userv3.OIDCProvider{
			ApiVersion: "usermgmt.k8smgmt.io/v3",
			Kind:       "OIDCProvider",
			Metadata: &commonv3.Metadata{
				Name:        entity.Name,
				Description: entity.Description,
				Id:          entity.Id.String(),
			},
			Spec: &userv3.OIDCProviderSpec{
				ProviderName:    entity.ProviderName,
				MapperUrl:       entity.MapperURL,
				MapperFilename:  entity.MapperFilename,
				ClientId:        entity.ClientId,
				Scopes:          entity.Scopes,
				IssuerUrl:       entity.IssuerURL,
				AuthUrl:         entity.AuthURL,
				TokenUrl:        entity.TokenURL,
				RequestedClaims: rclaims,
				Predefined:      entity.Predefined,
				CallbackUrl:     entity.CallbackURL,
			},
		}
		result = append(result, e)
	}

	rv := &userv3.OIDCProviderList{
		ApiVersion: "usermgmt.k8smgmt.io/v3",
		Kind:       "OIDCProviderList",
		Items:      result,
	}
	return rv, nil
}

func (s *oidcProvider) Update(ctx context.Context, provider *userv3.OIDCProvider) (*userv3.OIDCProvider, error) {
	return &userv3.OIDCProvider{}, nil
}

func (s *oidcProvider) Delete(ctx context.Context, provider *userv3.OIDCProvider) error {
	return nil
}
