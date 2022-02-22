package util

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509/pkix"
	"database/sql"
	"fmt"
	"time"

	"github.com/RafaySystems/rcloud-base/components/adminsrv/pkg/service"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/query"
	"github.com/rs/xid"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/sentry/cryptoutil"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/sentry/register"
	"github.com/RafaySystems/rcloud-base/components/common/proto/types/sentry"
)

// GetPeeringServerCreds returns peering cert, key, ca
func GetPeeringServerCreds(ctx context.Context, bs service.BootstrapService, rpcPort int, host string) (cert, key, ca []byte, err error) {
	nctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	var template *sentry.BootstrapAgentTemplate

	template, err = bs.GetBootstrapAgentTemplate(nctx, "rafay-sentry-peering-server")
	if err != nil {
		return
	}

	config := &register.Config{
		TemplateName: "rafay-sentry-peering-server",
		Addr:         fmt.Sprintf("localhost:%d", rpcPort),
		Name:         "rafay-sentry-peering-server",
		Scheme:       "grpc",
		Mode:         "server",
	}

	var privKey *ecdsa.PrivateKey

	privKey, err = cryptoutil.GenerateECDSAPrivateKey()
	if err != nil {
		return
	}

	config.PrivateKey, err = cryptoutil.EncodePrivateKey(privKey, cryptoutil.NoPassword)
	if err != nil {
		return
	}

	var csr []byte

	csr, err = cryptoutil.CreateCSR(pkix.Name{
		CommonName:         host,
		Country:            []string{"USA"},
		Organization:       []string{"Rafay Systems Inc"},
		OrganizationalUnit: []string{"Rafay Sentry Peering Server"},
		Province:           []string{"California"},
		Locality:           []string{"Sunnyvale"},
	}, privKey)
	if err != nil {
		return
	}

	config.CSR = csr

	var agent *sentry.BootstrapAgent

	agent, err = bs.GetBootstrapAgent(nctx, template.Metadata.Name, query.WithName("rafay-sentry-peering-server"), query.WithGlobalScope())

	if err != nil {
		if err != sql.ErrNoRows {
			return
		}
	}

	if agent != nil {
		config.ClientID = agent.Spec.Token
	} else {
		config.ClientID = xid.New().String()
	}

	err = register.Register(nctx, config)
	if err != nil {
		return
	}

	cert = config.Certificate
	key = config.PrivateKey
	ca = config.CACertificate

	return
}
