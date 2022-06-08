package runtime

import (
	"bytes"
	"errors"

	"github.com/paralus/paralus/pkg/controller/scheme"
	apiv2 "github.com/paralus/paralus/proto/types/controller"
	rbacv1 "k8s.io/api/rbac/v1"
	apixv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/paralus/paralus/pkg/controller/util"
)

var (
	// ErrInvalidObject is returned for invalid object
	ErrInvalidObject = errors.New("object interface not implemented")
)

// FromObject creates step object from runtime object
func FromObject(ro runtime.Object) (*apiv2.StepObject, error) {

	var so apiv2.StepObject
	var err error

	bb := new(bytes.Buffer)
	err = scheme.Serializer.Encode(ro, bb)
	if err != nil {
		return nil, err
	}

	// use step object accessor to get object gvk
	// so.SetGroupVersionKind(ro.GetObjectKind().GroupVersionKind())
	// if mo, ok := ro.(metav1.Object); ok {
	// 	so.Name = mo.GetName()
	// }
	so.Raw = bb.Bytes()

	// so.Raw, err = util.CleanPatch(so.Raw)
	// if err != nil {
	// 	return nil, err
	// }

	return &so, nil
}

// SetNamespace sets namespace for runtime object
func SetNamespace(ro runtime.Object, namespace string) error {

	switch ro.(type) {
	case *apixv1beta1.CustomResourceDefinition:
	case *rbacv1.ClusterRole:
	case *rbacv1.ClusterRoleBinding:
		crb := ro.(*rbacv1.ClusterRoleBinding)
		for i := range crb.Subjects {
			if crb.Subjects[i].Kind == rbacv1.ServiceAccountKind &&
				crb.Subjects[i].Namespace == "" {
				crb.Subjects[i].Namespace = namespace
			}
		}
	case *rbacv1.RoleBinding:
		rb := ro.(*rbacv1.RoleBinding)
		if rb.Namespace == "" {
			rb.Namespace = namespace
		}
		for i := range rb.Subjects {
			if rb.Subjects[i].Kind == rbacv1.ServiceAccountKind &&
				rb.Subjects[i].Namespace == "" {
				rb.Subjects[i].Namespace = namespace
			}
		}
	case *rbacv1.Role:
		rb := ro.(*rbacv1.Role)
		if rb.Namespace == "" {
			rb.Namespace = namespace
		}
	default:
		if mo, ok := ro.(metav1.Object); ok {
			mo.SetNamespace(namespace)
			return nil
		}
		return ErrInvalidObject

	}
	return nil
}

// ToObject converts step object to runtime object
func ToObject(so *apiv2.StepObject) (o runtime.Object, gvk *schema.GroupVersionKind, err error) {

	accessor, err := so.Accessor()
	if err != nil {
		return
	}

	eGVK, err := accessor.GroupVersionKind()
	if err != nil {
		return
	}

	if scheme.Scheme.Recognizes(eGVK) {
		o, gvk, err = scheme.Serializer.Decode(so.Raw, nil, nil)
	} else {
		o, gvk, err = scheme.Serializer.Decode(so.Raw, nil, &unstructured.Unstructured{})
	}

	return o, gvk, err
}

// ToStructuredObject converts unstructured object to structured object
func ToStructuredObject(obj runtime.Object) runtime.Object {

	if _, ok := obj.(*unstructured.Unstructured); ok {
		bb := new(bytes.Buffer)
		err := scheme.Serializer.Encode(obj, bb)
		if err != nil {
			return obj
		}

		o, _, err := scheme.Serializer.Decode(bb.Bytes(), nil, nil)
		if err != nil {
			return obj
		}
		return o

	}

	return obj
}

// ToUnstructuredObject converts step object to unstructured object,
// this is useful for preserving original json serialized input in step object.
// Note: while patching k8s resources, we should preserve the user input, as
// patching can remove fields; which are represented as null values in the patch
func ToUnstructuredObject(so *apiv2.StepObject) (*unstructured.Unstructured, *schema.GroupVersionKind, error) {
	var o runtime.Object
	var err error
	var gvk *schema.GroupVersionKind

	o, gvk, err = scheme.Serializer.Decode(so.Raw, nil, &unstructured.Unstructured{})

	return o.(*unstructured.Unstructured), gvk, err
}

// PatchMeta is the metadata to be added while patching
type PatchMeta struct {
	Annotations map[string]string
}

// PatchOption is the functional patch option
type PatchOption func(*PatchMeta)

// AddAnnotations adds annotations to object
func AddAnnotations(annotations map[string]string) PatchOption {
	return func(m *PatchMeta) {
		m.Annotations = annotations
	}
}

// Patch patches the step object with give step object
func Patch(input *apiv2.StepObject, with *apiv2.StepObject, opts ...PatchOption) error {
	pm := &PatchMeta{}
	for _, opt := range opts {
		opt(pm)
	}
	accessor, err := input.Accessor()
	if err != nil {
		return err
	}
	gvk, err := accessor.GroupVersionKind()
	if err != nil {
		return err
	}

	if util.IsStrategicMergePatch(gvk) {
		pb, err := util.CreateStrategicMergePatch(gvk, nil, input.Raw, with.Raw)
		if err != nil {
			return err
		}

		fb, err := util.ApplyStrategicMergePatch(gvk, input.Raw, pb)
		if err != nil {
			return err
		}

		input.Raw = fb
	} else {
		pb, err := util.CreateJSONMergePatch(nil, input.Raw, with.Raw)
		if err != nil {
			return err
		}

		fb, err := util.ApplyJSONMergePatch(input.Raw, pb)
		if err != nil {
			return err
		}

		input.Raw = fb

	}

	if pm.Annotations != nil {
		accessor, err := input.Accessor()
		if err != nil {
			return err
		}

		accessor.SetAnnotations(pm.Annotations)
		input.Raw = accessor.Bytes()
	}

	return nil
}
