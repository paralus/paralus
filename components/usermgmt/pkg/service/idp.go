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
	"os"
	"time"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/internal/models"
	userv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

var baseUrl *url.URL

func init() {
	base, ok := os.LookupEnv("APP_HOST_HTTP")
	if !ok || len(base) == 0 {
		panic("APP_HOST_HTTP env not set")
	}
	var err error
	baseUrl, err = url.Parse(base)
	if err != nil {
		panic("Failed to get application url")
	}
}

type IdpService interface {
	Create(context.Context, *userv3.Idp) (*userv3.Idp, error)
	GetByID(context.Context, *userv3.Idp) (*userv3.Idp, error)
	List(context.Context) (*userv3.IdpList, error)
	Update(context.Context, *userv3.Idp) (*userv3.Idp, error)
	Delete(context.Context, *userv3.Idp) error
}

type idpService struct {
	dao pg.EntityDAO
}

func NewIdpService(db *bun.DB) IdpService {
	return &idpService{
		dao: pg.NewEntityDAO(db),
	}
}

func generateAcsURL() (string, error) {
	uuid := uuid.New()
	return fmt.Sprintf("%s/%s/", baseUrl.String(), uuid.String()), nil
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

func (s *idpService) Create(ctx context.Context, idp *userv3.Idp) (*userv3.Idp, error) {
	name := idp.Metadata.GetName()
	domain := idp.Spec.GetDomain()

	// validate name and domain
	if len(name) == 0 {
		return &userv3.Idp{}, fmt.Errorf("EMPTY NAME")
	}
	if len(domain) == 0 {
		return &userv3.Idp{}, fmt.Errorf("EMPTY DOMAIN")
	}
	e := &models.Idp{}
	s.dao.GetByName(ctx, name, e)
	if e.Name == name {
		return &userv3.Idp{}, fmt.Errorf("DUPLICATE NAME")
	}
	s.dao.GetX(ctx, "domain", domain, e)
	if e.Domain == domain {
		return &userv3.Idp{}, fmt.Errorf("DUPLICATE DOMAIN")
	}

	acsURL, err := generateAcsURL()
	if err != nil {
		return &userv3.Idp{}, err
	}
	entity := &models.Idp{
		Name:               name,
		Description:        idp.Metadata.GetDescription(),
		CreatedAt:          time.Now(),
		IdpName:            idp.Spec.GetIdpName(),
		Domain:             domain,
		AcsURL:             acsURL,
		SsoURL:             idp.Spec.GetSsoUrl(),
		IdpCert:            idp.Spec.GetIdpCert(),
		MetadataURL:        idp.Spec.GetMetadataUrl(),
		MetadataFilename:   idp.Spec.GetMetadataFilename(),
		GroupAttributeName: idp.Spec.GetGroupAttributeName(),
		SaeEnabled:         idp.Spec.GetSaeEnabled(),
	}
	if entity.SaeEnabled {
		spcert, spkey, err := generateSpCert(baseUrl.Host)
		if err != nil {
			return &userv3.Idp{}, err
		}
		entity.SpCert = spcert
		entity.SpKey = spkey
	}
	_, err = s.dao.Create(ctx, entity)
	if err != nil {
		return &userv3.Idp{}, err
	}

	rv := &userv3.Idp{
		ApiVersion: "usermgmt.k8smgmt.io/v3",
		Kind:       "Idp",
		Metadata: &commonv3.Metadata{
			Name:         entity.Name,
			Organization: entity.OrganizationId.String(),
			Partner:      entity.PartnerId.String(),
			Id:           entity.Id.String(),
		},
		Spec: &userv3.IdpSpec{
			IdpName:            entity.IdpName,
			Domain:             entity.Domain,
			AcsUrl:             entity.AcsURL,
			SsoUrl:             entity.SsoURL,
			IdpCert:            entity.IdpCert,
			SpCert:             entity.SpCert,
			MetadataUrl:        entity.MetadataURL,
			MetadataFilename:   entity.MetadataFilename,
			SaeEnabled:         entity.SaeEnabled,
			GroupAttributeName: entity.GroupAttributeName,
			NameIdFormat:       "Email Address",
			ConsumerBinding:    "HTTP-POST",
			SpEntityId:         entity.AcsURL,
		},
	}
	return rv, nil
}

func (s *idpService) GetByID(ctx context.Context, idp *userv3.Idp) (*userv3.Idp, error) {
	id, err := uuid.Parse(idp.Metadata.GetId())
	if err != nil {
		return &userv3.Idp{}, err
	}
	entity := &models.Idp{}
	// TODO: Check for existance of id before GetByID
	_, err = s.dao.GetByID(ctx, id, entity)
	if err != nil {
		return &userv3.Idp{}, err
	}
	rv := &userv3.Idp{
		ApiVersion: "usermgmt.k8smgmt.io/v3",
		Kind:       "Idp",
		Metadata: &commonv3.Metadata{
			Name:         entity.Name,
			Organization: entity.OrganizationId.String(),
			Partner:      entity.PartnerId.String(),
			Id:           entity.Id.String(),
		},
		Spec: &userv3.IdpSpec{
			IdpName:            entity.IdpName,
			Domain:             entity.Domain,
			AcsUrl:             entity.AcsURL,
			SsoUrl:             entity.SsoURL,
			IdpCert:            entity.IdpCert,
			SpCert:             entity.SpCert,
			MetadataUrl:        entity.MetadataURL,
			MetadataFilename:   entity.MetadataFilename,
			SaeEnabled:         entity.SaeEnabled,
			GroupAttributeName: entity.GroupAttributeName,
			NameIdFormat:       "Email Address",
			ConsumerBinding:    "HTTP-POST",
			SpEntityId:         entity.AcsURL,
		},
	}
	return rv, nil
}

func (s *idpService) Update(ctx context.Context, idp *userv3.Idp) (*userv3.Idp, error) {
	var id, orgId, partId uuid.UUID
	id, err := uuid.Parse(idp.Metadata.GetId())
	// TODO: 400 Bad Request
	if err != nil {
		return &userv3.Idp{}, err
	}
	if len(idp.Metadata.GetOrganization()) != 0 {
		orgId, err = uuid.Parse(idp.Metadata.GetOrganization())
		if err != nil {
			return &userv3.Idp{}, err
		}
	}
	if len(idp.Metadata.GetPartner()) != 0 {
		partId, err = uuid.Parse(idp.Metadata.GetPartner())
		if err != nil {
			return &userv3.Idp{}, err
		}
	}
	_, err = s.dao.GetByID(ctx, id, &models.Idp{})
	// TODO: Return proper error for Id not exist
	if err != nil {
		return &userv3.Idp{}, err
	}

	entity := &models.Idp{
		Id:                 id,
		Name:               idp.Metadata.GetName(),
		Description:        idp.Metadata.GetDescription(),
		ModifiedAt:         time.Now(),
		IdpName:            idp.Spec.GetIdpName(),
		Domain:             idp.Spec.GetDomain(),
		AcsURL:             idp.Spec.GetAcsUrl(),
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
		spcert, spkey, err := generateSpCert(baseUrl.Host)
		if err != nil {
			return &userv3.Idp{}, err
		}
		entity.SpCert = spcert
		entity.SpKey = spkey
	}

	_, err = s.dao.Update(ctx, id, entity)
	if err != nil {
		return &userv3.Idp{}, err
	}
	rv := &userv3.Idp{
		ApiVersion: "usermgmt.k8smgmt.io/v3",
		Kind:       "Idp",
		Metadata: &commonv3.Metadata{
			Name:         entity.Name,
			Organization: entity.OrganizationId.String(),
			Partner:      entity.PartnerId.String(),
			Id:           entity.Id.String(),
		},
		Spec: &userv3.IdpSpec{
			IdpName:            entity.IdpName,
			Domain:             entity.Domain,
			AcsUrl:             entity.AcsURL,
			SsoUrl:             entity.SsoURL,
			IdpCert:            entity.IdpCert,
			SpCert:             entity.SpCert,
			MetadataUrl:        entity.MetadataURL,
			MetadataFilename:   entity.MetadataFilename,
			SaeEnabled:         entity.SaeEnabled,
			GroupAttributeName: entity.GroupAttributeName,
			NameIdFormat:       "Email Address",
			ConsumerBinding:    "HTTP-POST",
			SpEntityId:         entity.AcsURL,
		},
	}
	return rv, nil
}

func (s *idpService) List(ctx context.Context) (*userv3.IdpList, error) {
	var (
		entities []models.Idp
		orgID    uuid.NullUUID
		parID    uuid.NullUUID
	)
	_, err := s.dao.List(ctx, parID, orgID, &entities)
	if err != nil {
		return &userv3.IdpList{}, err
	}

	// Get idps only till limit
	var result []*userv3.Idp
	for _, entity := range entities {
		e := &userv3.Idp{
			ApiVersion: "usermgmt.k8smgmt.io/v3",
			Kind:       "Idp",
			Metadata: &commonv3.Metadata{
				Name:         entity.Name,
				Organization: entity.OrganizationId.String(),
				Partner:      entity.PartnerId.String(),
				Id:           entity.Id.String(),
			},
			Spec: &userv3.IdpSpec{
				IdpName:            entity.IdpName,
				Domain:             entity.Domain,
				AcsUrl:             entity.AcsURL,
				SsoUrl:             entity.SsoURL,
				IdpCert:            entity.IdpCert,
				SpCert:             entity.SpCert,
				MetadataUrl:        entity.MetadataURL,
				MetadataFilename:   entity.MetadataFilename,
				SaeEnabled:         entity.SaeEnabled,
				GroupAttributeName: entity.GroupAttributeName,
				NameIdFormat:       "Email Address",
				ConsumerBinding:    "HTTP-POST",
				SpEntityId:         entity.AcsURL,
			},
		}
		result = append(result, e)
	}

	rv := &userv3.IdpList{
		ApiVersion: "usermgmt.k8smgmt.io/v3",
		Kind:       "IdpList",
		Items:      result,
	}
	return rv, nil
}

func (s *idpService) Delete(ctx context.Context, idp *userv3.Idp) error {
	id, err := uuid.Parse(idp.Metadata.GetId())
	if err != nil {
		return err
	}
	entity := &models.Idp{}
	_, err = s.dao.GetByID(ctx, id, entity)
	if entity.Id != id {
		return fmt.Errorf("ID DOES NOT EXISTS")
	}

	err = s.dao.Delete(ctx, id, &models.Idp{})
	if err != nil {
		return err
	}
	return nil
}
