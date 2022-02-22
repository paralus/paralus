package cryptoutil

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
)

const (
	certType = "CERTIFICATE"
)

// EncodeCert encodes the DER encoded cert to PEM
func EncodeCert(cert []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  certType,
		Bytes: cert,
	})
}

// DecodeCert decodes PEM encoded cert
func DecodeCert(cert []byte) (c *x509.Certificate, err error) {
	var p *pem.Block
	p, err = decodePEM(cert)
	if err != nil {
		return
	}

	if p.Type != certType {
		err = errors.New("invalid pem type")
		return
	}

	c, err = x509.ParseCertificate(p.Bytes)

	return
}
