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
	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/internal/models"
	userv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/emptypb"
)

const TimeLayout = "2006-01-02T15:04:05.999999Z"

type IdpService interface {
	CreateIdp(context.Context, *userv3.NewIdp) (*userv3.Idp, error)
	UpdateIdp(context.Context, *userv3.UpdateIdp) (*userv3.Idp, error)
	GetSpConfigById(context.Context, *userv3.IdpID) (*userv3.SpConfig, error)
	ListIdps(context.Context, *userv3.ListIdpsRequest) (*userv3.ListIdpsResponse, error)
	DeleteIdp(context.Context, *userv3.IdpID) (*emptypb.Empty, error)
}

type idpService struct {
	dao pg.EntityDAO
}

func NewIdpService(db *bun.DB) IdpService {
	return &idpService{
		dao: pg.NewEntityDAO(db),
	}
}

func generateAcsURL(baseURL string) string {
	uuid := uuid.New()
	acsURL := fmt.Sprintf("%s/%s/", baseURL, uuid.String())
	return acsURL
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

func (s *idpService) CreateIdp(ctx context.Context, idp *userv3.NewIdp) (*userv3.Idp, error) {
	name := idp.GetName()
	domain := idp.GetDomain()

	e := &models.Idp{}
	s.dao.GetByName(ctx, name, e)
	if e.Name == name {
		return &userv3.Idp{}, fmt.Errorf("DUPLICATE NAME")
	}
	s.dao.GetX(ctx, "domain", domain, e)
	if e.Domain == domain {
		return &userv3.Idp{}, fmt.Errorf("DUPLICATE DOMAIN")
	}

	base, err := url.Parse(os.Getenv("APP_HOST_HTTP"))
	if err != nil {
		return &userv3.Idp{}, err
	}
	acsURL := generateAcsURL(base.String())
	entity := &models.Idp{
		Name:               name,
		IdpName:            idp.GetIdpName(),
		Domain:             domain,
		AcsURL:             acsURL,
		GroupAttributeName: idp.GetGroupAttributeName(),
		SaeEnabled:         idp.GetIsSaeEnabled(),
	}
	if entity.SaeEnabled {
		spcert, spkey, err := generateSpCert(base.Host)
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
		Id:                 entity.Id.String(),
		Name:               entity.Name,
		IdpName:            entity.IdpName,
		Domain:             entity.Domain,
		AcsUrl:             entity.AcsURL,
		SsoUrl:             entity.SsoURL,
		IdpCert:            entity.IdpCert,
		SpCert:             entity.SpCert,
		MetadataUrl:        entity.MetadataURL,
		MetadataFilename:   entity.MetadataFilename,
		IsSaeEnabled:       entity.SaeEnabled,
		GroupAttributeName: entity.GroupAttributeName,
		OrganizationId:     entity.OrganizationId,
		PartnerId:          entity.PartnerId,
		CreatedAt:          entity.CreatedAt.Format(TimeLayout),
		ModifiedAt:         entity.ModifiedAt.Format(TimeLayout),
	}
	return rv, nil
}

func (s *idpService) UpdateIdp(ctx context.Context, new *userv3.UpdateIdp) (*userv3.Idp, error) {
	id, err := uuid.Parse(new.GetId())
	if err != nil {
		return &userv3.Idp{}, err
	}
	entity := &models.Idp{
		Id:                 id,
		Name:               new.GetName(),
		ModifiedAt:         time.Now(),
		IdpName:            new.GetIdpName(),
		Domain:             new.GetDomain(),
		AcsURL:             new.GetAcsUrl(),
		MetadataURL:        new.GetMetadataUrl(),
		GroupAttributeName: new.GetGroupAttributeName(),
		SaeEnabled:         new.GetIsSaeEnabled(),
	}
	if entity.SaeEnabled {
		base, err := url.Parse(os.Getenv("APP_HOST_HTTP"))
		if err != nil {
			return &userv3.Idp{}, err
		}
		spcert, spkey, err := generateSpCert(base.Host)
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
		Id:                 entity.Id.String(),
		Name:               entity.Name,
		IdpName:            entity.IdpName,
		Domain:             entity.Domain,
		AcsUrl:             entity.AcsURL,
		SsoUrl:             entity.SsoURL,
		IdpCert:            entity.IdpCert,
		SpCert:             entity.SpCert,
		MetadataUrl:        entity.MetadataURL,
		MetadataFilename:   entity.MetadataFilename,
		IsSaeEnabled:       entity.SaeEnabled,
		GroupAttributeName: entity.GroupAttributeName,
		OrganizationId:     entity.OrganizationId,
		PartnerId:          entity.PartnerId,
		CreatedAt:          entity.CreatedAt.Format(TimeLayout),
		ModifiedAt:         entity.ModifiedAt.Format(TimeLayout),
	}
	return rv, nil
}

func (s *idpService) GetSpConfigById(ctx context.Context, idpID *userv3.IdpID) (*userv3.SpConfig, error) {
	id, err := uuid.Parse(idpID.GetId())
	if err != nil {
		return &userv3.SpConfig{}, err
	}

	entity := &models.Idp{}
	_, err = s.dao.GetByID(ctx, id, entity)
	if err != nil {
		return &userv3.SpConfig{}, err
	}
	if entity.Id != id {
		return &userv3.SpConfig{}, fmt.Errorf("IDP ID DOES NOT EXISTS")
	}
	rv := &userv3.SpConfig{
		NameidFormat:       "Email Address",
		ConsumerBinding:    "HTTP-POST",
		AcsUrl:             entity.AcsURL,
		EntityId:           entity.AcsURL,
		GroupAttributeName: entity.GroupAttributeName,
		SpCert:             entity.SpCert,
	}
	return rv, nil
}

func (s *idpService) ListIdps(ctx context.Context, req *userv3.ListIdpsRequest) (*userv3.ListIdpsResponse, error) {
	entities := []*models.Idp{}
	var orgID uuid.NullUUID
	var parID uuid.NullUUID
	s.dao.List(ctx, parID, orgID, entities)

	// Get idps only till limit
	var result []*userv3.Idp
	for _, entity := range entities {
		e := &userv3.Idp{
			Id:                 entity.Id.String(),
			Name:               entity.Name,
			IdpName:            entity.IdpName,
			Domain:             entity.Domain,
			AcsUrl:             entity.AcsURL,
			SsoUrl:             entity.SsoURL,
			IdpCert:            entity.IdpCert,
			SpCert:             entity.SpCert,
			MetadataUrl:        entity.MetadataURL,
			MetadataFilename:   entity.MetadataFilename,
			IsSaeEnabled:       entity.SaeEnabled,
			GroupAttributeName: entity.GroupAttributeName,
			OrganizationId:     entity.OrganizationId,
			PartnerId:          entity.PartnerId,
			CreatedAt:          entity.CreatedAt.Format(TimeLayout),
			ModifiedAt:         entity.ModifiedAt.Format(TimeLayout),
		}
		result = append(result, e)
	}

	rv := &userv3.ListIdpsResponse{
		Count:    int32(len(entities)),
		Next:     0,
		Previous: 0,
		Result:   result,
	}
	return rv, nil
}

func (s *idpService) DeleteIdp(ctx context.Context, idpID *userv3.IdpID) (*emptypb.Empty, error) {
	id, err := uuid.Parse(idpID.GetId())
	if err != nil {
		return &emptypb.Empty{}, err
	}

	entity := &models.Idp{}
	err = s.dao.Delete(ctx, id, entity)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}
