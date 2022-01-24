package credentials

import (
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"time"
)

// pem block types
const (
	CertificateRequestType = "CERTIFICATE REQUEST"
	CertificateType        = "CERTIFICATE"
	PrivateKeyType         = "PRIVATE KEY"
)

const (
	caCertName = "ca.pem"
	caKeyName  = "ca-key.pem"
)

// Signer signs cluster csr
type Signer interface {
	GetCACert() []byte
	Sign(csr []byte) ([]byte, error)
}

type signer struct {
	caCert    x509.Certificate
	caKey     crypto.PrivateKey
	caCertPem []byte
}

var _ Signer = (*signer)(nil)

// NewSigner returns new cluster credential signer
func NewSigner(path string) (Signer, error) {
	certPath := fmt.Sprintf("%s/%s", path, caCertName)
	keyPath := fmt.Sprintf("%s/%s", path, caKeyName)

	caPair, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	ca, err := x509.ParseCertificate(caPair.Certificate[0])
	if err != nil {
		return nil, err
	}

	f, err := os.Open(certPath)
	if err != nil {
		return nil, err
	}

	caCertBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return &signer{
		caCert:    *ca,
		caKey:     caPair.PrivateKey,
		caCertPem: caCertBytes,
	}, nil
}

func (s *signer) GetCACert() []byte {
	return s.caCertPem
}

func (s *signer) Sign(csr []byte) ([]byte, error) {

	block, _ := pem.Decode([]byte(csr))

	if block.Type != CertificateRequestType {
		return nil, errors.New("invalid type")
	}

	cr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{
		SerialNumber: getSerialNumber(),
		Issuer: pkix.Name{
			Country:      s.caCert.Subject.Country,
			Province:     s.caCert.Subject.Province,
			Locality:     s.caCert.Subject.Locality,
			Organization: s.caCert.Subject.Organization,
			CommonName:   s.caCert.Subject.CommonName,
		},
		Subject:     cr.Subject,
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0), // 10 years
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}

	// sign the certificate
	b, err := x509.CreateCertificate(rand.Reader, template, &s.caCert, cr.PublicKey, s.caKey)
	if err != nil {
		err = fmt.Errorf("unable to create certificate %s", err.Error())
		return nil, err
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: b,
	})
	if err != nil {
		return nil, err
	}

	return pemBytes, nil
}

func getSerialNumber() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)

	return serialNumber
}
