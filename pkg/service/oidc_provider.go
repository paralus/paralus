package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	systemv3 "github.com/paralus/paralus/proto/types/systempb/v3"
	bun "github.com/uptrace/bun"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type OIDCProviderService interface {
	Create(context.Context, *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error)
	GetByID(context.Context, *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error)
	GetByName(context.Context, *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error)
	List(context.Context) (*systemv3.OIDCProviderList, error)
	Update(context.Context, *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error)
	Delete(context.Context, *systemv3.OIDCProvider) error
}

type oidcProvider struct {
	db        *bun.DB
	kratosUrl string
	al        *zap.Logger
}

func NewOIDCProviderService(db *bun.DB, kratosUrl string, al *zap.Logger) OIDCProviderService {
	return &oidcProvider{db: db, kratosUrl: kratosUrl, al: al}
}

func generateCallbackUrl(id string, kUrl string) string {
	scheme := "http"
	host, port, err := net.SplitHostPort(kUrl)
	if err == nil {
		if port == "443" {
			scheme = "https"
		}
	}
	return fmt.Sprintf("%s://%s/self-service/methods/oidc/callback/%s", scheme, host, id)
}

func validateURL(rawURL string) bool {
	var valid bool
	pfx := []string{
		"http://",
		"https://",
		"base64://",
	}
	for _, p := range pfx {
		if strings.HasPrefix(rawURL, p) {
			valid = true
		}
	}
	return valid
}

func (s *oidcProvider) getPartnerOrganization(ctx context.Context, provider *systemv3.OIDCProvider) (uuid.UUID, uuid.UUID, error) {
	partner := provider.GetMetadata().GetPartner()
	org := provider.GetMetadata().GetOrganization()
	partnerId, err := dao.GetPartnerId(ctx, s.db, partner)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	organizationId, err := dao.GetOrganizationId(ctx, s.db, org)
	if err != nil {
		return partnerId, uuid.Nil, err
	}
	return partnerId, organizationId, nil
}

func (s *oidcProvider) Create(ctx context.Context, provider *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
	name := provider.GetMetadata().GetName()
	if len(name) == 0 {
		return &systemv3.OIDCProvider{}, fmt.Errorf("empty name for provider")
	}
	scopes := provider.GetSpec().GetScopes()
	if len(scopes) == 0 {
		return &systemv3.OIDCProvider{}, fmt.Errorf("no scopes present")
	}
	issUrl := provider.GetSpec().GetIssuerUrl()
	if len(issUrl) == 0 {
		return &systemv3.OIDCProvider{}, fmt.Errorf("empty issuer url")
	}

	partnerId, organizationId, err := s.getPartnerOrganization(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("unable to get partner and org id")
	}
	p, _ := dao.GetIdByNamePartnerOrg(
		ctx,
		s.db,
		name,
		uuid.NullUUID{UUID: partnerId, Valid: true},
		uuid.NullUUID{UUID: organizationId, Valid: true},
		&models.OIDCProvider{},
	)
	if p != nil {
		return nil, fmt.Errorf("provider %q already exists", name)
	}

	p, _ = dao.GetM(ctx, s.db, map[string]interface{}{
		"issuer_url":      issUrl,
		"partner_id":      partnerId,
		"organization_id": organizationId,
		"trash":           false,
	}, &models.OIDCProvider{})
	if p != nil {
		return nil, fmt.Errorf("duplicate issuer url")
	}
	if !validateURL(issUrl) {
		return &systemv3.OIDCProvider{}, fmt.Errorf("invalid issuer url")
	}

	mapUrl := provider.Spec.GetMapperUrl()
	authUrl := provider.Spec.GetAuthUrl()
	tknUrl := provider.Spec.GetTokenUrl()

	if len(mapUrl) != 0 && !validateURL(mapUrl) {
		return &systemv3.OIDCProvider{}, fmt.Errorf("invalid mapper url")
	}
	if len(authUrl) != 0 && !validateURL(authUrl) {
		return &systemv3.OIDCProvider{}, fmt.Errorf("invalid auth url")
	}
	if len(tknUrl) != 0 && !validateURL(tknUrl) {
		return &systemv3.OIDCProvider{}, fmt.Errorf("invalid token url")
	}

	entity := &models.OIDCProvider{
		Name:            name,
		Description:     provider.GetMetadata().GetDescription(),
		CreatedAt:       time.Time{},
		ModifiedAt:      time.Time{},
		PartnerId:       partnerId,
		OrganizationId:  organizationId,
		ProviderName:    provider.Spec.GetProviderName(),
		MapperURL:       mapUrl,
		MapperFilename:  provider.Spec.GetMapperFilename(),
		ClientId:        provider.Spec.GetClientId(),
		ClientSecret:    provider.Spec.GetClientSecret(),
		Scopes:          provider.Spec.GetScopes(),
		IssuerURL:       issUrl,
		AuthURL:         authUrl,
		TokenURL:        tknUrl,
		RequestedClaims: provider.Spec.GetRequestedClaims().AsMap(),
		Predefined:      provider.Spec.GetPredefined(),
	}
	_, err = dao.Create(ctx, s.db, entity)
	if err != nil {
		return &systemv3.OIDCProvider{}, err
	}

	rclaims, _ := structpb.NewStruct(entity.RequestedClaims)
	rv := &systemv3.OIDCProvider{
		ApiVersion: apiVersion,
		Kind:       "OIDCProvider",
		Metadata: &commonv3.Metadata{
			Name:        entity.Name,
			Description: entity.Description,
			Id:          entity.Id.String(),
		},
		Spec: &systemv3.OIDCProviderSpec{
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
			CallbackUrl:     generateCallbackUrl(entity.Name, s.kratosUrl),
		},
	}

	CreateOidcAuditEvent(ctx, s.al, AuditActionCreate, rv.GetMetadata().GetName(), entity.Id)
	return rv, nil
}

