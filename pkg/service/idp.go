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

	"github.com/RafaySystems/rcloud-base/internal/models"
	"github.com/RafaySystems/rcloud-base/internal/persistence/provider/pg"
	commonv3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
	systemv3 "github.com/RafaySystems/rcloud-base/proto/types/systempb/v3"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
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
	dao     pg.EntityDAO
	appHost string
}

func NewIdpService(db *bun.DB, hostUrl string) IdpService {
	return &idpService{
		dao:     pg.NewEntityDAO(db),
		appHost: hostUrl,
	}
}

func generateAcsURL(id string, hostUrl string) string {
	b, _ := url.Parse(hostUrl)
	return fmt.Sprintf("%s://%s/auth/v3/sso/acs/%s", b.Scheme, b.Host, id)
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
			Organization: []string{"Rafay"},
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

func (s *idpService) Create(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
	name := idp.Metadata.GetName()
	domain := idp.Spec.GetDomain()

	// validate name and domain
	if len(name) == 0 {
		return &systemv3.Idp{}, fmt.Errorf("EMPTY NAME")
	}
	if len(domain) == 0 {
		return &systemv3.Idp{}, fmt.Errorf("EMPTY DOMAIN")
	}
	e := &models.Idp{}
	s.dao.GetByName(ctx, name, e)
	if e.Name == name {
		return &systemv3.Idp{}, fmt.Errorf("DUPLICATE NAME")
	}
	s.dao.GetX(ctx, "domain", domain, e)
	if e.Domain == domain {
		return &systemv3.Idp{}, fmt.Errorf("DUPLICATE DOMAIN")
	}

	entity := &models.Idp{
		Name:               name,
		Description:        idp.Metadata.GetDescription(),
		CreatedAt:          time.Now(),
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
	_, err := s.dao.Create(ctx, entity)
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

func (s *idpService) GetByID(ctx context.Context, idp *systemv3.Idp) (*systemv3.Idp, error) {
	id, err := uuid.Parse(idp.Metadata.GetId())
	if err != nil {
		return &systemv3.Idp{}, err
	}
	entity := &models.Idp{}
	// TODO: Check for existance of id before GetByID
	_, err = s.dao.GetByID(ctx, id, entity)
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
	_, err := s.dao.GetByName(ctx, name, entity)
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

	_, err := s.dao.GetByName(ctx, name, existingIdp)
	if err != nil {
		// TODO: Handle both db and idp not exist errors
		// separately.
		return &systemv3.Idp{}, status.Errorf(codes.InvalidArgument, "IDP %q NOT EXIST", name)
	}

	s.dao.GetX(ctx, "domain", domain, existingIdp)
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

	_, err = s.dao.Update(ctx, existingIdp.Id, entity)
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

func (s *idpService) List(ctx context.Context) (*systemv3.IdpList, error) {
	var (
		entities []models.Idp
		orgID    uuid.NullUUID
		parID    uuid.NullUUID
	)
	_, err := s.dao.List(ctx, parID, orgID, &entities)
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

	_, err := s.dao.GetByName(ctx, name, entity)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "IDP %q NOT EXISTS", name)
	}

	err = s.dao.Delete(ctx, entity.Id, &models.Idp{})
	if err != nil {
		return err
	}
	return nil
}
