package cryptoutil

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
)

func getSerialNumber() *big.Int {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)

	return serialNumber
}

// GenerateCA Generates PEM encoded CA Certificate and Private key
// PasswordFunc should return non zero value to encrypt the PEM encoded private key
func GenerateCA(subject pkix.Name, f PasswordFunc) (cert, key []byte, err error) {
	var caCert *x509.Certificate
	var privKey *ecdsa.PrivateKey

	caCert = &x509.Certificate{
		SerialNumber:          getSerialNumber(),
		Subject:               subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	privKey, err = GenerateECDSAPrivateKey()
	if err != nil {
		return
	}

	var caBytes []byte

	caBytes, err = x509.CreateCertificate(rand.Reader, caCert, caCert, &privKey.PublicKey, privKey)
	if err != nil {
		return
	}

	cert = EncodeCert(caBytes)

	key, err = EncodePrivateKey(privKey, f)

	return
}