func (s *oidcProvider) GetByID(ctx context.Context, provider *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
	id, err := uuid.Parse(provider.Metadata.GetId())
	if err != nil {
		return &systemv3.OIDCProvider{}, err
	}

	entity := &models.OIDCProvider{}
	_, err = dao.GetByID(ctx, s.db, id, entity)
	// TODO: Return proper error for Id not exist
	if err != nil {
		return &systemv3.OIDCProvider{}, err
	}

	rclaims, _ := structpb.NewStruct(entity.RequestedClaims)
	rv := &systemv3.OIDCProvider{
		ApiVersion: apiVersion,
		Kind:       "OIDCProvider",
		Metadata: &commonv3.Metadata{
			Name:        entity.Name,
			Description: entity.Description,
			Id:          entity.Id.String(),
		},
		Spec: &systemv3.OIDCProviderSpec{
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
			CallbackUrl:     generateCallbackUrl(entity.Name, s.kratosUrl),
		},
	}
	return rv, nil
}

func (s *oidcProvider) GetByName(ctx context.Context, provider *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
	name := provider.Metadata.GetName()
	if len(name) == 0 {
		return &systemv3.OIDCProvider{}, status.Error(codes.InvalidArgument, "EMPTY NAME")
	}

	entity := &models.OIDCProvider{}
	_, err := dao.GetByName(ctx, s.db, name, entity)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &systemv3.OIDCProvider{}, status.Errorf(codes.InvalidArgument, "OIDC PROVIDER %q NOT EXIST", name)
		} else {
			return &systemv3.OIDCProvider{}, status.Errorf(codes.Internal, codes.Internal.String())
		}

	}

	rclaims, _ := structpb.NewStruct(entity.RequestedClaims)
	rv := &systemv3.OIDCProvider{
		ApiVersion: apiVersion,
		Kind:       "OIDCProvider",
		Metadata: &commonv3.Metadata{
			Name:         entity.Name,
			Description:  entity.Description,
			Id:           entity.Id.String(),
			Organization: entity.OrganizationId.String(),
			Partner:      entity.PartnerId.String(),
		},
		Spec: &systemv3.OIDCProviderSpec{
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
			CallbackUrl:     generateCallbackUrl(entity.Name, s.kratosUrl),
		},
	}
	return rv, nil
}

func (s *oidcProvider) List(ctx context.Context) (*systemv3.OIDCProviderList, error) {
	var (
		entities []models.OIDCProvider
		orgID    uuid.NullUUID
		parID    uuid.NullUUID
	)
	_, err := dao.List(ctx, s.db, parID, orgID, &entities)
	if err != nil {
		return &systemv3.OIDCProviderList{}, nil
	}
	var result []*systemv3.OIDCProvider
	for _, entity := range entities {
		rclaims, _ := structpb.NewStruct(entity.RequestedClaims)
		e := &systemv3.OIDCProvider{
			ApiVersion: apiVersion,
			Kind:       "OIDCProvider",
			Metadata: &commonv3.Metadata{
				Name:        entity.Name,
				Description: entity.Description,
				Id:          entity.Id.String(),
			},
			Spec: &systemv3.OIDCProviderSpec{
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
				CallbackUrl:     generateCallbackUrl(entity.Name, s.kratosUrl),
			},
		}
		result = append(result, e)
	}

	rv := &systemv3.OIDCProviderList{
		ApiVersion: "usermgmt.k8smgmt.io/v3",
		Kind:       "OIDCProviderList",
		Items:      result,
	}
	return rv, nil
}

