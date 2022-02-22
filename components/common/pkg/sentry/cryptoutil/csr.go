package cryptoutil

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
)

const (
	csrType = "CERTIFICATE REQUEST"
)

// EncodeCSR encodes DER encoded CSR to PEM
func EncodeCSR(csr []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: csrType, Bytes: csr})
}

// DecodeCSR decodes PEM encoded CSR
func DecodeCSR(csr []byte) (cr *x509.CertificateRequest, err error) {
	var p *pem.Block
	p, err = decodePEM(csr)
	if err != nil {
		return nil, err
	}

	if p.Type != csrType {
		err = errors.New("invalid type")
		return
	}

	cr, err = x509.ParseCertificateRequest(p.Bytes)
	if err != nil {
		return
	}

	return
}

// CreateCSR creates csr for commonName
func CreateCSR(subject pkix.Name, privKey crypto.PrivateKey) ([]byte, error) {
	req := &x509.CertificateRequest{
		Subject: subject,
	}
	switch privKey.(type) {
	case *ecdsa.PrivateKey:
		req.SignatureAlgorithm = x509.ECDSAWithSHA256
	default:
		return nil, fmt.Errorf("unsupported private keys %T", privKey)
	}

	b, err := x509.CreateCertificateRequest(rand.Reader, req, privKey)
	if err != nil {
		return nil, err
	}

	return EncodeCSR(b), nil

}
