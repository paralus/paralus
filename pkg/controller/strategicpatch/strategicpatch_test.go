package strategicpatch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE ON SCOPE: patch.go (~2160 lines) is a vendored fork of the actual
// Kubernetes strategic-merge-patch library (see its copyright header) and
// is already extensively tested upstream in kubernetes/apimachinery.
// Re-deriving exhaustive coverage of its internal merge algorithm here
// would be low-value, low-signal work; these tests instead smoke-test the
// two public entry points this repo actually calls (via pkg/patch) plus
// the small amount of custom scaffolding in errors.go/meta.go.

func TestErrorTypes(t *testing.T) {
	t.Run("LookupPatchMetaError", func(t *testing.T) {
		err := LookupPatchMetaError{Path: "spec.foo", Err: assertErr("boom")}
		assert.Contains(t, err.Error(), "spec.foo")
		assert.Contains(t, err.Error(), "boom")
	})

	t.Run("FieldNotFoundError", func(t *testing.T) {
		err := FieldNotFoundError{Path: "spec", Field: "foo"}
		assert.Contains(t, err.Error(), "spec")
		assert.Contains(t, err.Error(), "foo")
	})

	t.Run("InvalidTypeError", func(t *testing.T) {
		err := InvalidTypeError{Path: "spec.foo", Expected: "string", Actual: "int"}
		assert.Contains(t, err.Error(), "string")
		assert.Contains(t, err.Error(), "int")
	})
}

type assertErr string

func (e assertErr) Error() string { return string(e) }

type mergeTestStruct struct {
	Conditions []mergeCondition `json:"conditions" patchStrategy:"merge" patchMergeKey:"type"`
	Name       string           `json:"name"`
}

type mergeCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

func TestPatchMetaFromStruct(t *testing.T) {
	t.Run("resolves merge strategy and key for a tagged slice field", func(t *testing.T) {
		meta, err := NewPatchMetaFromStruct(&mergeTestStruct{})
		require.NoError(t, err)

		_, patchMeta, err := meta.LookupPatchMetadataForStruct("conditions")
		require.NoError(t, err)
		assert.Equal(t, "type", patchMeta.GetPatchMergeKey())
		assert.Contains(t, patchMeta.GetPatchStrategies(), "merge")
	})

	t.Run("errors on an unrecognized field", func(t *testing.T) {
		meta, err := NewPatchMetaFromStruct(&mergeTestStruct{})
		require.NoError(t, err)

		_, _, err = meta.LookupPatchMetadataForStruct("doesNotExist")
		assert.Error(t, err)
	})

	t.Run("errors when dataStruct is not a struct", func(t *testing.T) {
		_, err := NewPatchMetaFromStruct("not a struct")
		assert.Error(t, err)
	})

	t.Run("errors when dataStruct is nil", func(t *testing.T) {
		_, err := NewPatchMetaFromStruct(nil)
		assert.Error(t, err)
	})
}

func TestCreateAndApplyTwoWayMergePatch(t *testing.T) {
	t.Run("merges a changed field and preserves an unrelated one", func(t *testing.T) {
		original := []byte(`{"name":"a","conditions":[{"type":"Ready","status":"False"}]}`)
		modified := []byte(`{"name":"a","conditions":[{"type":"Ready","status":"True"}]}`)

		patch, err := CreateTwoWayMergePatch(original, modified, &mergeTestStruct{})
		require.NoError(t, err)
		assert.NotEmpty(t, patch)

		result, err := StrategicMergePatch(original, patch, &mergeTestStruct{})
		require.NoError(t, err)
		assert.JSONEq(t, string(modified), string(result))
	})

	t.Run("merging an added condition by key preserves the existing one", func(t *testing.T) {
		existing := []byte(`{"conditions":[{"type":"Ready","status":"False"}]}`)
		current := []byte(`{"conditions":[{"type":"Ready","status":"False"},{"type":"Approved","status":"True"}]}`)

		patch, err := CreateTwoWayMergePatch(existing, current, &mergeTestStruct{})
		require.NoError(t, err)

		result, err := StrategicMergePatch(existing, patch, &mergeTestStruct{})
		require.NoError(t, err)
		assert.JSONEq(t, string(current), string(result))
	})

	t.Run("no differences yields an empty (or trivial) patch", func(t *testing.T) {
		same := []byte(`{"name":"a"}`)
		patch, err := CreateTwoWayMergePatch(same, same, &mergeTestStruct{})
		require.NoError(t, err)
		assert.JSONEq(t, "{}", string(patch))
	})

	t.Run("propagates a malformed-JSON error", func(t *testing.T) {
		_, err := CreateTwoWayMergePatch([]byte(`not-json`), []byte(`{}`), &mergeTestStruct{})
		assert.Error(t, err)
	})
}
