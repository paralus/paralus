package cryptoutil

import (
	"crypto/x509/pkix"
	"testing"
)

func TestCreateCSR(t *testing.T) {
	privKey, err := GenerateECDSAPrivateKey()
	if err != nil {
		t.Error(err)
		return
	}

	csr, err := CreateCSR(pkix.Name{CommonName: "test-token"}, privKey)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(string(csr))
}
