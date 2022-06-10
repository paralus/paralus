package converter

import (
	"errors"
	"fmt"
	"strings"

	runtimeutil "github.com/paralus/paralus/pkg/controller/runtime"

	"github.com/paralus/paralus/pkg/log"
	controllerv2 "github.com/paralus/paralus/proto/types/controller"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	k8sapijson "sigs.k8s.io/kustomize/pseudo/k8s/apimachinery/pkg/runtime/serializer/json"
)

const (
	delimiter                         = "---"
	ingressAnnotationConfigSnippetKey = "nginx.ingress.kubernetes.io/configuration-snippet"
)

var (
	// ErrInvalidObject is retuned for invalid object
	ErrInvalidObject = errors.New("object does not support object interface")
	json             = k8sapijson.CaseSensitiveJsonIterator()
	dmf              = kjson.DefaultMetaFactory
)

var _log = log.GetLogger()

func getIngressAnnotations(name string, orgID, partnerID string) map[string]string {
	return map[string]string{
		ingressAnnotationConfigSnippetKey: fmt.Sprintf("set $workload_name \"%s\";set $orgId \"%s\";set $partnerId \"%s\";", name, orgID, partnerID),
	}
}

func addIngressAnnotations(annotations map[string]string, name string, orgId, partnerId string) {
	if _, ok := annotations[ingressAnnotationConfigSnippetKey]; !ok {
		annotations[ingressAnnotationConfigSnippetKey] = fmt.Sprintf("set $workload_name \"%s\";set $orgId \"%s\";set $partnerId \"%s\";",
			name, orgId, partnerId)
	}
}

func addDebugLabels(stepTemplate *controllerv2.StepTemplate, debugLabels []byte) error {

	if stepTemplate.Object != nil {
		accessor, err := stepTemplate.Object.Accessor()
		if err != nil {
			return err
		}
		kind, err := accessor.Kind()
		if err != nil {
			return err
		}
		_log.Infow("addDebugLabels", "kind", kind)
		switch strings.ToLower(kind) {
		case "pod":
			accessor.SetRaw(debugLabels, "metadata", "labels")
		case "deployment", "replicationcontroller", "replicaset", "statefulset", "daemonset", "job":
			accessor.SetRaw(debugLabels, "spec", "template", "metadata", "labels")
		case "cronjob":
			accessor.SetRaw(debugLabels, "spec", "jobTemplate", "spec", "template", "metadata", "labels")
		default:
			_log.Warnw("Unknown Install component in TaskSet. Debug is not possible.", "Kind:", kind)
			return nil
		}

		stepTemplate.Object.Raw = accessor.Bytes()

	}

	return nil
}

func getDebugLabelsMap(orgID, partnerID, projectID string, name string, isSystemWorkload bool) (map[string]string, error) {
	labels := make(map[string]string)
	labels["rep-organization"] = orgID
	labels["rep-partner"] = partnerID
	labels["rep-project"] = projectID
	if isSystemWorkload {
		labels["rep-addon"] = name
	} else {
		labels["rep-workload"] = name
	}

	return labels, nil
}

func resourceToStepTemplate(resource []byte) (sts []controllerv2.StepTemplate, err error) {

	so := &controllerv2.StepObject{Raw: resource}
	var accessor controllerv2.Accessor
	var gvk schema.GroupVersionKind

	accessor, err = so.Accessor()
	if err != nil {
		return
	}

	gvk, err = accessor.GroupVersionKind()
	if err != nil {
		return
	}

	if IsListGVK(gvk) {
		var ro runtime.Object
		ro, _, err = runtimeutil.ToObject(so)
		if err != nil {
			return
		}
		if list, ok := ro.(*corev1.List); ok {
			for _, item := range list.Items {

				var lso = &controllerv2.StepObject{Raw: item.Raw}
				var la controllerv2.Accessor
				var ln string
				var lgvk schema.GroupVersionKind
				la, err = lso.Accessor()
				if err != nil {
					return
				}
				ln, err = la.Name()
				if err != nil {
					return
				}
				lgvk, err = la.GroupVersionKind()
				if err != nil {
					return
				}
				var st controllerv2.StepTemplate
				st.Name = strings.ToLower(fmt.Sprintf("step-%s-%s", lgvk.Kind, ln))
				st.Object = lso
				sts = append(sts, st)
			}
		}
	} else {
		var name string
		var st controllerv2.StepTemplate

		name, err = accessor.Name()
		if err != nil {
			return
		}

		accessor.ResetAutoFields()

		so.Raw = accessor.Bytes()

		st.Name = strings.ToLower(fmt.Sprintf("step-%s-%s", gvk.Kind, name))
		st.Object = so
		sts = append(sts, st)
	}

	return
}

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

// tryConvert attempts to convert the given object to the provided versions in order. This function assumes
// the object is in internal version.
func tryConvert(converter runtime.ObjectConvertor, object runtime.Object, versions ...schema.GroupVersion) (runtime.Object, error) {
	var last error
	for _, version := range versions {
		if version.Empty() {
			return object, nil
		}
		obj, err := converter.ConvertToVersion(object, version)
		if err != nil {
			last = err
			continue
		}
		return obj, nil
	}
	return nil, last
}