func (s *oidcProvider) Update(ctx context.Context, provider *systemv3.OIDCProvider) (*systemv3.OIDCProvider, error) {
	name := provider.GetMetadata().GetName()
	if len(name) == 0 {
		return &systemv3.OIDCProvider{}, status.Error(codes.InvalidArgument, "empty name")
	}
	scopes := provider.GetSpec().GetScopes()
	if len(scopes) == 0 {
		return &systemv3.OIDCProvider{}, fmt.Errorf("no scopes")
	}
	issUrl := provider.GetSpec().GetIssuerUrl()
	if len(issUrl) == 0 {
		return &systemv3.OIDCProvider{}, fmt.Errorf("empty issuer url")
	}

	partnerId, organizationId, err := s.getPartnerOrganization(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("unable to get partner and org id")
	}

	existingP := &models.OIDCProvider{}
	_, err = dao.GetByName(ctx, s.db, name, existingP)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &systemv3.OIDCProvider{}, status.Errorf(codes.InvalidArgument, "oidc provider %q not exist", name)
		} else {
			return &systemv3.OIDCProvider{}, status.Error(codes.Internal, codes.Internal.String())
		}
	}

	mapUrl := provider.Spec.GetMapperUrl()
	authUrl := provider.Spec.GetAuthUrl()
	tknUrl := provider.Spec.GetTokenUrl()

	if !validateURL(issUrl) {
		return &systemv3.OIDCProvider{}, fmt.Errorf("invalid issuer url")
	}
	if len(mapUrl) != 0 && !validateURL(mapUrl) {
		return &systemv3.OIDCProvider{}, fmt.Errorf("invalid mapper url")
	}
	if len(authUrl) != 0 && !validateURL(authUrl) {
		return &systemv3.OIDCProvider{}, fmt.Errorf("invalid auth url")
	}
	if len(tknUrl) != 0 && !validateURL(tknUrl) {
		return &systemv3.OIDCProvider{}, fmt.Errorf("invalid token url")
	}

	entity := &models.OIDCProvider{
		Name:            provider.Metadata.GetName(),
		Description:     provider.Metadata.GetDescription(),
		OrganizationId:  organizationId,
		PartnerId:       partnerId,
		ModifiedAt:      time.Now(),
		ProviderName:    provider.Spec.GetProviderName(),
		MapperURL:       mapUrl,
		MapperFilename:  provider.Spec.GetMapperFilename(),
		ClientId:        provider.Spec.GetClientId(),
		ClientSecret:    provider.Spec.GetClientSecret(),
		Scopes:          provider.Spec.GetScopes(),
		IssuerURL:       issUrl,
		AuthURL:         authUrl,
		TokenURL:        tknUrl,
		RequestedClaims: provider.Spec.GetRequestedClaims().AsMap(),
		Predefined:      provider.Spec.GetPredefined(),
	}
	_, err = dao.Update(ctx, s.db, existingP.Id, entity)
	if err != nil {
		_log.Errorf("Unable to create oidc provider: %s", err)
		// TODO: catch already existing issuer url and return exact error
		return &systemv3.OIDCProvider{}, fmt.Errorf("unable to create oidc provider")
	}

	rclaims, _ := structpb.NewStruct(entity.RequestedClaims)
	rv := &systemv3.OIDCProvider{
		ApiVersion: apiVersion,
		Kind:       "OIDCProvider",
		Metadata: &commonv3.Metadata{
			Name:        entity.Name,
			Description: entity.Description,
			Id:          provider.GetMetadata().GetId(),
		},
		Spec: &systemv3.OIDCProviderSpec{
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
			CallbackUrl:     generateCallbackUrl(provider.GetMetadata().GetName(), s.kratosUrl),
		},
	}

	CreateOidcAuditEvent(ctx, s.al, AuditActionUpdate, rv.GetMetadata().GetName(), entity.Id)
	return rv, nil
}

func (s *oidcProvider) Delete(ctx context.Context, provider *systemv3.OIDCProvider) error {
	entity := &models.OIDCProvider{}
	name := provider.GetMetadata().GetName()
	if len(name) == 0 {
		return status.Error(codes.InvalidArgument, "EMPTY NAME")
	}
	_, err := dao.GetByName(ctx, s.db, name, entity)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "OIDC PROVIDER %q NOT EXIST", name)
	}

	err = dao.Delete(ctx, s.db, entity.Id, &models.OIDCProvider{})
	if err != nil {
		return err
	}

	CreateOidcAuditEvent(ctx, s.al, AuditActionDelete, name, entity.Id)
	return nil
}
