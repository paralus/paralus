package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	systemv3 "github.com/paralus/paralus/proto/types/systempb/v3"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IdpService interface {
	Create(context.Context, *systemv3.Idp) (*systemv3.Idp, error)
	GetByID(context.Context, *systemv3.Idp) (*systemv3.Idp, error)
	GetByName(context.Context, *systemv3.Idp) (*systemv3.Idp, error)
	List(context.Context) (*systemv3.IdpList, error)
	Update(context.Context, *systemv3.Idp) (*systemv3.Idp, error)
	Delete(context.Context, *systemv3.Idp) error
}

type idpService struct {
	db      *bun.DB
	appHost string
	al      *zap.Logger
}

func NewIdpService(db *bun.DB, hostUrl string, al *zap.Logger) IdpService {
	return &idpService{db: db, appHost: hostUrl, al: al}
}

func generateAcsURL(id string, hostUrl string) string {
	b, _ := url.Parse(hostUrl)
	return fmt.Sprintf("%s/auth/v3/sso/acs/%s", b.String(), id)
}

// generateSpCert generates self signed certificate. Returns cert and
// private key.
func generateSpCert(host string) (string, string, error) {
	// generate private key of type rsa
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", err
	}
	privPEM := new(bytes.Buffer)
	err = pem.Encode(privPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
	if err != nil {
		return "", "", err
	}
	privPEMBytes, err := ioutil.ReadAll(privPEM)
	if err != nil {
		return "", "", err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1000),
		Subject: pkix.Name{
			Organization: []string{"Paralus"},
			Country:      []string{"US"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(30, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		DNSNames:    []string{host},
	}
	// generate self sign certificate
	cBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", err
	}
	cPEM := new(bytes.Buffer)
	err = pem.Encode(cPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cBytes,
	})
	if err != nil {
		return "", "", err
	}
	cPEMBytes, err := ioutil.ReadAll(cPEM)
	if err != nil {
		return "", "", err
	}

	return string(cPEMBytes), string(privPEMBytes), nil
}

