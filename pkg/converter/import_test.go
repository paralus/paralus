package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ToRuntimeObject/ToStepTemplate/StepObjectToStepTemplate's happy paths
// depend on the full runtime scheme + fastjson-based Accessor machinery in
// pkg/controller/runtime and proto/types/controller; that combination
// wasn't practical to construct a faithful round-trip fixture for here
// without reverse-engineering the Accessor's exact expected byte layout for
// StepObject.Raw. These tests cover the straightforward, high-value error
// paths instead.
func TestToRuntimeObject_InvalidInput(t *testing.T) {
	gvk := schema.GroupVersionKind{Version: "v1", Kind: "Namespace"}

	_, err := ToRuntimeObject(gvk, []byte(`not-json`))
	assert.Error(t, err)
}

func TestToStepTemplate_NilObject(t *testing.T) {
	_, err := ToStepTemplate(nil)
	assert.Error(t, err)
}
