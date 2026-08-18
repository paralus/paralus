package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptAES_RoundTrip(t *testing.T) {
	// EncryptAES/DecryptAES call cipher.Block.Encrypt/Decrypt directly, which
	// only operates on exactly one AES block (16 bytes) at a time. Valid
	// round trips therefore require plaintext of exactly aes.BlockSize.
	tests := []struct {
		name string
		key  []byte
	}{
		{name: "AES-128 key", key: []byte("0123456789abcdef")},
		{name: "AES-192 key", key: []byte("0123456789abcdef01234567")},
		{name: "AES-256 key", key: []byte("0123456789abcdef0123456789abcdef")},
	}

	const plaintext = "sixteen-byte-pt!"
	require.Len(t, plaintext, 16)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct, err := EncryptAES(tt.key, plaintext)
			require.NoError(t, err)
			assert.NotEmpty(t, ct)

			pt, err := DecryptAES(tt.key, ct)
			require.NoError(t, err)
			assert.Equal(t, plaintext, pt)
		})
	}
}

func TestEncryptAES_InvalidKeySize(t *testing.T) {
	_, err := EncryptAES([]byte("too-short"), "sixteen-byte-pt!")
	assert.Error(t, err)
}

func TestDecryptAES_InvalidKeySize(t *testing.T) {
	// Known existing behavior: aes.NewCipher's error is discarded and
	// DecryptAES returns ("", nil) instead of surfacing the error.
	pt, err := DecryptAES([]byte("too-short"), "00")
	assert.NoError(t, err)
	assert.Empty(t, pt)
}

func TestDecryptAES_MalformedCiphertextPanics(t *testing.T) {
	// Known existing behavior: hex.DecodeString's error is discarded, and a
	// ciphertext that decodes to fewer than aes.BlockSize bytes reaches
	// cipher.Block.Decrypt, which panics instead of returning an error.
	key := []byte("0123456789abcdef")
	assert.Panics(t, func() {
		_, _ = DecryptAES(key, "zz")
	})
}

func TestGenerateSha1Key(t *testing.T) {
	k1 := GenerateSha1Key()
	k2 := GenerateSha1Key()

	assert.Len(t, k1, 40) // sha1 = 20 bytes = 40 hex chars
	assert.Regexp(t, "^[0-9a-f]{40}$", k1)
	assert.NotEqual(t, k1, k2, "each call embeds a fresh xid, so keys must differ")
}

func TestGenerateSha256Secret(t *testing.T) {
	s1 := GenerateSha256Secret()
	s2 := GenerateSha256Secret()

	assert.Len(t, s1, 64) // sha256 = 32 bytes = 64 hex chars
	assert.Regexp(t, "^[0-9a-f]{64}$", s1)
	assert.NotEqual(t, s1, s2, "each call embeds a fresh xid, so secrets must differ")
}
