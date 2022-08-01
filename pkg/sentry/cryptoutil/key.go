package cryptoutil

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

const (
	ecKeyType  = "EC PRIVATE KEY"
	rsaKeyType = "RSA PRIVATE KEY"
)

// PasswordFunc is the signature for passing password while
// PEM encoding/decoding private keys
type PasswordFunc func() ([]byte, error)

// NoPassword should be used when the private key need not be encrypted
var NoPassword = func() ([]byte, error) {
	return nil, nil
}

// EncodePrivateKey PEM encodes private key
// when password is not empty private key is encrypted with password
func EncodePrivateKey(privKey crypto.PrivateKey, f PasswordFunc) ([]byte, error) {
	password, err := f()
	if err != nil {
		return nil, fmt.Errorf("unable to get password %s", err.Error())
	}

	switch privKey.(type) {
	case *ecdsa.PrivateKey:
		ecPrivKey := privKey.(*ecdsa.PrivateKey)
		b, err := x509.MarshalECPrivateKey(ecPrivKey)
		if err != nil {
			return nil, err
		}

		var p *pem.Block

		if len(password) == 0 {
			p = &pem.Block{
				Type:  ecKeyType,
				Bytes: b,
			}
		} else {
			p, err = x509.EncryptPEMBlock(rand.Reader, ecKeyType, b, password, x509.PEMCipherAES256)
			if err != nil {
				return nil, err
			}
		}
		return pem.EncodeToMemory(p), nil
	case *rsa.PrivateKey:
		rsaPrivKey := privKey.(*rsa.PrivateKey)
		b := x509.MarshalPKCS1PrivateKey(rsaPrivKey)

		var p *pem.Block

		if len(password) == 0 {
			p = &pem.Block{
				Type:  rsaKeyType,
				Bytes: b,
			}
		} else {
			p, err = x509.EncryptPEMBlock(rand.Reader, rsaKeyType, b, password, x509.PEMCipherAES256)
			if err != nil {
				return nil, err
			}
		}
		return pem.EncodeToMemory(p), nil
	default:
		return nil, fmt.Errorf("unsupported private key %T", privKey)

	}

}

// DecodePrivateKey decodes PEM encoded private key
// when PasswordFunc is provied private key is decrypted with password
func DecodePrivateKey(privKey []byte, f PasswordFunc) (crypto.PrivateKey, error) {
	p, err := decodePEM(privKey)
	if err != nil {
		return nil, err
	}

	password, err := f()
	if err != nil {
		return nil, err
	}

	var b []byte

	switch {
	case x509.IsEncryptedPEMBlock(p):
		b, err = x509.DecryptPEMBlock(p, password)
		if err != nil {
			fmt.Print("DecryptPEMBlock here is the error", err.Error())
			return nil, err
		}
	default:
		b = p.Bytes
	}

	switch p.Type {
	case ecKeyType:
		return x509.ParseECPrivateKey(b)
	case rsaKeyType:
		return x509.ParsePKCS1PrivateKey(b)
	default:
		return nil, fmt.Errorf("type %s is not suported", p.Type)
	}

}

// DecryptPrivateKeyAsPem returns a decrypted private key in PEM encoding
func DecryptPrivateKeyAsPem(privKey []byte, f PasswordFunc) ([]byte, error) {
	pk, err := DecodePrivateKey(privKey, f)
	if err != nil {
		return nil, err
	}

	return EncodePrivateKey(pk, NoPassword)
}

// GenerateECDSAPrivateKey generates new ECDSA private key
func GenerateECDSAPrivateKey() (*ecdsa.PrivateKey, error) {
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return ecKey, err

}