func (s *idpService) getPartnerOrganization(ctx context.Context, provider *systemv3.Idp) (uuid.UUID, uuid.UUID, error) {
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

func (s *idpService) Create(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
	name := idp.Metadata.GetName()
	domain := idp.Spec.GetDomain()

	// validate name and domain
	if len(name) == 0 {
		return &systemv3.Idp{}, fmt.Errorf("empty name for idp provider")
	}
	if len(domain) == 0 {
		return &systemv3.Idp{}, fmt.Errorf("empty domain for idp provider")
	}

	partnerId, organizationId, err := s.getPartnerOrganization(ctx, idp)
	if err != nil {
		return nil, fmt.Errorf("unable to get partner and org id")
	}
	i, _ := dao.GetIdByNamePartnerOrg(
		ctx,
		s.db,
		idp.GetMetadata().GetName(),
		uuid.NullUUID{UUID: partnerId, Valid: true},
		uuid.NullUUID{UUID: organizationId, Valid: true},
		&models.Idp{},
	)
	if i != nil {
		return nil, fmt.Errorf("idp %q already exists", idp.GetMetadata().GetName())
	}

	e := &models.Idp{}
	dao.GetX(ctx, s.db, "domain", domain, e)
	if e.Domain == domain {
		return &systemv3.Idp{}, fmt.Errorf("duplicate idp domain")
	}

	entity := &models.Idp{
		Name:               name,
		Description:        idp.Metadata.GetDescription(),
		CreatedAt:          time.Now(),
		PartnerId:          partnerId,
		OrganizationId:     organizationId,
		IdpName:            idp.Spec.GetIdpName(),
		Domain:             domain,
		SsoURL:             idp.Spec.GetSsoUrl(),
		IdpCert:            idp.Spec.GetIdpCert(),
		MetadataURL:        idp.Spec.GetMetadataUrl(),
		MetadataFilename:   idp.Spec.GetMetadataFilename(),
		GroupAttributeName: idp.Spec.GetGroupAttributeName(),
		SaeEnabled:         idp.Spec.GetSaeEnabled(),
	}
	if entity.SaeEnabled {
		baseURL, err := url.Parse(s.appHost)
		if err != nil {
			return &systemv3.Idp{}, err
		}
		spcert, spkey, err := generateSpCert(baseURL.Host)
		if err != nil {
			return &systemv3.Idp{}, err
		}
		entity.SpCert = spcert
		entity.SpKey = spkey
	}
	_, err = dao.Create(ctx, s.db, entity)
	if err != nil {
		return &systemv3.Idp{}, err
	}

	acsURL := generateAcsURL(entity.Id.String(), s.appHost)
	rv := &systemv3.Idp{
		ApiVersion: apiVersion,
		Kind:       "Idp",
		Metadata: &commonv3.Metadata{
			Name: entity.Name,
			Id:   entity.Id.String(),
		},
		Spec: &systemv3.IdpSpec{
			IdpName:            entity.IdpName,
			Domain:             entity.Domain,
			AcsUrl:             acsURL,
			SsoUrl:             entity.SsoURL,
			IdpCert:            entity.IdpCert,
			SpCert:             entity.SpCert,
			MetadataUrl:        entity.MetadataURL,
			MetadataFilename:   entity.MetadataFilename,
			SaeEnabled:         entity.SaeEnabled,
			GroupAttributeName: entity.GroupAttributeName,
			NameIdFormat:       "Email Address",
			ConsumerBinding:    "HTTP-POST",
			SpEntityId:         acsURL,
		},
	}

	CreateIdpAuditEvent(ctx, s.al, AuditActionCreate, rv.GetMetadata().GetName(), entity.Id)

	return rv, nil
}

func (s *idpService) GetByID(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
	id, err := uuid.Parse(idp.Metadata.GetId())
	if err != nil {
		return &systemv3.Idp{}, err
	}
	entity := &models.Idp{}
	// TODO: Check for existence of id before GetByID
	_, err = dao.GetByID(ctx, s.db, id, entity)
	if err != nil {
		return &systemv3.Idp{}, err
	}

	acsURL := generateAcsURL(entity.Id.String(), s.appHost)
	rv := &systemv3.Idp{
		ApiVersion: apiVersion,
		Kind:       "Idp",
		Metadata: &commonv3.Metadata{
			Name:         entity.Name,
			Organization: entity.OrganizationId.String(),
			Partner:      entity.PartnerId.String(),
			Id:           entity.Id.String(),
		},
		Spec: &systemv3.IdpSpec{
			IdpName:            entity.IdpName,
			Domain:             entity.Domain,
			AcsUrl:             acsURL,
			SsoUrl:             entity.SsoURL,
			IdpCert:            entity.IdpCert,
			SpCert:             entity.SpCert,
			MetadataUrl:        entity.MetadataURL,
			MetadataFilename:   entity.MetadataFilename,
			SaeEnabled:         entity.SaeEnabled,
			GroupAttributeName: entity.GroupAttributeName,
			NameIdFormat:       "Email Address",
			ConsumerBinding:    "HTTP-POST",
			SpEntityId:         acsURL,
		},
	}
	return rv, nil
}

func (s *idpService) GetByName(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
	name := idp.Metadata.GetName()
	if len(name) == 0 {
		// TODO: Write helper functions for the server and client error
		return &systemv3.Idp{}, status.Error(codes.InvalidArgument, "EMPTY NAME")
	}
	entity := &models.Idp{}
	_, err := dao.GetByName(ctx, s.db, name, entity)
	if err != nil {
		return &systemv3.Idp{}, err
	}

	acsURL := generateAcsURL(entity.Id.String(), s.appHost)
	rv := &systemv3.Idp{
		ApiVersion: apiVersion,
		Kind:       "Idp",
		Metadata: &commonv3.Metadata{
			Name:         entity.Name,
			Organization: entity.OrganizationId.String(),
			Partner:      entity.PartnerId.String(),
			Id:           entity.Id.String(),
		},
		Spec: &systemv3.IdpSpec{
			IdpName:            entity.IdpName,
			Domain:             entity.Domain,
			AcsUrl:             acsURL,
			SsoUrl:             entity.SsoURL,
			IdpCert:            entity.IdpCert,
			SpCert:             entity.SpCert,
			MetadataUrl:        entity.MetadataURL,
			MetadataFilename:   entity.MetadataFilename,
			SaeEnabled:         entity.SaeEnabled,
			GroupAttributeName: entity.GroupAttributeName,
			NameIdFormat:       "Email Address",
			ConsumerBinding:    "HTTP-POST",
			SpEntityId:         acsURL,
		},
	}
	return rv, nil
}

func (s *idpService) Update(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
	name := idp.Metadata.GetName()
	domain := idp.Spec.GetDomain()
	existingIdp := &models.Idp{}

	if len(name) == 0 {
		return &systemv3.Idp{}, status.Error(codes.InvalidArgument, "EMPTY NAME")
	}
	if len(domain) == 0 {
		return &systemv3.Idp{}, status.Error(codes.InvalidArgument, "EMPTY DOMAIN")
	}

	_, err := dao.GetByName(ctx, s.db, name, existingIdp)
	if err != nil {
		// TODO: Handle both db and idp not exist errors
		// separately.
		return &systemv3.Idp{}, status.Errorf(codes.InvalidArgument, "IDP %q NOT EXIST", name)
	}

	dao.GetX(ctx, s.db, "domain", domain, existingIdp)
	if existingIdp.Domain == domain {
		return &systemv3.Idp{}, status.Error(codes.InvalidArgument, "DUPLICATE DOMAIN")
	}

	orgId, err := uuid.Parse(idp.Metadata.GetOrganization())
	if err != nil {
		return &systemv3.Idp{}, status.Errorf(codes.InvalidArgument,
			"ORG ID %q INCORRECT", idp.Metadata.GetOrganization())
	}
	partId, err := uuid.Parse(idp.Metadata.GetPartner())
	if err != nil {
		return &systemv3.Idp{}, status.Errorf(codes.InvalidArgument,
			"PARTNER ID %q INCORRECT", idp.Metadata.GetPartner())
	}
	entity := &models.Idp{
		Name:               idp.Metadata.GetName(),
		Description:        idp.Metadata.GetDescription(),
		ModifiedAt:         time.Now(),
		IdpName:            idp.Spec.GetIdpName(),
		Domain:             idp.Spec.GetDomain(),
		OrganizationId:     orgId,
		PartnerId:          partId,
		SsoURL:             idp.Spec.GetSsoUrl(),
		IdpCert:            idp.Spec.GetIdpCert(),
		MetadataURL:        idp.Spec.GetMetadataUrl(),
		MetadataFilename:   idp.Spec.GetMetadataFilename(),
		GroupAttributeName: idp.Spec.GetGroupAttributeName(),
		SaeEnabled:         idp.Spec.GetSaeEnabled(),
	}
	if entity.SaeEnabled {
		baseURL, err := url.Parse(s.appHost)
		if err != nil {
			return &systemv3.Idp{}, err
		}
		spcert, spkey, err := generateSpCert(baseURL.Host)
		if err != nil {
			return &systemv3.Idp{}, err
		}
		entity.SpCert = spcert
		entity.SpKey = spkey
	}

	_, err = dao.Update(ctx, s.db, existingIdp.Id, entity)
	if err != nil {
		return &systemv3.Idp{}, err
	}

	acsURL := generateAcsURL(idp.GetMetadata().GetId(), s.appHost)
	rv := &systemv3.Idp{
		ApiVersion: apiVersion,
		Kind:       "Idp",
		Metadata: &commonv3.Metadata{
			Name: entity.Name,
			Id:   idp.GetMetadata().GetId(),
		},
		Spec: &systemv3.IdpSpec{
			IdpName:            entity.IdpName,
			Domain:             entity.Domain,
			AcsUrl:             acsURL,
			SsoUrl:             entity.SsoURL,
			IdpCert:            entity.IdpCert,
			SpCert:             entity.SpCert,
			MetadataUrl:        entity.MetadataURL,
			MetadataFilename:   entity.MetadataFilename,
			SaeEnabled:         entity.SaeEnabled,
			GroupAttributeName: entity.GroupAttributeName,
			NameIdFormat:       "Email Address",
			ConsumerBinding:    "HTTP-POST",
			SpEntityId:         acsURL,
		},
	}

	CreateIdpAuditEvent(ctx, s.al, AuditActionUpdate, rv.GetMetadata().GetName(), entity.Id)
	return rv, nil
}

func (s *idpService) List(ctx context.Context) (*systemv3.IdpList, error) {
	var (
		entities []models.Idp
		orgID    uuid.NullUUID
		parID    uuid.NullUUID
	)
	_, err := dao.List(ctx, s.db, parID, orgID, &entities)
	if err != nil {
		return &systemv3.IdpList{}, err
	}

	// Get idps only till limit
	var result []*systemv3.Idp
	for _, entity := range entities {
		acsURL := generateAcsURL(entity.Id.String(), s.appHost)
		e := &systemv3.Idp{
			ApiVersion: apiVersion,
			Kind:       "Idp",
			Metadata: &commonv3.Metadata{
				Name:         entity.Name,
				Organization: entity.OrganizationId.String(),
				Partner:      entity.PartnerId.String(),
				Id:           entity.Id.String(),
			},
			Spec: &systemv3.IdpSpec{
				IdpName:            entity.IdpName,
				Domain:             entity.Domain,
				AcsUrl:             acsURL,
				SsoUrl:             entity.SsoURL,
				IdpCert:            entity.IdpCert,
				SpCert:             entity.SpCert,
				MetadataUrl:        entity.MetadataURL,
				MetadataFilename:   entity.MetadataFilename,
				SaeEnabled:         entity.SaeEnabled,
				GroupAttributeName: entity.GroupAttributeName,
				NameIdFormat:       "Email Address",
				ConsumerBinding:    "HTTP-POST",
				SpEntityId:         acsURL,
			},
		}
		result = append(result, e)
	}

	rv := &systemv3.IdpList{
		ApiVersion: apiVersion,
		Kind:       "IdpList",
		Items:      result,
	}
	return rv, nil
}

func (s *idpService) Delete(ctx context.Context, idp *systemv3.Idp) error {
	entity := &models.Idp{}
	name := idp.Metadata.GetName()

	if len(name) == 0 {
		return status.Error(codes.InvalidArgument, "EMPTY NAME")
	}

	_, err := dao.GetByName(ctx, s.db, name, entity)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "IDP %q NOT EXISTS", name)
	}

	err = dao.Delete(ctx, s.db, entity.Id, &models.Idp{})
	if err != nil {
		return err
	}

	CreateIdpAuditEvent(ctx, s.al, AuditActionDelete, name, entity.Id)
	return nil
}
