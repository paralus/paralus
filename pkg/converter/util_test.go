package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sampleData struct {
	Name string `json:"name"`
}

func TestConvertToJsonRawMessage(t *testing.T) {
	t.Run("marshals valid data", func(t *testing.T) {
		raw := ConvertToJsonRawMessage(sampleData{Name: "foo"})
		assert.JSONEq(t, `{"name":"foo"}`, string(raw))
	})

	t.Run("returns empty bytes for unmarshalable data, logging the error", func(t *testing.T) {
		raw := ConvertToJsonRawMessage(func() {})
		assert.Empty(t, raw)
	})
}

// BUG: ConvertToObject does `json.Unmarshal(data, &dest)` where `dest` is
// already the interface{} parameter -- so it decodes into a fresh
// interface{} slot (producing a generic map[string]interface{}) rather
// than through the pointer the caller actually passed in. The caller's
// original struct is never populated, and the returned value is not the
// caller's type.
func TestConvertToObject(t *testing.T) {
	t.Run("BUG: caller's destination struct is never populated", func(t *testing.T) {
		dest := sampleData{Name: "unchanged"}
		_ = ConvertToObject([]byte(`{"name":"foo"}`), &dest)
		assert.Equal(t, "unchanged", dest.Name)
	})

	t.Run("BUG: return value is a generic map, not the caller's struct type", func(t *testing.T) {
		dest := sampleData{}
		result := ConvertToObject([]byte(`{"name":"foo"}`), &dest)
		m, ok := result.(map[string]interface{})
		require.True(t, ok, "expected a generic map, got %T", result)
		assert.Equal(t, "foo", m["name"])
	})

	t.Run("BUG: invalid JSON returns nil instead of the original destination", func(t *testing.T) {
		dest := sampleData{Name: "unchanged"}
		result := ConvertToObject([]byte(`not-json`), &dest)
		assert.Nil(t, result)
		assert.Equal(t, "unchanged", dest.Name)
	})
}
