package cryptoutil

import (
	"crypto/x509/pkix"
	"testing"
)

func TestGenerateCA(t *testing.T) {
	cert, key, err := GenerateCA(pkix.Name{
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
	t.Log(string(cert))
	t.Log(string(key))
}
