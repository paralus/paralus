package cryptoutil

import (
	"reflect"
	"testing"
)

func TestEncryptDecryptPrivateKey(t *testing.T) {
	privKey, err := GenerateECDSAPrivateKey()

	if err != nil {
		t.Error(err)
		return
	}

	pf := func() ([]byte, error) {
		return []byte(`pass123`), nil
	}

	enc, err := EncodePrivateKey(privKey, pf)
	if err != nil {
		t.Error(err)
		return
	}

	privKey1, err := DecodePrivateKey(enc, pf)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(privKey1, privKey) {
		t.Error("expected same key")
	}

}
