package tail

import (
	"context"
	"crypto/x509/pkix"
	"fmt"
	"net/url"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/sentry/cryptoutil"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/sentry/register"
	sentryrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/sentry"
	"github.com/RafaySystems/rcloud-base/components/relay/pkg/utils"
)

// prepare config for outside relay boot strapping
func prepareConfigCSRForBootStrapOutSideCore(config *register.Config, CN string) error {
	privKey, err := cryptoutil.GenerateECDSAPrivateKey()
	if err != nil {
		_log.Infow("failed generate private key", "CN", CN)
		return err
	}

	key, err := cryptoutil.EncodePrivateKey(privKey, cryptoutil.NoPassword)
	if err != nil {
		_log.Infow("failed encode ecd private key", "cn", CN, "error", err)
		return err
	}

	config.PrivateKey = key

	csr, err := cryptoutil.CreateCSR(pkix.Name{
		CommonName:         CN,
		Country:            []string{"USA"},
		Organization:       []string{"Rafay Systems Inc"},
		OrganizationalUnit: []string{config.Name},
		Province:           []string{"California"},
		Locality:           []string{"Sunnyvale"},
	}, privKey)
	if err != nil {
		_log.Infow("failed to create CSR", "CN", CN, "error", err)
		return err
	}

	config.CSR = csr
	return nil
}

// registerRelayPeerService will register with  rafay-sentry-peering-client template token
// registration fetches client-certificate/root-ca to connect to sentry peer service
func registerRelayPeerService(ctx context.Context) (*register.Config, error) {
	cfg := &register.Config{
		ClientID: "tail-" + podName,
		Name:     "tail-" + podName,
		Mode:     "client",
	}

	if bootstrapAddr != "" {
		// outside core bootstrap
		cfg.Addr = bootstrapAddr
		if relayPeeringToken == "" {
			return nil, fmt.Errorf("empty peering token for bootstrap")
		}
		cfg.TemplateToken = relayPeeringToken
		cfg.Scheme = "http"
		// check port is 443 then set scheme as https
		if utils.IsHTTPS(bootstrapAddr) {
			cfg.Scheme = "https"
		}
		// this is a client certificate CN is same as ClientID
		err := prepareConfigCSRForBootStrapOutSideCore(cfg, cfg.ClientID)
		if err != nil {
			return nil, fmt.Errorf("failed in config csr for relay peering bootstrap")
		}
	} else {
		cfg.TemplateToken = "template/-"
		cfg.Addr = sentryAddr
		cfg.Scheme = "grpc"
		cfg.TemplateName = "rafay-sentry-peering-client"
	}

	if err := register.Register(ctx, cfg); err != nil {
		_log.Infow("unable to register", "error", err)
		return nil, err
	}

	_log.Infow("successfully registered", "cert", string(cfg.Certificate))

	return cfg, nil
}

func newSentryAuthorizationPool(ctx context.Context) (sentryrpc.SentryAuthorizationPool, error) {

	cfg, err := registerRelayPeerService(ctx)
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(relayPeeringURI)
	if err != nil {
		return nil, err
	}

	return sentryrpc.NewSentryAuthorizationPool(
		sentryrpc.WithAddr(url.Host),
		sentryrpc.WithClientCertPEM(cfg.Certificate),
		sentryrpc.WithClientKeyPEM(cfg.PrivateKey),
		sentryrpc.WithCaCertPEM(cfg.CACertificate),
	)
}
