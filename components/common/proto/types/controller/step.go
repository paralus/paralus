package controller

//jsoniter "github.com/json-iterator/go"
import (
	kjson "sigs.k8s.io/kustomize/pseudo/k8s/apimachinery/pkg/runtime/serializer/json"
)

var (
	jsonAPI = kjson.CaseSensitiveJsonIterator()
)

// UnmarshalJSON converts json to object
func (so *StepObject) UnmarshalJSON(b []byte) error {
	so.Raw = b
	return nil
}

// MarshalJSON marshals step object into json
func (so *StepObject) MarshalJSON() ([]byte, error) {
	return so.Raw, nil
}

// Accessor returns accessor for the step object
func (so *StepObject) Accessor() (Accessor, error) {
	return newAccessor(so.Raw)
}

// GetStepOnFail returns on failed
func GetStepOnFail(t StepTemplate) StepOnFailed {
	if t.OnFailed == "" {
		return StepBreak
	}
	return StepOnFailed(t.OnFailed)
}
