package apply

import (
	"errors"
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"

	"github.com/paralus/paralus/pkg/controller/scheme"
	clusterv2 "github.com/paralus/paralus/proto/types/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	cruntime "github.com/paralus/paralus/pkg/controller/runtime"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Errors that could be returned by Apply.
var (
	ErrNilObject  = errors.New("can't reference a nil object")
	ErrNoSelfLink = errors.New("selfLink was empty, can't make reference")
)

var (
	// ErrNoPreviousConfig is returned when no previous configuration is found in the annotations
	ErrNoPreviousConfig = errors.New("last applied configuration not found")
)

var (
	emptyGVK = schema.GroupVersionKind{}
)

func getBytes(o runtime.Object, withOriginal bool) ([]byte, error) {
	do := o.DeepCopyObject()
	if !withOriginal {
		if mo, ok := do.(metav1.Object); ok {
			annotations := mo.GetAnnotations()
			if annotations != nil {
				delete(annotations, clusterv2.OrignalConfig)
			}
			mo.SetAnnotations(annotations)
		}
	}
	return runtime.Encode(unstructured.UnstructuredJSONScheme, o)

}

// GetOriginalConfig returns previous config of the object
func GetOriginalConfig(o runtime.Object) ([]byte, error) {
	if mo, ok := o.(metav1.Object); ok {
		annotations := mo.GetAnnotations()
		if annotations != nil {
			if v, ok := annotations[clusterv2.OrignalConfig]; ok {
				return []byte(v), nil
			}
		}
	}
	return nil, nil
}

// GetGVK returns group version kind of a runtime object
func GetGVK(obj runtime.Object) (schema.GroupVersionKind, error) {
	gvks, _, err := scheme.Scheme.ObjectKinds(obj)
	if err != nil {
		return emptyGVK, err
	}
	return gvks[0], nil
}

// updateObject updates current object with modified object
func updateObject(current, modified runtime.Object) error {
	if reflect.TypeOf(current) != reflect.TypeOf(modified) {
		current = cruntime.ToStructuredObject(current)
		modified = cruntime.ToStructuredObject(modified)

		//return fmt.Errorf("current %T and modified %T of different types", current, modified)
	}

	switch current.(type) {
	case *clusterv2.Task:
		c := current.(*clusterv2.Task)
		m := modified.(*clusterv2.Task)
		c.ObjectMeta.Labels = m.ObjectMeta.Labels
		c.ObjectMeta.Annotations = m.ObjectMeta.Annotations
		c.ObjectMeta.Finalizers = m.ObjectMeta.Finalizers
		c.Spec = m.Spec

	case *clusterv2.Tasklet:
		c := current.(*clusterv2.Tasklet)
		m := modified.(*clusterv2.Tasklet)
		c.ObjectMeta.Labels = m.ObjectMeta.Labels
		c.ObjectMeta.Annotations = m.ObjectMeta.Annotations
		c.ObjectMeta.Finalizers = m.ObjectMeta.Finalizers
		c.Spec = m.Spec
	case *clusterv2.Namespace:
		c := current.(*clusterv2.Namespace)
		m := modified.(*clusterv2.Namespace)
		c.ObjectMeta.Labels = m.ObjectMeta.Labels
		c.ObjectMeta.Annotations = m.ObjectMeta.Annotations
		c.ObjectMeta.Finalizers = m.ObjectMeta.Finalizers
		c.Spec = m.Spec
	case *v1.Node:
		c := current.(*v1.Node)
		m := modified.(*v1.Node)
		c.Labels = m.Labels
		c.Annotations = m.Annotations
		c.Finalizers = m.Finalizers
		c.Spec.Unschedulable = m.Spec.Unschedulable
		c.Spec.Taints = m.Spec.Taints
	default:
		return fmt.Errorf("unhandled type for update %T", current)
	}

	return nil
}
