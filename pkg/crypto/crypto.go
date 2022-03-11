package crypto

import (
	"crypto/aes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"

	"github.com/rs/xid"
)

func EncryptAES(key []byte, plaintext string) (string, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	out := make([]byte, len(plaintext))
	c.Encrypt(out, []byte(plaintext))
	return hex.EncodeToString(out), nil
}

func DecryptAES(key []byte, ct string) (string, error) {
	ciphertext, _ := hex.DecodeString(ct)
	c, err := aes.NewCipher(key)
	if err != nil {
		return "", nil
	}
	pt := make([]byte, len(ciphertext))
	c.Decrypt(pt, ciphertext)
	s := string(pt[:])
	return s, nil
}

func GenerateSha1Key() string {
	sum := sha1.Sum([]byte(xid.New().String()))
	return hex.EncodeToString(sum[:])
}

func GenerateSha256Secret() string {
	sum := sha256.Sum256([]byte(xid.New().String()))
	return hex.EncodeToString(sum[:])
}
