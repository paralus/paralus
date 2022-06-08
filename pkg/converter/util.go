package converter

import (
	gojson "encoding/json"

	apiv2 "github.com/paralus/paralus/proto/types/controller"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func ConvertToJsonRawMessage(data interface{}) gojson.RawMessage {
	bytes, err := json.Marshal(data)
	if err != nil {
		_log.Errorw("failed to marshal", "err", err, "data", data)
	}
	return gojson.RawMessage(bytes)
}

func ConvertToObject(data []byte, dest interface{}) interface{} {
	err := json.Unmarshal(data, &dest)
	if err != nil {
		_log.Errorw("failed to unmarshal", "err", err, "data", data)
	}
	return dest
}

// ToRuntimeObject converts JSON bytes into runtime object of latest version
func ToRuntimeObject(gvk schema.GroupVersionKind, b []byte) (runtime.Object, error) {
	return toRuntimeObject(gvk, b)
}

// ToStepTemplate converts runtime.Object to StepTemplate
func ToStepTemplate(o runtime.Object) (apiv2.StepTemplate, error) {
	return toStepTemplate(o)
}

// ToStepTemplate converts runtime.Object to StepTemplate
func StepObjectToStepTemplate(so apiv2.StepObject) (apiv2.StepTemplate, error) {
	return stepObjectToStepTemplate(so)
}
