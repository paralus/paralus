package converter

import (
	"errors"
	"fmt"
	"strings"

	runtimeutil "github.com/RafaySystems/rcloud-base/components/common/pkg/controller/runtime"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/hasher"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	typesv2 "github.com/RafaySystems/rcloud-base/components/common/proto/types/config"
	controllerv2 "github.com/RafaySystems/rcloud-base/components/common/proto/types/controller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	k8sapijson "sigs.k8s.io/kustomize/pseudo/k8s/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/yaml"
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

// FromJSON converts data in JSON format into namespace and taskset
func FromJSON(name, data string) (ns *typesv2.Namespace, placement *typesv2.Placement, err error) {
	sresources := strings.Split(data, delimiter)

	var resources [][]byte

	for _, sresource := range sresources {
		resources = append(resources, []byte(sresource))
	}

	return fromResources(name, resources)
}

// FromYAML converts the data in YAML format into namespace and taskset
func FromYAML(name, data string) (ns *typesv2.Namespace, placement *typesv2.Placement, err error) {
	resources, err := getResourcesArrayFromYAML(data)
	if err != nil {
		return
	}
	return fromResources(name, resources)
}

func getResourcesArrayFromYAML(data string) ([][]byte, error) {
	sresources := strings.Split(data, delimiter)
	var resources [][]byte
	for _, sresource := range sresources {
		if strings.TrimSpace(sresource) == "" {
			continue
		}
		var b []byte
		b, err := yaml.YAMLToJSONStrict([]byte(sresource))
		if err != nil {
			return nil, err
		}
		resources = append(resources, b)
	}
	return resources, nil
}

func fromResources(name string, resources [][]byte) (ns *typesv2.Namespace, placement *typesv2.Placement, err error) {
	var remaining [][]byte

	ns = &typesv2.Namespace{}
	placement = &typesv2.Placement{}

	remaining, err = ToNamespace(resources, ns, "", "", "")
	if err != nil {
		return
	}

	remaining, err = ToPlacement(resources, name, placement, "", "", "")
	if err != nil {
		return
	}

	if len(remaining) > 0 {
	}

	return
}

// ToNamespace converts data to namespace
func ToNamespace(resources [][]byte, ns *typesv2.Namespace, orgID, partnerID, projectID string) (remaining [][]byte, err error) {
	ns.ApiVersion = typesv2.ConfigGroup
	ns.Kind = typesv2.NamespaceKind

	for _, resource := range resources {
		gvk, err := dmf.Interpret(resource)
		if err != nil {
			return nil, err
		}

		o, err := toRuntimeObject(*gvk, resource)
		if err != nil {
			return nil, err
		}

		switch {
		case IsNamespaceGVK(*gvk):
			if mo, ok := o.(metav1.Object); ok {
				ns.Metadata.Name = mo.GetName()
				ns.Metadata.Labels = mo.GetLabels()
				ns.Metadata.Annotations = mo.GetAnnotations()
			} else {
				return nil, ErrInvalidObject
			}
			ns.Spec.ObjectMeta.Name = ns.Metadata.Name
			ns.Spec.ObjectMeta.Labels = ns.Metadata.Labels
			ns.Spec.ObjectMeta.Annotations = ns.Metadata.Annotations
			ns.Metadata.Organization = orgID
			ns.Metadata.Partner = partnerID
			ns.Metadata.Project = projectID
		case IsNamespacePostCreate(*gvk):
			var st controllerv2.StepTemplate
			st, err = toStepTemplate(o)
			if err != nil {
				return nil, err
			}

			ns.Spec.Spec.PostCreate = append(ns.Spec.Spec.PostCreate, &st)

		default:
			remaining = append(remaining, resource)
		}

	}
	return
}

// ToPlacement converts data to placement
func ToPlacement(resources [][]byte, name string, placement *typesv2.Placement, orgID, partnerID, projectID string) (remaining [][]byte, err error) {
	for _, resource := range resources {
		gvk, err := dmf.Interpret(resource)
		if err != nil {
			return nil, err
		}
		switch {
		case IsPlacementGVK(*gvk):
			json.Unmarshal(resource, placement)
			if placement.Spec.ClusterSelector == "" || placement.Spec.PlacementType != typesv2.PlacementType_ClusterSelector {
				return nil, fmt.Errorf("placement spec is missing type and selector")
			}
			if placement.Metadata.Name == "" {
				placement.Metadata.Name = name
			}
			placement.ApiVersion = typesv2.ConfigGroup
			placement.Kind = typesv2.PlacementKind
			placement.Metadata.Organization = orgID
			placement.Metadata.Partner = partnerID
			placement.Metadata.Project = projectID
		default:
			remaining = append(remaining, resource)
		}
	}
	return
}

func getIngressAnnotations(name string, orgID, partnerID string) map[string]string {
	orgHash, err := hasher.HashFromHex(orgID)
	if err != nil {
		orgHash = "unknown"
	}
	partnerHash, err := hasher.HashFromHex(partnerID)
	if err != nil {
		partnerHash = "unknown"
	}
	return map[string]string{
		ingressAnnotationConfigSnippetKey: fmt.Sprintf("set $workload_name \"%s\";set $orgId \"%s\";set $partnerId \"%s\";", name, orgHash, partnerHash),
	}
}

func addIngressAnnotations(annotations map[string]string, name string, orgId, partnerId string) {
	if _, ok := annotations[ingressAnnotationConfigSnippetKey]; !ok {
		orgHash, err := hasher.HashFromHex(orgId)
		if err != nil {
			orgHash = "unknown"
		}
		partnerHash, err := hasher.HashFromHex(partnerId)
		if err != nil {
			partnerHash = "unknown"
		}
		annotations[ingressAnnotationConfigSnippetKey] = fmt.Sprintf("set $workload_name \"%s\";set $orgId \"%s\";set $partnerId \"%s\";",
			name, orgHash, partnerHash)
	}
}

func isLoggingEnabled(annotations map[string]string) bool {
	if _, ok := annotations[typesv2.LogEndpoint]; ok {
		return true
	}
	return false
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
	orgHashID, err := hasher.HashFromHex(orgID)
	if err != nil {
		err = fmt.Errorf("failed to convert Org ID %d %s", orgID, err.Error())
		return nil, err
	}
	partnerHashID, err := hasher.HashFromHex(partnerID)
	if err != nil {
		err = fmt.Errorf("failed to convert Partner ID %d %s", partnerID, err.Error())
		return nil, err
	}
	projectHashID, err := hasher.HashFromHex(projectID)
	if err != nil {
		err = fmt.Errorf("failed to convert ProjectID ID %d %s", orgID, err.Error())
		return nil, err
	}

	labels := make(map[string]string)
	labels["rep-organization"] = orgHashID
	labels["rep-partner"] = partnerHashID
	labels["rep-project"] = projectHashID
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
