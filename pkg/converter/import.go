package converter

import (
	"errors"
	"fmt"
	"strings"

	runtimeutil "github.com/paralus/paralus/pkg/controller/runtime"

	"github.com/paralus/paralus/pkg/log"
	controllerv2 "github.com/paralus/paralus/proto/types/controller"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sapijson "sigs.k8s.io/kustomize/pseudo/k8s/apimachinery/pkg/runtime/serializer/json"
)

var (
	// ErrInvalidObject is retuned for invalid object
	ErrInvalidObject = errors.New("object does not support object interface")
	json             = k8sapijson.CaseSensitiveJsonIterator()
)

var _log = log.GetLogger()

// toRuntimeObject converts JSON bytes into runtime object of
// latest version
func toRuntimeObject(gvk schema.GroupVersionKind, b []byte) (runtime.Object, error) {
	var sa controllerv2.StepObject

	err := json.Unmarshal(b, &sa)
	if err != nil {
		return nil, err
	}

	ro, _, err := runtimeutil.ToObject(&sa)
	if err != nil {
		return nil, err
	}

	return ro, nil

}

func toStepTemplate(o runtime.Object) (controllerv2.StepTemplate, error) {
	so, err := runtimeutil.FromObject(o)
	if err != nil {
		return controllerv2.StepTemplate{}, err
	}
	return stepObjectToStepTemplate(*so)
}

func stepObjectToStepTemplate(so controllerv2.StepObject) (controllerv2.StepTemplate, error) {
	var st controllerv2.StepTemplate

	accessor, err := so.Accessor()
	if err != nil {
		return st, err
	}

	gvk, err := accessor.GroupVersionKind()
	if err != nil {
		return st, err
	}

	name, err := accessor.Name()
	if err != nil {
		return st, err
	}

	st.Name = strings.ToLower(fmt.Sprintf("step-%s-%s", gvk.Kind, name))
	st.Object = &so

	return st, nil
}
