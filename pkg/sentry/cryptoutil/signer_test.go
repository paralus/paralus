package cryptoutil

import (
	"crypto/x509/pkix"
	"testing"
)

func TestSigner(t *testing.T) {
	certBytes, keyBytes, err := GenerateCA(pkix.Name{
		CommonName:   "Paralus Sentry Bootstrap CA",
		Organization: []string{"Paralus"},
		Country:      []string{"USA"},
		Province:     []string{"California"},
		Locality:     []string{"Sunnyvale"},
	}, NoPassword)
	if err != nil {
		t.Error(err)
		return
	}

	signer, err := NewSigner(certBytes, keyBytes)
	if err != nil {
		t.Error(err)
		return
	}

	privKey, err := GenerateECDSAPrivateKey()
	if err != nil {
		t.Error(err)
		return
	}

	csr, err := CreateCSR(pkix.Name{
		CommonName: "test-token",
	}, privKey)

	if err != nil {
		t.Error(err)
		return
	}

	signed, err := signer.Sign(csr)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(string(signed))

}
